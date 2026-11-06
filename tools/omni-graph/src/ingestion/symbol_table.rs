// In-Memory Symbol Table & SAIR Target Resolver (R-078, R-082, R-093, R-102, R-103, R-106)
// Decouples symbol declarations (Pass 1) from cross-file reference resolution (Pass 2)
// Provides reactive reverse dependency indexing and unresolved reference retention.

use crate::parser::{ExtractedEdge, ExtractedNode};
use std::collections::{HashMap, HashSet};
use std::path::Path;

#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct SymbolDef {
    pub id: String,
    pub file_path: String,
    pub kind: String,
    pub language: String,
    pub line_start: usize,
}

/// In-memory pool for retaining unresolved cross-file edges (R-103)
/// Automatically drained when newly registered symbols satisfy pending calls.
#[derive(Clone, Debug, Default)]
pub struct UnresolvedReferencePool {
    // (workspace, target_symbol_label) -> Vec<ExtractedEdge>
    pool: HashMap<(String, String), Vec<ExtractedEdge>>,
}

#[allow(dead_code)]
impl UnresolvedReferencePool {
    pub fn new() -> Self {
        Self::default()
    }

    /// Retain an unresolved cross-file edge waiting for target symbol declaration (R-103)
    pub fn retain_unresolved(&mut self, edge: ExtractedEdge) {
        self.pool
            .entry((edge.workspace.clone(), edge.target_label.clone()))
            .or_default()
            .push(edge);
    }

    /// Drain edges satisfied by a newly registered symbol declaration (R-103)
    pub fn drain_satisfied(&mut self, workspace: &str, symbol_label: &str) -> Vec<ExtractedEdge> {
        self.pool
            .remove(&(workspace.to_string(), symbol_label.to_string()))
            .unwrap_or_default()
    }

    /// Purge dangling edges originating from a file that was deleted or re-indexed
    pub fn remove_caller_file(&mut self, workspace: &str, file_path: &str) {
        self.pool.retain(|(ws, _), edges| {
            if ws == workspace {
                edges.retain(|e| {
                    let caller_file = e.source_id.split(':').nth(1).unwrap_or("");
                    caller_file != file_path
                });
                !edges.is_empty()
            } else {
                true
            }
        });
    }

    pub fn total_unresolved(&self) -> usize {
        self.pool.values().map(|v| v.len()).sum()
    }

    pub fn contains(&self, workspace: &str, symbol_label: &str) -> bool {
        self.pool
            .contains_key(&(workspace.to_string(), symbol_label.to_string()))
    }
}

#[derive(Clone, Debug, Default)]
pub struct SymbolTable {
    // (workspace, symbol_label) -> Vec<SymbolDef>
    symbols: HashMap<(String, String), Vec<SymbolDef>>,
    // (workspace, file_path) -> Vec<String> (imported labels)
    file_imports: HashMap<(String, String), Vec<String>>,
    // (workspace, target_symbol_label) -> HashSet<caller_file_path> (R-093)
    symbol_callers: HashMap<(String, String), HashSet<String>>,
    // Unresolved edge retention pool (R-103)
    unresolved_pool: UnresolvedReferencePool,
    // Unresolved cross-file edges accumulated during Pass 1
    unresolved_edges: Vec<ExtractedEdge>,
}

impl SymbolTable {
    pub fn new() -> Self {
        Self::default()
    }

    /// Detect programming language from file path extension (R-106)
    pub fn detect_language_from_path(path: &str) -> &str {
        let ext = path.rsplit('.').next().unwrap_or("");
        match ext {
            "rs" => "rust",
            "py" => "python",
            "js" | "jsx" => "javascript",
            "ts" | "tsx" => "typescript",
            "go" => "go",
            "java" => "java",
            "c" | "h" => "c",
            "cpp" | "cc" | "cxx" | "hpp" => "cpp",
            "sh" | "bash" => "shell",
            "sql" => "sql",
            "yaml" | "yml" => "yaml",
            "json" => "json",
            "toml" => "toml",
            "md" | "markdown" => "markdown",
            _ => "",
        }
    }

    /// Register extracted nodes for a file into the symbol table (Pass 1, R-106)
    pub fn register_file_nodes(&mut self, nodes: &[ExtractedNode]) {
        for node in nodes {
            if node.kind != "file" && node.kind != "workspace" && node.kind != "directory" {
                self.symbols
                    .entry((node.workspace.clone(), node.label.clone()))
                    .or_default()
                    .push(SymbolDef {
                        id: node.id.clone(),
                        file_path: node.file_path.clone(),
                        kind: node.kind.clone(),
                        language: if !node.language.is_empty() {
                            node.language.clone()
                        } else {
                            Self::detect_language_from_path(&node.file_path).to_string()
                        },
                        line_start: node.line_start,
                    });
            }
        }
    }

