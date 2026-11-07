// Graph RAG Analysis: Community Detection & Hybrid Graph-RAG Retrieval
// Implements: R-012 (Community Detection / Leiden clustering) and Graph-RAG Hybrid Synthesis

use crate::db::{DbClient, DbLink, DbNode, GalaxyRecord};
use crate::embedder::EmbedderClient;
use serde::{Deserialize, Serialize};
use std::collections::{HashMap, HashSet};

#[derive(Serialize, Deserialize, Debug, Clone)]
pub struct CommunitySummary {
    pub id: i32,
    pub name: String,
    pub node_count: usize,
    pub top_symbols: Vec<String>,
    pub files: Vec<String>,
}

#[derive(Serialize, Deserialize, Debug)]
pub struct GraphRagResponse {
    pub query: String,
    pub macroscopic_summary: String,
    pub communities_involved: Vec<CommunitySummary>,
    pub seed_symbols: Vec<String>,
    pub expanded_subgraph: String,
    pub token_estimate: usize,
}

/// Deterministic lightweight pseudo-random number generator (Xorshift64)
/// Used for shuffling node evaluation order in LPA to eliminate deterministic traversal bias (Finding #5)
pub struct SimpleRng(u64);

impl SimpleRng {
    pub fn new(seed: u64) -> Self {
        Self(if seed == 0 { 0x853c49e6748fea9b } else { seed })
    }

    pub fn next_u64(&mut self) -> u64 {
        let mut x = self.0;
        x ^= x << 13;
        x ^= x >> 7;
        x ^= x << 17;
        self.0 = x;
        x
    }

    pub fn shuffle<T>(&mut self, slice: &mut [T]) {
        for i in (1..slice.len()).rev() {
            let j = (self.next_u64() as usize) % (i + 1);
            slice.swap(i, j);
        }
    }
}

/// Assign weights to edges based on architectural significance
pub fn get_edge_weight(edge_type: &str) -> f32 {
    match edge_type {
        "IMPLEMENTS" => 3.0,
        "CALLS" => 2.0,
        "TYPE_REF" => 1.5,
        "IMPORTS" => 1.0,
        "CONTAINS" => 0.8,
        "CONTAINS_FILE" => 0.4,
        "CONTAINS_DIR" => 0.2,
        _ => 1.0,
    }
}

pub struct CommunityDetector;

impl CommunityDetector {
    /// Detect communities using iterative label propagation (high-speed O(E) modular clustering)
    pub fn detect(nodes: &[DbNode], links: &[DbLink], max_iterations: usize) -> HashMap<String, i32> {
        Self::detect_with_seeds(nodes, links, max_iterations, None)
    }

    /// Detect communities with warm-seeded initial assignments to prevent cluster ID thrashing across live edits
    pub fn detect_with_seeds(
        nodes: &[DbNode],
        links: &[DbLink],
        max_iterations: usize,
        seeds: Option<&HashMap<String, i32>>,
    ) -> HashMap<String, i32> {
        let mut labels: HashMap<String, i32> = HashMap::new();
        let mut adj: HashMap<String, Vec<(String, f32)>> = HashMap::new();

        // 1. Initialize node labels from seeds if available, otherwise allocate unique IDs
        for (i, node) in nodes.iter().enumerate() {
            let initial_label = seeds
                .and_then(|s| s.get(&node.id).copied())
                .unwrap_or(i as i32 + 100_000);
            labels.insert(node.id.clone(), initial_label);
            adj.insert(node.id.clone(), Vec::new());
        }

        // 2. Build weighted undirected adjacency list
        for link in links {
            if adj.contains_key(&link.source) && adj.contains_key(&link.target) {
                let weight = get_edge_weight(&link.edge_type);
                adj.get_mut(&link.source).unwrap().push((link.target.clone(), weight));
                adj.get_mut(&link.target).unwrap().push((link.source.clone(), weight));
            }
        }

        // 3. Iterative label propagation with randomized node ordering
        let mut rng = SimpleRng::new(0x4d595f5345454431);
        let mut node_indices: Vec<usize> = (0..nodes.len()).collect();

        let mut freq: HashMap<i32, f32> = HashMap::with_capacity(32);
        for _ in 0..max_iterations {
            let mut changed = false;
            rng.shuffle(&mut node_indices);
            for &idx in &node_indices {
                let node = &nodes[idx];
                let neighbors = match adj.get(&node.id) {
                    Some(n) if !n.is_empty() => n,
                    _ => continue,
                };

                // Accumulate weighted label frequencies
                freq.clear();
                for (neighbor_id, weight) in neighbors {
                    if let Some(lbl) = labels.get(neighbor_id) {
                        *freq.entry(*lbl).or_insert(0.0) += *weight;
                    }
                }

                // Select label with maximum accumulated edge weight, with deterministic tie-breaking on label ID
                if let Some((&best_label, _)) = freq.iter().max_by(|a, b| {
                    match a.1.partial_cmp(b.1).unwrap_or(std::cmp::Ordering::Equal) {
                        std::cmp::Ordering::Equal => a.0.cmp(b.0),
                        other => other,
                    }
                }) {
                    if let Some(curr_label) = labels.get_mut(&node.id) {
                        if *curr_label != best_label {
                            *curr_label = best_label;
                            changed = true;
                        }
                    }
                }
            }

            if !changed {
                break;
            }
        }

        // 4. Normalize labels to compact IDs while preserving established seed IDs
        let mut label_map: HashMap<i32, i32> = HashMap::new();
        let mut next_id = 0;

        // Reserve existing seed IDs first
        if let Some(seed_map) = seeds {
            for &seed_id in seed_map.values() {
                if let std::collections::hash_map::Entry::Vacant(e) = label_map.entry(seed_id) {
                    e.insert(seed_id);
                    if seed_id >= next_id {
                        next_id = seed_id + 1;
                    }
                }
            }
        }

        let mut normalized = HashMap::new();
        for (node_id, raw_label) in labels {
            let compact_id = *label_map.entry(raw_label).or_insert_with(|| {
                let id = next_id;
                next_id += 1;
                id
            });
            normalized.insert(node_id, compact_id);
        }

        normalized
    }

