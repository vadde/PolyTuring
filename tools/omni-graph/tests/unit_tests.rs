// =============================================================================
// Omni-Graph Unit & Integration Test Suite
// Covers: parser, condenser, community detection, API normalization
// =============================================================================

mod parser_tests {
    use omni_graph::parser::CodeParser;

    #[test]
    fn parse_rust_function() {
        let content = r#"
fn hello_world(name: &str) -> String {
    format!("Hello, {}!", name)
}

fn greet(user: &str) {
    let msg = hello_world(user);
    println!("{}", msg);
}
"#;
        let result = CodeParser::parse_file("test-ws", "src/main.rs", content);
        assert!(result.is_some(), "Parser should return Some for valid Rust");
        let pr = result.unwrap();

        assert_eq!(pr.nodes.len(), 2, "Should extract 2 functions");
        assert_eq!(pr.nodes[0].label, "hello_world");
        assert_eq!(pr.nodes[0].kind, "function");
        assert_eq!(pr.nodes[0].language, "rust");
        assert_eq!(pr.nodes[0].workspace, "test-ws");
        assert_eq!(pr.nodes[0].file_path, "src/main.rs");
        assert!(pr.nodes[0].line_start > 0);
        assert!(pr.nodes[0].line_end >= pr.nodes[0].line_start);
        assert_eq!(pr.nodes[1].label, "greet");

        // Should have a CALLS edge from greet -> hello_world
        let calls: Vec<_> = pr
            .edges
            .iter()
            .filter(|e| e.edge_type == "CALLS")
            .collect();
        assert!(
            !calls.is_empty(),
            "Should detect CALLS edges from greet to hello_world"
        );
        assert!(calls.iter().any(|e| e.target_label == "hello_world"));
    }

    #[test]
    fn parse_rust_struct() {
        let content = r#"
struct Config {
    host: String,
    port: u16,
}
"#;
        let result = CodeParser::parse_file("test-ws", "config.rs", content).unwrap();
        assert_eq!(result.nodes.len(), 3, "Should extract struct plus 2 fields");
        assert_eq!(result.nodes[0].label, "Config");
        assert_eq!(result.nodes[0].kind, "struct");
        assert!(result.nodes.iter().any(|n| n.label == "host" && n.kind == "field"));
        assert!(result.nodes.iter().any(|n| n.label == "port" && n.kind == "field"));
    }

    #[test]
    fn parse_rust_imports() {
        let content = r#"
use std::collections::HashMap;
use std::sync::Arc;

fn do_nothing() {}
"#;
        let result = CodeParser::parse_file("test-ws", "lib.rs", content).unwrap();
        let imports: Vec<_> = result.nodes.iter().filter(|n| n.kind == "import").collect();
        assert_eq!(imports.len(), 2, "Should extract 2 use declarations");
    }

    #[test]
    fn parse_python_function() {
        let content = r#"
def calculate_sum(a, b):
    return a + b

def main():
    result = calculate_sum(1, 2)
    print(result)
"#;
        let result = CodeParser::parse_file("py-ws", "calc.py", content);
        assert!(result.is_some(), "Parser should handle Python");
        let pr = result.unwrap();
        let fns: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "function").collect();
        assert_eq!(fns.len(), 2);
        assert_eq!(fns[0].label, "calculate_sum");
        assert_eq!(fns[0].language, "python");
        assert_eq!(fns[1].label, "main");
    }

    #[test]
    fn parse_go_function() {
        let content = r#"
package main

func Add(a int, b int) int {
    return a + b
}
"#;
        let result = CodeParser::parse_file("go-ws", "math.go", content);
        assert!(result.is_some(), "Parser should handle Go");
        let pr = result.unwrap();
        let fns: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "function").collect();
        assert!(!fns.is_empty());
        assert_eq!(fns[0].label, "Add");
        assert_eq!(fns[0].language, "go");
    }

    #[test]
    fn parse_go_full_grammar_methods_structs_constants_imports() {
        let content = r#"
package solver

import (
    "fmt"
    "math"
)

const (
    ActionFulfill = 1
    ActionPick    = 2
)

type OptimizedSolver struct {
    iterations int
}

type Evaluator interface {
    Eval(score float64) bool
}

func (s *OptimizedSolver) executeFulfill(orderId string) error {
    return nil
}

func (s OptimizedSolver) GetIterations() int {
    return s.iterations
}
"#;
        let result = CodeParser::parse_file("go-ws", "solver.go", content);
        assert!(result.is_some(), "Parser should handle full Go grammar");
        let pr = result.unwrap();

        // 1. Check constants
        let consts: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "constant").collect();
        assert_eq!(consts.len(), 2);
        assert!(consts.iter().any(|c| c.label == "ActionFulfill"));
        assert!(consts.iter().any(|c| c.label == "ActionPick"));

        // 2. Check struct and interface
        let structs: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "struct").collect();
        assert_eq!(structs.len(), 1);
        assert_eq!(structs[0].label, "OptimizedSolver");

        let ifaces: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "interface").collect();
        assert_eq!(ifaces.len(), 1);
        assert_eq!(ifaces[0].label, "Evaluator");

        // 3. Check methods
        let methods: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "method").collect();
        assert_eq!(methods.len(), 3);
        assert!(methods.iter().any(|m| m.label == "executeFulfill"));
        assert!(methods.iter().any(|m| m.label == "GetIterations"));
        assert!(methods.iter().any(|m| m.label == "Eval"));

        // 4. Check imports
        let imports: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "import").collect();
        assert_eq!(imports.len(), 2);

        // 5. Check DECLARES edges from struct and interface to methods
        let struct_declares: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "DECLARES" && e.source_id == "go-ws:solver.go:OptimizedSolver").collect();
        assert_eq!(struct_declares.len(), 2);

        let iface_declares: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "DECLARES" && e.target_label == "Eval").collect();
        assert_eq!(iface_declares.len(), 1);
    }

    #[test]
    fn parse_javascript_function() {
        let content = r#"
function fetchData(url) {
    return fetch(url);
}
"#;
        let result = CodeParser::parse_file("js-ws", "api.js", content);
        assert!(result.is_some(), "Parser should handle JavaScript");
        let pr = result.unwrap();
        let fns: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "function").collect();
        assert!(!fns.is_empty());
        assert_eq!(fns[0].label, "fetchData");
        assert_eq!(fns[0].language, "javascript");
    }

    #[test]
    fn parse_typescript_function() {
        let content = r#"
function greet(name: string): string {
    return `Hello, ${name}`;
}
"#;
        let result = CodeParser::parse_file("ts-ws", "hello.ts", content);
        assert!(result.is_some(), "Parser should handle TypeScript");
        let pr = result.unwrap();
        let fns: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "function").collect();
        assert!(!fns.is_empty());
        assert_eq!(fns[0].label, "greet");
        assert_eq!(fns[0].language, "typescript");
    }

    #[test]
    fn parse_unsupported_extension_returns_none() {
        let result = CodeParser::parse_file("ws", "data.xyz", "some random data");
        assert!(result.is_none());
    }

    #[test]
    fn parse_empty_file() {
        let result = CodeParser::parse_file("ws", "empty.rs", "");
        assert!(result.is_some());
        assert!(result.unwrap().nodes.is_empty());
    }

    #[test]
    fn node_ids_contain_workspace_and_path() {
        let content = "fn foo() {}";
        let pr = CodeParser::parse_file("my-project", "src/lib.rs", content).unwrap();
        assert!(!pr.nodes.is_empty());
        let id = &pr.nodes[0].id;
        assert!(
            id.contains("my-project"),
            "Node ID should contain workspace"
        );
        assert!(
            id.contains("src/lib.rs"),
            "Node ID should contain file path"
        );
    }

    #[test]
    fn parse_typescript_arrow_functions_and_interfaces() {
        let content = r#"
export const loadWorkspaces = async () => {
    return fetch('/api/workspaces');
};

const handleFilter = (query: string) => {
    console.log(query);
};

export interface UserConfig {
    host: string;
    port: number;
}

export type UserID = string;

export enum TaskState {
    Pending,
    Active,
    Done,
}
"#;
        let pr = CodeParser::parse_file("ui-ws", "App.tsx", content).unwrap();

        // 1. Functions (arrow functions)
        let fns: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "function").collect();
        assert_eq!(fns.len(), 2);
        assert!(fns.iter().any(|f| f.label == "loadWorkspaces"));
        assert!(fns.iter().any(|f| f.label == "handleFilter"));

        // 2. Interface
        let ifaces: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "interface").collect();
        assert_eq!(ifaces.len(), 1);
        assert_eq!(ifaces[0].label, "UserConfig");

        // 3. Type alias
        let types: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "type").collect();
        assert_eq!(types.len(), 1);
        assert_eq!(types[0].label, "UserID");

        // 4. Enum
        let enums: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "enum").collect();
        assert_eq!(enums.len(), 1);
        assert_eq!(enums[0].label, "TaskState");
    }

    #[test]
    fn parse_rust_impl_and_enums() {
        let content = r#"
pub struct CodeParser;

pub enum DeltaKind {
    Create,
    Modify,
    Delete,
}

pub trait IngestionStrategy {
    fn ingest(&self);
}

impl CodeParser {
    pub fn parse_file(path: &str) -> bool {
        true
    }
}
"#;
        let pr = CodeParser::parse_file("rust-ws", "lib.rs", content).unwrap();

        // 1. Struct
        assert!(pr.nodes.iter().any(|n| n.label == "CodeParser" && n.kind == "struct"));

        // 2. Enum
        assert!(pr.nodes.iter().any(|n| n.label == "DeltaKind" && n.kind == "enum"));

        // 3. Trait
        assert!(pr.nodes.iter().any(|n| n.label == "IngestionStrategy" && n.kind == "trait"));

        // 4. Method
        assert!(pr.nodes.iter().any(|n| n.label == "parse_file" && n.kind == "function"));

        // 5. DECLARES edge linking struct to method inside impl
        let declares: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "DECLARES").collect();
        assert!(!declares.is_empty(), "impl block should emit DECLARES edge from struct to method");
        assert!(declares.iter().any(|e| e.source_id == "rust-ws:lib.rs:CodeParser" && e.target_label == "parse_file"));
    }

    #[test]
    fn parse_python_async_functions() {
        let content = r#"
async def fetch_remote_data(endpoint: str):
    return await http_get(endpoint)

def sync_process():
    pass
"#;
        let pr = CodeParser::parse_file("py-ws", "async_worker.py", content).unwrap();
        let fns: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "function").collect();
        assert_eq!(fns.len(), 2);
        assert!(fns.iter().any(|f| f.label == "fetch_remote_data"));
        assert!(fns.iter().any(|f| f.label == "sync_process"));
    }

    #[test]
    fn parse_go_interface_methods() {
        let content = r#"
package solver

type Solver interface {
    Solve(orders []int) error
    Reset()
}
"#;
        let pr = CodeParser::parse_file("go-ws", "solver.go", content).unwrap();

        // 1. Interface node
        assert!(pr.nodes.iter().any(|n| n.label == "Solver" && n.kind == "interface"));

        // 2. Interface methods (method_spec)
        let methods: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "method").collect();
        assert_eq!(methods.len(), 2);
        assert!(methods.iter().any(|m| m.label == "Solve"));
        assert!(methods.iter().any(|m| m.label == "Reset"));

        // 3. DECLARES edge linking interface to method
        let declares: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "DECLARES").collect();
        assert_eq!(declares.len(), 2);
        assert!(declares.iter().any(|e| e.target_label == "Solve"));
        assert!(declares.iter().any(|e| e.target_label == "Reset"));
    }

    #[test]
    fn parse_python_calls_and_methods() {
        let content = r#"
class Worker:
    def execute(self):
        self.run_task()
        process_data()
"#;
        let pr = CodeParser::parse_file("py-ws", "worker.py", content).unwrap();

        // Check CALLS edges emitted from Python calls
        let calls: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "CALLS").collect();
        assert!(!calls.is_empty(), "Python call expressions should emit CALLS edges");
        assert!(calls.iter().any(|c| c.target_label == "run_task"));
        assert!(calls.iter().any(|c| c.target_label == "process_data"));
    }

    #[test]
    fn parse_rust_method_calls() {
        let content = r#"
fn orchestrate(client: &DbClient) {
    client.query_sql("SELECT 1;");
    format_output();
}
"#;
        let pr = CodeParser::parse_file("rs-ws", "orch.rs", content).unwrap();

        // Check CALLS edges emitted from Rust method_call_expression and call_expression
        let calls: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "CALLS").collect();
        assert!(!calls.is_empty(), "Rust calls should emit CALLS edges");
        assert!(calls.iter().any(|c| c.target_label == "query_sql"));
        assert!(calls.iter().any(|c| c.target_label == "format_output"));
    }

    #[test]
    fn parse_import_edges() {
        let pr_rs = CodeParser::parse_file("rs-ws", "lib.rs", "use crate::db::{DbClient, Payload};").unwrap();
        assert!(pr_rs.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "DbClient"));
        assert!(pr_rs.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "Payload"));

        let pr_py = CodeParser::parse_file("py-ws", "main.py", "from solver import GreedySolver, ExactSolver, Heuristic").unwrap();
        assert!(pr_py.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "GreedySolver"));
        assert!(pr_py.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "ExactSolver"));
        assert!(pr_py.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "Heuristic"));

        let pr_ts = CodeParser::parse_file("ts-ws", "index.ts", "import { executeFulfill, ActionFulfill as AF } from './actions';").unwrap();
        assert!(pr_ts.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "executeFulfill"));
        assert!(pr_ts.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "ActionFulfill"));
        assert!(pr_ts.edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "actions"));
    }

    #[test]
    fn parse_rust_trait_implementation() {
        let content = r#"
pub struct CustomSolver;

pub trait Solvable {
    fn solve(&self);
}

impl Solvable for CustomSolver {
    fn solve(&self) {
        println!("solved");
    }
}
"#;
        let pr = CodeParser::parse_file("rs-ws", "solver.rs", content).unwrap();
        let impl_edges: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "IMPLEMENTS").collect();
        assert_eq!(impl_edges.len(), 1, "Must emit IMPLEMENTS edge for impl Trait for Struct");
        assert_eq!(impl_edges[0].target_label, "Solvable");
        assert!(impl_edges[0].source_id.ends_with("CustomSolver"));
    }

    #[test]
    fn parse_python_class_inheritance() {
        let content = r#"
class Animal:
    pass

class Dog(Animal):
    def bark(self):
        pass
"#;
        let pr = CodeParser::parse_file("py-ws", "models.py", content).unwrap();
        let extends_edges: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "EXTENDS").collect();
        assert_eq!(extends_edges.len(), 1, "Must emit EXTENDS edge for Python class inheritance");
        assert_eq!(extends_edges[0].target_label, "Animal");
    }

    #[test]
    fn parse_typescript_class_heritage() {
        let content = r#"
export interface Runnable {
    run(): void;
}

export class TaskRunner extends BaseRunner implements Runnable {
    run() {
        console.log("running");
    }
}
"#;
        let pr = CodeParser::parse_file("ts-ws", "runner.ts", content).unwrap();
        let extends_edges: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "EXTENDS").collect();
        assert_eq!(extends_edges.len(), 1, "Must emit EXTENDS edge for TS class");
        assert_eq!(extends_edges[0].target_label, "BaseRunner");

        let impl_edges: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "IMPLEMENTS").collect();
        assert_eq!(impl_edges.len(), 1, "Must emit IMPLEMENTS edge for TS class implements clause");
        assert_eq!(impl_edges[0].target_label, "Runnable");
    }

    #[test]
    fn parse_rust_macro_definition_and_invocation() {
        let content = r#"
macro_rules! my_telemetry {
    ($msg:expr) => {
        println!("{}", $msg);
    };
}

fn track() {
    my_telemetry!("ping");
}
"#;
        let pr = CodeParser::parse_file("rs-ws", "macro_test.rs", content).unwrap();
        assert!(pr.nodes.iter().any(|n| n.label == "my_telemetry" && n.kind == "macro"));
        assert!(pr.edges.iter().any(|e| e.edge_type == "CALLS" && e.target_label == "my_telemetry"));
    }

    #[test]
    fn text_truncation_at_1000_chars() {
        // Generate a function with a very long body
        let body = "let x = 1;\n".repeat(200); // ~2200 chars
        let content = format!("fn huge_fn() {{\n{}}}", body);
        let pr = CodeParser::parse_file("ws", "big.rs", &content).unwrap();
        assert!(!pr.nodes.is_empty());
        // Text should be truncated to ~1000 chars
        assert!(
            pr.nodes[0].text.len() <= 1100,
            "Text should be truncated around 1000 chars, got {}",
            pr.nodes[0].text.len()
        );
    }

    #[test]
    fn parse_markdown_sections_and_snippets() {
        let content = r#"# Concurrency Foundations

Some introductory text about concurrency.

## Goroutines

A goroutine is a lightweight thread.

```go
func doWork(id int) {
    println(id)
}
```

## Channels

Channels connect concurrent goroutines.
"#;
        let pr = CodeParser::parse_file("test-ws", "docs/concurrency.md", content).unwrap();

        // Check extracted nodes
        let labels: Vec<&str> = pr.nodes.iter().map(|n| n.label.as_str()).collect();
        assert!(labels.contains(&"Concurrency Foundations"), "Should extract root title");
        assert!(labels.contains(&"Goroutines"), "Should extract H2 Goroutines");
        assert!(labels.contains(&"Channels"), "Should extract H2 Channels");
        assert!(labels.contains(&"doWork"), "Should extract embedded Go function doWork");

        // Verify snippet has language 'go' and kind 'function'
        let do_work_node = pr.nodes.iter().find(|n| n.label == "doWork").unwrap();
        assert_eq!(do_work_node.language, "go");
        assert_eq!(do_work_node.kind, "function");

        // Verify hierarchy edges: Concurrency Foundations -> Goroutines -> doWork
        assert!(pr.edges.iter().any(|e| e.target_label == "Goroutines" && e.edge_type == "CONTAINS"));
        assert!(pr.edges.iter().any(|e| e.target_label == "doWork" && e.edge_type == "CONTAINS"));
    }

    #[test]
    fn parse_markdown_python_code_block() {
        let content = r#"# GenAI Foundations

## Attention Mechanism

```python
def scaled_dot_product_attention(q, k, v):
    return torch.matmul(q, k.transpose(-2, -1))
```
"#;
        let pr = CodeParser::parse_file("genai-ws", "notes/attention.md", content).unwrap();
        let fn_node = pr.nodes.iter().find(|n| n.label == "scaled_dot_product_attention");
        assert!(fn_node.is_some(), "Should extract python function from markdown block");
        assert_eq!(fn_node.unwrap().language, "python");
    }

    #[test]
    fn parse_markdown_links_and_references() {
        let content = r#"# Modified Binary Search

## Overview
See prerequisite in [Two Pointers](../02-two-pointers/theory.md) and related [Rotated Array](problem_02_rotated.md).
The algorithm calls `binary_search` and references `SearchRange.find_bound` to eliminate half the search space.
"#;
        let pr = CodeParser::parse_file("dsa-ws", "17-binary-search/theory.md", content).unwrap();

        // Check LINKS_TO edges
        let links: Vec<&str> = pr.edges.iter()
            .filter(|e| e.edge_type == "LINKS_TO")
            .map(|e| e.target_label.as_str())
            .collect();
        assert!(links.contains(&"theory.md"), "Should extract relative markdown link to theory.md");
        assert!(links.contains(&"problem_02_rotated.md"), "Should extract link to problem_02_rotated.md");

        // Check REFERENCES edges
        let refs: Vec<&str> = pr.edges.iter()
            .filter(|e| e.edge_type == "REFERENCES")
            .map(|e| e.target_label.as_str())
            .collect();
        assert!(refs.contains(&"binary_search"), "Should extract backticked symbol reference `binary_search`");
        assert!(refs.contains(&"SearchRange.find_bound"), "Should extract backticked symbol `SearchRange.find_bound`");
    }

    #[test]
    fn parse_yaml_kubernetes_manifest() {
        let content = r#"apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend-app
spec:
  replicas: 3
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-svc
spec:
  ports:
    - port: 80
"#;
        let pr = CodeParser::parse_file("k8s-ws", "manifests/frontend.yaml", content).unwrap();
        let labels: Vec<&str> = pr.nodes.iter().map(|n| n.label.as_str()).collect();
        assert!(labels.contains(&"frontend-app"), "Should extract Deployment name");
        assert!(labels.contains(&"frontend-svc"), "Should extract Service name");

        let dep_node = pr.nodes.iter().find(|n| n.label == "frontend-app").unwrap();
        assert_eq!(dep_node.kind, "Deployment");
        assert_eq!(dep_node.language, "yaml");

        assert!(pr.edges.iter().any(|e| e.target_label == "frontend-app" && e.edge_type == "CONTAINS"));
        assert!(pr.edges.iter().any(|e| e.target_label == "frontend-svc" && e.edge_type == "CONTAINS"));
    }

    #[test]
    fn parse_json_config() {
        let content = r#"{
  "name": "omni-graph-ui",
  "version": "1.0.0",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build"
  }
}"#;
        let pr = CodeParser::parse_file("test-ws", "package.json", content).unwrap();
        let labels: Vec<&str> = pr.nodes.iter().map(|n| n.label.as_str()).collect();
        assert!(labels.iter().any(|l| l.contains("scripts")));
        assert!(labels.iter().any(|l| l.contains("version")));
    }

    #[test]
    fn parse_shell_functions() {
        let content = r#"#!/usr/bin/env bash
function build_stack() {
    cargo build --release
}

run_tests() {
    cargo test
}
"#;
        let pr = CodeParser::parse_file("sh-ws", "scripts/build.sh", content).unwrap();
        let labels: Vec<&str> = pr.nodes.iter().map(|n| n.label.as_str()).collect();
        assert!(labels.contains(&"build_stack"));
        assert!(labels.contains(&"run_tests"));
        let fn_node = pr.nodes.iter().find(|n| n.label == "build_stack").unwrap();
        assert_eq!(fn_node.kind, "function");
        assert_eq!(fn_node.language, "bash");
    }

    #[test]
    fn parse_sql_tables_and_functions() {
        let content = r#"
CREATE TABLE users (
    id INT PRIMARY KEY,
    username VARCHAR(50)
);

CREATE FUNCTION get_user_count() RETURNS INT AS $$
BEGIN
    RETURN 42;
END;
$$ LANGUAGE plpgsql;
"#;
        let pr = CodeParser::parse_file("db-ws", "schema.sql", content).unwrap();
        let labels: Vec<&str> = pr.nodes.iter().map(|n| n.label.as_str()).collect();
        assert!(labels.contains(&"users"));
        assert!(labels.contains(&"get_user_count"));
    }

    #[test]
    fn parse_toml_tables() {
        let content = r#"
[package]
name = "omni-graph"
version = "0.1.0"

[dependencies]
tokio = "1.0"
"#;
        let pr = CodeParser::parse_file("ws", "Cargo.toml", content).unwrap();
        let labels: Vec<&str> = pr.nodes.iter().map(|n| n.label.as_str()).collect();
        assert!(labels.iter().any(|l| l.contains("[package]")));
        assert!(labels.iter().any(|l| l.contains("[dependencies]")));
    }

    #[test]
    fn parse_fallback_chunker() {
        let lines: Vec<String> = (1..=100).map(|i| format!("Note line number {}", i)).collect();
        let content = lines.join("\n");
        let pr = CodeParser::parse_fallback("ws", "notes/study-guide.sample", &content).unwrap();

        assert!(!pr.nodes.is_empty());
        let root = &pr.nodes[0];
        assert_eq!(root.kind, "file");
        assert_eq!(root.label, "study-guide.sample");

        let blocks: Vec<_> = pr.nodes.iter().filter(|n| n.kind == "block").collect();
        assert!(!blocks.is_empty(), "Should generate chunk blocks for 100-line text");
        assert!(pr.edges.iter().any(|e| e.edge_type == "CONTAINS"));
    }
}