    /// Register explicit imports for a file into the symbol table (Pass 1, SAIR Tier 2)
    pub fn register_file_imports(
        &mut self,
        workspace: &str,
        file_path: &str,
        imported_symbols: &[String],
    ) {
        if !imported_symbols.is_empty() {
            self.file_imports
                .entry((workspace.to_string(), file_path.to_string()))
                .or_default()
                .extend(imported_symbols.iter().cloned());
        }
    }

    /// Register raw symbol definitions (used to populate SymbolTable from database on watcher boot)
    pub fn register_raw_symbols(
        &mut self,
        workspace: &str,
        raw_symbols: Vec<(String, SymbolDef)>,
    ) {
        for (label, mut def) in raw_symbols {
            if def.language.is_empty() {
                def.language = Self::detect_language_from_path(&def.file_path).to_string();
            }
            self.symbols
                .entry((workspace.to_string(), label))
                .or_default()
                .push(def);
        }
    }

    /// Record that a caller file depends on a target symbol (R-093)
    pub fn record_caller_dependency(
        &mut self,
        workspace: &str,
        caller_file: &str,
        target_label: &str,
    ) {
        self.symbol_callers
            .entry((workspace.to_string(), target_label.to_string()))
            .or_default()
            .insert(caller_file.to_string());
    }

    /// Retrieve all foreign caller files dependent on an exported symbol (R-093)
    #[allow(dead_code)]
    pub fn get_dependent_files(&self, workspace: &str, symbol_label: &str) -> HashSet<String> {
        self.symbol_callers
            .get(&(workspace.to_string(), symbol_label.to_string()))
            .cloned()
            .unwrap_or_default()
    }

    /// Retain an unresolved edge into the retention pool (R-103)
    #[allow(dead_code)]
    pub fn retain_unresolved(&mut self, edge: ExtractedEdge) {
        self.unresolved_pool.retain_unresolved(edge);
    }

    /// Drain pending unresolved edges satisfied by a target symbol (R-103)
    #[allow(dead_code)]
    pub fn drain_satisfied(&mut self, workspace: &str, symbol_label: &str) -> Vec<ExtractedEdge> {
        self.unresolved_pool.drain_satisfied(workspace, symbol_label)
    }

    /// Access the unresolved reference pool (R-103)
    #[allow(dead_code)]
    pub fn unresolved_pool(&self) -> &UnresolvedReferencePool {
        &self.unresolved_pool
    }

    /// Mutable access to the unresolved reference pool (R-103)
    #[allow(dead_code)]
    pub fn unresolved_pool_mut(&mut self) -> &mut UnresolvedReferencePool {
        &mut self.unresolved_pool
    }

    /// Remove all symbol definitions and imports associated with a specific file (incremental update)
    pub fn remove_file(&mut self, workspace: &str, file_path: &str) {
        self.symbols.retain(|(ws, _), defs| {
            if ws == workspace {
                defs.retain(|d| d.file_path != file_path);
                !defs.is_empty()
            } else {
                true
            }
        });
        self.file_imports
            .remove(&(workspace.to_string(), file_path.to_string()));

        // Prune caller dependencies originating from this file
        for callers in self.symbol_callers.values_mut() {
            callers.remove(file_path);
        }

        // Clean up unresolved pool entries originating from this file
        self.unresolved_pool.remove_caller_file(workspace, file_path);
    }

    /// Update file path references when a file is renamed
    pub fn rename_file(&mut self, workspace: &str, old_path: &str, new_path: &str) {
        for ((ws, _), defs) in self.symbols.iter_mut() {
            if ws == workspace {
                for def in defs.iter_mut() {
                    if def.file_path == old_path {
                        def.file_path = new_path.to_string();
                        def.language = Self::detect_language_from_path(new_path).to_string();
                    }
                }
            }
        }
        if let Some(imports) =
            self.file_imports.remove(&(workspace.to_string(), old_path.to_string()))
        {
            self.file_imports
                .insert((workspace.to_string(), new_path.to_string()), imports);
        }
        for callers in self.symbol_callers.values_mut() {
            if callers.remove(old_path) {
                callers.insert(new_path.to_string());
            }
        }
    }