    /// Hierarchical directory- and topology-aware community consolidation (R-064).
    /// Merges micro-clusters (< min_size nodes) into dominant macro-clusters based on:
    /// 1. Direct cross-boundary link weights (topological affinity)
    /// 2. Common directory / module path affinity (spatial locality)
    /// 3. Dominant parent cluster fallback
    pub fn consolidate_communities(
        nodes: &[DbNode],
        links: &[DbLink],
        assignments: &mut HashMap<String, i32>,
        min_size: usize,
    ) {
        if nodes.is_empty() || assignments.is_empty() || min_size <= 1 {
            return;
        }

        // 1. Group nodes by their current community ID
        let mut groups: HashMap<i32, Vec<&DbNode>> = HashMap::new();
        for node in nodes {
            if let Some(&cid) = assignments.get(&node.id) {
                groups.entry(cid).or_default().push(node);
            }
        }

        // 2. Partition into macro-clusters (>= min_size) and micro-clusters (< min_size)
        type NodeGroup<'a> = HashMap<i32, Vec<&'a DbNode>>;
        let (macro_groups, micro_groups): (NodeGroup, NodeGroup) =
            groups.into_iter().partition(|(_, members)| members.len() >= min_size);

        // If no cluster meets min_size, merge everything into cluster 0
        if macro_groups.is_empty() {
            for val in assignments.values_mut() {
                *val = 0;
            }
            return;
        }

        // If there are no micro-clusters, simply ensure compact 0..K indexing
        if micro_groups.is_empty() {
            return;
        }

        // 3. Precompute directory and file sets for each macro cluster
        let mut macro_dirs: HashMap<i32, HashSet<String>> = HashMap::new();
        let mut macro_node_ids: HashMap<i32, HashSet<&str>> = HashMap::new();
        let mut macro_sizes: Vec<(i32, usize)> = Vec::new();

        for (&cid, members) in &macro_groups {
            let mut dirs = HashSet::new();
            let mut nids = HashSet::new();
            for m in members {
                nids.insert(m.id.as_str());
                if let Some(parent) = std::path::Path::new(&m.file_path).parent() {
                    let p = parent.to_string_lossy().to_string();
                    if !p.is_empty() && p != "." {
                        dirs.insert(p);
                    }
                }
            }
            macro_sizes.push((cid, members.len()));
            macro_dirs.insert(cid, dirs);
            macro_node_ids.insert(cid, nids);
        }
        // 4. Build node-to-macro map for fast topological edge lookup
        let mut node_to_macro: HashMap<&str, i32> = HashMap::new();
        for (&cid, nids) in &macro_node_ids {
            for &nid in nids {
                node_to_macro.insert(nid, cid);
            }
        }

        // 5. For each micro-group, determine target macro cluster
        let mut merge_map: HashMap<i32, i32> = HashMap::new();

        for (&micro_cid, members) in &micro_groups {
            let member_nids: HashSet<&str> = members.iter().map(|m| m.id.as_str()).collect();

            // Criterion 1: Topological affinity (edge connections to any macro cluster)
            let mut edge_affinity: HashMap<i32, f32> = HashMap::new();
            for link in links {
                let s_in = member_nids.contains(link.source.as_str());
                let t_in = member_nids.contains(link.target.as_str());
                if s_in && !t_in {
                    if let Some(&target_macro) = node_to_macro.get(link.target.as_str()) {
                        *edge_affinity.entry(target_macro).or_insert(0.0) += get_edge_weight(&link.edge_type);
                    }
                } else if !s_in && t_in {
                    if let Some(&source_macro) = node_to_macro.get(link.source.as_str()) {
                        *edge_affinity.entry(source_macro).or_insert(0.0) += get_edge_weight(&link.edge_type);
                    }
                }
            }

            if let Some((&best_macro, _)) = edge_affinity
                .iter()
                .max_by(|a, b| a.1.partial_cmp(b.1).unwrap_or(std::cmp::Ordering::Equal))
            {
                merge_map.insert(micro_cid, best_macro);
                continue;
            }

            // Criterion 2: Directory / Spatial locality
            let mut dir_match_counts: HashMap<i32, usize> = HashMap::new();
            for m in members.as_slice() {
                if let Some(parent) = std::path::Path::new(&m.file_path).parent() {
                    let p = parent.to_string_lossy().to_string();
                    if !p.is_empty() && p != "." {
                        for (&mcid, mdirs) in &macro_dirs {
                            for md in mdirs {
                                if p == *md || p.starts_with(md) || md.starts_with(&p) {
                                    *dir_match_counts.entry(mcid).or_insert(0) += 1;
                                }
                            }
                        }
                    }
                }
            }

            if let Some((best_macro, _)) = dir_match_counts.into_iter().max_by_key(|(_, count)| *count) {
                merge_map.insert(micro_cid, best_macro);
                continue;
            }

            // Criterion 3: Group isolated unconnected micro-orphans into a dedicated utilities domain (R-090)
            // NEVER dump them into largest_macro_cid!
            merge_map.insert(micro_cid, 999_999);
        }

        // 6. Apply merges to assignments
        for val in assignments.values_mut() {
            if let Some(&target_cid) = merge_map.get(val) {
                *val = target_cid;
            }
        }

        // 7. Compact community IDs into contiguous 0..K space
        let mut id_map: HashMap<i32, i32> = HashMap::new();
        let mut sorted_cids: Vec<i32> = assignments.values().copied().collect::<HashSet<_>>().into_iter().collect();
        sorted_cids.sort();

        for (next_id, cid) in sorted_cids.into_iter().enumerate() {
            id_map.insert(cid, next_id as i32);
        }

        for val in assignments.values_mut() {
            if let Some(&compact) = id_map.get(val) {
                *val = compact;
            }
        }
    }

