// In-Memory Symbol Table & SAIR Target Resolver (R-078, R-082)
// Decouples symbol declarations (Pass 1) from cross-file reference resolution (Pass 2)

use crate::parser::{ExtractedEdge, ExtractedNode};
use std::collections::HashMap;
use std::path::Path;

#[derive(Clone, Debug)]
pub struct SymbolDef {
    pub id: String,
    pub file_path: String,
    #[allow(dead_code)]
    pub kind: String,
    #[allow(dead_code)]
    pub line_start: usize,
}

#[derive(Clone, Debug, Default)]
pub struct SymbolTable {
    // (workspace, symbol_label) -> Vec<SymbolDef>
    symbols: HashMap<(String, String), Vec<SymbolDef>>,
    // (workspace, file_path) -> Vec<String> (imported labels)
    file_imports: HashMap<(String, String), Vec<String>>,
    // Unresolved cross-file edges accumulated during Pass 1
    unresolved_edges: Vec<ExtractedEdge>,
}

impl SymbolTable {
    pub fn new() -> Self {
        Self::default()
    }

    /// Register extracted nodes for a file into the symbol table (Pass 1)
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
                        line_start: node.line_start,
                    });
            }
        }
    }

    /// Register explicit imports for a file into the symbol table (Pass 1, SAIR Tier 2)
    pub fn register_file_imports(&mut self, workspace: &str, file_path: &str, imported_symbols: &[String]) {
        if !imported_symbols.is_empty() {
            self.file_imports
                .entry((workspace.to_string(), file_path.to_string()))
                .or_default()
                .extend(imported_symbols.iter().cloned());
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
            "new" | "run" | "init" | "start" | "stop" | "handle" | "execute"
                | "get" | "set" | "update" | "create" | "delete" | "parse"
                | "build" | "render" | "test" | "main"
        )
    }

    /// Resolve an individual edge against the symbol table using 4-tier SAIR (Pass 2)
    pub fn resolve_edge(&self, edge: &ExtractedEdge) -> Option<String> {
        let candidates = self.symbols.get(&(edge.workspace.clone(), edge.target_label.clone()))?;
        if candidates.is_empty() {
            return None;
        }

        let source_file = edge.source_id.split(':').nth(1).unwrap_or("");
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

        // Tier 2: Explicit Imports Scope (SAIR R-065 / R-082)
        if let Some(imports) = self.file_imports.get(&(edge.workspace.clone(), source_file.to_string())) {
            if imports.iter().any(|imp| imp == &edge.target_label) {
                // If the file explicitly imported this symbol, check if candidate file matches import path or return best candidate
                return Some(candidates[0].id.clone());
            }
        }

        // Tier 3: Sibling Directory / Package Scope
        if !parent_dir.is_empty() {
            for cand in candidates {
                if cand.file_path.starts_with(&parent_dir) {
                    return Some(cand.id.clone());
                }
            }
        }

        // Tier 4: Global Workspace Scope (excluding generic verbs)
        if !Self::is_generic_verb(&edge.target_label) {
            return Some(candidates[0].id.clone());
        }

        None
    }

    /// Resolve all queued edges in Pass 2 and return resolved edges ready for bulk database write
    pub fn resolve_all_queued_edges(&mut self) -> (Vec<ExtractedEdge>, usize, usize) {
        let edges_to_resolve = std::mem::take(&mut self.unresolved_edges);
        let mut resolved_edges = Vec::with_capacity(edges_to_resolve.len());
        let mut resolved_count = 0;
        let mut dropped_count = 0;

        for mut edge in edges_to_resolve {
            if let Some(target_id) = self.resolve_edge(&edge) {
                edge.target_id = Some(target_id);
                resolved_edges.push(edge);
                resolved_count += 1;
            } else {
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
}
