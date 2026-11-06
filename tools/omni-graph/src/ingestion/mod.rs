// Ingestion pipeline: Walk directory, parse AST, embed vectors, store in SurrealDB
// Implements: R-010, R-011

pub mod worktree;
#[allow(unused_imports)]
pub use worktree::{detect_git_worktree, discover_parent_worktrees, WorktreeInfo};
pub mod symbol_table;
pub use symbol_table::SymbolTable;

use crate::analysis::CommunityDetector;
use crate::db::DbClient;
use crate::embedder::EmbedderClient;
use crate::parser::{CodeParser, ExtractedEdge, ExtractedNode};
use sha2::{Digest, Sha256};
use std::collections::{HashMap, HashSet};
use std::fs;
use std::path::{Path, PathBuf};
use std::sync::Arc;
use tokio::sync::{mpsc, oneshot, Mutex, RwLock};
use tracing::{error, info, warn};
use walkdir::WalkDir;

#[derive(Clone, Debug, Default)]
pub struct FileCache {
    hashes: Arc<Mutex<HashMap<String, String>>>,
}

impl FileCache {
    pub fn new() -> Self {
        Self {
            hashes: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub async fn populate(&self, hashes: HashMap<String, String>) {
        let mut map = self.hashes.lock().await;
        map.extend(hashes);
    }

    pub async fn is_stale(&self, file_path: &str, current_hash: &str) -> bool {
        let map = self.hashes.lock().await;
        if let Some(prev_hash) = map.get(file_path) {
            if prev_hash == current_hash {
                return false; // Unchanged
            }
        }
        true
    }

    pub async fn update(&self, file_path: &str, current_hash: &str) {
        let mut map = self.hashes.lock().await;
        map.insert(file_path.to_string(), current_hash.to_string());
    }

    pub async fn remove(&self, file_path: &str) {
        let mut map = self.hashes.lock().await;
        map.remove(file_path);
    }

    pub async fn rename(&self, old_path: &str, new_path: &str) {
        let mut map = self.hashes.lock().await;
        if let Some(h) = map.remove(old_path) {
            map.insert(new_path.to_string(), h);
        }
    }

    pub async fn get_all_paths(&self) -> Vec<String> {
        let map = self.hashes.lock().await;
        map.keys().cloned().collect()
    }

    /// Clear all cached hashes. Used by refresh-ingestion to force full re-scan.
    pub async fn clear(&self) {
        let mut map = self.hashes.lock().await;
        map.clear();
    }
}

#[derive(Clone, Debug, serde::Serialize, serde::Deserialize)]
pub struct IngestionJobStatus {
    pub workspace: String,
    pub path: String,
    pub is_refresh: bool,
    pub phase: String, // "queued", "purging", "scanning", "indexing", "clustering", "completed", "failed", "cancelled"
    pub queue_position: usize,
    pub files_scanned: usize,
    pub files_indexed: usize,
    pub files_skipped: usize,
    pub total_files: usize,
    pub nodes_created: usize,
    pub edges_created: usize,
    pub clusters_computed: usize,
    pub started_at: u64,
    pub elapsed_ms: u64,
    pub completed: bool,
    pub error: Option<String>,
}

#[allow(dead_code)]
#[derive(Debug)]
pub struct IngestTask {
    pub root_dir: String,
    pub project: Option<String>,
    pub refresh: bool,
    pub auto_watch: bool,
    pub workspace_name: String,
    pub is_worktree: bool,
    pub parent_ws: Option<String>,
    pub branch: Option<String>,
    pub response_tx: Option<oneshot::Sender<Result<IngestResult, String>>>,
}

#[allow(dead_code)]
pub struct IngestionPipeline {
    db: DbClient,
    embedder: EmbedderClient,
    cache: FileCache,
    jobs: Arc<RwLock<HashMap<String, IngestionJobStatus>>>,
    queue_tx: mpsc::Sender<IngestTask>,
    watcher: Arc<RwLock<Option<Arc<crate::watcher::WatchManager>>>>,
}

#[derive(serde::Serialize, Debug)]
pub struct IngestResult {
    pub workspace: String,
    pub files_scanned: usize,
    pub files_indexed: usize,
    pub files_skipped: usize,
    pub nodes_created: usize,
    pub edges_created: usize,
    pub clusters_computed: usize,
    pub duration_ms: u64,
    pub is_worktree: bool,
    pub parent_workspace: Option<String>,
    pub branch: Option<String>,
}

impl IngestionPipeline {
    pub fn new(db: DbClient, embedder: EmbedderClient, cache: FileCache) -> Self {
        let (queue_tx, mut queue_rx) = mpsc::channel::<IngestTask>(64);
        let jobs: Arc<RwLock<HashMap<String, IngestionJobStatus>>> = Arc::new(RwLock::new(HashMap::new()));
        let watcher: Arc<RwLock<Option<Arc<crate::watcher::WatchManager>>>> = Arc::new(RwLock::new(None));

        let worker_jobs = jobs.clone();
        let worker_db = db.clone();
        let worker_embedder = embedder.clone();
        let worker_cache = cache.clone();
        let worker_watcher = watcher.clone();
        let worker_queue_tx = queue_tx.clone();

        tokio::spawn(async move {
            info!("Ingestion task queue worker active (FIFO serialized mode).");
            while let Some(task) = queue_rx.recv().await {
                // Check if cancelled before starting
                let is_cancelled = {
                    let j_lock = worker_jobs.read().await;
                    j_lock.get(&task.workspace_name).map(|j| j.phase == "cancelled").unwrap_or(false)
                };
                if is_cancelled {
                    info!("Ingestion task for '{}' was cancelled; skipping execution.", task.workspace_name);
                    if let Some(tx) = task.response_tx {
                        let _ = tx.send(Err("Ingestion cancelled by user".to_string()));
                    }
                    continue;
                }

                // Update queue positions
                {
                    let mut j_lock = worker_jobs.write().await;
                    if let Some(j) = j_lock.get_mut(&task.workspace_name) {
                        j.phase = if task.refresh { "purging".to_string() } else { "scanning".to_string() };
                        j.queue_position = 0;
                        j.started_at = std::time::SystemTime::now()
                            .duration_since(std::time::UNIX_EPOCH)
                            .unwrap_or_default()
                            .as_millis() as u64;
                    }
                    let mut pos = 1;
                    for other in j_lock.values_mut() {
                        if !other.completed && other.workspace != task.workspace_name {
                            other.queue_position = pos;
                            pos += 1;
                        }
                    }
                }

                // Execute ingestion
                execute_ingest(
                    &worker_db,
                    &worker_embedder,
                    &worker_cache,
                    &worker_jobs,
                    &worker_watcher,
                    &worker_queue_tx,
                    task,
                ).await;
            }
        });

        Self {
            db,
            embedder,
            cache,
            jobs,
            queue_tx,
            watcher,
        }
    }

    pub async fn set_watcher(&self, watcher: Arc<crate::watcher::WatchManager>) {
        let mut lock = self.watcher.write().await;
        *lock = Some(watcher);
    }

    /// Retrieve real-time status of all active and recently completed ingestion jobs
    pub async fn get_status(&self) -> Vec<IngestionJobStatus> {
        let jobs = self.jobs.read().await;
        let mut list: Vec<IngestionJobStatus> = jobs.values().cloned().collect();
        list.sort_by(|a, b| {
            let a_active = !a.completed;
            let b_active = !b.completed;
            b_active.cmp(&a_active).then_with(|| b.started_at.cmp(&a.started_at))
        });
        list
    }

    /// Check if a workspace is actively being scanned, indexed, purged, or clustered (R-080)
    pub async fn is_indexing(&self, workspace: Option<&str>) -> Option<IngestionJobStatus> {
        let jobs = self.jobs.read().await;
        if let Some(ws) = workspace {
            if let Some(job) = jobs.get(ws) {
                if !job.completed && matches!(job.phase.as_str(), "queued" | "purging" | "scanning" | "indexing" | "clustering") {
                    return Some(job.clone());
                }
            }
        } else {
            for job in jobs.values() {
                if !job.completed && matches!(job.phase.as_str(), "queued" | "purging" | "scanning" | "indexing" | "clustering") {
                    return Some(job.clone());
                }
            }
        }
        None
    }

    /// Set or update a job status directly (used for testing and status overrides)
    #[allow(dead_code)]
    pub async fn set_job_status(&self, job: IngestionJobStatus) {
        let mut map = self.jobs.write().await;
        map.insert(job.workspace.clone(), job);
    }

    /// Dismiss a completed or failed ingestion job from the tracker
    pub async fn dismiss_job(&self, workspace: &str) {
        let mut jobs = self.jobs.write().await;
        if let Some(job) = jobs.get_mut(workspace) {
            if !job.completed {
                job.phase = "cancelled".to_string();
                job.completed = true;
                job.error = Some("Cancelled by user".to_string());
                return;
            }
        }
        jobs.remove(workspace);
    }

    #[allow(clippy::result_large_err)]
    pub async fn enqueue_ingest(
        &self,
        root_dir: &str,
        project: Option<&str>,
        refresh: bool,
        auto_watch: bool,
        is_async: bool,
    ) -> Result<oneshot::Receiver<Result<IngestResult, String>>, IngestionJobStatus> {
        let root = Path::new(root_dir);
        let now_ms = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_millis() as u64;

        if !root.exists() || !root.is_dir() {
            let err_status = IngestionJobStatus {
                workspace: project.unwrap_or("unknown").to_string(),
                path: root_dir.to_string(),
                is_refresh: refresh,
                phase: "failed".to_string(),
                queue_position: 0,
                files_scanned: 0,
                files_indexed: 0,
                files_skipped: 0,
                total_files: 0,
                nodes_created: 0,
                edges_created: 0,
                clusters_computed: 0,
                started_at: now_ms,
                elapsed_ms: 0,
                completed: true,
                error: Some(format!("Directory does not exist or is not a dir: {}", root_dir)),
            };
            return Err(err_status);
        }

        let worktree_info = detect_git_worktree(root);
        let (is_worktree, parent_ws, branch, default_ws) = if let Some(ref wt) = worktree_info {
            (true, Some(wt.parent_workspace.clone()), wt.branch.clone(), wt.worktree_name.clone())
        } else {
            let abs_path = fs::canonicalize(root).unwrap_or_else(|_| root.to_path_buf());
            let name = abs_path
                .file_name()
                .map(|n| n.to_string_lossy().to_string())
                .unwrap_or_else(|| "default".to_string());
            (false, None, None, name)
        };

        let workspace_name = match project {
            Some(p) if !p.trim().is_empty() => p.trim().to_string(),
            _ => default_ws,
        };

        let abs_root = fs::canonicalize(root)
            .map(|p| p.to_string_lossy().to_string())
            .unwrap_or_else(|_| root_dir.to_string());

        // Check if workspace is already running or queued
        {
            let jobs = self.jobs.read().await;
            if let Some(existing) = jobs.get(&workspace_name) {
                if !existing.completed {
                    info!("Workspace '{}' is already in-flight or queued (phase: {})", workspace_name, existing.phase);
                    return Err(existing.clone());
                }
            }
        }

        let queue_pos = {
            let jobs = self.jobs.read().await;
            jobs.values().filter(|j| !j.completed).count()
        };

        let initial_phase = if queue_pos == 0 {
            if refresh { "purging" } else { "scanning" }
        } else {
            "queued"
        };

        let job_status = IngestionJobStatus {
            workspace: workspace_name.clone(),
            path: abs_root.clone(),
            is_refresh: refresh,
            phase: initial_phase.to_string(),
            queue_position: queue_pos,
            files_scanned: 0,
            files_indexed: 0,
            files_skipped: 0,
            total_files: 0,
            nodes_created: 0,
            edges_created: 0,
            clusters_computed: 0,
            started_at: now_ms,
            elapsed_ms: 0,
            completed: false,
            error: None,
        };

        {
            let mut jobs = self.jobs.write().await;
            jobs.insert(workspace_name.clone(), job_status.clone());
        }

        if is_async {
            let task = IngestTask {
                root_dir: abs_root,
                project: project.map(|s| s.to_string()),
                refresh,
                auto_watch,
                workspace_name,
                is_worktree,
                parent_ws,
                branch,
                response_tx: None,
            };
            let _ = self.queue_tx.send(task).await;
            Err(job_status)
        } else {
            let (tx, rx) = oneshot::channel();
            let task = IngestTask {
                root_dir: abs_root,
                project: project.map(|s| s.to_string()),
                refresh,
                auto_watch,
                workspace_name,
                is_worktree,
                parent_ws,
                branch,
                response_tx: Some(tx),
            };
            let _ = self.queue_tx.send(task).await;
            Ok(rx)
        }
    }

    pub async fn ingest_directory(
        &self,
        root_dir: &str,
        project: Option<&str>,
        refresh: bool,
    ) -> Result<IngestResult, String> {
        match self.enqueue_ingest(root_dir, project, refresh, true, false).await {
            Ok(rx) => match rx.await {
                Ok(inner) => inner,
                Err(e) => Err(format!("Ingestion task terminated unexpectedly: {}", e)),
            },
            Err(status) => Err(status.error.unwrap_or_else(|| "Failed to enqueue task".to_string())),
        }
    }
}

struct PendingBatchFile {
    rel_path: String,
    hash: String,
    nodes: Vec<ExtractedNode>,
    containment_edges: Vec<ExtractedEdge>,
    text_count: usize,
}

#[allow(clippy::too_many_arguments)]
async fn flush_vector_batch(
    workspace_name: &str,
    pending_files: &mut Vec<PendingBatchFile>,
    pending_texts: &mut Vec<String>,
    embedder: &EmbedderClient,
    db: &DbClient,
    cache: &FileCache,
    files_indexed: &mut usize,
    total_nodes: &mut usize,
    total_edges: &mut usize,
) {
    if pending_files.is_empty() {
        return;
    }
    let text_refs: Vec<&str> = pending_texts.iter().map(|s| s.as_str()).collect();
    let embeddings = match embedder.embed_batch(&text_refs).await {
        Ok(embs) => embs,
        Err(e) => {
            warn!("Batch embedding generation failed for {} texts: {}", text_refs.len(), e);
            vec![vec![0.0f32; 384]; pending_texts.len()]
        }
    };

    let mut offset = 0;
    for file in pending_files.drain(..) {
        let end = offset + file.text_count;
        let file_embeddings = if end <= embeddings.len() {
            &embeddings[offset..end]
        } else {
            &[]
        };
        offset = end;

        // Prune any prior nodes/edges for this file to avoid zombie duplicates
        let _ = db.delete_file(workspace_name, &file.rel_path).await;

        let mut stored_nodes_ok = false;
        if let Err(e) = db.store_nodes(&file.nodes, file_embeddings, Some(&file.hash)).await {
            error!("Failed to store nodes for {}: {}", file.rel_path, e);
        } else {
            *total_nodes += file.nodes.len();
            stored_nodes_ok = true;
        }

        if let Err(e) = db.store_resolved_edges(&file.containment_edges).await {
            error!("Failed to store containment edges for {}: {}", file.rel_path, e);
        } else {
            *total_edges += file.containment_edges.len();
        }

        if stored_nodes_ok {
            cache.update(&file.rel_path, &file.hash).await;
        }
        *files_indexed += 1;
    }
    pending_texts.clear();
}

async fn execute_ingest(
    db: &DbClient,
    embedder: &EmbedderClient,
    cache: &FileCache,
    jobs: &Arc<RwLock<HashMap<String, IngestionJobStatus>>>,
    watcher_slot: &Arc<RwLock<Option<Arc<crate::watcher::WatchManager>>>>,
    queue_tx: &mpsc::Sender<IngestTask>,
    task: IngestTask,
) {
    let start_time = std::time::Instant::now();
    let root_dir = &task.root_dir;
    let workspace_name = &task.workspace_name;
    info!("FIFO Worker: Starting ingestion for directory: {} (workspace: '{}')", root_dir, workspace_name);

    let root = Path::new(root_dir);
    if !root.exists() || !root.is_dir() {
        let err = format!("Directory does not exist or is not a dir: {}", root_dir);
        let mut j_map = jobs.write().await;
        if let Some(j) = j_map.get_mut(workspace_name) {
            j.phase = "failed".to_string();
            j.completed = true;
            j.error = Some(err.clone());
            j.elapsed_ms = start_time.elapsed().as_millis() as u64;
        }
        if let Some(tx) = task.response_tx {
            let _ = tx.send(Err(err));
        }
        return;
    }

    let abs_root = fs::canonicalize(root)
        .map(|p| p.to_string_lossy().to_string())
        .unwrap_or_else(|_| root_dir.to_string());

    let _ = db.record_workspace_meta(
        workspace_name,
        &abs_root,
        task.is_worktree,
        task.parent_ws.as_deref(),
        task.branch.as_deref(),
        if task.is_worktree { Some(workspace_name.as_str()) } else { None },
    ).await;

    // Guaranteed Clean Invariant: purge existing workspace data in SurrealDB
    info!("Running clean purge for workspace '{}' to guarantee zero stale/orphaned nodes", workspace_name);
    match db.purge_workspace(workspace_name).await {
        Ok(purged) => info!("Cleanly purged {} prior nodes for workspace '{}'", purged, workspace_name),
        Err(e) => warn!("Clean purge notice for '{}': {}", workspace_name, e),
    }
    cache.clear().await;
    {
        let mut j_map = jobs.write().await;
        if let Some(j) = j_map.get_mut(workspace_name) {
            j.phase = "scanning".to_string();
            j.elapsed_ms = start_time.elapsed().as_millis() as u64;
        }
    }

        let mut files_scanned = 0;
        let mut files_indexed = 0;
        let mut files_skipped = 0;
        let mut total_nodes = 0;
        let mut total_edges = 0;
        let mut discovered_dirs: HashSet<String> = HashSet::new();
        let mut symbol_table = SymbolTable::new();
        let mut pending_files: Vec<PendingBatchFile> = Vec::new();
        let mut pending_texts: Vec<String> = Vec::new();

        let root_path_buf = PathBuf::from(root);
        let gitignore = crate::watcher::load_workspace_gitignore(&root_path_buf);

        for entry in WalkDir::new(root)
            .into_iter()
            .filter_entry(|e| {
                if e.depth() == 0 {
                    return true;
                }
                let p = e.path();
                let is_dir = e.file_type().is_dir();
                !crate::watcher::is_path_ignored(&root_path_buf, &gitignore, p, is_dir)
            })
            .filter_map(|e| e.ok())
        {
            let path = entry.path();
            if !path.is_file() {
                continue;
            }

            let file_name = path.file_name().and_then(|n| n.to_str()).unwrap_or("");
            if file_name == ".DS_Store" || file_name == ".gitignore" || file_name == "Thumbs.db" {
                continue;
            }

            // Skip oversized files (> 512 KB) to prevent OOM / computational explosion
            if entry.metadata().map(|m| m.len()).unwrap_or(0) > 512 * 1024 {
                continue;
            }

            let ext = path.extension().and_then(|s| s.to_str()).unwrap_or("");
            // Skip common binary artifacts, data dumps, and noise
            if matches!(
                ext,
                "png" | "jpg" | "jpeg" | "gif" | "webp" | "ico" | "svg" | "pdf"
                    | "zip" | "tar" | "gz" | "bz2" | "xz" | "7z"
                    | "exe" | "bin" | "dll" | "dylib" | "so" | "a" | "o" | "obj"
                    | "wasm" | "pyc" | "pyo" | "pyd" | "class" | "jar"
                    | "lock" | "map" | "rlib" | "rmeta" | "timestamp"
                    | "out" | "parquet" | "arrow" | "npy" | "npz" | "h5" | "pkl" | "pt" | "pth"
            ) {
                continue;
            }

            // For extensionless files or scripts, peek first 512 bytes for binary null bytes
            let raw_bytes = match fs::read(path) {
                Ok(b) => b,
                Err(_) => continue,
            };
            let peek_len = raw_bytes.len().min(512);
            if raw_bytes[..peek_len].contains(&0) {
                // Binary file containing null bytes (ELF, Mach-O, raw data), skip cleanly
                continue;
            }

            let content = match String::from_utf8(raw_bytes) {
                Ok(c) => c,
                Err(_) => {
                    // Non-UTF-8 binary or unreadable file, skip cleanly
                    continue;
                }
            };

            files_scanned += 1;

            // Compute hash for staleness tracking (R-011)
            let mut hasher = Sha256::new();
            hasher.update(content.as_bytes());
            let hash = hex::encode(hasher.finalize());

            // Derive relative file path for clean and portable node IDs
            let rel_path = path
                .strip_prefix(root)
                .unwrap_or(path)
                .to_string_lossy()
                .to_string();

            // Collect directory hierarchy for Topological Containment Backbone (R-063)
            let mut cur_dir = Path::new(&rel_path).parent();
            while let Some(parent) = cur_dir {
                let p_str = parent.to_string_lossy().to_string();
                if p_str.is_empty() || p_str == "." {
                    break;
                }
                discovered_dirs.insert(p_str);
                cur_dir = parent.parent();
            }

            if !cache.is_stale(&rel_path, &hash).await {
                files_skipped += 1;
                continue;
            }

            // Parse AST or structured document, with universal fallback chunker (Guarantees zero dropped files)
            let parse_res = CodeParser::parse_file(workspace_name, &rel_path, &content)
                .filter(|pr| !pr.nodes.is_empty())
                .or_else(|| CodeParser::parse_fallback(workspace_name, &rel_path, &content));

            if let Some(parse_res) = parse_res {
                if parse_res.nodes.is_empty() {
                    continue;
                }

                // Topological Containment Backbone (R-063): Synthesize File node and containment edges
                let filename = Path::new(&rel_path)
                    .file_name()
                    .map(|n| n.to_string_lossy().to_string())
                    .unwrap_or_else(|| rel_path.clone());
                let ext = Path::new(&rel_path)
                    .extension()
                    .map(|e| e.to_string_lossy().to_string())
                    .unwrap_or_default();
                let total_lines = content.lines().count().max(1);
                let preview_text = match content.char_indices().nth(500) {
                    Some((idx, _)) => &content[..idx],
                    None => &content,
                };
                let file_node_id = format!("{}:file:{}", workspace_name, rel_path);
                let file_node = ExtractedNode {
                    id: file_node_id.clone(),
                    workspace: workspace_name.clone(),
                    label: filename.clone(),
                    kind: "file".to_string(),
                    file_path: rel_path.clone(),
                    language: ext,
                    line_start: 1,
                    line_end: total_lines,
                    text: preview_text.to_string(),
                };

                let mut all_nodes = Vec::with_capacity(parse_res.nodes.len() + 1);
                all_nodes.push(file_node);
                all_nodes.extend(parse_res.nodes);

                let mut all_edges = parse_res.edges;
                // Emit CONTAINS edge from File to each extracted symbol
                for sym in &all_nodes[1..] {
                    all_edges.push(ExtractedEdge {
                        workspace: workspace_name.clone(),
                        source_id: file_node_id.clone(),
                        target_label: sym.label.clone(),
                        target_id: Some(sym.id.clone()),
                        edge_type: "CONTAINS".to_string(),
                        category: "CONTAINMENT".to_string(),
                    });
                }

                // Emit CONTAINS_FILE edge from parent directory (or workspace) to File
                let parent_dir = Path::new(&rel_path)
                    .parent()
                    .map(|p| p.to_string_lossy().to_string())
                    .unwrap_or_default();
                let dir_source_id = if parent_dir.is_empty() || parent_dir == "." {
                    format!("{}:workspace", workspace_name)
                } else {
                    format!("{}:dir:{}", workspace_name, parent_dir)
                };
                all_edges.push(ExtractedEdge {
                    workspace: workspace_name.clone(),
                    source_id: dir_source_id,
                    target_label: filename.clone(),
                    target_id: Some(file_node_id),
                    edge_type: "CONTAINS_FILE".to_string(),
                    category: "CONTAINMENT".to_string(),
                });

                // Register nodes in in-memory symbol table (R-078 Pass 1)
                symbol_table.register_file_nodes(&all_nodes);

                // Register explicit imports for SAIR Tier 2 (R-082)
                let imported_symbols: Vec<String> = all_edges
                    .iter()
                    .filter(|e| e.edge_type == "IMPORTS")
                    .map(|e| e.target_label.clone())
                    .collect();
                symbol_table.register_file_imports(workspace_name, &rel_path, &imported_symbols);

                // Separate containment edges (immediate) from cross-file edges (deferred to Pass 2)
                let mut containment_edges = Vec::new();
                let mut cross_file_edges = Vec::new();
                for edge in all_edges {
                    if edge.target_id.is_some() {
                        containment_edges.push(edge);
                    } else {
                        cross_file_edges.push(edge);
                    }
                }
                symbol_table.queue_edges(cross_file_edges);

                // Batch vector embeddings across files into chunks of up to 256 (R-081)
                let text_count = all_nodes.len();
                for n in &all_nodes {
                    pending_texts.push(format!("[{}] {} {} in {}\n{}", n.language, n.kind, n.label, n.file_path, n.text));
                }
                pending_files.push(PendingBatchFile {
                    rel_path,
                    hash,
                    nodes: all_nodes,
                    containment_edges,
                    text_count,
                });

                if pending_texts.len() >= 256 {
                    flush_vector_batch(
                        workspace_name,
                        &mut pending_files,
                        &mut pending_texts,
                        embedder,
                        db,
                        cache,
                        &mut files_indexed,
                        &mut total_nodes,
                        &mut total_edges,
                    ).await;
                }

                if files_scanned % 5 == 0 || files_indexed % 2 == 0 {
                    let mut j_map = jobs.write().await;
                    if let Some(job) = j_map.get_mut(workspace_name) {
                        job.phase = "indexing".to_string();
                        job.files_scanned = files_scanned;
                        job.files_indexed = files_indexed;
                        job.files_skipped = files_skipped;
                        job.nodes_created = total_nodes;
                        job.edges_created = total_edges;
                        job.elapsed_ms = start_time.elapsed().as_millis() as u64;
                    }
                }
            }
        }

        // Flush any remaining pending files in vector batch (R-081)
        flush_vector_batch(
            workspace_name,
            &mut pending_files,
            &mut pending_texts,
            embedder,
            db,
            cache,
            &mut files_indexed,
            &mut total_nodes,
            &mut total_edges,
        ).await;

        // Store Topological Containment Backbone: Workspace root, directories, and CONTAINS_DIR edges (R-063)
        let ws_root_node = ExtractedNode {
            id: format!("{}:workspace", workspace_name),
            workspace: workspace_name.clone(),
            label: workspace_name.clone(),
            kind: "workspace".to_string(),
            file_path: String::new(),
            language: String::new(),
            line_start: 0,
            line_end: 0,
            text: format!("Workspace root: {}", workspace_name),
        };

        let mut dir_nodes = Vec::with_capacity(discovered_dirs.len() + 1);
        let mut dir_edges = Vec::with_capacity(discovered_dirs.len());
        dir_nodes.push(ws_root_node);

        for dir in &discovered_dirs {
            let dir_name = Path::new(dir)
                .file_name()
                .map(|n| n.to_string_lossy().to_string())
                .unwrap_or_else(|| dir.clone());
            dir_nodes.push(ExtractedNode {
                id: format!("{}:dir:{}", workspace_name, dir),
                workspace: workspace_name.clone(),
                label: dir_name.clone(),
                kind: "directory".to_string(),
                file_path: dir.clone(),
                language: String::new(),
                line_start: 0,
                line_end: 0,
                text: format!("Directory: {}", dir),
            });

            let parent_of_dir = Path::new(dir)
                .parent()
                .map(|p| p.to_string_lossy().to_string())
                .unwrap_or_default();
            let parent_id = if parent_of_dir.is_empty() || parent_of_dir == "." {
                format!("{}:workspace", workspace_name)
            } else {
                format!("{}:dir:{}", workspace_name, parent_of_dir)
            };
            dir_edges.push(ExtractedEdge {
                workspace: workspace_name.clone(),
                source_id: parent_id,
                target_label: dir_name,
                target_id: Some(format!("{}:dir:{}", workspace_name, dir)),
                edge_type: "CONTAINS_DIR".to_string(),
                category: "CONTAINMENT".to_string(),
            });
        }

        if let Err(e) = db.store_nodes(&dir_nodes, &[], None).await {
            warn!("Failed to store directory backbone nodes: {}", e);
        } else {
            total_nodes += dir_nodes.len();
        }
        if let Err(e) = db.store_edges(&dir_edges).await {
            warn!("Failed to store directory backbone edges: {}", e);
        } else {
            total_edges += dir_edges.len();
        }

        // PASS 2: DEFERRED LINKAGE PASS (R-078)
        // Resolve all cross-file forward references in-memory against complete SymbolTable
        info!(
            "Starting Pass 2 (Deferred SAIR Linkage) for '{}': {} symbols declared, {} cross-file edges queued",
            workspace_name, symbol_table.total_symbols(), symbol_table.total_unresolved_queued()
        );
        let (resolved_edges, resolved_count, dropped_count) = symbol_table.resolve_all_queued_edges();
        info!(
            "Pass 2 in-memory resolution completed for '{}': {} edges resolved, {} unresolvable dropped",
            workspace_name, resolved_count, dropped_count
        );

        if !resolved_edges.is_empty() {
            if let Err(e) = db.store_resolved_edges(&resolved_edges).await {
                error!("Failed to bulk store Pass 2 resolved edges for '{}': {}", workspace_name, e);
            } else {
                total_edges += resolved_count;
                info!("Pass 2 successfully stored {} cross-file edges for '{}'", resolved_count, workspace_name);
            }
        }

        // Auto-cluster upon ingestion: Compute Louvain/Leiden modularity communities
        {
            let mut j_map = jobs.write().await;
            if let Some(job) = j_map.get_mut(workspace_name) {
                job.phase = "clustering".to_string();
                job.files_scanned = files_scanned;
                job.files_indexed = files_indexed;
                job.files_skipped = files_skipped;
                job.nodes_created = total_nodes;
                job.edges_created = total_edges;
                job.elapsed_ms = start_time.elapsed().as_millis() as u64;
            }
        }

        let mut clusters_computed = 0;
        if total_nodes > 0 {
            if let Ok((all_nodes, all_links)) = db.get_graph(Some(workspace_name)).await {
                if !all_nodes.is_empty() {
                    let seeds = if let Some(ref p_ws) = task.parent_ws {
                        // R-053: Seed-Preserving Galaxy Inheritance
                        match db.get_community_seeds_for_parent(p_ws, workspace_name).await {
                            Ok(s) if !s.is_empty() => {
                                info!("Inherited {} community seed priors from parent workspace '{}'", s.len(), p_ws);
                                Some(s)
                            },
                            _ => None,
                        }
                    } else {
                        None
                    };

                    let mut assignments = CommunityDetector::detect_with_seeds(&all_nodes, &all_links, 15, seeds.as_ref());
                    CommunityDetector::consolidate_communities(&all_nodes, &all_links, &mut assignments, 3);
                    CommunityDetector::coarsen_to_macro_galaxies(&all_nodes, &all_links, &mut assignments, 0);
                    let (summaries, galaxy_records) = CommunityDetector::compute_galaxy_metrics(workspace_name, &all_nodes, &all_links, &assignments);
                    clusters_computed = summaries.len();
                    if let Err(e) = db.update_communities(&assignments).await {
                        warn!("Auto-clustering failed after ingestion for '{}': {}", workspace_name, e);
                    } else {
                        let _ = db.store_galaxies(workspace_name, &galaxy_records).await;
                        info!("Auto-clustered '{}' into {} modular communities", workspace_name, clusters_computed);
                    }
                }
            }
        }

        let duration_ms = start_time.elapsed().as_millis() as u64;
        info!(
            "Ingestion completed for '{}': {} files scanned, {} indexed, {} skipped, {} clusters in {}ms",
            workspace_name, files_scanned, files_indexed, files_skipped, clusters_computed, duration_ms
        );

        {
            let mut j_map = jobs.write().await;
            if let Some(job) = j_map.get_mut(workspace_name) {
                job.phase = "completed".to_string();
                job.completed = true;
                job.files_scanned = files_scanned;
                job.files_indexed = files_indexed;
                job.files_skipped = files_skipped;
                job.nodes_created = total_nodes;
                job.edges_created = total_edges;
                job.clusters_computed = clusters_computed;
                job.elapsed_ms = duration_ms;
                job.error = None;
            }
        }

        // Live watch enrollment
        if task.auto_watch {
            let watcher_opt = {
                let lock = watcher_slot.read().await;
                lock.clone()
            };
            if let Some(watcher) = watcher_opt {
                match watcher.start_watch(&task.root_dir, workspace_name, None).await {
                    Ok(_) => info!("Auto-started live watch for workspace '{}'", workspace_name),
                    Err(e) => warn!("Auto-watch registration returned for '{}': {}", workspace_name, e),
                }
            }
        }

        // Worktree Auto-Cascade: Enqueue discovered worktrees sequentially
        if !task.is_worktree {
            let parent_path = std::path::PathBuf::from(&task.root_dir);
            let discovered_worktrees = discover_parent_worktrees(&parent_path);
            for wt in discovered_worktrees {
                if let Some(ref wt_root) = wt.root_path {
                    info!("Enqueuing child worktree '{}' for sequential cascade...", wt.worktree_name);
                    let wt_task = IngestTask {
                        root_dir: wt_root.clone(),
                        project: Some(wt.worktree_name.clone()),
                        refresh: true,
                        auto_watch: task.auto_watch,
                        workspace_name: wt.worktree_name.clone(),
                        is_worktree: true,
                        parent_ws: Some(workspace_name.clone()),
                        branch: wt.branch.clone(),
                        response_tx: None,
                    };
                    let _ = queue_tx.send(wt_task).await;
                }
            }
        }

        let ingest_res = IngestResult {
            workspace: workspace_name.clone(),
            files_scanned,
            files_indexed,
            files_skipped,
            nodes_created: total_nodes,
            edges_created: total_edges,
            clusters_computed,
            duration_ms,
            is_worktree: task.is_worktree,
            parent_workspace: task.parent_ws.clone(),
            branch: task.branch.clone(),
        };

        if let Some(tx) = task.response_tx {
            let _ = tx.send(Ok(ingest_res));
        }
}