    /// Calculate dynamic scale-adaptive macro-galaxy ceiling based on graph complexity (R-088)
    pub fn calculate_adaptive_target(node_count: usize, user_limit: Option<usize>) -> usize {
        if let Some(limit) = user_limit {
            if limit > 0 {
                return limit;
            }
        }
        if node_count == 0 {
            return 5;
        }
        // Sub-linear scale scaling: K = clamp(round(3.5 * N^0.35), 5, 150)
        let dynamic = (3.5 * (node_count as f64).powf(0.35)).round() as usize;
        dynamic.clamp(5, 150)
    }

    /// Hierarchical Leiden macro-coarsening (R-084, R-086, R-087, R-088)
    /// Contracts fine-grained communities into cohesive macro subsystems using:
    /// 1. Modularity-guarded affinity (k_ij / m - (k_i * k_j) / (2 * m^2))
    /// 2. Hard affinity floor: NEVER merges decoupled domains with zero edge/directory coupling (R-086)
    /// 3. Modularity delta guard: Halts if delta_Q < -0.05 (R-087)
    /// 4. Scale-adaptive ceiling: Dynamically targets natural scale (R-088)
    pub fn coarsen_to_macro_galaxies(
        nodes: &[DbNode],
        links: &[DbLink],
        assignments: &mut HashMap<String, i32>,
        max_target: usize,
    ) {
        if nodes.is_empty() || assignments.is_empty() {
            return;
        }

        let effective_target = if max_target == 0 {
            Self::calculate_adaptive_target(nodes.len(), None)
        } else {
            max_target
        };

        // 1. Determine unique communities
        let mut unique_cids: HashSet<i32> = assignments.values().copied().collect();
        if unique_cids.len() <= effective_target {
            return;
        }

        // Build node-to-community fast lookup
        let mut node_comm: HashMap<&str, i32> = HashMap::with_capacity(nodes.len());
        for node in nodes {
            if let Some(&c) = assignments.get(&node.id) {
                node_comm.insert(node.id.as_str(), c);
            }
        }

        // Build community dominant directory
        let mut comm_dirs: HashMap<i32, String> = HashMap::new();
        let mut comm_members: HashMap<i32, Vec<&DbNode>> = HashMap::new();
        for node in nodes {
            if let Some(&c) = assignments.get(&node.id) {
                comm_members.entry(c).or_default().push(node);
            }
        }
        for (&c, members) in &comm_members {
            let mut dir_counts: HashMap<String, usize> = HashMap::new();
            for m in members {
                if let Some(p) = std::path::Path::new(&m.file_path).parent() {
                    let ps = p.to_string_lossy().to_string();
                    if !ps.is_empty() && ps != "." {
                        *dir_counts.entry(ps).or_insert(0) += 1;
                    }
                }
            }
            let dom = dir_counts.into_iter().max_by_key(|(_, cnt)| *cnt).map(|(d, _)| d).unwrap_or_default();
            comm_dirs.insert(c, dom);
        }

        // 2. Compute total edge weight m and community degrees for Newman modularity differential
        let mut total_m = 0.0f32;
        let mut comm_degree: HashMap<i32, f32> = HashMap::new();
        for link in links {
            let w = get_edge_weight(&link.edge_type);
            total_m += w;
            if let Some(&sc) = node_comm.get(link.source.as_str()) {
                *comm_degree.entry(sc).or_insert(0.0) += w;
            }
            if let Some(&tc) = node_comm.get(link.target.as_str()) {
                *comm_degree.entry(tc).or_insert(0.0) += w;
            }
        }
        let total_m2 = (2.0 * total_m).max(1.0);
        let total_m_sq = (2.0 * total_m * total_m).max(1.0);

        // Pre-compute initial inter-community edge affinity matrix once (R-095: O(|E|) initialization)
        let mut meta_affinity: HashMap<(i32, i32), f32> = HashMap::new();
        for link in links {
            let sc = node_comm.get(link.source.as_str());
            let tc = node_comm.get(link.target.as_str());
            if let (Some(&s), Some(&t)) = (sc, tc) {
                if s != t {
                    let key = if s < t { (s, t) } else { (t, s) };
                    *meta_affinity.entry(key).or_insert(0.0) += get_edge_weight(&link.edge_type);
                }
            }
        }

        // 3. Iteratively merge smallest community with its highest affinity neighbor
        while unique_cids.len() > effective_target {
            // Find sizes of remaining communities
            let mut sizes: Vec<(i32, usize)> = unique_cids.iter().map(|&c| {
                let sz = comm_members.get(&c).map(|m| m.len()).unwrap_or(0);
                (c, sz)
            }).collect();
            // Sort ascending by size to merge smallest community first
            sizes.sort_by_key(|(_, sz)| *sz);

            let candidate_c = sizes[0].0;
            let deg_cand = comm_degree.get(&candidate_c).copied().unwrap_or(0.0);

            // Find best merge target for candidate_c:
            // First check topological edge affinity maximizing modularity gain delta_Q
            let mut best_target: Option<i32> = None;
            let mut best_delta_q = -1000.0f32;
            let mut max_weight = -1.0f32;

            for &other_c in &unique_cids {
                if other_c == candidate_c {
                    continue;
                }
                let key = if candidate_c < other_c { (candidate_c, other_c) } else { (other_c, candidate_c) };
                if let Some(&w) = meta_affinity.get(&key) {
                    let deg_other = comm_degree.get(&other_c).copied().unwrap_or(0.0);
                    let delta_q = (w / total_m2) - (deg_cand * deg_other / total_m_sq);
                    if delta_q > best_delta_q {
                        best_delta_q = delta_q;
                        best_target = Some(other_c);
                        max_weight = w;
                    }
                }
            }

            // If no edge affinity, check directory affinity
            if best_target.is_none() {
                if let Some(cand_dir) = comm_dirs.get(&candidate_c) {
                    if !cand_dir.is_empty() {
                        for &other_c in &unique_cids {
                            if other_c == candidate_c {
                                continue;
                            }
                            if let Some(other_dir) = comm_dirs.get(&other_c) {
                                if !other_dir.is_empty() && (cand_dir.starts_with(other_dir) || other_dir.starts_with(cand_dir)) {
                                    best_target = Some(other_c);
                                    break;
                                }
                            }
                        }
                    }
                }
            }

            // R-086 & R-087: AFFINITY FLOOR & MODULARITY GUARD
            // Invariant: If candidate has zero topological coupling AND zero directory ancestry,
            // OR if merging would cause severe modularity degradation (delta_Q < -0.05),
            // HALT COARSENING! DO NOT MERGE UNRELATED DOMAINS!
            let target_c = match best_target {
                Some(tgt) => {
                    if max_weight > 0.0 && best_delta_q < -0.05 {
                        // Merging causes severe modularity penalty; preserve domain independence
                        break;
                    }
                    tgt
                }
                None => {
                    // Zero edge affinity and zero directory lineage.
                    // Preserve as independent architectural domain!
                    break;
                }
            };

            if target_c == candidate_c {
                break; // Cannot merge further
            }

            // Execute merge candidate_c -> target_c
            for node in comm_members.remove(&candidate_c).unwrap_or_default() {
                node_comm.insert(node.id.as_str(), target_c);
                if let Some(val) = assignments.get_mut(&node.id) {
                    *val = target_c;
                }
                comm_members.entry(target_c).or_default().push(node);
            }
            *comm_degree.entry(target_c).or_insert(0.0) += deg_cand;
            unique_cids.remove(&candidate_c);

            // Incrementally update inter-community affinities (R-095: O(K) per merge, not O(|E|))
            let self_key = if candidate_c < target_c { (candidate_c, target_c) } else { (target_c, candidate_c) };
            meta_affinity.remove(&self_key);

            for &other_c in &unique_cids {
                if other_c == target_c {
                    continue;
                }
                let old_key = if candidate_c < other_c { (candidate_c, other_c) } else { (other_c, candidate_c) };
                if let Some(w) = meta_affinity.remove(&old_key) {
                    let new_key = if target_c < other_c { (target_c, other_c) } else { (other_c, target_c) };
                    *meta_affinity.entry(new_key).or_insert(0.0) += w;
                }
            }
        }

        // 4. Compact final community IDs to contiguous 0..K space
        let mut id_map: HashMap<i32, i32> = HashMap::new();
        let mut sorted_cids: Vec<i32> = assignments.values().copied().collect::<HashSet<_>>().into_iter().collect();
        sorted_cids.sort();

        for (next_id, cid) in sorted_cids.into_iter().enumerate() {
            id_map.insert(cid, next_id as i32);
        }

        for val in assignments.values_mut() {
            if let Some(&compact) = id_map.get(val) {
                *val = compact;
            }
        }
    }