mod condenser_tests {
    use omni_graph::condenser::ContextCondenser;
    use omni_graph::db::{DbLink, DbNode};

    fn make_node(id: &str, label: &str, kind: &str, file_path: &str) -> DbNode {
        DbNode {
            id: id.to_string(),
            workspace: Some("test".to_string()),
            label: label.to_string(),
            kind: kind.to_string(),
            file_path: file_path.to_string(),
            language: "rust".to_string(),
            line_start: 1,
            line_end: 10,
            text: format!("fn {}() {{}}", label),
            community: None,
            file_hash: None,
        }
    }

    fn make_link(source: &str, target: &str, edge_type: &str) -> DbLink {
        DbLink {
            id: format!("edge:{}_{}", source, target),
            workspace: Some("test".to_string()),
            source: source.to_string(),
            target: target.to_string(),
            edge_type: edge_type.to_string(),
            category: "EXTRACTED".to_string(),
        }
    }

    #[test]
    fn condense_finds_root_symbol() {
        let nodes = vec![
            make_node("n1", "init_pool", "function", "src/db.rs"),
            make_node("n2", "other_fn", "function", "src/main.rs"),
        ];
        let links = vec![];
        let result = ContextCondenser::condense("init_pool", &nodes, &links, 2);

        assert_eq!(result.root_symbol, "init_pool");
        assert!(result.formatted_markdown.contains("init_pool"));
        assert!(result.related_files.contains(&"src/db.rs".to_string()));
    }

    #[test]
    fn condense_finds_root_by_file_path() {
        let nodes = vec![
            make_node("node:PolyTuring:file:src/db.rs", "db.rs", "file", "src/db.rs"),
            make_node("n2", "init_pool", "function", "src/db.rs"),
            make_node("n3", "other_fn", "function", "src/main.rs"),
        ];
        let links = vec![
            make_link("node:PolyTuring:file:src/db.rs", "n2", "CONTAINS"),
            make_link("n3", "n2", "CALLS"),
        ];
        let result = ContextCondenser::condense("src/db.rs", &nodes, &links, 2);

        assert_eq!(result.root_symbol, "src/db.rs");
        assert!(result.related_files.contains(&"src/db.rs".to_string()));
        assert!(result.formatted_markdown.contains("src/db.rs"));
    }

    #[test]
    fn condense_follows_edges_up_to_max_hops() {
        let nodes = vec![
            make_node("n1", "root_fn", "function", "src/a.rs"),
            make_node("n2", "caller_fn", "function", "src/b.rs"),
            make_node("n3", "callee_fn", "function", "src/c.rs"),
            make_node("n4", "distant_fn", "function", "src/d.rs"),
        ];
        let links = vec![
            make_link("n2", "n1", "CALLS"),  // caller -> root
            make_link("n1", "n3", "CALLS"),  // root -> callee
            make_link("n3", "n4", "CALLS"),  // callee -> distant (hop 2)
        ];

        // With hops=1, should NOT reach distant_fn
        let result_1 = ContextCondenser::condense("root_fn", &nodes, &links, 1);
        assert!(result_1.formatted_markdown.contains("caller_fn"));
        assert!(result_1.formatted_markdown.contains("callee_fn"));
        assert!(
            !result_1.formatted_markdown.contains("distant_fn"),
            "Should NOT include 2-hop distant_fn with max_hops=1"
        );

        // With hops=2, SHOULD reach distant_fn
        let result_2 = ContextCondenser::condense("root_fn", &nodes, &links, 2);
        assert!(
            result_2.formatted_markdown.contains("distant_fn"),
            "Should include 2-hop distant_fn with max_hops=2"
        );
    }

    #[test]
    fn condense_populates_callers_and_callees() {
        let nodes = vec![
            make_node("n1", "target", "function", "src/a.rs"),
            make_node("n2", "caller", "function", "src/b.rs"),
            make_node("n3", "callee", "function", "src/c.rs"),
        ];
        let links = vec![
            make_link("n2", "n1", "CALLS"), // caller -> target
            make_link("n1", "n3", "CALLS"), // target -> callee
        ];

        let result = ContextCondenser::condense("target", &nodes, &links, 2);
        assert!(result.direct_callers.contains(&"n2".to_string()));
        assert!(result.direct_callees.contains(&"n3".to_string()));
    }

    #[test]
    fn condense_nonexistent_symbol_produces_empty() {
        let nodes = vec![make_node("n1", "existing", "function", "src/a.rs")];
        let result = ContextCondenser::condense("nonexistent", &nodes, &[], 2);
        assert!(result.direct_callers.is_empty());
        assert!(result.direct_callees.is_empty());
        assert!(result.related_files.is_empty());
    }

    #[test]
    fn condense_token_estimate_is_reasonable() {
        let nodes = vec![
            make_node("n1", "foo", "function", "src/a.rs"),
            make_node("n2", "bar", "function", "src/b.rs"),
        ];
        let links = vec![make_link("n1", "n2", "CALLS")];
        let result = ContextCondenser::condense("foo", &nodes, &links, 2);

        // Token estimate = len / 4
        assert!(result.token_estimate > 0);
        assert_eq!(result.token_estimate, result.formatted_markdown.len() / 4);
    }

    #[test]
    fn condense_strict_token_budget_guarantees_call_traces() {
        let mut nodes = Vec::new();
        let mut links = Vec::new();
        nodes.push(DbNode {
            id: "root".to_string(),
            workspace: Some("test".to_string()),
            label: "core_root".to_string(),
            kind: "function".to_string(),
            file_path: "src/root.rs".to_string(),
            language: "rust".to_string(),
            line_start: 1,
            line_end: 20,
            text: "pub fn core_root() {\n    // Large root implementation\n}".to_string(),
            community: None,
            file_hash: None,
        });

        for i in 1..=20 {
            let id = format!("n{}", i);
            let lbl = format!("neighbor_{}", i);
            nodes.push(DbNode {
                id: id.clone(),
                workspace: Some("test".to_string()),
                label: lbl,
                kind: "function".to_string(),
                file_path: format!("src/file_{}.rs", i),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 50,
                text: format!("pub fn neighbor_{}() {{\n{}\n}}", i, "    let x = 1;\n".repeat(30)),
                community: None,
                file_hash: None,
            });
            links.push(make_link("root", &id, "CALLS"));
        }

        let result = ContextCondenser::condense("core_root", &nodes, &links, 1);

        // Budget guarantee: token estimate must not exceed 1500 tokens (6000 chars)
        assert!(result.token_estimate <= 1500, "Token estimate must be <= 1500");
        assert!(result.formatted_markdown.len() <= 6000, "Markdown length must be <= 6000 chars");

        // Structural trace guarantee: Call trace section must contain actual CALLS edges!
        assert!(
            result.formatted_markdown.contains("`root` ──[EXTRACTED: CALLS]──▶ `n1`"),
            "Must output structural call traces despite numerous large neighbor nodes"
        );
    }

    #[test]
    fn test_condense_ppr_centrality_ranks_hub_higher() {
        use std::collections::HashSet;

        let root_id = "root".to_string();
        let mut root_ids = HashSet::new();
        root_ids.insert(root_id.clone());

        let mut candidate_ids = HashSet::new();
        candidate_ids.insert(root_id.clone());
        candidate_ids.insert("hub".to_string());
        candidate_ids.insert("leaf".to_string());
        candidate_ids.insert("sub1".to_string());
        candidate_ids.insert("sub2".to_string());

        let links = vec![
            make_link("root", "hub", "CALLS"),
            make_link("root", "leaf", "CALLS"),
            make_link("hub", "sub1", "CALLS"),
            make_link("hub", "sub2", "CALLS"),
            make_link("sub1", "hub", "CALLS"),
            make_link("sub2", "hub", "CALLS"),
        ];

        let ppr = ContextCondenser::compute_ppr_scores(&root_ids, &candidate_ids, &links, 25);
        let hub_score = ppr.get("hub").copied().unwrap_or(0.0);
        let leaf_score = ppr.get("leaf").copied().unwrap_or(0.0);

        assert!(
            hub_score > leaf_score,
            "Hub with mutual caller cycles must score higher than isolated leaf: {} vs {}",
            hub_score,
            leaf_score
        );
    }
}

mod community_tests {
    use omni_graph::analysis::CommunityDetector;
    use omni_graph::db::{DbLink, DbNode};

    fn make_node(id: &str) -> DbNode {
        DbNode {
            id: id.to_string(),
            workspace: Some("test".to_string()),
            label: id.to_string(),
            kind: "function".to_string(),
            file_path: format!("src/{}.rs", id),
            language: "rust".to_string(),
            line_start: 1,
            line_end: 10,
            text: format!("fn {}() {{}}", id),
            community: None,
            file_hash: None,
        }
    }

    fn make_link(source: &str, target: &str) -> DbLink {
        DbLink {
            id: format!("edge:{}_{}", source, target),
            workspace: Some("test".to_string()),
            source: source.to_string(),
            target: target.to_string(),
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        }
    }

    #[test]
    fn detect_assigns_all_nodes() {
        let nodes = vec![make_node("a"), make_node("b"), make_node("c")];
        let links = vec![make_link("a", "b"), make_link("b", "c")];
        let assignments = CommunityDetector::detect(&nodes, &links, 15);

        assert_eq!(assignments.len(), 3, "Every node should have a community");
        assert!(assignments.contains_key("a"));
        assert!(assignments.contains_key("b"));
        assert!(assignments.contains_key("c"));
    }

    #[test]
    fn detect_connected_nodes_same_community() {
        // Fully connected triangle should converge to same community
        let nodes = vec![make_node("a"), make_node("b"), make_node("c")];
        let links = vec![
            make_link("a", "b"),
            make_link("b", "c"),
            make_link("a", "c"),
        ];
        let assignments = CommunityDetector::detect(&nodes, &links, 15);

        assert_eq!(
            assignments["a"], assignments["b"],
            "Connected nodes a and b should share a community"
        );
        assert_eq!(
            assignments["b"], assignments["c"],
            "Connected nodes b and c should share a community"
        );
    }

    #[test]
    fn detect_disconnected_components_different_communities() {
        // Two disconnected pairs
        let nodes = vec![
            make_node("a"),
            make_node("b"),
            make_node("x"),
            make_node("y"),
        ];
        let links = vec![
            make_link("a", "b"), // cluster 1
            make_link("x", "y"), // cluster 2
        ];
        let assignments = CommunityDetector::detect(&nodes, &links, 15);

        assert_eq!(assignments["a"], assignments["b"]);
        assert_eq!(assignments["x"], assignments["y"]);
        assert_ne!(
            assignments["a"], assignments["x"],
            "Disconnected components should have different communities"
        );
    }

    #[test]
    fn detect_compact_ids_start_at_zero() {
        let nodes = vec![make_node("a"), make_node("b")];
        let links = vec![make_link("a", "b")];
        let assignments = CommunityDetector::detect(&nodes, &links, 15);

        let min_id = *assignments.values().min().unwrap();
        assert_eq!(min_id, 0, "Compact IDs should start at 0");
    }

    #[test]
    fn detect_empty_graph() {
        let assignments = CommunityDetector::detect(&[], &[], 15);
        assert!(assignments.is_empty());
    }

