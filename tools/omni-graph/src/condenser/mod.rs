// Subgraph Context Condenser: Multi-hop AST trace compression
// Implements: R-028, R-068, R-069 (Sub-1500 token high-fidelity context with PPR & clean Markdown)

use crate::db::{clean_record_id, DbLink, DbNode};
use serde::Serialize;
use std::collections::{HashMap, HashSet, VecDeque};

#[derive(Serialize, Debug)]
pub struct CondensedContext {
    pub root_symbol: String,
    pub token_estimate: usize,
    pub formatted_markdown: String,
    pub direct_callers: Vec<String>,
    pub direct_callees: Vec<String>,
    pub related_files: Vec<String>,
}

pub struct ContextCondenser;

impl ContextCondenser {
    /// Compute Personalized PageRank (PPR) restart vector centered on root nodes (R-069)
    /// Runs power iteration with alpha = 0.85 to prioritize high-centrality call chain neighbors
    pub fn compute_ppr_scores(
        root_node_ids: &HashSet<String>,
        candidate_ids: &HashSet<String>,
        links: &[DbLink],
        iterations: usize,
    ) -> HashMap<String, f32> {
        let alpha = 0.85f32;
        let mut adj: HashMap<&str, Vec<&str>> = HashMap::new();

        for link in links {
            let s = link.source.as_str();
            let t = link.target.as_str();
            if candidate_ids.contains(s) && candidate_ids.contains(t) {
                adj.entry(s).or_default().push(t);
                adj.entry(t).or_default().push(s); // bidirectional graph view for call context
            }
        }

        let num_roots = root_node_ids.len().max(1) as f32;
        let mut p: HashMap<&str, f32> = HashMap::new();
        let mut p0: HashMap<&str, f32> = HashMap::new();

        for id in candidate_ids {
            let initial = if root_node_ids.contains(id) {
                1.0 / num_roots
            } else {
                0.0
            };
            p.insert(id.as_str(), initial);
            if initial > 0.0 {
                p0.insert(id.as_str(), initial);
            }
        }

        let mut next_p: HashMap<&str, f32> = HashMap::with_capacity(candidate_ids.len());
        for _ in 0..iterations {
            next_p.clear();

            for (&u, neighbors) in &adj {
                let u_score = p.get(u).copied().unwrap_or(0.0);
                if u_score <= 0.0 || neighbors.is_empty() {
                    continue;
                }
                let share = (alpha * u_score) / (neighbors.len() as f32);
                for &v in neighbors {
                    *next_p.entry(v).or_insert(0.0) += share;
                }
            }

            // Inject restart mass from roots
            for (&root_id, &init_val) in &p0 {
                *next_p.entry(root_id).or_insert(0.0) += (1.0 - alpha) * init_val;
            }

            p = next_p.clone();
        }

        p.into_iter().map(|(k, v)| (k.to_string(), v)).collect()
    }