    /// Compute full architectural metrics (Ca, Ce, Instability, roles) and generate galaxy records
    pub fn compute_galaxy_metrics(
        workspace: &str,
        nodes: &[DbNode],
        links: &[DbLink],
        assignments: &HashMap<String, i32>,
    ) -> (Vec<CommunitySummary>, Vec<GalaxyRecord>) {
        let mut groups: HashMap<i32, Vec<&DbNode>> = HashMap::new();
        for node in nodes {
            if let Some(&cid) = assignments.get(&node.id) {
                groups.entry(cid).or_default().push(node);
            }
        }

        // Count internal and cross-boundary edges per community
        let mut internal_edges_map: HashMap<i32, usize> = HashMap::new();
        let mut afferent_coupling_map: HashMap<i32, usize> = HashMap::new(); // Ca: foreign -> this
        let mut efferent_coupling_map: HashMap<i32, usize> = HashMap::new(); // Ce: this -> foreign

        for link in links {
            let src_comm = assignments.get(&link.source);
            let tgt_comm = assignments.get(&link.target);

            if let (Some(&src_c), Some(&tgt_c)) = (src_comm, tgt_comm) {
                if src_c == tgt_c {
                    *internal_edges_map.entry(src_c).or_insert(0) += 1;
                } else {
                    *efferent_coupling_map.entry(src_c).or_insert(0) += 1;
                    *afferent_coupling_map.entry(tgt_c).or_insert(0) += 1;
                }
            }
        }

        let mut summaries = Vec::new();
        let mut galaxy_records = Vec::new();

        for (cid, members) in groups {
            let mut files_set = HashSet::new();
            let mut top_symbols = Vec::new();
            let mut languages = HashSet::new();
            let mut dir_counts: HashMap<String, usize> = HashMap::new();

            // Semantic Kind-Weighted Centroid Election & Label Sanitization (R-111, SPEC-OMNI-006)
            // Filter out import nodes, files, directories, manifests, and multi-line statements.
            // Weight candidate symbols: class/struct/interface (100) > function/fn (75) > method (50) > const/enum (20).
            let is_invalid_label = |label: &str| -> bool {
                let trimmed = label.trim();
                trimmed.is_empty()
                    || trimmed.contains('\n')
                    || trimmed.starts_with("from ")
                    || trimmed.starts_with("import ")
                    || trimmed.starts_with('.')
                    || trimmed.starts_with("use ")
            };

            let get_kind_weight = |kind: &str| -> usize {
                match kind.to_lowercase().as_str() {
                    "class" | "struct" | "interface" | "trait" | "type" | "classdef" => 100,
                    "function" | "fn" => 75,
                    "method" | "constructor" => 50,
                    "constant" | "enum" | "macro" | "property" => 20,
                    _ => 0,
                }
            };

            let mut scored_symbols: Vec<(&DbNode, usize)> = Vec::new();

            for m in &members {
                files_set.insert(m.file_path.clone());
                if !m.language.is_empty() {
                    languages.insert(m.language.clone());
                }

                if let Some(parent) = std::path::Path::new(&m.file_path).parent() {
                    let p = parent.to_string_lossy().to_string();
                    if !p.is_empty() && p != "." {
                        *dir_counts.entry(p).or_insert(0) += 1;
                    }
                }

                if !is_invalid_label(&m.label) {
                    let weight = get_kind_weight(&m.kind);
                    if weight > 0 {
                        scored_symbols.push((m, weight));
                    }
                }
            }

            // Sort by kind weight descending, then label alphabetical for determinism
            scored_symbols.sort_by(|(a_node, a_weight), (b_node, b_weight)| {
                b_weight.cmp(a_weight).then_with(|| a_node.label.cmp(&b_node.label))
            });

            for (m, _) in &scored_symbols {
                if top_symbols.len() < 8 && !top_symbols.contains(&m.label) {
                    top_symbols.push(m.label.clone());
                }
            }

            // If no weighted code symbols qualified, populate fallback top_symbols from sanitized non-import nodes
            if top_symbols.is_empty() {
                for m in &members {
                    if !is_invalid_label(&m.label)
                        && m.kind != "import"
                        && m.kind != "import_from"
                        && top_symbols.len() < 8
                        && !top_symbols.contains(&m.label)
                    {
                        top_symbols.push(m.label.clone());
                    }
                }
            }

            let dominant = dir_counts
                .into_iter()
                .max_by_key(|(_, c)| *c)
                .map(|(d, _)| d)
                .unwrap_or_else(|| {
                    members.first().map(|m| m.file_path.clone()).unwrap_or_else(|| "root".to_string())
                });

            let primary_symbol = top_symbols.first().map(|s| s.as_str()).unwrap_or("");
            let name = if dominant.is_empty() || dominant == "." {
                if !primary_symbol.is_empty() {
                    format!("Galaxy #{}: {}", cid, primary_symbol)
                } else {
                    format!("Galaxy #{}", cid)
                }
            } else if !primary_symbol.is_empty() && primary_symbol != dominant {
                format!("{} ({})", dominant, primary_symbol)
            } else {
                dominant.clone()
            };

            let internal_edges = internal_edges_map.get(&cid).copied().unwrap_or(0);
            let ca = afferent_coupling_map.get(&cid).copied().unwrap_or(0);
            let ce = efferent_coupling_map.get(&cid).copied().unwrap_or(0);
            let total_cut = ca + ce;

            let instability = if total_cut == 0 {
                0.5
            } else {
                (ce as f64) / (total_cut as f64)
            };

            let role = if instability <= 0.25 {
                "Core Foundation".to_string()
            } else if instability <= 0.65 {
                "Domain Service".to_string()
            } else {
                "Orchestrator / Leaf".to_string()
            };

            // Uncle Bob Martin Package Metrics (R-085):
            // Abstractness: A = N_abstract / N_total
            // Distance from Main Sequence: D = |A + I - 1|
            let abstract_count = members.iter().filter(|m| {
                matches!(
                    m.kind.to_lowercase().as_str(),
                    "interface" | "trait" | "protocol" | "abstract_class" | "abstract" | "type_alias"
                )
            }).count();

            let abstractness = if members.is_empty() {
                0.0
            } else {
                (abstract_count as f64) / (members.len() as f64)
            };

            let distance = (abstractness + instability - 1.0).abs();

            let langs_vec: Vec<String> = languages.into_iter().collect();

            summaries.push(CommunitySummary {
                id: cid,
                name: format!("Galaxy #{}: {}", cid, name),
                node_count: members.len(),
                top_symbols: top_symbols.clone(),
                files: files_set.into_iter().collect(),
            });

            galaxy_records.push(GalaxyRecord {
                workspace: workspace.to_string(),
                galaxy_id: cid,
                name: name.clone(),
                dominant_path: dominant,
                node_count: members.len(),
                internal_edges,
                external_edges: total_cut,
                afferent_coupling: ca,
                efferent_coupling: ce,
                instability,
                abstractness,
                distance,
                role,
                key_symbols: top_symbols,
                languages: langs_vec,
                updated_at: None,
            });
        }

        summaries.sort_by_key(|s| std::cmp::Reverse(s.node_count));
        galaxy_records.sort_by_key(|g| std::cmp::Reverse(g.node_count));

        (summaries, galaxy_records)
    }