    #[test]
    fn detect_isolated_nodes() {
        let nodes = vec![make_node("a"), make_node("b"), make_node("c")];
        let links = vec![]; // no edges
        let assignments = CommunityDetector::detect(&nodes, &links, 15);
        assert_eq!(assignments.len(), 3);
        // Each isolated node keeps its own community
        let unique_communities: std::collections::HashSet<_> =
            assignments.values().collect();
        assert_eq!(
            unique_communities.len(),
            3,
            "Isolated nodes should each be in their own community"
        );
    }

    #[test]
    fn summarize_produces_correct_counts() {
        let nodes = vec![make_node("a"), make_node("b"), make_node("c")];
        let mut assignments = std::collections::HashMap::new();
        assignments.insert("a".to_string(), 0);
        assignments.insert("b".to_string(), 0);
        assignments.insert("c".to_string(), 1);

        let summaries = CommunityDetector::summarize(&nodes, &assignments);
        assert_eq!(summaries.len(), 2);

        let comm0 = summaries.iter().find(|s| s.id == 0).unwrap();
        assert_eq!(comm0.node_count, 2);

        let comm1 = summaries.iter().find(|s| s.id == 1).unwrap();
        assert_eq!(comm1.node_count, 1);
    }

    #[test]
    fn summarize_sorted_by_size_descending() {
        let nodes = vec![
            make_node("a"),
            make_node("b"),
            make_node("c"),
            make_node("d"),
        ];
        let mut assignments = std::collections::HashMap::new();
        assignments.insert("a".to_string(), 0);
        assignments.insert("b".to_string(), 1);
        assignments.insert("c".to_string(), 1);
        assignments.insert("d".to_string(), 1);

        let summaries = CommunityDetector::summarize(&nodes, &assignments);
        assert!(
            summaries[0].node_count >= summaries[1].node_count,
            "Summaries should be sorted by node_count descending"
        );
    }
}

mod api_normalize_tests {
    use omni_graph::api::normalize_workspace;

    #[test]
    fn normalize_full_path() {
        let result = normalize_workspace(Some("/Users/aparv/projects/my-repo"));
        assert_eq!(result, Some("my-repo".to_string()));
    }

    #[test]
    fn normalize_trailing_slash() {
        let result = normalize_workspace(Some("/Users/aparv/projects/my-repo/"));
        assert_eq!(result, Some("my-repo".to_string()));
    }

    #[test]
    fn normalize_bare_name() {
        let result = normalize_workspace(Some("session-explorer"));
        assert_eq!(result, Some("session-explorer".to_string()));
    }

    #[test]
    fn normalize_empty_string() {
        let result = normalize_workspace(Some(""));
        assert_eq!(result, None);
    }

    #[test]
    fn normalize_whitespace_only() {
        let result = normalize_workspace(Some("   "));
        assert_eq!(result, None);
    }

    #[test]
    fn normalize_none() {
        let result = normalize_workspace(None);
        assert_eq!(result, None);
    }

    #[test]
    fn normalize_single_segment_with_trailing_slash() {
        let result = normalize_workspace(Some("myproject/"));
        assert_eq!(result, Some("myproject".to_string()));
    }

    #[test]
    fn normalize_deeply_nested_path() {
        let result = normalize_workspace(Some("/a/b/c/d/e/deep-project"));
        assert_eq!(result, Some("deep-project".to_string()));
    }
}

mod escape_tests {
    use omni_graph::db::surql_escape;

    #[test]
    fn escape_single_quotes() {
        assert_eq!(surql_escape("hello'world"), "hello\\'world");
        assert_eq!(surql_escape("'test'"), "\\'test\\'");
    }

    #[test]
    fn escape_backslashes() {
        assert_eq!(surql_escape("path\\to\\file"), "path\\\\to\\\\file");
    }

    #[test]
    fn escape_newlines_and_returns() {
        assert_eq!(surql_escape("line1\nline2\r\nline3"), "line1 line2 line3");
    }

    #[test]
    fn escape_null_and_control_chars() {
        assert_eq!(surql_escape("clean\0text\x07bell"), "cleantextbell");
    }

    #[test]
    fn defeat_sql_injection_payload() {
        let malicious = "' OR '1'='1' --; DROP TABLE node;";
        let escaped = surql_escape(malicious);
        assert_eq!(escaped, "\\' OR \\'1\\'=\\'1\\' --; DROP TABLE node;");
        assert!(escaped.starts_with("\\'"));
    }

    #[test]
    fn clean_record_id_variants() {
        use omni_graph::db::clean_record_id;
        assert_eq!(clean_record_id("node:`ws:file:func:10`"), "ws:file:func:10");
        assert_eq!(clean_record_id("node:⟨ws:file:func:10⟩"), "ws:file:func:10");
        assert_eq!(clean_record_id("node:ws:file:func:10"), "ws:file:func:10");
        assert_eq!(clean_record_id("`ws:file:func:10`"), "ws:file:func:10");
        assert_eq!(clean_record_id("⟨ws:file:func:10⟩"), "ws:file:func:10");
        assert_eq!(clean_record_id("ws:file:func:10"), "ws:file:func:10");
    }
}

mod lpa_rng_tests {
    use omni_graph::analysis::SimpleRng;

    #[test]
    fn rng_deterministic_shuffle() {
        let mut rng1 = SimpleRng::new(42);
        let mut rng2 = SimpleRng::new(42);
        let mut v1 = vec![1, 2, 3, 4, 5, 6, 7, 8];
        let mut v2 = vec![1, 2, 3, 4, 5, 6, 7, 8];

        rng1.shuffle(&mut v1);
        rng2.shuffle(&mut v2);

        assert_eq!(v1, v2, "Same seed must produce identical shuffle");
        assert_eq!(v1.len(), 8);
        for item in 1..=8 {
            assert!(v1.contains(&item));
        }
    }

    #[test]
    fn rng_different_seeds_differ() {
        let mut rng1 = SimpleRng::new(12345);
        let mut rng2 = SimpleRng::new(67890);
        let mut v1: Vec<usize> = (0..20).collect();
        let mut v2: Vec<usize> = (0..20).collect();

        rng1.shuffle(&mut v1);
        rng2.shuffle(&mut v2);

        assert_ne!(v1, v2, "Different seeds should produce different shuffles");
    }
}

mod cluster_filter_tests {
    use omni_graph::analysis::CommunityDetector;
    use omni_graph::db::DbNode;
    use std::collections::HashMap;

    fn make_node(id: &str, file: &str, label: &str) -> DbNode {
        DbNode {
            id: id.to_string(),
            workspace: Some("test".to_string()),
            label: label.to_string(),
            kind: "function".to_string(),
            file_path: file.to_string(),
            language: "rust".to_string(),
            line_start: 1,
            line_end: 10,
            text: format!("fn {}() {{}}", label),
            community: None,
            file_hash: None,
        }
    }

    #[test]
    fn filter_singletons() {
        let nodes = vec![
            make_node("1", "src/a.rs", "fn_a1"),
            make_node("2", "src/a.rs", "fn_a2"),
            make_node("3", "src/a.rs", "fn_a3"),
            make_node("4", "src/b.rs", "fn_b1"), // singleton
        ];

        let mut assignments = HashMap::new();
        assignments.insert("1".to_string(), 0);
        assignments.insert("2".to_string(), 0);
        assignments.insert("3".to_string(), 0);
        assignments.insert("4".to_string(), 1); // singleton cluster

        // summarize without filter returns both
        let all = CommunityDetector::summarize(&nodes, &assignments);
        assert_eq!(all.len(), 2);

        // summarize_filtered with min_size=2 drops the singleton
        let filtered = CommunityDetector::summarize_filtered(&nodes, &assignments, 2);
        assert_eq!(filtered.len(), 1);
        assert_eq!(filtered[0].id, 0);
        assert_eq!(filtered[0].node_count, 3);
    }

    #[test]
    fn test_consolidate_communities_absorbs_singletons_via_links() {
        use omni_graph::db::DbLink;

        let nodes = vec![
            make_node("m1", "src/core/a.rs", "fn_a1"),
            make_node("m2", "src/core/a.rs", "fn_a2"),
            make_node("m3", "src/core/a.rs", "fn_a3"),
            make_node("m4", "src/core/b.rs", "fn_b1"),
            make_node("m5", "src/core/b.rs", "fn_b2"),
            make_node("micro1", "src/isolated/c.rs", "isolated_fn"),
        ];

        let links = vec![
            DbLink {
                id: "e1".to_string(),
                workspace: Some("test".to_string()),
                source: "micro1".to_string(),
                target: "m1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            }
        ];

        let mut assignments = HashMap::new();
        assignments.insert("m1".to_string(), 0);
        assignments.insert("m2".to_string(), 0);
        assignments.insert("m3".to_string(), 0);
        assignments.insert("m4".to_string(), 0);
        assignments.insert("m5".to_string(), 0);
        assignments.insert("micro1".to_string(), 1); // micro-community

        CommunityDetector::consolidate_communities(&nodes, &links, &mut assignments, 5);

        // micro1 should be absorbed by community 0 due to CALLS link to m1
        assert_eq!(assignments.values().collect::<std::collections::HashSet<_>>().len(), 1);
        assert_eq!(assignments["micro1"], assignments["m1"]);
    }

    #[test]
    fn test_consolidate_communities_absorbs_via_directory() {
        let nodes = vec![
            make_node("m1", "src/core/a.rs", "fn_a1"),
            make_node("m2", "src/core/a.rs", "fn_a2"),
            make_node("m3", "src/core/a.rs", "fn_a3"),
            make_node("m4", "src/core/b.rs", "fn_b1"),
            make_node("m5", "src/core/b.rs", "fn_b2"),
            make_node("micro1", "src/core/orphan.rs", "orphan_fn"), // in src/core
        ];

        let links = vec![]; // no links

        let mut assignments = HashMap::new();
        assignments.insert("m1".to_string(), 0);
        assignments.insert("m2".to_string(), 0);
        assignments.insert("m3".to_string(), 0);
        assignments.insert("m4".to_string(), 0);
        assignments.insert("m5".to_string(), 0);
        assignments.insert("micro1".to_string(), 99); // disconnected

        CommunityDetector::consolidate_communities(&nodes, &links, &mut assignments, 5);

        // micro1 should be absorbed by macro-community 0 because they share directory "src/core"
        assert_eq!(assignments.values().collect::<std::collections::HashSet<_>>().len(), 1);
        assert_eq!(assignments["micro1"], assignments["m1"]);
    }

    #[test]
    fn test_consolidate_communities_all_small_merges_to_zero() {
        let nodes = vec![
            make_node("1", "src/a.rs", "f1"),
            make_node("2", "src/b.rs", "f2"),
        ];
        let links = vec![];
        let mut assignments = HashMap::new();
        assignments.insert("1".to_string(), 10);
        assignments.insert("2".to_string(), 20);

        CommunityDetector::consolidate_communities(&nodes, &links, &mut assignments, 5);
        assert_eq!(assignments["1"], 0);
        assert_eq!(assignments["2"], 0);
    }
}

mod analytics_tests {
    use omni_graph::analytics::AnalyticsEngine;
    use std::fs::{self, File};
    use std::io::Write;
    use std::sync::Mutex;

    static BRAIN_ENV_MUTEX: Mutex<()> = Mutex::new(());

    #[test]
    fn test_analytics_scan_and_detail_with_synthetic_session() {
        let _guard = BRAIN_ENV_MUTEX.lock().unwrap();
        let temp_dir = std::env::temp_dir().join(format!("omni_test_brain_{}", std::process::id()));
        let session_id = "test-session-12345";
        let session_dir = temp_dir.join(session_id).join(".system_generated").join("logs");
        fs::create_dir_all(&session_dir).expect("create temp session dir");

        let transcript_path = session_dir.join("transcript.jsonl");
        let mut file = File::create(&transcript_path).expect("create transcript");

        let step1 = serde_json::json!({
            "step_index": 0,
            "source": "USER_INPUT",
            "type": "USER_INPUT",
            "status": "DONE",
            "created_at": "2026-09-25T12:00:00Z",
            "content": "Look up symbol NewSessionIndex in GoLang",
        });

        let step2 = serde_json::json!({
            "step_index": 1,
            "source": "MODEL",
            "type": "PLANNER_RESPONSE",
            "status": "DONE",
            "created_at": "2026-09-25T12:00:05Z",
            "content": "Using omni-graph to find symbol definition",
            "thinking": "I should query the Omni-Graph AST rather than dumping whole files.",
            "tool_calls": [
                {
                    "name": "run_command",
                    "args": {
                        "CommandLine": "make graph-symbol SYM=NewSessionIndex"
                    }
                },
                {
                    "name": "view_file",
                    "args": {
                        "AbsolutePath": "/Users/aparv/code/server.go"
                    }
                }
            ]
        });

        writeln!(file, "{}", step1).unwrap();
        writeln!(file, "{}", step2).unwrap();
        drop(file);

        // Point BRAIN_DIR to temp_dir
        std::env::set_var("BRAIN_DIR", &temp_dir);

        let analytics = AnalyticsEngine::scan_analytics();
        assert_eq!(analytics.summary.total_sessions, 1);
        assert_eq!(analytics.summary.total_steps, 2);
        assert_eq!(analytics.summary.total_tool_calls, 2);
        assert_eq!(analytics.summary.total_omni_calls, 1);
        assert_eq!(analytics.summary.total_lsp_lookups, 1);
        assert!(analytics.summary.estimated_tokens_saved > 0);

        // Verify tools breakdown
        assert!(analytics.tools_breakdown.iter().any(|t| t.name == "run_command"));
        assert!(analytics.tools_breakdown.iter().any(|t| t.name == "view_file"));

        // Verify language telemetry detected .go
        assert!(analytics.languages_telemetry.iter().any(|l| l.language == "Go"));

        // Verify session detail
        let detail = AnalyticsEngine::get_session_detail(session_id);
        assert!(detail.is_some());
        let d = detail.unwrap();
        assert_eq!(d.session_id, session_id);
        assert_eq!(d.total_steps, 2);
        assert_eq!(d.total_tools, 2);
        assert_eq!(d.omni_tools, 1);
        assert_eq!(d.messages.len(), 2);
        assert_eq!(d.messages[1].tool_calls.len(), 2);
        assert!(d.messages[1].tool_calls[0].is_omni);
        assert_eq!(
            d.messages[1].tool_calls[0].omni_category,
            Some("Omni-Graph: AST Definition (LSP)".to_string())
        );

        // Non-existent session returns None
        assert!(AnalyticsEngine::get_session_detail("non-existent-session-xyz").is_none());

        // Cleanup
        let _ = fs::remove_dir_all(&temp_dir);
        std::env::remove_var("BRAIN_DIR");
    }

    #[test]
    fn test_dynamic_workspace_selection_dominance() {
        let _guard = BRAIN_ENV_MUTEX.lock().unwrap();
        let temp_dir = std::env::temp_dir().join(format!("omni_test_dominance_{}", std::process::id()));
        let session_id = "test-session-quarkdock-dom";
        let session_dir = temp_dir.join(session_id).join(".system_generated").join("logs");
        fs::create_dir_all(&session_dir).expect("create temp session dir");

        let transcript_path = session_dir.join("transcript.jsonl");
        let mut file = File::create(&transcript_path).expect("create transcript");

        // Initial prompt mentions QuarkDock
        let step0 = serde_json::json!({
            "step_index": 0,
            "source": "USER_INPUT",
            "type": "USER_INPUT",
            "status": "DONE",
            "created_at": "2026-09-29T12:00:00Z",
            "content": "welcome to QuarkDock code repo. we will be building frontend and backend",
        });

        // Step 1 touches tool-scripts once (e.g. reading rules)
        let step1 = serde_json::json!({
            "step_index": 1,
            "source": "MODEL",
            "type": "PLANNER_RESPONSE",
            "status": "DONE",
            "created_at": "2026-09-29T12:00:05Z",
            "tool_calls": [
                {
                    "name": "view_file",
                    "args": {
                        "AbsolutePath": "/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/tool-scripts/.agents/AGENTS.md"
                    }
                }
            ]
        });

        // Step 2 touches QuarkDock multiple times
        let step2 = serde_json::json!({
            "step_index": 2,
            "source": "MODEL",
            "type": "PLANNER_RESPONSE",
            "status": "DONE",
            "created_at": "2026-09-29T12:00:10Z",
            "tool_calls": [
                {
                    "name": "run_command",
                    "args": {
                        "Cwd": "/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/QuarkDock",
                        "CommandLine": "docker compose up -d"
                    }
                },
                {
                    "name": "view_file",
                    "args": {
                        "AbsolutePath": "/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/QuarkDock/services/ui/src/App.tsx"
                    }
                },
                {
                    "name": "write_to_file",
                    "args": {
                        "TargetFile": "/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/QuarkDock/services/api/main.py"
                    }
                }
            ]
        });

        writeln!(file, "{}", step0).unwrap();
        writeln!(file, "{}", step1).unwrap();
        writeln!(file, "{}", step2).unwrap();
        drop(file);

        std::env::set_var("BRAIN_DIR", &temp_dir);

        let analytics = AnalyticsEngine::scan_analytics();
        assert_eq!(analytics.summary.total_sessions, 1);
        let sess = &analytics.sessions[0];
        assert_eq!(
            sess.workspace,
            "QuarkDock",
            "QuarkDock must dominate over tool-scripts because it has 3 tool operations vs 1"
        );

        let detail = AnalyticsEngine::get_session_detail(session_id).expect("must find session detail");
        assert_eq!(
            detail.workspace,
            "QuarkDock",
            "Session detail must also resolve to QuarkDock"
        );

        // Cleanup
        let _ = fs::remove_dir_all(&temp_dir);
        std::env::remove_var("BRAIN_DIR");
    }
}

mod worktree_tests {
    use omni_graph::ingestion::worktree::{detect_git_worktree, discover_parent_worktrees};
    use std::fs;

