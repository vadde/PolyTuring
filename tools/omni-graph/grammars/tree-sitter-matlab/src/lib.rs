use tree_sitter::Language;

extern "C" {
    fn tree_sitter_matlab() -> *const ();
}

/// Returns the tree-sitter Language for MATLAB
pub fn language() -> Language {
    unsafe { Language::from_raw(tree_sitter_matlab() as *const _) }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_can_load_grammar() {
        let mut parser = tree_sitter::Parser::new();
        parser
            .set_language(&language())
            .expect("Error loading Matlab parser");
    }

    #[test]
    fn test_parse_sample_matlab() {
        let mut parser = tree_sitter::Parser::new();
        parser.set_language(&language()).unwrap();

        let code = r#"
import control.guidance.NavSystem;
import math.filters.*;

classdef KalmanFilter < control.BaseFilter & math.Observer
    properties (Access = private)
        CovarianceMatrix double
        StateEstimate = [0; 0]
    end
    methods
        function obj = KalmanFilter(initialCov)
            obj.CovarianceMatrix = initialCov;
        end
        function [predState, predCov] = predict(obj, u)
            predState = obj.StateEstimate + u;
            predCov = obj.CovarianceMatrix;
            helperFunction(predState);
            obj.stepMethod(predState);
        end
    end
end

function [out1, out2] = standaloneFunc(arg1, arg2)
    out1 = arg1 * 2;
    out2 = arg2 + 10;
end
"#;

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
        let tree = parser.parse(code, None).unwrap();
        println!("AST:\n{}", tree.root_node().to_sexp());

        let root = tree.root_node();
        let mut cursor = root.walk();
        for child in root.children(&mut cursor) {
            if child.kind() == "function_definition" {
                println!("Top-level function_definition:");
                println!("  by_field('name'): {:?}", child.child_by_field_name("name").map(|n| &code[n.byte_range()]));
                for c in 0..child.child_count() {
                    let sub = child.child(c).unwrap();
                    println!("  child {}: kind='{}', field_name={:?}, text='{}'", 
                        c, sub.kind(), child.field_name_for_child(c as u32), &code[sub.byte_range()]);
                }
            }
        }
    }
}