    /// Infer community for a newly created AST node based on modal connection to existing graph
    #[allow(dead_code)]
    pub fn infer_neighbor_community(
        node_id: &str,
        all_links: &[DbLink],
        node_communities: &HashMap<String, i32>,
    ) -> Option<i32> {
        let mut votes: HashMap<i32, f32> = HashMap::new();
        for l in all_links {
            if l.source == node_id {
                if let Some(&c) = node_communities.get(&l.target) {
                    *votes.entry(c).or_insert(0.0) += get_edge_weight(&l.edge_type);
                }
            } else if l.target == node_id {
                if let Some(&c) = node_communities.get(&l.source) {
                    *votes.entry(c).or_insert(0.0) += get_edge_weight(&l.edge_type);
                }
            }
        }

        votes
            .into_iter()
            .max_by(|a, b| a.1.partial_cmp(&b.1).unwrap_or(std::cmp::Ordering::Equal))
            .map(|(c, _)| c)
    }

    /// Calculate closed-form local modularity delta for moving node u from community C_A to C_B (R-095)
    /// Delta Q = [ k_{u, C_B}/(2m) - (Sigma_{tot}(C_B) * k_u)/(2m^2) ] - [ k_{u, C_A}/(2m) - (Sigma_{tot}(C_A) * k_u)/(2m^2) ]
    #[allow(dead_code)]
    pub fn calculate_local_delta_q(
        k_u_target: f32,
        k_u_source: f32,
        sigma_tot_target: f32,
        sigma_tot_source: f32,
        k_u: f32,
        total_m: f32,
    ) -> f32 {
        let total_m2 = (2.0 * total_m).max(1.0);
        let total_m_sq = (2.0 * total_m * total_m).max(1.0);

        let target_gain = (k_u_target / total_m2) - (sigma_tot_target * k_u / total_m_sq);
        let source_loss = (k_u_source / total_m2) - (sigma_tot_source * k_u / total_m_sq);

        target_gain - source_loss
    }