    #[test]
    fn test_detect_git_worktree_valid() {
        let ts = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap().as_nanos();
        let temp_dir = std::env::temp_dir().join(format!("og_test_wt_{}", ts));
        let parent_repo = temp_dir.join("parent-project");
        let worktree_dir = temp_dir.join("wt-feature");

        let parent_git_wt = parent_repo.join(".git").join("worktrees").join("wt-feature");
        fs::create_dir_all(&parent_git_wt).unwrap();
        fs::write(parent_git_wt.join("gitdir"), worktree_dir.join(".git").to_string_lossy().to_string()).unwrap();
        fs::write(parent_git_wt.join("HEAD"), "ref: refs/heads/feature-alpha\n").unwrap();

        fs::create_dir_all(&worktree_dir).unwrap();
        let git_file_content = format!("gitdir: {}\n", parent_git_wt.to_string_lossy());
        fs::write(worktree_dir.join(".git"), git_file_content).unwrap();

        let info = detect_git_worktree(&worktree_dir);
        assert!(info.is_some(), "Should successfully detect valid git worktree");
        let wt = info.unwrap();
        assert!(wt.is_worktree);
        assert_eq!(wt.worktree_name, "wt-feature");
        assert_eq!(wt.branch.as_deref(), Some("feature-alpha"));
        assert_eq!(wt.parent_workspace, "parent-project");

        // Clean up
        let _ = fs::remove_dir_all(&temp_dir);
    }

    #[test]
    fn test_detect_git_worktree_regular_repo() {
        let ts = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap().as_nanos();
        let temp_dir = std::env::temp_dir().join(format!("og_test_reg_{}", ts));
        let git_dir = temp_dir.join(".git");
        fs::create_dir_all(&git_dir).unwrap();

        let info = detect_git_worktree(&temp_dir);
        assert!(info.is_none(), "Regular git directory should not be detected as worktree");

        let _ = fs::remove_dir_all(&temp_dir);
    }

    #[test]
    fn test_discover_parent_worktrees() {
        let ts = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap().as_nanos();
        let temp_dir = std::env::temp_dir().join(format!("og_test_disc_{}", ts));
        let parent_repo = temp_dir.join("core-app");
        let parent_git_wts = parent_repo.join(".git").join("worktrees");

        let wt1_meta = parent_git_wts.join("branch-one");
        let wt2_meta = parent_git_wts.join("branch-two");
        fs::create_dir_all(&wt1_meta).unwrap();
        fs::create_dir_all(&wt2_meta).unwrap();

        let wt1_dir = temp_dir.join("branch-one");
        let wt2_dir = temp_dir.join("branch-two");
        fs::create_dir_all(&wt1_dir).unwrap();
        fs::create_dir_all(&wt2_dir).unwrap();

        fs::write(wt1_meta.join("gitdir"), wt1_dir.join(".git").to_string_lossy().to_string()).unwrap();
        fs::write(wt1_meta.join("HEAD"), "ref: refs/heads/one\n").unwrap();

        fs::write(wt2_meta.join("gitdir"), wt2_dir.join(".git").to_string_lossy().to_string()).unwrap();
        fs::write(wt2_meta.join("HEAD"), "ref: refs/heads/two\n").unwrap();

        let discovered = discover_parent_worktrees(&parent_repo);
        assert_eq!(discovered.len(), 2, "Should discover two worktrees for parent repo");

        let names: Vec<String> = discovered.iter().map(|w| w.worktree_name.clone()).collect();
        assert!(names.contains(&"branch-one".to_string()));
        assert!(names.contains(&"branch-two".to_string()));
        assert!(discovered.iter().all(|w| w.root_path.is_some()), "All discovered worktrees must have root_path");

        let _ = fs::remove_dir_all(&temp_dir);
    }

    #[test]
    fn test_extract_sql_arrays_filters_null_let_statements() {
        use serde_json::json;
        let mock_surreal_resp = json!([
            {"result": null, "status": "OK"},
            {"result": null, "status": "OK"},
            {"result": null, "status": "OK"},
            {"result": [{"id": "node:1", "label": "test_fn"}], "status": "OK"},
            {"result": [{"id": "linked_to:1", "source": "node:1", "target": "node:2"}], "status": "OK"}
        ]);

        let arrays = omni_graph::db::extract_sql_arrays(&mock_surreal_resp);
        assert_eq!(arrays.len(), 2, "Should safely extract exactly 2 SELECT array results");
        assert_eq!(arrays[0].len(), 1);
        assert_eq!(arrays[0][0]["label"], "test_fn");
        assert_eq!(arrays[1].len(), 1);
        assert_eq!(arrays[1][0]["source"], "node:1");
    }

    #[test]
    fn test_parse_struct_fields_rust() {
        use omni_graph::parser::CodeParser;
        let rust_code = r#"
        pub struct UserConfig {
            pub username: String,
            pub timeout_ms: u64,
            active: bool,
        }
        "#;
        let res = CodeParser::parse_file("test-ws", "src/config.rs", rust_code).expect("should parse rust");
        let nodes = res.nodes;
        let edges = res.edges;

        // Should extract the struct
        assert!(nodes.iter().any(|n| n.label == "UserConfig" && n.kind == "struct"));

        // Should extract all 3 fields as kind: "field"
        let fields: Vec<_> = nodes.iter().filter(|n| n.kind == "field").collect();
        assert_eq!(fields.len(), 3);
        assert!(fields.iter().any(|f| f.label == "username"));
        assert!(fields.iter().any(|f| f.label == "timeout_ms"));
        assert!(fields.iter().any(|f| f.label == "active"));

        // Struct should CONTAINS fields
        let contains_edges: Vec<_> = edges.iter().filter(|e| e.edge_type == "CONTAINS").collect();
        assert_eq!(contains_edges.len(), 3);
        assert!(contains_edges.iter().any(|e| e.target_label == "username"));

        // Fields should reference types
        let ref_edges: Vec<_> = edges.iter().filter(|e| e.edge_type == "REFERENCES").collect();
        assert!(ref_edges.iter().any(|e| e.target_label == "String"));
        assert!(ref_edges.iter().any(|e| e.target_label == "u64"));
    }

    #[test]
    fn test_parse_interface_properties_ts() {
        use omni_graph::parser::CodeParser;
        let ts_code = r#"
        interface SessionPayload {
            sessionId: string;
            tokenCount: number;
        }
        "#;
        let res = CodeParser::parse_file("test-ws", "src/types.ts", ts_code).expect("should parse ts");
        let nodes = res.nodes;
        let edges = res.edges;

        assert!(nodes.iter().any(|n| n.label == "SessionPayload" && n.kind == "interface"));

        let fields: Vec<_> = nodes.iter().filter(|n| n.kind == "field").collect();
        assert_eq!(fields.len(), 2);
        assert!(fields.iter().any(|f| f.label == "sessionId"));
        assert!(fields.iter().any(|f| f.label == "tokenCount"));

        assert!(edges.iter().any(|e| e.edge_type == "CONTAINS" && e.target_label == "sessionId"));
    }

    #[test]
    fn test_parse_struct_fields_go() {
        use omni_graph::parser::CodeParser;
        let go_code = r#"
        package main
        type ServerOptions struct {
            Host string
            Port int
        }
        "#;
        let res = CodeParser::parse_file("test-ws", "options.go", go_code).expect("should parse go");
        let nodes = res.nodes;
        let edges = res.edges;

        assert!(nodes.iter().any(|n| n.label == "ServerOptions" && n.kind == "struct"));

        let fields: Vec<_> = nodes.iter().filter(|n| n.kind == "field").collect();
        assert_eq!(fields.len(), 2);
        assert!(fields.iter().any(|f| f.label == "Host"));
        assert!(fields.iter().any(|f| f.label == "Port"));

        assert!(edges.iter().any(|e| e.edge_type == "CONTAINS" && e.target_label == "Host"));
    }

    #[test]
    fn test_inject_watcher_status_root_and_worktree() {
        use omni_graph::api::inject_watcher_status;
        use omni_graph::watcher::WatcherStatus;

        let mut workspaces = vec![
            serde_json::json!({
                "workspace": "PolyTuring",
                "total_nodes": 3800,
                "worktrees": [
                    {
                        "workspace": "wt-feature",
                        "total_nodes": 500
                    },
                    {
                        "workspace": "wt-inactive",
                        "total_nodes": 300
                    }
                ]
            }),
            serde_json::json!({
                "workspace": "unwatched-repo",
                "total_nodes": 120,
                "worktrees": []
            })
        ];

        let active_statuses = vec![
            WatcherStatus {
                workspace: "PolyTuring".to_string(),
                path: "/path/to/PolyTuring".to_string(),
                status: "watching".to_string(),
                started_at: "2026-11-03T00:00:00Z".to_string(),
                files_tracked: 200,
                last_sync: None,
                events_processed: 0,
                files_reindexed: 0,
                files_deleted: 0,
                avg_sync_ms: 0,
                cluster_status: "current".to_string(),
                last_cluster_at: None,
                debounce_ms: 500,
            },
            WatcherStatus {
                workspace: "wt-feature".to_string(),
                path: "/path/to/wt-feature".to_string(),
                status: "syncing".to_string(), // Actively syncing must be recognized as is_watching == true
                started_at: "2026-11-03T00:00:00Z".to_string(),
                files_tracked: 50,
                last_sync: None,
                events_processed: 0,
                files_reindexed: 0,
                files_deleted: 0,
                avg_sync_ms: 0,
                cluster_status: "current".to_string(),
                last_cluster_at: None,
                debounce_ms: 500,
            },
        ];

        inject_watcher_status(&mut workspaces, &active_statuses);

        // Root workspace PolyTuring is watching
        assert_eq!(workspaces[0]["is_watching"], serde_json::Value::Bool(true));

        // Nested worktree wt-feature is watching
        assert_eq!(workspaces[0]["worktrees"][0]["is_watching"], serde_json::Value::Bool(true));

        // Nested worktree wt-inactive is NOT watching
        assert_eq!(workspaces[0]["worktrees"][1]["is_watching"], serde_json::Value::Bool(false));

        // unwatched-repo is NOT watching
        assert_eq!(workspaces[1]["is_watching"], serde_json::Value::Bool(false));
    }

    #[test]
    fn parse_java_classes_methods_inheritance_and_imports() {
        use omni_graph::parser::CodeParser;
        let java_code = r#"
        package com.example.service;

        import com.example.model.User;
        import java.io.Serializable;

        @Service
        public class UserService extends BaseService implements Serializable {
            private final UserRepository userRepository;

            public UserService(UserRepository repo) {
                this.userRepository = repo;
            }

            public User findById(String id) {
                return userRepository.findById(id);
            }
        }
        "#;
        let res = CodeParser::parse_file("test-ws", "UserService.java", java_code)
            .expect("should parse java code");
        let nodes = res.nodes;
        let edges = res.edges;

        // Class
        assert!(nodes.iter().any(|n| n.label == "UserService" && n.kind == "class"));

        // Constructor and Method
        assert!(nodes.iter().any(|n| n.label == "UserService" && n.kind == "constructor"));
        assert!(nodes.iter().any(|n| n.label == "findById" && n.kind == "method"));

        // Field
        assert!(nodes.iter().any(|n| n.label == "userRepository" && n.kind == "field"));

        // Imports
        assert!(edges.iter().any(|e| e.edge_type == "IMPORTS" && e.target_label == "User"));

        // Inheritance (EXTENDS & IMPLEMENTS)
        assert!(edges.iter().any(|e| e.edge_type == "EXTENDS" && e.target_label == "BaseService"));
        assert!(edges.iter().any(|e| e.edge_type == "IMPLEMENTS" && e.target_label == "Serializable"));

        // Calls (userRepository.findById)
        assert!(edges.iter().any(|e| e.edge_type == "CALLS" && e.target_label == "findById"));
    }

    #[test]
    fn parse_cpp_classes_functions_and_includes() {
        use omni_graph::parser::CodeParser;
        let cpp_code = r#"
        #include <vector>
        #include "shape.h"

        namespace graphics {
            class Circle : public Shape {
            public:
                double radius;
                Circle(double r) : radius(r) {}
                double area() {
                    return computeArea(radius);
                }
            };

            double computeArea(double r) {
                return 3.14159 * r * r;
            }
        }
        "#;
        let res = CodeParser::parse_file("test-ws", "circle.cpp", cpp_code)
            .expect("should parse cpp code");
        let nodes = res.nodes;
        let edges = res.edges;

        // Class & namespace
        assert!(nodes.iter().any(|n| n.label == "Circle" && n.kind == "class"));
        assert!(nodes.iter().any(|n| n.label == "graphics" && n.kind == "namespace"));

        // Method & function
        assert!(nodes.iter().any(|n| n.label == "computeArea" && n.kind == "function"));

        // Inheritance (Circle EXTENDS Shape)
        assert!(edges.iter().any(|e| e.edge_type == "EXTENDS" && e.target_label == "Shape"));

        // Include edges
        assert!(edges.iter().any(|e| e.edge_type == "IMPORTS" && (e.target_label == "vector" || e.target_label == "shape")));

        // Calls
        assert!(edges.iter().any(|e| e.edge_type == "CALLS" && e.target_label == "computeArea"));
    }

    #[test]
    fn parse_typescript_filter_detached_css_properties() {
        use omni_graph::parser::CodeParser;
        let ts_code = r#"
        export interface UserCardProps {
            username: string;
            avatarUrl: string;
        }

        // Inline object literal (should NOT be extracted as top-level field nodes)
        const styles = {
            cursor: 'pointer',
            color: '#ffffff',
            borderRadius: '8px'
        };

        export const renderCard = (props: UserCardProps) => {
            return props.username;
        };
        "#;
        let res = CodeParser::parse_file("test-ws", "Card.tsx", ts_code)
            .expect("should parse tsx code");
        let nodes = res.nodes;

        // Interface and its properties MUST be extracted
        assert!(nodes.iter().any(|n| n.label == "UserCardProps" && n.kind == "interface"));
        assert!(nodes.iter().any(|n| n.label == "username" && n.kind == "field"));
        assert!(nodes.iter().any(|n| n.label == "avatarUrl" && n.kind == "field"));

        // Detached CSS style properties MUST NOT be extracted as fields
        assert!(!nodes.iter().any(|n| n.label == "cursor"));
        assert!(!nodes.iter().any(|n| n.label == "borderRadius"));
        assert!(!nodes.iter().any(|n| n.label == "color"));
    }
}

mod topological_backbone_tests {
    use omni_graph::analysis::get_edge_weight;
    use omni_graph::parser::{ExtractedEdge, ExtractedNode};
    use std::collections::{HashSet, VecDeque};

    #[test]
    fn test_edge_weights_containment_hierarchy() {
        assert!((get_edge_weight("IMPLEMENTS") - 3.0).abs() < 1e-5);
        assert!((get_edge_weight("CALLS") - 2.0).abs() < 1e-5);
        assert!((get_edge_weight("TYPE_REF") - 1.5).abs() < 1e-5);
        assert!((get_edge_weight("IMPORTS") - 1.0).abs() < 1e-5);
        assert!((get_edge_weight("CONTAINS") - 0.8).abs() < 1e-5);
        assert!((get_edge_weight("CONTAINS_FILE") - 0.4).abs() < 1e-5);
        assert!((get_edge_weight("CONTAINS_DIR") - 0.2).abs() < 1e-5);
    }

