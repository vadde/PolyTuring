// Incremental pipeline: single-file re-indexing and pruning
// Re-uses tree-sitter CodeParser and EmbedderClient for fine-grained AST updates

use crate::db::DbClient;
use crate::embedder::EmbedderClient;
use crate::ingestion::FileCache;
use crate::parser::{CodeParser, ExtractedEdge, ExtractedNode};
use sha2::{Digest, Sha256};
use std::fs;
use std::path::Path;
use tracing::{error, info, warn};

pub struct IncrementalPipeline {
    db: DbClient,
    embedder: EmbedderClient,
    cache: FileCache,
}

impl IncrementalPipeline {
    pub fn new(db: DbClient, embedder: EmbedderClient, cache: FileCache) -> Self {
        Self {
            db,
            embedder,
            cache,
        }
    }

    /// Re-index a single created or modified file incrementally
    pub async fn reindex_file(
        &self,
        workspace: &str,
        root_dir: &str,
        rel_path: &str,
    ) -> Result<(), String> {
        let full_path = Path::new(root_dir).join(rel_path);

        if !full_path.exists() {
            // File might have been deleted right after a modify event
            return self.delete_file(workspace, rel_path).await;
        }

        let content = match fs::read_to_string(&full_path) {
            Ok(c) => c,
            Err(e) => {
                // Could be non-UTF-8 or lock file, skip silently
                warn!("Cannot read file '{}' for live sync: {}", rel_path, e);
                return Ok(());
            }
        };

        // Compute sha256 hash
        let mut hasher = Sha256::new();
        hasher.update(content.as_bytes());
        let hash = hex::encode(hasher.finalize());

        // Check staleness against cache
        if !self.cache.is_stale(rel_path, &hash).await {
            return Ok(());
        }

        // Parse AST or fallback chunker
        let parse_res = CodeParser::parse_file(workspace, rel_path, &content)
            .filter(|pr| !pr.nodes.is_empty())
            .or_else(|| CodeParser::parse_fallback(workspace, rel_path, &content));

        let parse_res = match parse_res {
            Some(pr) if !pr.nodes.is_empty() => pr,
            _ => {
                // If now empty, prune any old nodes
                return self.delete_file(workspace, rel_path).await;
            }
        };

        // Topological Containment Backbone (R-063): Synthesize File node and containment edges
        let filename = Path::new(rel_path)
            .file_name()
            .map(|n| n.to_string_lossy().to_string())
            .unwrap_or_else(|| rel_path.to_string());
        let ext = Path::new(rel_path)
            .extension()
            .map(|e| e.to_string_lossy().to_string())
            .unwrap_or_default();
        let total_lines = content.lines().count().max(1);
        let preview_text = match content.char_indices().nth(500) {
            Some((idx, _)) => &content[..idx],
            None => &content,
        };
        let file_node_id = format!("{}:file:{}", workspace, rel_path);
        let file_node = ExtractedNode {
            id: file_node_id.clone(),
            workspace: workspace.to_string(),
            label: filename.clone(),
            kind: "file".to_string(),
            file_path: rel_path.to_string(),
            language: ext,
            line_start: 1,
            line_end: total_lines,
            text: preview_text.to_string(),
        };

        let mut all_nodes = Vec::with_capacity(parse_res.nodes.len() + 1);
        all_nodes.push(file_node);
        all_nodes.extend(parse_res.nodes);

        let mut all_edges = parse_res.edges;
        for sym in &all_nodes[1..] {
            all_edges.push(ExtractedEdge {
                workspace: workspace.to_string(),
                source_id: file_node_id.clone(),
                target_label: sym.label.clone(),
                target_id: Some(sym.id.clone()),
                edge_type: "CONTAINS".to_string(),
                category: "CONTAINMENT".to_string(),
            });
        }

        let parent_dir = Path::new(rel_path)
            .parent()
            .map(|p| p.to_string_lossy().to_string())
            .unwrap_or_default();
        let dir_source_id = if parent_dir.is_empty() || parent_dir == "." {
            format!("{}:workspace", workspace)
        } else {
            format!("{}:dir:{}", workspace, parent_dir)
        };
        all_edges.push(ExtractedEdge {
            workspace: workspace.to_string(),
            source_id: dir_source_id,
            target_label: filename.clone(),
            target_id: Some(file_node_id),
            edge_type: "CONTAINS_FILE".to_string(),
            category: "CONTAINMENT".to_string(),
        });

        // 1. Look up existing community for this file before pruning to preserve community assignment
        let existing_comm = self.db.get_file_community(workspace, rel_path).await.unwrap_or(None);

        // 1b. Capture existing inbound caller edges (R-079) before pruning old nodes
        let inbound_edges = self.db.get_inbound_edges_for_file(workspace, rel_path).await.unwrap_or_default();

        // 2. Generate vector embeddings for newly parsed nodes FIRST (eliminates 404 blackout window)
        let enriched_texts: Vec<String> = all_nodes
            .iter()
            .map(|n| format!("[{}] {} {} in {}\n{}", n.language, n.kind, n.label, n.file_path, n.text))
            .collect();
        let text_refs: Vec<&str> = enriched_texts.iter().map(|s| s.as_str()).collect();
        let embeddings = match self.embedder.embed_batch(&text_refs).await {
            Ok(embs) => embs,
            Err(e) => {
                warn!("Embedding generation failed for live sync on '{}': {}", rel_path, e);
                vec![vec![0.0f32; 384]; all_nodes.len()]
            }
        };

        // 3. Atomic swap: prune previous nodes immediately before storing new ones
        if let Err(e) = self.db.delete_file(workspace, rel_path).await {
            warn!("Failed to prune previous nodes for '{}': {}", rel_path, e);
        }

        // 4. Store new nodes with inherited community
        if let Err(e) = self.db.store_nodes_with_community(&all_nodes, &embeddings, Some(&hash), existing_comm).await {
            error!("Failed to store nodes during live sync for '{}': {}", rel_path, e);
            return Err(e);
        }

        // 5. Store new edges
        if let Err(e) = self.db.store_edges(&all_edges).await {
            warn!("Failed to store edges during live sync for '{}': {}", rel_path, e);
        }

        // 5b. Re-link surviving inbound caller edges to new symbol node IDs (R-079)
        if !inbound_edges.is_empty() {
            if let Ok(relinked) = self.db.relink_inbound_edges(workspace, &inbound_edges, &all_nodes).await {
                if relinked > 0 {
                    info!("Preserved and relinked {} inbound caller edges for '{}'", relinked, rel_path);
                }
            }
        }

        // 5. Update staleness cache
        self.cache.update(rel_path, &hash).await;

        info!(
            "Live synced file '{}' in workspace '{}' ({} nodes, {} edges)",
            rel_path,
            workspace,
            all_nodes.len(),
            all_edges.len()
        );

        Ok(())
    }