    /// Queue unresolved cross-file edges for deferred resolution in Pass 2
    pub fn queue_edges(&mut self, edges: Vec<ExtractedEdge>) {
        for edge in edges {
            if edge.target_id.is_none() {
                self.unresolved_edges.push(edge);
            }
        }
    }

    /// Check if target_label is a generic common verb (R-065)
    pub fn is_generic_verb(label: &str) -> bool {
        matches!(
            label.to_lowercase().as_str(),
            "new"
                | "run"
                | "init"
                | "start"
                | "stop"
                | "handle"
                | "execute"
                | "get"
                | "set"
                | "update"
                | "create"
                | "delete"
                | "parse"
                | "build"
                | "render"
                | "test"
                | "main"
        )
    }

    /// Resolve an individual edge against the symbol table using 4-tier SAIR (Pass 2, R-106)
    pub fn resolve_edge(&self, edge: &ExtractedEdge) -> Option<String> {
        let candidates =
            self.symbols.get(&(edge.workspace.clone(), edge.target_label.clone()))?;
        if candidates.is_empty() {
            return None;
        }

        let source_file = edge.source_id.split(':').nth(1).unwrap_or("");
        let source_lang = Self::detect_language_from_path(source_file);
        let parent_dir = Path::new(source_file)
            .parent()
            .map(|p| p.to_string_lossy().to_string())
            .unwrap_or_default();

        // Tier 1: Local In-File Scope
        for cand in candidates {
            if cand.file_path == source_file {
                return Some(cand.id.clone());
            }
        }

        // Tier 2: Explicit Imports Scope (SAIR R-065 / R-082 / R-106)
        if let Some(imports) =
            self.file_imports.get(&(edge.workspace.clone(), source_file.to_string()))
        {
            if imports.iter().any(|imp| imp == &edge.target_label) {
                // If candidate matches same language, prioritize it (R-106)
                if !source_lang.is_empty() {
                    if let Some(cand) = candidates.iter().find(|c| c.language == source_lang) {
                        return Some(cand.id.clone());
                    }
                }
                return Some(candidates[0].id.clone());
            }
        }

        // Tier 3: Sibling Directory / Package Scope (R-106)
        if !parent_dir.is_empty() {
            if !source_lang.is_empty() {
                for cand in candidates {
                    if cand.file_path.starts_with(&parent_dir) && cand.language == source_lang {
                        return Some(cand.id.clone());
                    }
                }
            }
            for cand in candidates {
                if cand.file_path.starts_with(&parent_dir) {
                    return Some(cand.id.clone());
                }
            }
        }

        // Tier 4: Global Workspace Scope (excluding generic verbs, R-106 language-aware)
        if !Self::is_generic_verb(&edge.target_label) {
            if !source_lang.is_empty() {
                if let Some(cand) = candidates.iter().find(|c| c.language == source_lang) {
                    return Some(cand.id.clone());
                }
            }
            return Some(candidates[0].id.clone());
        }

        None
    }

    /// Resolve all queued edges in Pass 2 and return resolved edges ready for bulk database write (R-078, R-103)
    pub fn resolve_all_queued_edges(&mut self) -> (Vec<ExtractedEdge>, usize, usize) {
        let edges_to_resolve = std::mem::take(&mut self.unresolved_edges);
        let mut resolved_edges = Vec::with_capacity(edges_to_resolve.len());
        let mut resolved_count = 0;
        let mut dropped_count = 0;

        for mut edge in edges_to_resolve {
            let caller_file = edge.source_id.split(':').nth(1).unwrap_or("").to_string();
            if !caller_file.is_empty() {
                self.record_caller_dependency(&edge.workspace, &caller_file, &edge.target_label);
            }
            if let Some(target_id) = self.resolve_edge(&edge) {
                edge.target_id = Some(target_id);
                resolved_edges.push(edge);
                resolved_count += 1;
            } else {
                // Retain in unresolved pool instead of dropping into void (R-103)
                self.unresolved_pool.retain_unresolved(edge);
                dropped_count += 1;
            }
        }

        (resolved_edges, resolved_count, dropped_count)
    }

    #[allow(dead_code)]
    pub fn total_symbols(&self) -> usize {
        self.symbols.values().map(|v| v.len()).sum()
    }

    #[allow(dead_code)]
    pub fn total_unresolved_queued(&self) -> usize {
        self.unresolved_edges.len()
    }

    #[allow(dead_code)]
    pub fn total_unresolved_retained(&self) -> usize {
        self.unresolved_pool.total_unresolved()
    }
}