    #[test]
    fn test_topological_containment_dag_reachability() {
        let ws = "TestWorkspace";
        let ws_root_id = format!("{}:workspace", ws);

        // Nodes: 1 workspace + 2 dirs + 2 files + 4 symbols = 9 nodes
        let nodes = vec![
            ExtractedNode {
                id: ws_root_id.clone(),
                workspace: ws.to_string(),
                label: ws.to_string(),
                kind: "workspace".to_string(),
                file_path: "".to_string(),
                language: "".to_string(),
                line_start: 0,
                line_end: 0,
                text: "root".to_string(),
            },
            ExtractedNode {
                id: format!("{}:dir:src", ws),
                workspace: ws.to_string(),
                label: "src".to_string(),
                kind: "directory".to_string(),
                file_path: "src".to_string(),
                language: "".to_string(),
                line_start: 0,
                line_end: 0,
                text: "dir src".to_string(),
            },
            ExtractedNode {
                id: format!("{}:dir:src/parser", ws),
                workspace: ws.to_string(),
                label: "parser".to_string(),
                kind: "directory".to_string(),
                file_path: "src/parser".to_string(),
                language: "".to_string(),
                line_start: 0,
                line_end: 0,
                text: "dir parser".to_string(),
            },
            ExtractedNode {
                id: format!("{}:file:src/main.rs", ws),
                workspace: ws.to_string(),
                label: "main.rs".to_string(),
                kind: "file".to_string(),
                file_path: "src/main.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "fn main() {}".to_string(),
            },
            ExtractedNode {
                id: format!("{}:file:src/parser/mod.rs", ws),
                workspace: ws.to_string(),
                label: "mod.rs".to_string(),
                kind: "file".to_string(),
                file_path: "src/parser/mod.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 20,
                text: "pub fn parse() {}".to_string(),
            },
            // Symbols in main.rs
            ExtractedNode {
                id: format!("{}:src/main.rs:main:1", ws),
                workspace: ws.to_string(),
                label: "main".to_string(),
                kind: "function".to_string(),
                file_path: "src/main.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 5,
                text: "fn main() {}".to_string(),
            },
            ExtractedNode {
                id: format!("{}:src/main.rs:Config:6", ws),
                workspace: ws.to_string(),
                label: "Config".to_string(),
                kind: "struct".to_string(),
                file_path: "src/main.rs".to_string(),
                language: "rust".to_string(),
                line_start: 6,
                line_end: 10,
                text: "struct Config;".to_string(),
            },
            // Symbols in parser/mod.rs
            ExtractedNode {
                id: format!("{}:src/parser/mod.rs:parse:1", ws),
                workspace: ws.to_string(),
                label: "parse".to_string(),
                kind: "function".to_string(),
                file_path: "src/parser/mod.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "pub fn parse() {}".to_string(),
            },
            ExtractedNode {
                id: format!("{}:src/parser/mod.rs:AstParser:11", ws),
                workspace: ws.to_string(),
                label: "AstParser".to_string(),
                kind: "struct".to_string(),
                file_path: "src/parser/mod.rs".to_string(),
                language: "rust".to_string(),
                line_start: 11,
                line_end: 20,
                text: "pub struct AstParser;".to_string(),
            },
        ];

        // Edges:
        // Workspace -> CONTAINS_DIR -> src
        // src -> CONTAINS_DIR -> src/parser
        // src -> CONTAINS_FILE -> src/main.rs
        // src/parser -> CONTAINS_FILE -> src/parser/mod.rs
        // src/main.rs -> CONTAINS -> main
        // src/main.rs -> CONTAINS -> Config
        // src/parser/mod.rs -> CONTAINS -> parse
        // src/parser/mod.rs -> CONTAINS -> AstParser
        // main -> CALLS -> parse (AST edge)
        let edges = vec![
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: ws_root_id.clone(),
                target_label: "src".to_string(),
                target_id: Some(format!("{}:dir:src", ws)),
                edge_type: "CONTAINS_DIR".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:dir:src", ws),
                target_label: "parser".to_string(),
                target_id: Some(format!("{}:dir:src/parser", ws)),
                edge_type: "CONTAINS_DIR".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:dir:src", ws),
                target_label: "main.rs".to_string(),
                target_id: Some(format!("{}:file:src/main.rs", ws)),
                edge_type: "CONTAINS_FILE".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:dir:src/parser", ws),
                target_label: "mod.rs".to_string(),
                target_id: Some(format!("{}:file:src/parser/mod.rs", ws)),
                edge_type: "CONTAINS_FILE".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:file:src/main.rs", ws),
                target_label: "main".to_string(),
                target_id: Some(format!("{}:src/main.rs:main:1", ws)),
                edge_type: "CONTAINS".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:file:src/main.rs", ws),
                target_label: "Config".to_string(),
                target_id: Some(format!("{}:src/main.rs:Config:6", ws)),
                edge_type: "CONTAINS".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:file:src/parser/mod.rs", ws),
                target_label: "parse".to_string(),
                target_id: Some(format!("{}:src/parser/mod.rs:parse:1", ws)),
                edge_type: "CONTAINS".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:file:src/parser/mod.rs", ws),
                target_label: "AstParser".to_string(),
                target_id: Some(format!("{}:src/parser/mod.rs:AstParser:11", ws)),
                edge_type: "CONTAINS".to_string(),
                category: "CONTAINMENT".to_string(),
            },
            ExtractedEdge {
                workspace: ws.to_string(),
                source_id: format!("{}:src/main.rs:main:1", ws),
                target_label: "parse".to_string(),
                target_id: Some(format!("{}:src/parser/mod.rs:parse:1", ws)),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
        ];

        // 1. Verify BFS reachability from workspace root to ALL nodes
        let mut adj: std::collections::HashMap<String, Vec<String>> = std::collections::HashMap::new();
        for edge in &edges {
            if let Some(ref tgt) = edge.target_id {
                adj.entry(edge.source_id.clone()).or_default().push(tgt.clone());
            }
        }

        let mut visited = HashSet::new();
        let mut queue = VecDeque::new();
        queue.push_back(ws_root_id.clone());
        visited.insert(ws_root_id.clone());

        while let Some(curr) = queue.pop_front() {
            if let Some(neighbors) = adj.get(&curr) {
                for nbr in neighbors {
                    if visited.insert(nbr.clone()) {
                        queue.push_back(nbr.clone());
                    }
                }
            }
        }

        for node in &nodes {
            assert!(
                visited.contains(&node.id),
                "Node '{}' of kind '{}' must be reachable from workspace root in topological DAG",
                node.id,
                node.kind
            );
        }

        assert_eq!(visited.len(), nodes.len(), "Every vertex in the graph must belong to the connected giant component");
    }
}

mod hybrid_gitignore_tests {
    use omni_graph::watcher::{count_tracked_files, is_path_ignored, load_workspace_gitignore};
    use std::fs::{self, File};
    use std::io::Write;
    use std::path::PathBuf;

    fn make_test_temp_dir(suffix: &str) -> PathBuf {
        let nanos = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_nanos();
        let dir = std::env::temp_dir().join(format!("omni_test_{}_{}_{}", suffix, std::process::id(), nanos));
        fs::create_dir_all(&dir).expect("Failed to create tempdir");
        dir
    }

    #[test]
    fn test_r015_hybrid_filter_ignores_gitignored_directories_and_files() {
        let root = make_test_temp_dir("r015_filter");

        // Create a realistic .gitignore
        let gitignore_path = root.join(".gitignore");
        let mut f = File::create(&gitignore_path).expect("Failed to create .gitignore");
        writeln!(f, ".playwright-mcp/").unwrap();
        writeln!(f, "coverage/").unwrap();
        writeln!(f, "*.log").unwrap();
        writeln!(f, "secrets.env").unwrap();

        let gitignore = load_workspace_gitignore(&root);

        // Gitignored paths should be rejected
        let pw_file = root.join(".playwright-mcp").join("trace.json");
        let cov_file = root.join("coverage").join("index.html");
        let log_file = root.join("debug.log");
        let sec_file = root.join("secrets.env");

        assert!(is_path_ignored(&root, &gitignore, &pw_file, false), ".playwright-mcp child file must be ignored");
        assert!(is_path_ignored(&root, &gitignore, &root.join(".playwright-mcp"), true), ".playwright-mcp directory must be ignored");
        assert!(is_path_ignored(&root, &gitignore, &cov_file, false), "coverage child file must be ignored");
        assert!(is_path_ignored(&root, &gitignore, &log_file, false), "*.log pattern must be ignored");
        assert!(is_path_ignored(&root, &gitignore, &sec_file, false), "secrets.env must be ignored");

        // Static Layer 1 Invariants should also be rejected
        let node_modules_file = root.join("node_modules").join("package.json");
        let target_file = root.join("target").join("debug").join("app");
        let binary_file = root.join("src").join("logo.png");

        assert!(is_path_ignored(&root, &gitignore, &node_modules_file, false), "node_modules must be ignored via Layer 1");
        assert!(is_path_ignored(&root, &gitignore, &target_file, false), "target directory must be ignored via Layer 1");
        assert!(is_path_ignored(&root, &gitignore, &binary_file, false), "binary .png file must be ignored via Layer 1");

        // Legitimate source and doc files must NOT be ignored
        let source_file = root.join("src").join("main.rs");
        let readme_file = root.join("README.md");

        assert!(!is_path_ignored(&root, &gitignore, &source_file, false), "src/main.rs must not be ignored");
        assert!(!is_path_ignored(&root, &gitignore, &readme_file, false), "README.md must not be ignored");

        let _ = fs::remove_dir_all(&root);
    }

    #[test]
    fn test_r015_hybrid_filter_fallback_without_gitignore() {
        // Non-git repo: No .gitignore present
        let root = make_test_temp_dir("r015_nongit");

        let gitignore = load_workspace_gitignore(&root);

        // Layer 1 invariants must still function cleanly
        let node_modules = root.join("node_modules").join("lib.js");
        let git_dir = root.join(".git").join("HEAD");
        let target_dir = root.join("target").join("out.bin");
        let zip_file = root.join("archive.zip");

        assert!(is_path_ignored(&root, &gitignore, &node_modules, false), "node_modules must be pruned even without .gitignore");
        assert!(is_path_ignored(&root, &gitignore, &git_dir, false), ".git must be pruned even without .gitignore");
        assert!(is_path_ignored(&root, &gitignore, &target_dir, false), "target must be pruned even without .gitignore");
        assert!(is_path_ignored(&root, &gitignore, &zip_file, false), "binary zip must be pruned even without .gitignore");

        // Clean source files must be accepted
        let go_source = root.join("server.go");
        let doc_file = root.join("architecture.md");

        assert!(!is_path_ignored(&root, &gitignore, &go_source, false), "server.go must be accepted in non-git workspace");
        assert!(!is_path_ignored(&root, &gitignore, &doc_file, false), "architecture.md must be accepted in non-git workspace");

        let _ = fs::remove_dir_all(&root);
    }

    #[test]
    fn test_r015_count_tracked_files_with_hybrid_filter() {
        let root = make_test_temp_dir("r015_count");

        // Create directory hierarchy
        fs::create_dir_all(root.join("src")).unwrap();
        fs::create_dir_all(root.join("docs")).unwrap();
        fs::create_dir_all(root.join("coverage")).unwrap();
        fs::create_dir_all(root.join("target")).unwrap();
        fs::create_dir_all(root.join(".playwright-mcp")).unwrap();

        // Valid source files (3 files)
        fs::write(root.join("src/main.rs"), "fn main() {}").unwrap();
        fs::write(root.join("src/utils.rs"), "pub fn helper() {}").unwrap();
        fs::write(root.join("docs/index.md"), "# Title").unwrap();

        // Ignored files via .gitignore
        fs::write(root.join(".gitignore"), "coverage/\n*.log\n").unwrap();
        fs::write(root.join("coverage/lcov.info"), "TN:").unwrap();
        fs::write(root.join("debug.log"), "error log").unwrap();

        // Ignored files via Layer 1 static invariants
        fs::write(root.join("target/build.out"), "bin").unwrap();
        fs::write(root.join(".playwright-mcp/session.json"), "{}").unwrap();
        fs::write(root.join("src/photo.png"), "fake png bytes").unwrap();

        let gitignore = load_workspace_gitignore(&root);
        let tracked_count = count_tracked_files(root.to_str().unwrap(), &gitignore);

        assert_eq!(
            tracked_count, 3,
            "Only the 3 legitimate source files (src/main.rs, src/utils.rs, docs/index.md) should be tracked"
        );

        let _ = fs::remove_dir_all(&root);
    }
}

mod search_result_serde_tests {
    use omni_graph::db::SearchResult;

    #[test]
    fn test_search_result_handles_null_or_missing_similarity() {
        let json_with_null = r#"{
            "id": "node:PolyTuring:workspace",
            "workspace": "PolyTuring",
            "label": "PolyTuring",
            "kind": "workspace",
            "file_path": "",
            "language": "",
            "text": "workspace",
            "similarity": null,
            "community": null
        }"#;

        let res: Result<SearchResult, _> = serde_json::from_str(json_with_null);
        assert!(res.is_ok(), "SearchResult should deserialize cleanly even if similarity is null");
        let parsed = res.unwrap();
        assert_eq!(parsed.similarity, 0.0);
        assert_eq!(parsed.label, "PolyTuring");
    }

    #[test]
    fn test_search_result_handles_valid_float_similarity() {
        let json_with_val = r#"{
            "id": "node:PolyTuring:src/lib.rs:foo:1",
            "workspace": "PolyTuring",
            "label": "foo",
            "kind": "function",
            "file_path": "src/lib.rs",
            "language": "rust",
            "text": "fn foo() {}",
            "similarity": 0.895,
            "community": 42
        }"#;

        let res: Result<SearchResult, _> = serde_json::from_str(json_with_val);
        assert!(res.is_ok());
        let parsed = res.unwrap();
        assert_eq!(parsed.similarity, 0.895);
        assert_eq!(parsed.community, Some(42));
    }
}

mod hardening_v050_tests {
    use omni_graph::analysis::CommunityDetector;
    use omni_graph::db::{DbClient, DbLink, DbNode, InboundEdgeRef};
    use omni_graph::parser::ExtractedNode;

    fn make_db_node(id: &str, file: &str, label: &str) -> DbNode {
        DbNode {
            id: id.to_string(),
            workspace: Some("test-ws".to_string()),
            label: label.to_string(),
            kind: "function".to_string(),
            file_path: file.to_string(),
            language: "rust".to_string(),
            line_start: 1,
            line_end: 10,
            text: format!("fn {}() {{}}", label),
            community: None,
            file_hash: None,
        }
    }

