use std::collections::HashMap;
use std::io::Cursor;
use std::path::Path;
use quick_xml::events::Event;
use quick_xml::reader::Reader;
use zip::ZipArchive;

use crate::parser::{ExtractedEdge, ExtractedNode, ParseResult};

/// Parser for MathWorks Simulink models (.slx and .mdl)
pub struct SimulinkParser;

impl SimulinkParser {
    /// Ingests a Simulink .slx archive (Open Packaging Conventions ZIP containing simulink/blockdiagram.xml)
    pub fn parse_slx(workspace: &str, file_path: &str, raw_bytes: &[u8]) -> Option<ParseResult> {
        let cursor = Cursor::new(raw_bytes);
        let mut zip = match ZipArchive::new(cursor) {
            Ok(z) => z,
            Err(_) => return None,
        };

        let mut xml_content = String::new();
        for i in 0..zip.len() {
            if let Ok(mut file) = zip.by_index(i) {
                if let Ok(name) = file.name() {
                    if &*name == "simulink/blockdiagram.xml"
                        || name.ends_with("/blockdiagram.xml")
                        || &*name == "blockdiagram.xml"
                    {
                        use std::io::Read;
                        let _ = file.read_to_string(&mut xml_content);
                        break;
                    }
                }
            }
        }

        if xml_content.is_empty() {
            return None;
        }

        Self::parse_blockdiagram_xml(workspace, file_path, &xml_content)
    }