    /// Condense a multi-hop neighborhood around a node or symbol into an agent-ready markdown block
    pub fn condense(
        target_symbol: &str,
        nodes: &[DbNode],
        links: &[DbLink],
        max_hops: usize,
    ) -> CondensedContext {
        let mut direct_callers = Vec::new();
        let mut direct_callees = Vec::new();
        let mut related_files_set = HashSet::new();

        // 1. Locate root node(s) matching symbol or label
        let root_nodes: Vec<&DbNode> = nodes
            .iter()
            .filter(|n| n.label.eq_ignore_ascii_case(target_symbol) || n.id == target_symbol)
            .collect();

        let mut root_node_ids = HashSet::new();
        for root in &root_nodes {
            root_node_ids.insert(root.id.clone());
            related_files_set.insert(root.file_path.clone());
        }

        // 2. BFS traversal up to max_hops
        const MAX_CONDENSED_NODES: usize = 60;
        let mut candidate_node_ids = HashSet::new();
        let mut queue = VecDeque::new();
        for root in &root_nodes {
            candidate_node_ids.insert(root.id.clone());
            queue.push_back((root.id.clone(), 0));
        }

        while let Some((curr_id, depth)) = queue.pop_front() {
            if depth >= max_hops {
                continue;
            }

            for link in links {
                if link.source == curr_id {
                    // Outgoing: curr CALLS target
                    if candidate_node_ids.insert(link.target.clone()) {
                        queue.push_back((link.target.clone(), depth + 1));
                    }
                    if depth == 0 {
                        direct_callees.push(link.target.clone());
                    }
                } else if link.target == curr_id {
                    // Incoming: source CALLS curr
                    if candidate_node_ids.insert(link.source.clone()) {
                        queue.push_back((link.source.clone(), depth + 1));
                    }
                    if depth == 0 {
                        direct_callers.push(link.source.clone());
                    }
                }
            }
        }

        // Personalized PageRank (PPR) Neighborhood Slicing (R-069)
        // If candidates exceed MAX_CONDENSED_NODES, prioritize nodes with highest PPR centrality
        let mut relevant_node_ids = HashSet::new();
        let truncated_count;

        if candidate_node_ids.len() > MAX_CONDENSED_NODES {
            let ppr_scores = Self::compute_ppr_scores(&root_node_ids, &candidate_node_ids, links, 25);
            let mut sorted_candidates: Vec<String> = candidate_node_ids.into_iter().collect();
            sorted_candidates.sort_by(|a, b| {
                let sa = ppr_scores.get(a.as_str()).copied().unwrap_or(0.0);
                let sb = ppr_scores.get(b.as_str()).copied().unwrap_or(0.0);
                sb.partial_cmp(&sa).unwrap_or(std::cmp::Ordering::Equal)
            });

            // Retain all root nodes plus top central neighbors
            for root_id in &root_node_ids {
                relevant_node_ids.insert(root_id.clone());
            }
            for nid in sorted_candidates {
                if relevant_node_ids.len() >= MAX_CONDENSED_NODES {
                    break;
                }
                relevant_node_ids.insert(nid);
            }
            truncated_count = links.iter().filter(|l| !relevant_node_ids.contains(&l.source) || !relevant_node_ids.contains(&l.target)).count();
        } else {
            relevant_node_ids = candidate_node_ids;
            truncated_count = 0;
        }

        // 3. Format compact Markdown block with strict two-tier token budgeting (R-028, R-068)
        // Reserves at least 35% of token budget (~2100 chars) for Section 4 structural traces
        const MAX_OUTPUT_CHARS: usize = 5800;
        const MAX_SYMBOLS_CHARS: usize = 3600;

        let mut md = String::new();
        md.push_str(&format!("### 🧭 Omni-Graph AST Subgraph: `{}`\n\n", target_symbol));
        md.push_str("> High-fidelity deterministic AST slice (condensed for agent reasoning).\n\n");

        md.push_str("#### Identified Symbols & Signatures\n");
        let mut symbols_overflow = false;
        for n in nodes.iter().filter(|n| relevant_node_ids.contains(&n.id)) {
            related_files_set.insert(n.file_path.clone());

            let is_root = root_node_ids.contains(&n.id);
            let snippet = if is_root {
                n.text.trim().to_string()
            } else {
                let first_line = n.text.lines().next().unwrap_or("").trim();
                if first_line.len() > 140 {
                    format!("{}...", &first_line[..137])
                } else {
                    first_line.to_string()
                }
            };

            let entry = format!(
                "- **`{}`** ({}) in [`{}:{}`]({})\n  ```{}\n  {}\n  ```\n",
                n.label, n.kind, n.file_path, n.line_start, n.file_path, n.language, snippet
            );

            if md.len() + entry.len() > MAX_SYMBOLS_CHARS {
                symbols_overflow = true;
                break;
            }
            md.push_str(&entry);
        }

        if symbols_overflow {
            md.push_str("\n> ℹ️ Additional neighbor signatures condensed to preserve call trace budget.\n");
        }

        if truncated_count > 0 {
            md.push_str(&format!(
                "\n> ⚠️ Graph truncated: {} additional edges skipped (node cap: {}).\n",
                truncated_count, MAX_CONDENSED_NODES
            ));
        }

        md.push_str("\n#### Structural Relationships (Call / Import Traces)\n");
        for link in links.iter().filter(|l| relevant_node_ids.contains(&l.source) || relevant_node_ids.contains(&l.target)) {
            let src_clean = clean_record_id(&link.source);
            let tgt_clean = clean_record_id(&link.target);
            let edge_str = format!(
                "- `{}` ──[{}: {}]──▶ `{}`\n",
                src_clean, link.category, link.edge_type, tgt_clean
            );
            if md.len() + edge_str.len() > MAX_OUTPUT_CHARS {
                md.push_str(
                    "\n> ⚠️ Output truncated at ~1500 tokens. Use smaller HOPS or narrower workspace filter.\n"
                );
                break;
            }
            md.push_str(&edge_str);
        }

        let mut token_estimate = md.len() / 4; // Standard heuristic: 4 chars/token
        if token_estimate > 1495 {
            let max_bytes = 1490 * 4;
            if md.len() > max_bytes {
                let truncated = &md[..max_bytes];
                if let Some(last_nl) = truncated.rfind('\n') {
                    md.truncate(last_nl);
                    md.push_str("\n> ⚠️ Strict 1500 token ceiling applied.\n");
                }
                token_estimate = md.len() / 4;
            }
        }

        CondensedContext {
            root_symbol: target_symbol.to_string(),
            token_estimate,
            formatted_markdown: md,
            direct_callers,
            direct_callees,
            related_files: related_files_set.into_iter().collect(),
        }
    }
}