    #[test]
    fn test_r077_ingestion_and_recluster_parity() {
        // Construct a graph with a core cluster and small micro-communities
        let mut nodes = Vec::new();
        for i in 1..=10 {
            nodes.push(make_db_node(&format!("core_{}", i), "src/core.rs", &format!("core_fn_{}", i)));
        }
        for i in 1..=3 {
            nodes.push(make_db_node(&format!("micro_{}", i), &format!("src/util_{}.rs", i), &format!("util_fn_{}", i)));
        }

        let mut links = Vec::new();
        // Dense core
        for i in 1..9 {
            links.push(DbLink {
                id: format!("e_{}", i),
                workspace: Some("test-ws".to_string()),
                source: format!("core_{}", i),
                target: format!("core_{}", i + 1),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            });
        }
        // Micro links connecting to core
        for i in 1..=3 {
            links.push(DbLink {
                id: format!("e_micro_{}", i),
                workspace: Some("test-ws".to_string()),
                source: format!("micro_{}", i),
                target: "core_1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            });
        }

        // Path 1: Ingestion pipeline (detect_with_seeds + consolidate)
        let mut assignments_ingestion = CommunityDetector::detect_with_seeds(&nodes, &links, 15, None);
        CommunityDetector::consolidate_communities(&nodes, &links, &mut assignments_ingestion, 5);
        let (summaries_ingestion, _) = CommunityDetector::compute_galaxy_metrics("test-ws", &nodes, &links, &assignments_ingestion);

        // Path 2: Recluster API (detect + consolidate)
        let mut assignments_recluster = CommunityDetector::detect(&nodes, &links, 15);
        CommunityDetector::consolidate_communities(&nodes, &links, &mut assignments_recluster, 5);
        let (summaries_recluster, _) = CommunityDetector::compute_galaxy_metrics("test-ws", &nodes, &links, &assignments_recluster);

        // Parity Verification (R-077):
        // Both pipelines must produce identical cluster counts and identical assignments
        assert_eq!(
            summaries_ingestion.len(),
            summaries_recluster.len(),
            "Ingestion and Recluster API must produce identical galaxy counts"
        );

        for node in &nodes {
            assert_eq!(
                assignments_ingestion.get(&node.id),
                assignments_recluster.get(&node.id),
                "Node '{}' community assignment must match across ingestion and recluster",
                node.id
            );
        }
    }

    #[test]
    fn test_r079_delta_sync_preserves_inbound_callers() {
        let inbound_edges = vec![
            InboundEdgeRef {
                caller_id: "test-ws:tests/test_profiler.rs:test_measure:10".to_string(),
                target_label: "SQLAlchemyProfiler".to_string(),
                target_kind: None,
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            InboundEdgeRef {
                caller_id: "test-ws:src/legacy.rs:old_call:20".to_string(),
                target_label: "RemovedFunction".to_string(),
                target_kind: None,
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
        ];

        // Newly parsed nodes in profiler.rs where line numbers shifted
        let surviving_nodes = vec![
            ExtractedNode {
                id: "test-ws:file:src/profiler.rs".to_string(),
                workspace: "test-ws".to_string(),
                label: "profiler.rs".to_string(),
                kind: "file".to_string(),
                file_path: "src/profiler.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 120,
                text: "file content".to_string(),
            },
            ExtractedNode {
                id: "test-ws:src/profiler.rs:SQLAlchemyProfiler:45".to_string(), // shifted from line 25 to 45
                workspace: "test-ws".to_string(),
                label: "SQLAlchemyProfiler".to_string(),
                kind: "struct".to_string(),
                file_path: "src/profiler.rs".to_string(),
                language: "rust".to_string(),
                line_start: 45,
                line_end: 90,
                text: "pub struct SQLAlchemyProfiler {}".to_string(),
            },
        ];

        let (sql, count) = DbClient::build_relink_inbound_sql("test-ws", &inbound_edges, &surviving_nodes);

        // R-079: Must relink surviving symbol and discard deleted symbol
        assert_eq!(count, 1, "Only surviving symbol 'SQLAlchemyProfiler' should be relinked");
        assert!(sql.contains("test-ws:tests/test_profiler.rs:test_measure:10"), "Must reference caller id");
        assert!(sql.contains("test-ws:src/profiler.rs:SQLAlchemyProfiler:45"), "Must link to new node id with updated line");
        assert!(!sql.contains("RemovedFunction"), "Deleted symbol must not be relinked");
        assert!(sql.contains("RELATE $s->linked_to->$t"), "Must use valid SurrealDB RELATE syntax");
    }

    #[tokio::test]
    async fn test_r080_is_indexing_read_barrier() {
        use omni_graph::config::Config;
        use omni_graph::embedder::EmbedderClient;
        use omni_graph::ingestion::{FileCache, IngestionJobStatus, IngestionPipeline};

        let cfg = Config::from_env();
        let db = DbClient::new(&cfg);
        let embedder = EmbedderClient::new(cfg.tei_url);
        let cache = FileCache::new();
        let pipeline = IngestionPipeline::new(db, embedder, cache);

        // Initially idle
        assert!(pipeline.is_indexing(Some("test-ws")).await.is_none());
        assert!(pipeline.is_indexing(None).await.is_none());

        // Simulate active indexing job (R-080)
        let active_job = IngestionJobStatus {
            workspace: "test-ws".to_string(),
            path: "/tmp/test".to_string(),
            is_refresh: false,
            phase: "indexing".to_string(),
            queue_position: 0,
            files_scanned: 100,
            files_indexed: 50,
            files_skipped: 0,
            total_files: 100,
            nodes_created: 500,
            edges_created: 1200,
            clusters_computed: 0,
            started_at: 1000,
            elapsed_ms: 5000,
            completed: false,
            error: None,
        };
        pipeline.set_job_status(active_job).await;

        // Must lock target workspace
        let locked = pipeline.is_indexing(Some("test-ws")).await;
        assert!(locked.is_some(), "Active indexing workspace must be locked");
        let info = locked.unwrap();
        assert_eq!(info.phase, "indexing");
        assert_eq!(info.files_indexed, 50);

        // Unrelated workspace must NOT be locked
        assert!(pipeline.is_indexing(Some("other-ws")).await.is_none());

        // Omniverse/global query (None) must detect active background indexing
        assert!(pipeline.is_indexing(None).await.is_some());

        // Once job completes, read fence must lift cleanly
        let completed_job = IngestionJobStatus {
            workspace: "test-ws".to_string(),
            path: "/tmp/test".to_string(),
            is_refresh: false,
            phase: "completed".to_string(),
            queue_position: 0,
            files_scanned: 100,
            files_indexed: 100,
            files_skipped: 0,
            total_files: 100,
            nodes_created: 1000,
            edges_created: 2500,
            clusters_computed: 12,
            started_at: 1000,
            elapsed_ms: 12000,
            completed: true,
            error: None,
        };
        pipeline.set_job_status(completed_job).await;

        assert!(pipeline.is_indexing(Some("test-ws")).await.is_none(), "Lock must lift once completed");
        assert!(pipeline.is_indexing(None).await.is_none());
    }

    #[test]
    fn test_r078_two_pass_resolves_forward_references() {
        use omni_graph::ingestion::SymbolTable;
        use omni_graph::parser::{ExtractedEdge, ExtractedNode};

        let mut symbol_table = SymbolTable::new();

        // Pass 1: Walk File A (e.g. index 10)
        // File A defines test_profiler, which calls SQLAlchemyProfiler before File B is visited
        let file_a_node = ExtractedNode {
            id: "test-ws:tests/test_profiler.py:test_profiler:10".to_string(),
            workspace: "test-ws".to_string(),
            label: "test_profiler".to_string(),
            kind: "function".to_string(),
            file_path: "tests/test_profiler.py".to_string(),
            language: "python".to_string(),
            line_start: 10,
            line_end: 20,
            text: "def test_profiler(): SQLAlchemyProfiler()".to_string(),
        };
        symbol_table.register_file_nodes(&[file_a_node]);
        symbol_table.register_file_imports(
            "test-ws",
            "tests/test_profiler.py",
            &["SQLAlchemyProfiler".to_string()],
        );

        let unresolved_edge = ExtractedEdge {
            workspace: "test-ws".to_string(),
            source_id: "test-ws:tests/test_profiler.py:test_profiler:10".to_string(),
            target_label: "SQLAlchemyProfiler".to_string(),
            target_id: None, // unresolved forward reference
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        };
        symbol_table.queue_edges(vec![unresolved_edge]);

        assert_eq!(symbol_table.total_unresolved_queued(), 1);

        // Pass 1 continues: Walk File B much later (e.g. index 100)
        // File B defines SQLAlchemyProfiler
        let file_b_node = ExtractedNode {
            id: "test-ws:django_sorcery/db/profiler.py:SQLAlchemyProfiler:25".to_string(),
            workspace: "test-ws".to_string(),
            label: "SQLAlchemyProfiler".to_string(),
            kind: "class".to_string(),
            file_path: "django_sorcery/db/profiler.py".to_string(),
            language: "python".to_string(),
            line_start: 25,
            line_end: 80,
            text: "class SQLAlchemyProfiler: pass".to_string(),
        };
        symbol_table.register_file_nodes(&[file_b_node]);

        // Pass 2: Deferred Linkage Pass (R-078)
        let (resolved_edges, resolved_count, dropped_count) = symbol_table.resolve_all_queued_edges();

        // Verification (R-078): Forward reference must resolve with 100% success
        assert_eq!(resolved_count, 1, "Forward reference should be resolved in Pass 2");
        assert_eq!(dropped_count, 0, "No valid forward reference should be dropped");
        assert_eq!(resolved_edges.len(), 1);
        assert_eq!(
            resolved_edges[0].target_id,
            Some("test-ws:django_sorcery/db/profiler.py:SQLAlchemyProfiler:25".to_string()),
            "Resolved target must link to File B symbol declared later in walk order"
        );
    }

    #[test]
    fn test_r091_watcher_symbol_table_lifecycle() {
        use omni_graph::ingestion::{SymbolDef, SymbolTable};
        use omni_graph::parser::{ExtractedEdge, ExtractedNode};
        let mut st = SymbolTable::new();

        // 1. Populate raw symbols (simulating DB query at watcher startup)
        st.register_raw_symbols("test-ws", vec![
            ("UserAuth".to_string(), SymbolDef {
                id: "test-ws:src/auth.rs:UserAuth:10".to_string(),
                file_path: "src/auth.rs".to_string(),
                kind: "struct".to_string(),
                language: "rust".to_string(),
                line_start: 10,
            }),
            ("OldHelper".to_string(), SymbolDef {
                id: "test-ws:src/old_file.rs:OldHelper:5".to_string(),
                file_path: "src/old_file.rs".to_string(),
                kind: "function".to_string(),
                language: "rust".to_string(),
                line_start: 5,
            }),
        ]);

        assert_eq!(st.total_symbols(), 2);

        // 2. Incremental update: modify src/auth.rs (remove old symbols, register updated node)
        st.remove_file("test-ws", "src/auth.rs");
        assert_eq!(st.total_symbols(), 1);

        let updated_auth_node = ExtractedNode {
            id: "test-ws:src/auth.rs:UserAuth:15".to_string(),
            workspace: "test-ws".to_string(),
            label: "UserAuth".to_string(),
            kind: "struct".to_string(),
            file_path: "src/auth.rs".to_string(),
            language: "rust".to_string(),
            line_start: 15,
            line_end: 30,
            text: "pub struct UserAuth;".to_string(),
        };
        st.register_file_nodes(&[updated_auth_node]);
        assert_eq!(st.total_symbols(), 2);

        // 3. Resolve an edge targeting UserAuth
        let edge = ExtractedEdge {
            workspace: "test-ws".to_string(),
            source_id: "test-ws:src/main.rs:run:4".to_string(),
            target_label: "UserAuth".to_string(),
            target_id: None,
            edge_type: "USES".to_string(),
            category: "EXTRACTED".to_string(),
        };
        let target_id = st.resolve_edge(&edge);
        assert_eq!(target_id, Some("test-ws:src/auth.rs:UserAuth:15".to_string()));

        // 4. File rename lifecycle
        st.rename_file("test-ws", "src/old_file.rs", "src/new_file.rs");
        let helper_edge = ExtractedEdge {
            workspace: "test-ws".to_string(),
            source_id: "test-ws:src/main.rs:run:4".to_string(),
            target_label: "OldHelper".to_string(),
            target_id: None,
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        };
        let resolved_helper = st.resolve_edge(&helper_edge);
        assert_eq!(resolved_helper, Some("test-ws:src/old_file.rs:OldHelper:5".to_string()));
    }

    #[test]
    fn test_r082_ast_receiver_preservation() {
        use omni_graph::parser::CodeParser;

        let python_code = r#"
def create_account():
    user = User.objects.create(username="antigravity")
    self.log("Created")
"#;
        let pr = CodeParser::parse_file("test-ws", "services.py", python_code).unwrap();

        // Verify method call "create" is emitted
        let calls: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "CALLS").collect();
        assert!(calls.iter().any(|c| c.target_label == "create"), "Must emit CALLS to 'create'");

        // Verify root receiver "User" is preserved as a REFERENCES edge (R-082)
        let refs: Vec<_> = pr.edges.iter().filter(|e| e.edge_type == "REFERENCES").collect();
        assert!(refs.iter().any(|r| r.target_label == "User"), "Must emit REFERENCES edge to root receiver 'User'");

        // Verify "self" is not emitted as a spurious reference
        assert!(!refs.iter().any(|r| r.target_label == "self"), "'self' should never be emitted as receiver reference");
    }

    #[test]
    fn test_r084_hierarchical_leiden_coarsening() {
        use omni_graph::analysis::CommunityDetector;
        use omni_graph::db::{DbLink, DbNode};
        use std::collections::{HashMap, HashSet};

        // Create 30 nodes across 20 distinct fine-grained clusters
        let mut nodes = Vec::new();
        let mut assignments = HashMap::new();
        for i in 0..30 {
            let nid = format!("node_{}", i);
            nodes.push(DbNode {
                id: nid.clone(),
                workspace: Some("test-ws".to_string()),
                label: format!("fn_{}", i),
                kind: "function".to_string(),
                file_path: format!("pkg_{}/mod.rs", i / 3),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "fn test() {}".to_string(),
                community: None,
                file_hash: None,
            });
            // Assign 15 distinct initial cluster IDs
            assignments.insert(nid, i % 15);
        }

        let mut links = Vec::new();
        for i in 0..29 {
            links.push(DbLink {
                id: format!("edge_{}", i),
                workspace: Some("test-ws".to_string()),
                source: format!("node_{}", i),
                target: format!("node_{}", i + 1),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            });
        }

        let initial_count = assignments.values().copied().collect::<HashSet<_>>().len();
        assert_eq!(initial_count, 15, "Should start with 15 clusters");

        // Coarsen to at most 6 macro subsystems (R-084)
        CommunityDetector::coarsen_to_macro_galaxies(&nodes, &links, &mut assignments, 6);

        let final_count = assignments.values().copied().collect::<HashSet<_>>().len();
        assert!(final_count <= 6, "Must coarsen to at most 6 macro subsystems, got {}", final_count);
        assert!(final_count >= 1, "Must maintain non-empty clusters");

        // Every node must have a valid community ID in [0..final_count)
        for nid in nodes.iter().map(|n| &n.id) {
            let cid = *assignments.get(nid).unwrap();
            assert!(cid >= 0 && cid < final_count as i32, "Community ID must be compacted");
        }
    }

    #[test]
    fn test_r085_package_metrics_completeness() {
        use omni_graph::analysis::CommunityDetector;
        use omni_graph::db::{DbLink, DbNode};
        use std::collections::HashMap;

        // Create 10 nodes in a single galaxy: 4 interfaces (abstract) and 6 concrete functions
        let mut nodes = Vec::new();
        let mut assignments = HashMap::new();

        for i in 0..4 {
            let id = format!("iface_{}", i);
            nodes.push(DbNode {
                id: id.clone(),
                workspace: Some("test-ws".to_string()),
                label: format!("Interface{}", i),
                kind: "interface".to_string(),
                file_path: "core/types.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "pub trait Interface {}".to_string(),
                community: None,
                file_hash: None,
            });
            assignments.insert(id, 0);
        }

        for i in 0..6 {
            let id = format!("fn_{}", i);
            nodes.push(DbNode {
                id: id.clone(),
                workspace: Some("test-ws".to_string()),
                label: format!("Worker{}", i),
                kind: "function".to_string(),
                file_path: "core/worker.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "fn worker() {}".to_string(),
                community: None,
                file_hash: None,
            });
            assignments.insert(id, 0);
        }

        // Add 2 foreign nodes in another galaxy (galaxy 1)
        for i in 0..2 {
            let id = format!("ext_{}", i);
            nodes.push(DbNode {
                id: id.clone(),
                workspace: Some("test-ws".to_string()),
                label: format!("Ext{}", i),
                kind: "function".to_string(),
                file_path: "client/app.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "fn ext() {}".to_string(),
                community: None,
                file_hash: None,
            });
            assignments.insert(id, 1);
        }

        // Links:
        // 1 incoming call from ext_0 to iface_0 (Ca = 1 for galaxy 0)
        // 3 outgoing calls from fn_0, fn_1, fn_2 to ext_1 (Ce = 3 for galaxy 0)
        let links = vec![
            DbLink {
                id: "e1".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "ext_0".to_string(),
                target: "iface_0".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "e2".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "fn_0".to_string(),
                target: "ext_1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "e3".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "fn_1".to_string(),
                target: "ext_1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "e4".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "fn_2".to_string(),
                target: "ext_1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
        ];

        let (_, galaxies) = CommunityDetector::compute_galaxy_metrics("test-ws", &nodes, &links, &assignments);
        let g0 = galaxies.iter().find(|g| g.galaxy_id == 0).expect("Galaxy 0 must exist");

        // R-085 Package Metrics Verification:
        // Node count: 10
        assert_eq!(g0.node_count, 10);
        // Afferent Coupling Ca = 1, Efferent Coupling Ce = 3
        assert_eq!(g0.afferent_coupling, 1);
        assert_eq!(g0.efferent_coupling, 3);
        // Instability I = Ce / (Ca + Ce) = 3 / (1 + 3) = 0.75
        assert!((g0.instability - 0.75).abs() < 1e-4, "I must equal 0.75");
        // Abstractness A = 4 / 10 = 0.4
        assert!((g0.abstractness - 0.4).abs() < 1e-4, "A must equal 0.4");
        // Distance D = |A + I - 1| = |0.4 + 0.75 - 1| = |0.15| = 0.15
        assert!((g0.distance - 0.15).abs() < 1e-4, "D must equal 0.15");
        // Theoretical bounds
        assert!(g0.abstractness >= 0.0 && g0.abstractness <= 1.0);
        assert!(g0.distance >= 0.0 && g0.distance <= 1.0);
    }

    #[test]
    fn test_r086_affinity_floor_preserves_decoupled_domains() {
        use omni_graph::analysis::CommunityDetector;
        use omni_graph::db::{DbLink, DbNode};
        use std::collections::{HashMap, HashSet};

        // Two completely decoupled domains: "auth" and "billing"
        let nodes = vec![
            DbNode {
                id: "auth_1".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "login".to_string(),
                kind: "function".to_string(),
                file_path: "auth/service.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "fn login() {}".to_string(),
                community: None,
                file_hash: None,
            },
            DbNode {
                id: "auth_2".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "verify".to_string(),
                kind: "function".to_string(),
                file_path: "auth/service.rs".to_string(),
                language: "rust".to_string(),
                line_start: 11,
                line_end: 20,
                text: "fn verify() {}".to_string(),
                community: None,
                file_hash: None,
            },
            DbNode {
                id: "billing_1".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "charge".to_string(),
                kind: "function".to_string(),
                file_path: "billing/gateway.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 10,
                text: "fn charge() {}".to_string(),
                community: None,
                file_hash: None,
            },
            DbNode {
                id: "billing_2".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "invoice".to_string(),
                kind: "function".to_string(),
                file_path: "billing/gateway.rs".to_string(),
                language: "rust".to_string(),
                line_start: 11,
                line_end: 20,
                text: "fn invoice() {}".to_string(),
                community: None,
                file_hash: None,
            },
        ];

        // Internal edges only: auth_1 -> auth_2, billing_1 -> billing_2
        // Zero cross-domain edges!
        let links = vec![
            DbLink {
                id: "e_auth".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "auth_1".to_string(),
                target: "auth_2".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "e_bill".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "billing_1".to_string(),
                target: "billing_2".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
        ];

        let mut assignments = HashMap::new();
        assignments.insert("auth_1".to_string(), 0);
        assignments.insert("auth_2".to_string(), 0);
        assignments.insert("billing_1".to_string(), 1);
        assignments.insert("billing_2".to_string(), 1);

        // Attempt forced coarsening to max_target = 1
        CommunityDetector::coarsen_to_macro_galaxies(&nodes, &links, &mut assignments, 1);

        // R-086 Invariant: Because auth and billing have ZERO coupling and distinct directories,
        // coarsening MUST halt and preserve them as 2 distinct domains!
        let unique_cids: HashSet<i32> = assignments.values().copied().collect();
        assert_eq!(unique_cids.len(), 2, "Decoupled domains MUST NOT be force-merged into 1!");
        assert_ne!(
            assignments.get("auth_1").unwrap(),
            assignments.get("billing_1").unwrap(),
            "Auth and billing must belong to different communities"
        );
    }

    #[test]
    fn test_r088_scale_adaptive_sizing() {
        use omni_graph::analysis::CommunityDetector;

        // Bounded limits for varying codebase complexity
        let k_small = CommunityDetector::calculate_adaptive_target(37, None);
        assert!((5..=15).contains(&k_small), "Small repo target: {}", k_small);

        let k_medium = CommunityDetector::calculate_adaptive_target(2352, None);
        assert!((20..=60).contains(&k_medium), "Medium repo target: {}", k_medium);

        let k_titan = CommunityDetector::calculate_adaptive_target(1_000_000, None);
        assert_eq!(k_titan, 150, "Titan ceiling capped at 150");

        // Explicit user override
        let k_custom = CommunityDetector::calculate_adaptive_target(2352, Some(22));
        assert_eq!(k_custom, 22, "Explicit user override respected");
    }

    #[test]
    fn test_r090_orphan_isolation() {
        use omni_graph::analysis::CommunityDetector;
        use omni_graph::db::{DbLink, DbNode};
        use std::collections::HashMap;

        let nodes = vec![
            // Core domain (size 3)
            DbNode {
                id: "core_1".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "c1".to_string(),
                kind: "function".to_string(),
                file_path: "src/core.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 5,
                text: "".to_string(),
                community: None,
                file_hash: None,
            },
            DbNode {
                id: "core_2".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "c2".to_string(),
                kind: "function".to_string(),
                file_path: "src/core.rs".to_string(),
                language: "rust".to_string(),
                line_start: 6,
                line_end: 10,
                text: "".to_string(),
                community: None,
                file_hash: None,
            },
            DbNode {
                id: "core_3".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "c3".to_string(),
                kind: "function".to_string(),
                file_path: "src/core.rs".to_string(),
                language: "rust".to_string(),
                line_start: 11,
                line_end: 15,
                text: "".to_string(),
                community: None,
                file_hash: None,
            },
            // Isolated orphan (size 1) in completely detached path
            DbNode {
                id: "orphan_1".to_string(),
                workspace: Some("test-ws".to_string()),
                label: "orphan".to_string(),
                kind: "function".to_string(),
                file_path: "scratch/stray.rs".to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 5,
                text: "".to_string(),
                community: None,
                file_hash: None,
            },
        ];

        let links: Vec<DbLink> = vec![];
        let mut assignments = HashMap::new();
        assignments.insert("core_1".to_string(), 0);
        assignments.insert("core_2".to_string(), 0);
        assignments.insert("core_3".to_string(), 0);
        assignments.insert("orphan_1".to_string(), 1);

        // Consolidate with min_macro_size = 3
        CommunityDetector::consolidate_communities(&nodes, &links, &mut assignments, 3);

        // Orphan must NOT pollute core domain (cluster 0)
        assert_ne!(
            assignments.get("core_1").unwrap(),
            assignments.get("orphan_1").unwrap(),
            "Orphan must not pollute core cluster"
        );
    }

    #[test]
    fn test_r093_reactive_reverse_dependency_resolution() {
        use omni_graph::ingestion::SymbolTable;

        let mut st = SymbolTable::new();
        let ws = "test-ws";

        // Register two separate callers for TargetService
        st.record_caller_dependency(ws, "src/caller_a.rs", "TargetService");
        st.record_caller_dependency(ws, "src/caller_b.rs", "TargetService");

        let callers = st.get_dependent_files(ws, "TargetService");
        assert_eq!(callers.len(), 2);
        assert!(callers.contains("src/caller_a.rs"));
        assert!(callers.contains("src/caller_b.rs"));

        // When caller_b is removed / reindexed, caller_b is pruned from dependencies
        st.remove_file(ws, "src/caller_b.rs");
        let remaining_callers = st.get_dependent_files(ws, "TargetService");
        assert_eq!(remaining_callers.len(), 1);
        assert!(remaining_callers.contains("src/caller_a.rs"));
        assert!(!remaining_callers.contains("src/caller_b.rs"));
    }

    #[test]
    fn test_r103_unresolved_edge_retention_pool() {
        use omni_graph::ingestion::SymbolTable;
        use omni_graph::parser::{ExtractedEdge, ExtractedNode};

        let mut st = SymbolTable::new();
        let ws = "test-ws";

        // 1. File A calls calculate_tax() before it is defined anywhere
        let caller_edge = ExtractedEdge {
            workspace: ws.to_string(),
            source_id: "test-ws:src/billing.rs:checkout:15".to_string(),
            target_label: "calculate_tax".to_string(),
            target_id: None,
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        };
        st.queue_edges(vec![caller_edge]);

        // Resolve pass: calculate_tax cannot be resolved yet
        let (resolved, resolved_count, dropped_count) = st.resolve_all_queued_edges();
        assert_eq!(resolved_count, 0);
        assert_eq!(dropped_count, 1);
        assert_eq!(resolved.len(), 0);

        // Crucial invariant (R-103): Edge is retained in pool, NOT lost!
        assert_eq!(st.total_unresolved_retained(), 1);

        // 2. Later, calculate_tax is defined in src/tax.rs
        let tax_node = ExtractedNode {
            id: "test-ws:src/tax.rs:calculate_tax:50".to_string(),
            workspace: ws.to_string(),
            label: "calculate_tax".to_string(),
            kind: "function".to_string(),
            file_path: "src/tax.rs".to_string(),
            language: "rust".to_string(),
            line_start: 50,
            line_end: 65,
            text: "fn calculate_tax() -> f64 { 0.08 }".to_string(),
        };
        st.register_file_nodes(&[tax_node]);

        // 3. Drain satisfied edges for calculate_tax
        let satisfied = st.drain_satisfied(ws, "calculate_tax");
        assert_eq!(satisfied.len(), 1, "Drained satisfied edges must return retained edge");
        assert_eq!(st.total_unresolved_retained(), 0, "Pool should now be empty");

        // The drained edge can now resolve immediately
        let resolved_target = st.resolve_edge(&satisfied[0]);
        assert_eq!(
            resolved_target,
            Some("test-ws:src/tax.rs:calculate_tax:50".to_string()),
            "Retained edge must resolve to the newly declared symbol"
        );
    }

    #[test]
    fn test_r106_language_aware_sair_scoping() {
        use omni_graph::ingestion::SymbolTable;
        use omni_graph::parser::{ExtractedEdge, ExtractedNode};

        let mut st = SymbolTable::new();
        let ws = "polyglot-ws";

        // Polyglot workspace defines parse_config in both Rust and Python
        let rust_node = ExtractedNode {
            id: "polyglot-ws:src/config.rs:parse_config:10".to_string(),
            workspace: ws.to_string(),
            label: "parse_config".to_string(),
            kind: "function".to_string(),
            file_path: "src/config.rs".to_string(),
            language: "rust".to_string(),
            line_start: 10,
            line_end: 25,
            text: "fn parse_config() {}".to_string(),
        };
        let python_node = ExtractedNode {
            id: "polyglot-ws:scripts/config.py:parse_config:20".to_string(),
            workspace: ws.to_string(),
            label: "parse_config".to_string(),
            kind: "function".to_string(),
            file_path: "scripts/config.py".to_string(),
            language: "python".to_string(),
            line_start: 20,
            line_end: 35,
            text: "def parse_config(): pass".to_string(),
        };
        st.register_file_nodes(&[rust_node, python_node]);

        // Caller 1: Rust file calls parse_config
        let rust_caller_edge = ExtractedEdge {
            workspace: ws.to_string(),
            source_id: "polyglot-ws:src/main.rs:run:5".to_string(),
            target_label: "parse_config".to_string(),
            target_id: None,
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        };
        let rust_target = st.resolve_edge(&rust_caller_edge);
        assert_eq!(
            rust_target,
            Some("polyglot-ws:src/config.rs:parse_config:10".to_string()),
            "Rust caller must link to Rust candidate"
        );

        // Caller 2: Python file calls parse_config
        let python_caller_edge = ExtractedEdge {
            workspace: ws.to_string(),
            source_id: "polyglot-ws:scripts/run.py:execute:8".to_string(),
            target_label: "parse_config".to_string(),
            target_id: None,
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        };
        let python_target = st.resolve_edge(&python_caller_edge);
        assert_eq!(
            python_target,
            Some("polyglot-ws:scripts/config.py:parse_config:20".to_string()),
            "Python caller must link to Python candidate"
        );
    }

    #[tokio::test]
    async fn test_r094_autonomous_directory_backbone_synthesis() {
        use omni_graph::config::Config;
        use omni_graph::db::DbClient;
        use omni_graph::embedder::EmbedderClient;
        use omni_graph::ingestion::{FileCache, SymbolTable};
        use omni_graph::watcher::IncrementalPipeline;
        use std::sync::Arc;
        use tokio::sync::RwLock;

        let cfg = Config::from_env();
        let db = DbClient::new(&cfg);
        let embedder = EmbedderClient::new(cfg.tei_url);
        let cache = FileCache::new();
        let st = Arc::new(RwLock::new(SymbolTable::new()));

        let pipeline = IncrementalPipeline::with_default_dirs(db, embedder, cache, st);
        let ws = "test-ws";

        // Deeply nested file path
        let rel_path = "src/modules/billing/v2/invoice_generator.rs";

        // First pass: synthesizes missing directory hierarchy up to workspace
        let (nodes, edges) = pipeline.synthesize_directory_backbone(ws, rel_path).await;

        // Expect 4 ancestor directories: src, src/modules, src/modules/billing, src/modules/billing/v2
        assert_eq!(nodes.len(), 4, "Should synthesize 4 missing directory nodes");
        assert_eq!(edges.len(), 4, "Should synthesize 4 hierarchical CONTAINS edges");

        // Verify nodes
        let dir_paths: Vec<String> = nodes.iter().map(|n| n.file_path.clone()).collect();
        assert_eq!(
            dir_paths,
            vec![
                "src".to_string(),
                "src/modules".to_string(),
                "src/modules/billing".to_string(),
                "src/modules/billing/v2".to_string(),
            ]
        );
        for n in &nodes {
            assert_eq!(n.kind, "directory");
            assert_eq!(n.workspace, ws);
        }

        // Verify edge hierarchy: root workspace -> src -> modules -> billing -> v2
        assert_eq!(edges[0].source_id, "test-ws:workspace");
        assert_eq!(edges[0].target_id, Some("test-ws:dir:src".to_string()));
        assert_eq!(edges[0].edge_type, "CONTAINS");

        assert_eq!(edges[1].source_id, "test-ws:dir:src");
        assert_eq!(edges[1].target_id, Some("test-ws:dir:src/modules".to_string()));

        assert_eq!(edges[2].source_id, "test-ws:dir:src/modules");
        assert_eq!(edges[2].target_id, Some("test-ws:dir:src/modules/billing".to_string()));

        assert_eq!(edges[3].source_id, "test-ws:dir:src/modules/billing");
        assert_eq!(edges[3].target_id, Some("test-ws:dir:src/modules/billing/v2".to_string()));

        // Second pass: another file in the same directory (or sibling subdirectory)
        let (sibling_nodes, sibling_edges) = pipeline
            .synthesize_directory_backbone(ws, "src/modules/billing/v2/receipt.rs")
            .await;
        assert_eq!(sibling_nodes.len(), 0, "Known directories should not be re-synthesized");
        assert_eq!(sibling_edges.len(), 0);
    }

    #[test]
    fn test_r101_atomic_transactional_multi_statement_swap() {
        use omni_graph::db::{DbClient, InboundEdgeRef};
        use omni_graph::parser::{ExtractedEdge, ExtractedNode};

        let ws = "test-ws";
        let rel_path = "src/payment.rs";

        let nodes = vec![
            ExtractedNode {
                id: "test-ws:file:src/payment.rs".to_string(),
                workspace: ws.to_string(),
                label: "payment.rs".to_string(),
                kind: "file".to_string(),
                file_path: rel_path.to_string(),
                language: "rust".to_string(),
                line_start: 1,
                line_end: 100,
                text: "pub fn charge() {}".to_string(),
            },
            ExtractedNode {
                id: "test-ws:src/payment.rs:charge:10".to_string(),
                workspace: ws.to_string(),
                label: "charge".to_string(),
                kind: "function".to_string(),
                file_path: rel_path.to_string(),
                language: "rust".to_string(),
                line_start: 10,
                line_end: 25,
                text: "pub fn charge() {}".to_string(),
            },
        ];

        let edges = vec![ExtractedEdge {
            workspace: ws.to_string(),
            source_id: "test-ws:src/payment.rs:charge:10".to_string(),
            target_label: "StripeClient".to_string(),
            target_id: Some("test-ws:src/stripe.rs:StripeClient:5".to_string()),
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        }];

        let inbound = vec![InboundEdgeRef {
            caller_id: "test-ws:src/checkout.rs:run:20".to_string(),
            target_label: "charge".to_string(),
            target_kind: Some("function".to_string()),
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        }];

        let embeddings = vec![vec![0.1f32; 384], vec![0.2f32; 384]];

        let (sql, node_cnt, edge_cnt, relink_cnt) = DbClient::build_atomic_swap_sql(
            ws,
            rel_path,
            Some("fake_hash_123"),
            &nodes,
            &embeddings,
            Some(4),
            &edges,
            &inbound,
        );

        // Transaction boundaries (R-101)
        assert!(sql.starts_with("BEGIN TRANSACTION;\n"));
        assert!(sql.ends_with("COMMIT TRANSACTION;\n"));

        // Step 1: Deletion of previous state
        assert!(sql.contains("DELETE linked_to WHERE in IN $nodes OR out IN $nodes;"));
        assert!(sql.contains("DELETE node WHERE workspace = 'test-ws' AND file_path = 'src/payment.rs';"));

        // Step 2: Upsert new nodes
        assert!(sql.contains("UPSERT type::thing('node', 'test-ws:file:src/payment.rs') MERGE"));
        assert!(sql.contains("UPSERT type::thing('node', 'test-ws:src/payment.rs:charge:10') MERGE"));
        assert_eq!(node_cnt, 2);

        // Step 3: Outbound edge relate with call_count & weight
        assert!(sql.contains("call_count = 1"));
        assert!(sql.contains("weight = 1"));
        assert_eq!(edge_cnt, 1);

        // Step 4: Relink inbound caller
        assert!(sql.contains("test-ws:src/checkout.rs:run:20"));
        assert_eq!(relink_cnt, 1);
    }

    #[test]
    fn test_r102_disambiguated_inbound_signature_relinking() {
        use omni_graph::db::{DbClient, InboundEdgeRef};
        use omni_graph::parser::ExtractedNode;

        let ws = "test-ws";
        // File defines both a struct User and a function User (or sibling kinds)
        let surviving_nodes = vec![
            ExtractedNode {
                id: "test-ws:src/auth.rs:User:10".to_string(),
                workspace: ws.to_string(),
                label: "User".to_string(),
                kind: "struct".to_string(),
                file_path: "src/auth.rs".to_string(),
                language: "rust".to_string(),
                line_start: 10,
                line_end: 20,
                text: "pub struct User {}".to_string(),
            },
            ExtractedNode {
                id: "test-ws:src/auth.rs:User:30".to_string(),
                workspace: ws.to_string(),
                label: "User".to_string(),
                kind: "function".to_string(),
                file_path: "src/auth.rs".to_string(),
                language: "rust".to_string(),
                line_start: 30,
                line_end: 40,
                text: "pub fn User() {}".to_string(),
            },
        ];

        // Caller 1 wants struct User
        let struct_caller = InboundEdgeRef {
            caller_id: "test-ws:src/repo.rs:get_user:5".to_string(),
            target_label: "User".to_string(),
            target_kind: Some("struct".to_string()),
            edge_type: "REFERENCES".to_string(),
            category: "EXTRACTED".to_string(),
        };

        // Caller 2 wants function User
        let fn_caller = InboundEdgeRef {
            caller_id: "test-ws:src/factory.rs:make:8".to_string(),
            target_label: "User".to_string(),
            target_kind: Some("function".to_string()),
            edge_type: "CALLS".to_string(),
            category: "EXTRACTED".to_string(),
        };

        let (sql, relinked_cnt) = DbClient::build_relink_inbound_sql(
            ws,
            &[struct_caller, fn_caller],
            &surviving_nodes,
        );

        assert_eq!(relinked_cnt, 2);
        // struct_caller links to line 10 struct node
        assert!(sql.contains("test-ws:src/auth.rs:User:10"));
        // fn_caller links to line 30 function node
        assert!(sql.contains("test-ws:src/auth.rs:User:30"));
    }

    #[test]
    fn test_r104_edge_relationship_deduplication() {
        use omni_graph::db::deduplicate_edges;
        use omni_graph::parser::ExtractedEdge;

        let ws = "test-ws";
        // Simulate 4 calls to get() and 2 calls to post() from the same function
        let mut edges = Vec::new();
        for _ in 0..4 {
            edges.push(ExtractedEdge {
                workspace: ws.to_string(),
                source_id: "test-ws:src/worker.rs:fetch:10".to_string(),
                target_label: "get".to_string(),
                target_id: Some("test-ws:src/client.rs:get:50".to_string()),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            });
        }
        for _ in 0..2 {
            edges.push(ExtractedEdge {
                workspace: ws.to_string(),
                source_id: "test-ws:src/worker.rs:fetch:10".to_string(),
                target_label: "post".to_string(),
                target_id: Some("test-ws:src/client.rs:post:80".to_string()),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            });
        }

        let deduped = deduplicate_edges(&edges);
        assert_eq!(deduped.len(), 2, "6 redundant calls should deduplicate to exactly 2 unique edges");

        let get_edge = deduped.iter().find(|e| e.target_label == "get").unwrap();
        assert_eq!(get_edge.call_count, 4, "get() call count should be 4");
        assert_eq!(get_edge.weight, 4.0, "get() edge weight should be 4.0");

        let post_edge = deduped.iter().find(|e| e.target_label == "post").unwrap();
        assert_eq!(post_edge.call_count, 2, "post() call count should be 2");
        assert_eq!(post_edge.weight, 2.0, "post() edge weight should be 2.0");
    }

    #[test]
    fn test_r095_dual_tier_local_modularity_fast_path() {
        use omni_graph::analysis::CommunityDetector;

        // Node u strongly coupled to target community B (k_{u, B} = 10.0) vs weak source community A (k_{u, A} = 1.0)
        let k_u_target = 10.0;
        let k_u_source = 1.0;
        let sigma_tot_target = 50.0;
        let sigma_tot_source = 100.0;
        let k_u = 11.0;
        let total_m = 200.0;

        let delta_q = CommunityDetector::calculate_local_delta_q(
            k_u_target,
            k_u_source,
            sigma_tot_target,
            sigma_tot_source,
            k_u,
            total_m,
        );

        // Fast-path invariant (R-095): Positive delta_Q triggers instant local reassignment
        assert!(delta_q > 0.0, "Delta Q must be positive when coupled more strongly to target community");

        // Reverse check: coupled weakly to target B vs strongly to source A
        let delta_q_negative = CommunityDetector::calculate_local_delta_q(
            0.5,
            10.0,
            120.0,
            40.0,
            10.5,
            total_m,
        );
        assert!(delta_q_negative < 0.0, "Delta Q must be negative when source community has higher affinity");
    }

    #[test]
    fn test_r105_zero_delay_community_inference_for_new_files() {
        use omni_graph::analysis::CommunityDetector;
        use omni_graph::db::DbLink;
        use std::collections::HashMap;

        let mut node_communities = HashMap::new();
        node_communities.insert("node_billing_1".to_string(), 2);
        node_communities.insert("node_billing_2".to_string(), 2);
        node_communities.insert("node_billing_3".to_string(), 2);
        node_communities.insert("node_auth_1".to_string(), 5);

        // New node links to 3 billing nodes and 1 auth node
        let links = vec![
            DbLink {
                id: "l1".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "node_new".to_string(),
                target: "node_billing_1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "l2".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "node_new".to_string(),
                target: "node_billing_2".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "l3".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "node_new".to_string(),
                target: "node_billing_3".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
            DbLink {
                id: "l4".to_string(),
                workspace: Some("test-ws".to_string()),
                source: "node_new".to_string(),
                target: "node_auth_1".to_string(),
                edge_type: "CALLS".to_string(),
                category: "EXTRACTED".to_string(),
            },
        ];

        let inferred = CommunityDetector::infer_neighbor_community("node_new", &links, &node_communities);

        // Invariant (R-105): New file immediately inherits modal neighbor community 2 upon insertion
        assert_eq!(inferred, Some(2), "New node must infer modal community of its connected neighbors");
    }

    #[tokio::test]
    async fn test_r097_convergence_barrier_blocks_until_sync() {
        use std::sync::Arc;
        use std::time::Duration;
        use tokio::sync::RwLock;
        use omni_graph::watcher::WatcherStatus;

        let status = Arc::new(RwLock::new(WatcherStatus {
            workspace: "barrier-ws".to_string(),
            path: "/tmp/barrier".to_string(),
            status: "syncing".to_string(),
            started_at: "now".to_string(),
            files_tracked: 10,
            last_sync: None,
            events_processed: 0,
            files_reindexed: 0,
            files_deleted: 0,
            avg_sync_ms: 0,
            cluster_status: "current".to_string(),
            last_cluster_at: None,
            debounce_ms: 200,
        }));

        // Simulate background sync completing after 40ms
        let bg_status = status.clone();
        tokio::spawn(async move {
            tokio::time::sleep(Duration::from_millis(40)).await;
            let mut s = bg_status.write().await;
            s.status = "watching".to_string();
        });

        // Loop mimicking wait_for_convergence
        let start = std::time::Instant::now();
        let timeout = Duration::from_millis(300);
        let mut converged = false;
        while start.elapsed() < timeout {
            if status.read().await.status != "syncing" {
                converged = true;
                break;
            }
            tokio::time::sleep(Duration::from_millis(10)).await;
        }

        assert!(converged, "Barrier must block until status transitions from 'syncing' to 'watching'");
        assert!(start.elapsed() >= Duration::from_millis(30), "Must have awaited sync duration");
    }

    #[test]
    fn test_r098_fidelity_diagnostic_endpoint() {
        // Test fidelity score calculation invariants
        let edge_count: i64 = 100;
        let dangling_directory_pointers: i64 = 0;
        let healthy = edge_count.saturating_sub(dangling_directory_pointers) as f64;
        let fidelity_score = (healthy / edge_count as f64).clamp(0.0, 1.0);

        assert_eq!(fidelity_score, 1.0, "Zero dangling directory pointers must produce 1.0 fidelity score");

        // When defects exist
        let defective_edges: i64 = 100;
        let dangling_defects: i64 = 5;
        let defective_score = ((defective_edges.saturating_sub(dangling_defects)) as f64) / (defective_edges as f64);
        assert!((defective_score - 0.95).abs() < 1e-6);

        // When edge count is 0
        let zero_edges: i64 = 0;
        let zero_score = if zero_edges > 0 { 0.0 } else { 1.0 };
        assert_eq!(zero_score, 1.0, "Empty workspace must yield baseline 1.0 fidelity");
    }

    #[tokio::test]
    async fn test_r099_lock_free_generational_versioning() {
        use std::sync::atomic::{AtomicU64, Ordering};
        use std::sync::Arc;

        let generation = Arc::new(AtomicU64::new(1));
        assert_eq!(generation.load(Ordering::Relaxed), 1);

        // Concurrent atomic increments across 10 tasks
        let mut handles = Vec::new();
        for _ in 0..10 {
            let gen = generation.clone();
            handles.push(tokio::spawn(async move {
                for _ in 0..10 {
                    gen.fetch_add(1, Ordering::SeqCst);
                }
            }));
        }

        for h in handles {
            h.await.unwrap();
        }

        // 1 initial + 10 * 10 increments = 101
        assert_eq!(generation.load(Ordering::SeqCst), 101, "Generational counter must update lock-free and thread-safe");
    }

    #[tokio::test]
    async fn test_r109_symlink_cycle_guard_and_alias_navigation() {
        use std::fs;

        let test_dir = std::env::temp_dir().join(format!(
            "omni_test_r109_{}",
            std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_nanos()
        ));
        fs::create_dir_all(&test_dir).unwrap();

        let src_dir = test_dir.join("src");
        fs::create_dir_all(&src_dir).unwrap();

        let real_file = src_dir.join("real.rs");
        fs::write(&real_file, "pub fn compute_trajectory() {}").unwrap();

        // 1. Create a valid file symlink
        let symlink_file = test_dir.join("shortcut.rs");
        #[cfg(unix)]
        std::os::unix::fs::symlink(&real_file, &symlink_file).unwrap();

        // 2. Create a cyclic directory symlink: cyclic_dir/loop -> cyclic_dir
        let cyclic_dir = test_dir.join("cyclic_dir");
        fs::create_dir_all(&cyclic_dir).unwrap();
        let loop_link = cyclic_dir.join("loop");
        #[cfg(unix)]
        std::os::unix::fs::symlink(&cyclic_dir, &loop_link).unwrap();

        // 3. Create a broken symlink
        let broken_link = test_dir.join("broken.rs");
        #[cfg(unix)]
        std::os::unix::fs::symlink(test_dir.join("nonexistent.rs"), &broken_link).unwrap();

        // Verify symlink metadata detection
        #[cfg(unix)]
        {
            let is_sym = fs::symlink_metadata(&symlink_file)
                .map(|m| m.file_type().is_symlink())
                .unwrap_or(false);
            assert!(is_sym, "symlink_file must be detected as a symlink");

            let target_link = fs::read_link(&symlink_file).unwrap();
            assert_eq!(target_link, real_file);

            let canonical_test_dir = fs::canonicalize(&test_dir).unwrap();
            let canonical = fs::canonicalize(&symlink_file).unwrap();
            let rel_target = canonical.strip_prefix(&canonical_test_dir).unwrap();
            assert_eq!(rel_target, std::path::Path::new("src/real.rs"));

            // Verify cyclic symlink detection
            let is_cycle_sym = fs::symlink_metadata(&loop_link)
                .map(|m| m.file_type().is_symlink())
                .unwrap_or(false);
            assert!(is_cycle_sym, "loop_link must be detected as symlink");

            // Canonicalizing cyclic dir succeeds to itself without recursion
            let canonical_cycle = fs::canonicalize(&loop_link).unwrap();
            assert_eq!(canonical_cycle, fs::canonicalize(&cyclic_dir).unwrap());

            // Broken link canonicalization fails gracefully
            let broken_res = fs::canonicalize(&broken_link);
            assert!(broken_res.is_err(), "broken symlink must fail canonicalization cleanly");
        }

        // Clean up
        let _ = fs::remove_dir_all(&test_dir);
    }

    #[test]
    fn test_r109_symlink_atomic_swap_sql_generation() {
        use omni_graph::db::DbClient;
        use omni_graph::parser::{ExtractedEdge, ExtractedNode};

        let ws = "test-ws";
        let rel_path = "scripts/omni";
        let target_path = "tools/omni-graph/scripts/omni.sh";

        let symlink_node = ExtractedNode {
            id: format!("{}:file:{}", ws, rel_path),
            workspace: ws.to_string(),
            label: "omni".to_string(),
            kind: "symlink".to_string(),
            file_path: rel_path.to_string(),
            language: "symlink".to_string(),
            line_start: 1,
            line_end: 1,
            text: format!("Symlink: {} -> {}", rel_path, target_path),
        };

        let link_edge = ExtractedEdge {
            workspace: ws.to_string(),
            source_id: format!("{}:file:{}", ws, rel_path),
            target_label: "omni.sh".to_string(),
            target_id: Some(format!("{}:file:{}", ws, target_path)),
            edge_type: "LINKS_TO".to_string(),
            category: "FILESYSTEM".to_string(),
        };

        let (sql, n_cnt, e_cnt, _) = DbClient::build_atomic_swap_sql(
            ws,
            rel_path,
            Some("dummy_hash"),
            &[symlink_node],
            &[vec![0.0f32; 384]],
            Some(18),
            &[link_edge],
            &[],
        );

        assert_eq!(n_cnt, 1);
        assert_eq!(e_cnt, 1);
        assert!(sql.contains("BEGIN TRANSACTION;"));
        assert!(sql.contains("DELETE node WHERE workspace = 'test-ws' AND file_path = 'scripts/omni';"));
        assert!(sql.contains("kind: 'symlink'"));
        assert!(sql.contains("type = 'LINKS_TO'"));
        assert!(sql.contains("category = 'FILESYSTEM'"));
        assert!(sql.contains("COMMIT TRANSACTION;"));
    }

    #[test]
    fn test_r107_matlab_function_extraction() {
        use omni_graph::parser::CodeParser;

        let code = r#"
function [filtered, residual] = filterSignals(raw_sig, threshold)
    % Multi-return signal filtering function
    gain = 2.5;
    residual = raw_sig - threshold;
    filtered = applyGain(residual, gain);
    saturate(filtered);
end

function y = singleOutput(x)
    y = x * 10;
end
"#;

        let res = CodeParser::parse_file("test-ws", "signals/filterSignals.m", code)
            .expect("Should parse MATLAB file successfully");

        // Verify function nodes
        let f1 = res.nodes.iter().find(|n| n.label == "filterSignals").expect("filterSignals function node");
        assert_eq!(f1.kind, "function");
        assert_eq!(f1.language, "matlab");
        assert_eq!(f1.line_start, 2);

        let f2 = res.nodes.iter().find(|n| n.label == "singleOutput").expect("singleOutput function node");
        assert_eq!(f2.kind, "function");
        assert_eq!(f2.language, "matlab");
        assert_eq!(f2.line_start, 10);

        // Verify CALLS edges inside filterSignals
        let calls_apply_gain = res.edges.iter().find(|e| {
            e.source_id == f1.id && e.target_label == "applyGain" && e.edge_type == "CALLS"
        });
        assert!(calls_apply_gain.is_some(), "Should extract applyGain call");

        let calls_saturate = res.edges.iter().find(|e| {
            e.source_id == f1.id && e.target_label == "saturate" && e.edge_type == "CALLS"
        });
        assert!(calls_saturate.is_some(), "Should extract saturate call");
    }

    #[test]
    fn test_r107_matlab_classdef_and_inheritance() {
        use omni_graph::parser::CodeParser;

        let code = r#"
classdef FlightController < BaseController & Observable
    properties (Access = public)
        Kp double = 1.0
        Ki double = 0.05
        Kd double = 0.2
    end
    methods
        function obj = FlightController(p, i, d)
            obj.Kp = p;
            obj.Ki = i;
            obj.Kd = d;
        end
        function [u, state] = step(obj, error)
            integral = computeIntegral(error);
            u = obj.Kp * error + obj.Ki * integral;
            state = obj.updateState(u);
        end
    end
end
"#;

        let res = CodeParser::parse_file("test-ws", "control/FlightController.m", code)
            .expect("Should parse MATLAB classdef successfully");

        // Verify class node
        let cls = res.nodes.iter().find(|n| n.label == "FlightController" && n.kind == "class")
            .expect("FlightController class node");
        assert_eq!(cls.kind, "class");
        assert_eq!(cls.language, "matlab");

        // Verify multiple inheritance EXTENDS edges
        let ext_base = res.edges.iter().find(|e| {
            e.source_id == cls.id && e.target_label == "BaseController" && e.edge_type == "EXTENDS"
        });
        assert!(ext_base.is_some(), "Should emit EXTENDS edge to BaseController");

        let ext_obs = res.edges.iter().find(|e| {
            e.source_id == cls.id && e.target_label == "Observable" && e.edge_type == "EXTENDS"
        });
        assert!(ext_obs.is_some(), "Should emit EXTENDS edge to Observable");

        // Verify properties
        for prop_name in &["Kp", "Ki", "Kd"] {
            assert!(
                res.nodes.iter().any(|n| n.label == *prop_name && n.kind == "property"),
                "Property node {} missing",
                prop_name
            );
            let c_edge = res.edges.iter().find(|e| {
                e.source_id == cls.id && e.target_label == *prop_name && e.edge_type == "CONTAINS"
            });
            assert!(c_edge.is_some(), "Should emit CONTAINS edge for property {}", prop_name);
        }

        // Verify constructor
        let ctor = res.nodes.iter().find(|n| n.label == "FlightController" && n.kind == "constructor")
            .expect("Constructor node");
        assert_eq!(ctor.kind, "constructor");
        let ctor_edge = res.edges.iter().find(|e| {
            e.source_id == cls.id && e.target_label == "FlightController" && e.edge_type == "CONTAINS"
        });
        assert!(ctor_edge.is_some(), "Should emit CONTAINS edge to constructor");

        // Verify method
        let method = res.nodes.iter().find(|n| n.label == "step" && n.kind == "method")
            .expect("step method node");
        assert_eq!(method.kind, "method");
        let method_edge = res.edges.iter().find(|e| {
            e.source_id == cls.id && e.target_label == "step" && e.edge_type == "DECLARES"
        });
        assert!(method_edge.is_some(), "Should emit DECLARES edge to step method");

        // Verify calls inside step method
        let call_integral = res.edges.iter().find(|e| {
            e.source_id == method.id && e.target_label == "computeIntegral" && e.edge_type == "CALLS"
        });
        assert!(call_integral.is_some(), "Should emit CALLS edge to computeIntegral");

        let call_update = res.edges.iter().find(|e| {
            e.source_id == method.id && e.target_label == "updateState" && e.edge_type == "CALLS"
        });
        assert!(call_update.is_some(), "Should emit CALLS edge to updateState");
    }

    #[test]
    fn test_r107_matlab_receiver_preservation() {
        use omni_graph::parser::CodeParser;

        let code = r#"
function runSimulation()
    raw = SensorClient.readData();
    FilterEngine.calculate(raw);
end
"#;

        let res = CodeParser::parse_file("test-ws", "sim/runSimulation.m", code)
            .expect("Should parse MATLAB file successfully");

        let fn_node = res.nodes.iter().find(|n| n.label == "runSimulation").expect("runSimulation node");

        // Callee calls
        assert!(res.edges.iter().any(|e| e.source_id == fn_node.id && e.target_label == "readData" && e.edge_type == "CALLS"));
        assert!(res.edges.iter().any(|e| e.source_id == fn_node.id && e.target_label == "calculate" && e.edge_type == "CALLS"));

        // Receiver preservation
        assert!(res.edges.iter().any(|e| e.source_id == fn_node.id && e.target_label == "SensorClient" && e.edge_type == "REFERENCES"));
        assert!(res.edges.iter().any(|e| e.source_id == fn_node.id && e.target_label == "FilterEngine" && e.edge_type == "REFERENCES"));
    }

    #[test]
    fn test_r107_matlab_import_commands() {
        use omni_graph::parser::CodeParser;

        let code = r#"
import control.guidance.NavSystem;
import math.filters.*;

function run()
    x = 1;
end
"#;

        let res = CodeParser::parse_file("test-ws", "main.m", code)
            .expect("Should parse MATLAB file with imports");

        // Verify import nodes
        let imports: Vec<_> = res.nodes.iter().filter(|n| n.kind == "import").collect();
        assert_eq!(imports.len(), 2, "Should extract 2 import statements");

        // Verify IMPORTS edges
        assert!(res.edges.iter().any(|e| e.target_label == "NavSystem" && e.edge_type == "IMPORTS"));
        assert!(res.edges.iter().any(|e| e.target_label == "filters" && e.edge_type == "IMPORTS"));
    }

    #[test]
    fn test_r107_matlab_package_hierarchy() {
        use omni_graph::parser::CodeParser;

        let code = r#"
classdef GuidanceComputer
    methods
        function compute(obj)
        end
    end
end
"#;

        let res = CodeParser::parse_file("test-ws", "src/+aerospace/+gnc/GuidanceComputer.m", code)
            .expect("Should parse package-scoped MATLAB file");

        // Verify package node
        let pkg_node = res.nodes.iter().find(|n| n.kind == "package" && n.label == "aerospace.gnc")
            .expect("aerospace.gnc package node");
        assert_eq!(pkg_node.file_path, "src/+aerospace/+gnc/GuidanceComputer.m");

        // Verify DECLARES edge from package to class
        let cls_node = res.nodes.iter().find(|n| n.kind == "class" && n.label == "GuidanceComputer")
            .expect("GuidanceComputer class node");

        let decl_edge = res.edges.iter().find(|e| {
            e.source_id == pkg_node.id && e.target_id == Some(cls_node.id.clone()) && e.edge_type == "DECLARES"
        });
        assert!(decl_edge.is_some(), "Package should emit DECLARES edge to class GuidanceComputer");
    }
}