    /// Parses Simulink blockdiagram.xml XML stream
    pub fn parse_blockdiagram_xml(workspace: &str, file_path: &str, xml_content: &str) -> Option<ParseResult> {
        let mut reader = Reader::from_str(xml_content);
        reader.config_mut().trim_text(true);

        let model_name = Path::new(file_path)
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("Model");

        let model_id = format!("{}:{}:model:{}", workspace, file_path, model_name);

        let mut nodes = Vec::new();
        let mut edges = Vec::new();

        // Emit root model node
        nodes.push(ExtractedNode {
            id: model_id.clone(),
            workspace: workspace.to_string(),
            label: model_name.to_string(),
            kind: "model".to_string(),
            file_path: file_path.to_string(),
            language: "simulink".to_string(),
            line_start: 1,
            line_end: 1,
            text: format!("Simulink Model: {}", model_name),
        });

        // SID map: sid -> (node_id, label)
        let mut sid_map: HashMap<String, (String, String)> = HashMap::new();

        // System hierarchy stack: current parent for child blocks
        let mut parent_stack: Vec<String> = vec![model_id.clone()];

        // Current block parsing context
        struct CurrentBlock {
            block_type: String,
            _sid: String,
            is_subsystem: bool,
            node_id: String,
        }
        let mut current_block: Option<CurrentBlock> = None;

        // Line parsing context
        struct CurrentLine {
            src: Option<String>,
            dsts: Vec<String>,
        }
        let mut current_line: Option<CurrentLine> = None;

        // Current <P Name="..."> context for property text
        let mut current_p_name: Option<String> = None;

        let mut buf = Vec::new();

        loop {
            match reader.read_event_into(&mut buf) {
                Ok(Event::Start(ref e)) => {
                    let tag = e.name().into_inner();
                    match tag {
                        "Block" => {
                            let mut block_type = String::new();
                            let mut name = String::new();
                            let mut sid = String::new();

                            for attr in e.attributes().flatten() {
                                let key = attr.key.into_inner();
                                let val = attr.value.to_string();
                                if key == "BlockType" {
                                    block_type = val;
                                } else if key == "Name" {
                                    name = val;
                                } else if key == "SID" {
                                    sid = val;
                                }
                            }

                            if !sid.is_empty() {
                                let is_subsystem = block_type == "SubSystem";
                                let block_id = format!("{}:{}:block:{}", workspace, file_path, sid);
                                let parent_id = parent_stack.last().cloned().unwrap_or_else(|| model_id.clone());

                                let node_kind = match block_type.as_str() {
                                    "SubSystem" => "subsystem",
                                    "ModelReference" => "model_reference",
                                    "Inport" => "inport",
                                    "Outport" => "outport",
                                    _ => "block",
                                };

                                let display_name = if name.is_empty() { format!("Block_{}", sid) } else { name.clone() };

                                nodes.push(ExtractedNode {
                                    id: block_id.clone(),
                                    workspace: workspace.to_string(),
                                    label: display_name.clone(),
                                    kind: node_kind.to_string(),
                                    file_path: file_path.to_string(),
                                    language: "simulink".to_string(),
                                    line_start: 1,
                                    line_end: 1,
                                    text: format!("<Block BlockType=\"{}\" Name=\"{}\" SID=\"{}\">", block_type, display_name, sid),
                                });

                                edges.push(ExtractedEdge {
                                    workspace: workspace.to_string(),
                                    source_id: parent_id,
                                    target_label: display_name.clone(),
                                    target_id: Some(block_id.clone()),
                                    edge_type: "CONTAINS_BLOCK".to_string(),
                                    category: "MBSE_CONTAINMENT".to_string(),
                                });

                                sid_map.insert(sid.clone(), (block_id.clone(), display_name));

                                current_block = Some(CurrentBlock {
                                    block_type,
                                    _sid: sid,
                                    is_subsystem,
                                    node_id: block_id,
                                });
                            }
                        }
                        "System" => {
                            if let Some(ref cb) = current_block {
                                if cb.is_subsystem {
                                    parent_stack.push(cb.node_id.clone());
                                }
                            }
                        }
                        "Line" => {
                            current_line = Some(CurrentLine {
                                src: None,
                                dsts: Vec::new(),
                            });
                        }
                        "P" => {
                            for attr in e.attributes().flatten() {
                                if attr.key.into_inner() == "Name" {
                                    current_p_name = Some(attr.value.to_string());
                                }
                            }
                        }
                        _ => {}
                    }
                }
                Ok(Event::Empty(ref e)) => {
                    let tag = e.name().into_inner();
                    if tag == "Block" {
                        let mut block_type = String::new();
                        let mut name = String::new();
                        let mut sid = String::new();
                        let mut model_name = String::new();

                        for attr in e.attributes().flatten() {
                            let key = attr.key.into_inner();
                            let val = attr.value.to_string();
                            if key == "BlockType" {
                                block_type = val;
                            } else if key == "Name" {
                                name = val;
                            } else if key == "SID" {
                                sid = val;
                            } else if key == "ModelName" || key == "ModelFile" {
                                model_name = val;
                            }
                        }

                        if !sid.is_empty() {
                            let block_id = format!("{}:{}:block:{}", workspace, file_path, sid);
                            let parent_id = parent_stack.last().cloned().unwrap_or_else(|| model_id.clone());

                            let node_kind = match block_type.as_str() {
                                "SubSystem" => "subsystem",
                                "ModelReference" => "model_reference",
                                "Inport" => "inport",
                                "Outport" => "outport",
                                _ => "block",
                            };

                            let display_name = if name.is_empty() { format!("Block_{}", sid) } else { name.clone() };

                            nodes.push(ExtractedNode {
                                id: block_id.clone(),
                                workspace: workspace.to_string(),
                                label: display_name.clone(),
                                kind: node_kind.to_string(),
                                file_path: file_path.to_string(),
                                language: "simulink".to_string(),
                                line_start: 1,
                                line_end: 1,
                                text: format!("<Block BlockType=\"{}\" Name=\"{}\" SID=\"{}\" />", block_type, display_name, sid),
                            });

                            edges.push(ExtractedEdge {
                                workspace: workspace.to_string(),
                                source_id: parent_id,
                                target_label: display_name.clone(),
                                target_id: Some(block_id.clone()),
                                edge_type: "CONTAINS_BLOCK".to_string(),
                                category: "MBSE_CONTAINMENT".to_string(),
                            });

                            if block_type == "ModelReference" && !model_name.is_empty() {
                                let clean_target = model_name
                                    .trim_end_matches(".slx")
                                    .trim_end_matches(".mdl")
                                    .trim();
                                if !clean_target.is_empty() {
                                    edges.push(ExtractedEdge {
                                        workspace: workspace.to_string(),
                                        source_id: block_id.clone(),
                                        target_label: clean_target.to_string(),
                                        target_id: None,
                                        edge_type: "REFERENCES_MODEL".to_string(),
                                        category: "MBSE_REFERENCE".to_string(),
                                    });
                                }
                            }

                            sid_map.insert(sid.clone(), (block_id, display_name));
                        }
                    }
                }
                Ok(Event::Text(ref e)) => {
                    if let Some(ref p_name) = current_p_name {
                        let text = e.trim().to_string();
                        if !text.is_empty() {
                            if let Some(ref mut line) = current_line {
                                if p_name == "Src" {
                                    line.src = Some(text.clone());
                                } else if p_name == "Dst" {
                                    line.dsts.push(text.clone());
                                }
                            }

                            if let Some(ref cb) = current_block {
                                if cb.block_type == "ModelReference" && (p_name == "ModelName" || p_name == "ModelFile") {
                                    let clean_target = text
                                        .trim_end_matches(".slx")
                                        .trim_end_matches(".mdl")
                                        .trim();
                                    if !clean_target.is_empty() {
                                        edges.push(ExtractedEdge {
                                            workspace: workspace.to_string(),
                                            source_id: cb.node_id.clone(),
                                            target_label: clean_target.to_string(),
                                            target_id: None,
                                            edge_type: "REFERENCES_MODEL".to_string(),
                                            category: "MBSE_REFERENCE".to_string(),
                                        });
                                    }
                                }
                            }
                        }
                    }
                }
                Ok(Event::End(ref e)) => {
                    let tag = e.name().into_inner();
                    match tag {
                        "Block" => {
                            current_block = None;
                        }
                        "System" => {
                            if parent_stack.len() > 1 {
                                parent_stack.pop();
                            }
                        }
                        "Line" => {
                            if let Some(line) = current_line.take() {
                                if let Some(src_str) = line.src {
                                    let src_sid = src_str.split('#').next().unwrap_or(&src_str).trim();
                                    let category = if line.dsts.len() > 50 {
                                        "DATAFLOW_BROADCAST"
                                    } else {
                                        "DATAFLOW"
                                    };

                                    if let Some((src_id, _)) = sid_map.get(src_sid) {
                                        for dst_str in line.dsts {
                                            let dst_sid = dst_str.split('#').next().unwrap_or(&dst_str).trim();
                                            if let Some((dst_id, dst_label)) = sid_map.get(dst_sid) {
                                                edges.push(ExtractedEdge {
                                                    workspace: workspace.to_string(),
                                                    source_id: src_id.clone(),
                                                    target_label: dst_label.clone(),
                                                    target_id: Some(dst_id.clone()),
                                                    edge_type: "SIGNAL_TO".to_string(),
                                                    category: category.to_string(),
                                                });
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        "P" => {
                            current_p_name = None;
                        }
                        _ => {}
                    }
                }
                Ok(Event::Eof) => break,
                Err(_) => break,
                _ => {}
            }
            buf.clear();
        }

        Some(ParseResult { nodes, edges })
    }

    /// Parses legacy plain-text Simulink .mdl files
    pub fn parse_mdl(workspace: &str, file_path: &str, content: &str) -> Option<ParseResult> {
        let model_name = Path::new(file_path)
            .file_stem()
            .and_then(|s| s.to_str())
            .unwrap_or("Model");

        let model_id = format!("{}:{}:model:{}", workspace, file_path, model_name);

        let mut nodes = Vec::new();
        let mut edges = Vec::new();

        nodes.push(ExtractedNode {
            id: model_id.clone(),
            workspace: workspace.to_string(),
            label: model_name.to_string(),
            kind: "model".to_string(),
            file_path: file_path.to_string(),
            language: "simulink".to_string(),
            line_start: 1,
            line_end: 1,
            text: format!("Simulink Model: {}", model_name),
        });

        let mut sid_map: HashMap<String, (String, String)> = HashMap::new();
        let mut parent_stack: Vec<String> = vec![model_id.clone()];

        let mut in_block = false;
        let mut in_line = false;
        let mut block_type = String::new();
        let mut block_name = String::new();
        let mut block_sid = String::new();
        let mut model_ref_name = String::new();
        let mut block_line_start = 1;

        let mut line_src = String::new();
        let mut line_dsts = Vec::new();

        for (idx, line) in content.lines().enumerate() {
            let line_num = idx + 1;
            let trimmed = line.trim();

            if trimmed.starts_with("Block {") {
                in_block = true;
                block_type.clear();
                block_name.clear();
                block_sid.clear();
                model_ref_name.clear();
                block_line_start = line_num;
            } else if in_block && trimmed.starts_with("BlockType") {
                block_type = Self::extract_mdl_value(trimmed);
            } else if in_block && trimmed.starts_with("Name") {
                block_name = Self::extract_mdl_value(trimmed);
            } else if in_block && trimmed.starts_with("SID") {
                block_sid = Self::extract_mdl_value(trimmed);
            } else if in_block && (trimmed.starts_with("ModelName") || trimmed.starts_with("ModelFile")) {
                model_ref_name = Self::extract_mdl_value(trimmed);
            } else if in_block && trimmed == "System {" {
                if !block_sid.is_empty() {
                    let block_id = format!("{}:{}:block:{}", workspace, file_path, block_sid);
                    parent_stack.push(block_id);
                }
            } else if in_block && trimmed == "}" {
                if !block_sid.is_empty() {
                    let block_id = format!("{}:{}:block:{}", workspace, file_path, block_sid);
                    let parent_id = parent_stack.last().cloned().unwrap_or_else(|| model_id.clone());

                    let node_kind = match block_type.as_str() {
                        "SubSystem" => "subsystem",
                        "ModelReference" => "model_reference",
                        "Inport" => "inport",
                        "Outport" => "outport",
                        _ => "block",
                    };

                    let display_name = if block_name.is_empty() {
                        format!("Block_{}", block_sid)
                    } else {
                        block_name.clone()
                    };

                    nodes.push(ExtractedNode {
                        id: block_id.clone(),
                        workspace: workspace.to_string(),
                        label: display_name.clone(),
                        kind: node_kind.to_string(),
                        file_path: file_path.to_string(),
                        language: "simulink".to_string(),
                        line_start: block_line_start,
                        line_end: line_num,
                        text: format!("Block {{ BlockType: {}, Name: {}, SID: {} }}", block_type, display_name, block_sid),
                    });

                    edges.push(ExtractedEdge {
                        workspace: workspace.to_string(),
                        source_id: parent_id,
                        target_label: display_name.clone(),
                        target_id: Some(block_id.clone()),
                        edge_type: "CONTAINS_BLOCK".to_string(),
                        category: "MBSE_CONTAINMENT".to_string(),
                    });

                    if block_type == "ModelReference" && !model_ref_name.is_empty() {
                        let clean_target = model_ref_name
                            .trim_end_matches(".slx")
                            .trim_end_matches(".mdl")
                            .trim();
                        edges.push(ExtractedEdge {
                            workspace: workspace.to_string(),
                            source_id: block_id.clone(),
                            target_label: clean_target.to_string(),
                            target_id: None,
                            edge_type: "REFERENCES_MODEL".to_string(),
                            category: "MBSE_REFERENCE".to_string(),
                        });
                    }

                    sid_map.insert(block_sid.clone(), (block_id, display_name));
                }
                in_block = false;
            } else if trimmed.starts_with("Line {") {
                in_line = true;
                line_src.clear();
                line_dsts.clear();
            } else if in_line && trimmed.starts_with("Src") {
                line_src = Self::extract_mdl_value(trimmed);
            } else if in_line && trimmed.starts_with("Dst") {
                line_dsts.push(Self::extract_mdl_value(trimmed));
            } else if in_line && trimmed == "}" {
                in_line = false;
                if !line_src.is_empty() {
                    let src_sid = line_src.split('#').next().unwrap_or(&line_src).trim();
                    let category = if line_dsts.len() > 50 {
                        "DATAFLOW_BROADCAST"
                    } else {
                        "DATAFLOW"
                    };

                    if let Some((src_id, _)) = sid_map.get(src_sid) {
                        for dst in &line_dsts {
                            let dst_sid = dst.split('#').next().unwrap_or(dst).trim();
                            if let Some((dst_id, dst_label)) = sid_map.get(dst_sid) {
                                edges.push(ExtractedEdge {
                                    workspace: workspace.to_string(),
                                    source_id: src_id.clone(),
                                    target_label: dst_label.clone(),
                                    target_id: Some(dst_id.clone()),
                                    edge_type: "SIGNAL_TO".to_string(),
                                    category: category.to_string(),
                                });
                            }
                        }
                    }
                }
            } else if trimmed == "}" && parent_stack.len() > 1 {
                parent_stack.pop();
            }
        }

        Some(ParseResult { nodes, edges })
    }

    fn extract_mdl_value(line: &str) -> String {
        let mut parts = line.split_whitespace();
        let _key = parts.next();
        let rest = parts.collect::<Vec<&str>>().join(" ");
        rest.trim_matches(|c: char| c == '"' || c == '\'').to_string()
    }
}
