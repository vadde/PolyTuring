// Incremental pipeline: single-file re-indexing and pruning
// Re-uses tree-sitter CodeParser and EmbedderClient for fine-grained AST updates

use crate::db::DbClient;
use crate::embedder::EmbedderClient;
use crate::ingestion::{FileCache, SymbolTable};
use crate::parser::{CodeParser, ExtractedEdge, ExtractedNode};
use sha2::{Digest, Sha256};
use std::collections::HashSet;
use std::fs;
use std::path::Path;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing::{info, warn};

pub struct IncrementalPipeline {
    db: DbClient,
    embedder: EmbedderClient,
    cache: FileCache,
    symbol_table: Arc<RwLock<SymbolTable>>,
    known_dirs: Arc<RwLock<HashSet<String>>>,
}

impl IncrementalPipeline {
    pub fn new(
        db: DbClient,
        embedder: EmbedderClient,
        cache: FileCache,
        symbol_table: Arc<RwLock<SymbolTable>>,
        known_dirs: Arc<RwLock<HashSet<String>>>,
    ) -> Self {
        Self {
            db,
            embedder,
            cache,
            symbol_table,
            known_dirs,
        }
    }

    #[allow(dead_code)]
    pub fn with_default_dirs(
        db: DbClient,
        embedder: EmbedderClient,
        cache: FileCache,
        symbol_table: Arc<RwLock<SymbolTable>>,
    ) -> Self {
        Self::new(
            db,
            embedder,
            cache,
            symbol_table,
            Arc::new(RwLock::new(HashSet::new())),
        )
    }

    /// Ensure all ancestor directory nodes and hierarchy edges exist for rel_path (R-094)
    pub async fn synthesize_directory_backbone(
        &self,
        workspace: &str,
        rel_path: &str,
    ) -> (Vec<ExtractedNode>, Vec<ExtractedEdge>) {
        let mut cur_dir = Path::new(rel_path).parent();
        let mut missing_dirs = Vec::new();

        {
            let known = self.known_dirs.read().await;
            while let Some(dir) = cur_dir {
                let dir_str = dir.to_string_lossy().to_string();
                if dir_str.is_empty() || dir_str == "." {
                    break;
                }
                if !known.contains(&dir_str) {
                    missing_dirs.push(dir_str);
                }
                cur_dir = dir.parent();
            }
        }

        if missing_dirs.is_empty() {
            return (Vec::new(), Vec::new());
        }

        // Sort from shallowest to deepest directory so hierarchy is built top-down
        missing_dirs.sort_by_key(|d| d.len());

        let mut synth_nodes = Vec::with_capacity(missing_dirs.len());
        let mut synth_edges = Vec::with_capacity(missing_dirs.len());

        {
            let mut known = self.known_dirs.write().await;
            for dir_str in &missing_dirs {
                if known.insert(dir_str.clone()) {
                    let dir_name = Path::new(dir_str)
                        .file_name()
                        .map(|n| n.to_string_lossy().to_string())
                        .unwrap_or_else(|| dir_str.clone());

                    synth_nodes.push(ExtractedNode {
                        id: format!("{}:dir:{}", workspace, dir_str),
                        workspace: workspace.to_string(),
                        label: dir_name.clone(),
                        kind: "directory".to_string(),
                        file_path: dir_str.clone(),
                        language: String::new(),
                        line_start: 0,
                        line_end: 0,
                        text: format!("Directory: {}", dir_str),
                    });

                    let parent_of_dir = Path::new(dir_str)
                        .parent()
                        .map(|p| p.to_string_lossy().to_string())
                        .unwrap_or_default();

                    let parent_id = if parent_of_dir.is_empty() || parent_of_dir == "." {
                        format!("{}:workspace", workspace)
                    } else {
                        format!("{}:dir:{}", workspace, parent_of_dir)
                    };

                    synth_edges.push(ExtractedEdge {
                        workspace: workspace.to_string(),
                        source_id: parent_id,
                        target_label: dir_name,
                        target_id: Some(format!("{}:dir:{}", workspace, dir_str)),
                        edge_type: "CONTAINS".to_string(),
                        category: "CONTAINMENT".to_string(),
                    });
                }
            }
        }

        (synth_nodes, synth_edges)
    }