    /// Summarize detected communities (macroscopic view)
    pub fn summarize(nodes: &[DbNode], assignments: &HashMap<String, i32>) -> Vec<CommunitySummary> {
        Self::summarize_filtered(nodes, assignments, 1)
    }

    /// Summarize detected communities with minimum node count filter to eliminate singleton noise (Finding #6)
    pub fn summarize_filtered(
        nodes: &[DbNode],
        assignments: &HashMap<String, i32>,
        min_size: usize,
    ) -> Vec<CommunitySummary> {
        let mut groups: HashMap<i32, Vec<&DbNode>> = HashMap::new();
        for node in nodes {
            if let Some(&cid) = assignments.get(&node.id) {
                groups.entry(cid).or_default().push(node);
            }
        }

        let mut summaries = Vec::new();
        for (cid, members) in groups {
            if members.len() < min_size {
                continue;
            }
            let mut files_set = HashSet::new();
            let mut top_symbols = Vec::new();

            for m in &members {
                files_set.insert(m.file_path.clone());
                if top_symbols.len() < 5 {
                    top_symbols.push(m.label.clone());
                }
            }

            // Derive representative name from most common directory or top symbol
            let sample_file = members.first().map(|m| m.file_path.as_str()).unwrap_or("unknown");
            let dir_name = sample_file
                .rsplit_once('/')
                .map(|(dir, _)| dir)
                .unwrap_or(sample_file);

            let comm_name = format!("Cluster #{}: {}", cid, dir_name);

            summaries.push(CommunitySummary {
                id: cid,
                name: comm_name,
                node_count: members.len(),
                top_symbols,
                files: files_set.into_iter().collect(),
            });
        }

        summaries.sort_by_key(|s| std::cmp::Reverse(s.node_count));
        summaries
    }
}