    /// Prune nodes and relations for a deleted file
    pub async fn delete_file(&self, workspace: &str, rel_path: &str) -> Result<(), String> {
        info!("Pruning deleted file '{}' from workspace '{}'", rel_path, workspace);
        self.db.delete_file(workspace, rel_path).await?;
        self.cache.remove(rel_path).await;
        Ok(())
    }

    /// Rename file path references in the graph
    pub async fn rename_file(
        &self,
        workspace: &str,
        old_path: &str,
        new_path: &str,
    ) -> Result<(), String> {
        info!(
            "Renaming file '{}' -> '{}' in workspace '{}'",
            old_path, new_path, workspace
        );
        self.db.rename_file(workspace, old_path, new_path).await?;
        self.cache.rename(old_path, new_path).await;
        Ok(())
    }

    /// Prune any files from SurrealDB that were cached but no longer exist on disk
    pub async fn prune_missing_files(&self, workspace: &str, root_dir: &str) -> Vec<String> {
        let cached_paths = self.cache.get_all_paths().await;
        let mut pruned = Vec::new();
        for rel_path in cached_paths {
            let full_p = Path::new(root_dir).join(&rel_path);
            if !full_p.exists() {
                if let Ok(()) = self.delete_file(workspace, &rel_path).await {
                    pruned.push(rel_path);
                }
            }
        }
        pruned
    }
}