    /// Access the known directory cache (R-094)
    #[allow(dead_code)]
    pub fn known_dirs(&self) -> Arc<RwLock<HashSet<String>>> {
        Arc::clone(&self.known_dirs)
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

        // 0. Ensure directory backbone DAG exists up to root (R-094)
        let (synth_dir_nodes, synth_dir_edges) = self
            .synthesize_directory_backbone(workspace, rel_path)
            .await;
        if !synth_dir_nodes.is_empty() {
            let (store_nodes_res, store_edges_res) = tokio::join!(
                self.db.store_nodes(&synth_dir_nodes, &[], None),
                self.db.store_edges(&synth_dir_edges)
            );
            if let Err(e) = store_nodes_res {
                warn!("Failed to store synthesized directory backbone nodes for '{}': {}", rel_path, e);
            }
            if let Err(e) = store_edges_res {
                warn!("Failed to store synthesized directory backbone edges for '{}': {}", rel_path, e);
            }
        }

        // 1. Look up existing community & inbound edges concurrently (R-100)
        let (comm_res, inbound_res) = tokio::join!(
            self.db.get_file_community(workspace, rel_path),
            self.db.get_inbound_edges_for_file(workspace, rel_path)
        );
        let mut existing_comm = comm_res.unwrap_or(None);
        let inbound_edges = inbound_res.unwrap_or_default();

        // 1a. Zero-delay community inference for brand-new files (R-105)
        // If file is brand-new (existing_comm is None), infer modal community from neighbor connections or directory siblings
        if existing_comm.is_none() {
            let target_ids: Vec<String> = all_edges
                .iter()
                .filter_map(|e| e.target_id.clone())
                .collect();
            if !target_ids.is_empty() {
                existing_comm = self.db.infer_community_from_targets(workspace, &target_ids).await.unwrap_or(None);
            }
            if existing_comm.is_none() {
                let parent_dir = Path::new(rel_path).parent().map(|p| p.to_string_lossy().to_string()).unwrap_or_default();
                if !parent_dir.is_empty() && parent_dir != "." {
                    existing_comm = self.db.get_directory_modal_community(workspace, &parent_dir).await.unwrap_or(None);
                }
            }
        }

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

        // 3. In-Memory SAIR Edge Resolution & Unresolved Pool Retention (R-078, R-093, R-103)
        let mut containment_edges = Vec::new();
        let mut cross_file_edges = Vec::new();
        for edge in all_edges {
            if edge.target_id.is_some() {
                containment_edges.push(edge);
            } else {
                cross_file_edges.push(edge);
            }
        }

        let imported_symbols: Vec<String> = cross_file_edges
            .iter()
            .filter(|e| e.edge_type == "IMPORTS")
            .map(|e| e.target_label.clone())
            .collect();

        let mut resolved_edges = containment_edges;
        {
            let mut st = self.symbol_table.write().await;
            st.remove_file(workspace, rel_path);
            st.register_file_nodes(&all_nodes);
            st.register_file_imports(workspace, rel_path, &imported_symbols);

            for mut edge in cross_file_edges {
                let caller_file = edge.source_id.split(':').nth(1).unwrap_or("").to_string();
                if !caller_file.is_empty() {
                    st.record_caller_dependency(workspace, &caller_file, &edge.target_label);
                }
                if let Some(target_id) = st.resolve_edge(&edge) {
                    edge.target_id = Some(target_id);
                    resolved_edges.push(edge);
                } else {
                    // Retain unresolved forward reference in pool (R-103)
                    st.retain_unresolved(edge);
                }
            }

            // Drain any pending unsatisfied edges from other files now satisfied by our new symbols (R-103)
            for node in &all_nodes {
                let satisfied = st.drain_satisfied(workspace, &node.label);
                for mut sat_edge in satisfied {
                    if let Some(target_id) = st.resolve_edge(&sat_edge) {
                        sat_edge.target_id = Some(target_id);
                        resolved_edges.push(sat_edge);
                    }
                }
            }
        }

        // 4. Atomic Multi-Statement Swap (R-101, R-102, R-104)
        // Combines file pruning, node upsert, edge persistence, and inbound relinking into a single transaction.
        let (_node_cnt, _edge_cnt, relink_cnt) = self
            .db
            .execute_atomic_swap(
                workspace,
                rel_path,
                Some(&hash),
                &all_nodes,
                &embeddings,
                existing_comm,
                &resolved_edges,
                &inbound_edges,
            )
            .await?;

        if relink_cnt > 0 {
            info!("Preserved and relinked {} inbound caller edges for '{}'", relink_cnt, rel_path);
        }

        // 5. Update staleness cache
        self.cache.update(rel_path, &hash).await;

        info!(
            "Live synced file '{}' in workspace '{}' ({} nodes, {} edges)",
            rel_path,
            workspace,
            all_nodes.len(),
            resolved_edges.len()
        );

        Ok(())
    }

    /// Prune nodes and relations for a deleted file
    pub async fn delete_file(&self, workspace: &str, rel_path: &str) -> Result<(), String> {
        info!("Pruning deleted file '{}' from workspace '{}'", rel_path, workspace);
        {
            let mut st = self.symbol_table.write().await;
            st.remove_file(workspace, rel_path);
        }
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
        {
            let mut st = self.symbol_table.write().await;
            st.rename_file(workspace, old_path, new_path);
        }
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