pub struct GraphRagEngine;

impl GraphRagEngine {
    /// Hybrid Graph-RAG Retrieval: Combines Micro Vector Search + Macro Community Summaries + AST Traces
    pub async fn query(
        db: &DbClient,
        embedder: &EmbedderClient,
        prompt: &str,
        top_k: usize,
        workspace: Option<&str>,
        include_docs: bool,
    ) -> Result<GraphRagResponse, String> {
        // 1. Microscopic Vector Search with Content-Tier Partitioning (R-067)
        let query_vec = embedder.embed_single(prompt).await?;
        let vector_hits = db.search_vector(&query_vec, top_k, workspace, include_docs, false, None).await?;

        // 2. Collect Seed Symbols and IDs
        let mut seed_symbols = Vec::new();
        let mut seed_node_ids = HashSet::new();
        let mut seed_node_ids_vec = Vec::new();
        let mut touched_communities = HashSet::new();

        for hit in &vector_hits {
            seed_symbols.push(hit.label.clone());
            seed_node_ids.insert(hit.id.clone());
            seed_node_ids_vec.push(hit.id.clone());
            if let Some(c) = hit.community {
                touched_communities.insert(c);
            }
        }

        // 3. Fetch Localized Neighborhood Subgraph (Targeted 1-hop BFS, avoids full graph download)
        let (all_nodes, all_links) = db.get_neighborhood_subgraph(&seed_node_ids_vec, workspace).await?;

        // 4. One-hop AST neighbor expansion
        let mut expanded_node_ids = seed_node_ids.clone();
        for link in &all_links {
            if seed_node_ids.contains(&link.source) {
                expanded_node_ids.insert(link.target.clone());
            } else if seed_node_ids.contains(&link.target) {
                expanded_node_ids.insert(link.source.clone());
            }
        }

        // 5. Build Subgraph Markdown with strict budgeting (< 3,800 chars for definitions, >= 1,800 chars reserved for edges)
        let mut sub_md = String::new();
        sub_md.push_str("#### Microscopic AST Slice (Seeds & Direct Callers/Callees)\n");

        // Sort nodes so seed hits appear first with snippet, followed by 1-line neighbor signatures
        let mut matched_nodes: Vec<&DbNode> = all_nodes.iter().filter(|n| expanded_node_ids.contains(&n.id)).collect();
        matched_nodes.sort_by_key(|n| if seed_node_ids.contains(&n.id) { 0 } else { 1 });

        let mut defs_chars = 0;
        const MAX_DEFS_CHARS: usize = 3800;

        for n in matched_nodes {
            if defs_chars >= MAX_DEFS_CHARS {
                sub_md.push_str("- *(+ additional AST symbols truncated to preserve context budget)*\n");
                break;
            }

            let entry = if seed_node_ids.contains(&n.id) {
                let snippet = if n.text.len() > 300 {
                    format!("{}...", n.text[..300].trim_end())
                } else {
                    n.text.trim().to_string()
                };
                format!(
                    "- **`{}`** ({}) in `{}:{}`\n  ```{}\n  {}\n  ```\n",
                    n.label, n.kind, n.file_path, n.line_start, n.language, snippet
                )
            } else {
                let sig = n.text.lines().find(|l| !l.trim().is_empty()).unwrap_or(&n.text).trim();
                format!(
                    "- `{}` ({}) in `{}:{}`: `{}`\n",
                    n.label, n.kind, n.file_path, n.line_start, sig
                )
            };

            defs_chars += entry.len();
            sub_md.push_str(&entry);
        }

        sub_md.push_str("\n#### Direct Structural Relations\n");
        for (link_count, link) in all_links
            .iter()
            .filter(|l| expanded_node_ids.contains(&l.source) && expanded_node_ids.contains(&l.target))
            .enumerate()
        {
            if link_count >= 25 {
                sub_md.push_str("- *(additional relations truncated)*\n");
                break;
            }
            let src_clean = crate::db::clean_record_id(&link.source);
            let tgt_clean = crate::db::clean_record_id(&link.target);
            sub_md.push_str(&format!(
                "- `{}` ──[{}: {}]──▶ `{}`\n",
                src_clean, link.category, link.edge_type, tgt_clean
            ));
        }

        // 6. Community summaries for touched clusters from pre-computed galaxy table (R-036)
        let all_galaxy_records = db.get_galaxies_records(workspace).await.unwrap_or_default();
        let relevant_summaries: Vec<CommunitySummary> = if !all_galaxy_records.is_empty() {
            all_galaxy_records
                .into_iter()
                .filter(|g| touched_communities.contains(&g.galaxy_id))
                .map(|g| CommunitySummary {
                    id: g.galaxy_id,
                    node_count: g.node_count,
                    name: g.name,
                    top_symbols: g.key_symbols,
                    files: vec![g.dominant_path],
                })
                .collect()
        } else {
            let mut comm_assignments = HashMap::new();
            for n in &all_nodes {
                if let Some(c) = n.community {
                    comm_assignments.insert(n.id.clone(), c);
                }
            }
            let all_summaries = CommunityDetector::summarize(&all_nodes, &comm_assignments);
            all_summaries
                .into_iter()
                .filter(|s| touched_communities.contains(&s.id))
                .collect()
        };

        let mut macro_summary = String::new();
        macro_summary.push_str(&format!(
            "Query matched {} high-confidence AST seeds across {} architectural community clusters.\n",
            vector_hits.len(),
            relevant_summaries.len()
        ));
        for comm in &relevant_summaries {
            macro_summary.push_str(&format!(
                "- **{}** ({} symbols): Top keys: {}\n",
                comm.name,
                comm.node_count,
                comm.top_symbols.join(", ")
            ));
        }

        let total_chars = sub_md.len() + macro_summary.len();
        let token_estimate = total_chars / 4;

        Ok(GraphRagResponse {
            query: prompt.to_string(),
            macroscopic_summary: macro_summary,
            communities_involved: relevant_summaries,
            seed_symbols,
            expanded_subgraph: sub_md,
            token_estimate,
        })
    }
}
