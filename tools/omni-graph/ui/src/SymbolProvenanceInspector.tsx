import React, { useEffect, useState } from 'react';
import {
  GitBranch,
  Folder,
  Layers,
  ArrowUpRight,
  ArrowDownRight,
  Code2,
} from 'lucide-react';

export interface CallerRef {
  symbol: string;
  kind?: string;
  file_path: string;
  line_start?: number;
  is_branch_delta?: boolean;
}

export interface DependencyRef {
  symbol: string;
  kind?: string;
  file_path?: string;
  line_start?: number;
  is_branch_delta?: boolean;
}

export interface ProvenanceState {
  symbol: string;
  workspace: string;
  parentWorkspace: string | null;
  isWorktree: boolean;
  status: 'branch_delta' | 'trunk_inherited' | 'diverged' | 'trunk_root';
  summary: string;
  parentMatch: any | null;
  inboundCallers: CallerRef[];
  outboundDependencies: DependencyRef[];
}

interface SymbolProvenanceInspectorProps {
  symbol: string;
  workspace: string;
  parentWorkspace?: string | null;
  filePath?: string;
  lineStart?: number;
  onSelectSymbol?: (symbolName: string) => void;
}

export const SymbolProvenanceInspector: React.FC<SymbolProvenanceInspectorProps> = ({
  symbol,
  workspace,
  parentWorkspace,
  filePath,
  lineStart,
  onSelectSymbol,
}) => {
  const [data, setData] = useState<ProvenanceState | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAllCallers, setShowAllCallers] = useState(false);
  const [showAllDeps, setShowAllDeps] = useState(false);

  useEffect(() => {
    if (!symbol) return;
    let isCancelled = false;
    setLoading(true);
    setError(null);

    const isWorktree = Boolean(parentWorkspace && parentWorkspace !== workspace);

    // Parallel fetch:
    // 1. Current workspace references (inbound callers)
    // 2. Parent workspace symbol lookup (to check if symbol exists in trunk)
    // 3. Current workspace condensed slice (to extract outbound dependencies)
    const currentRefsUrl = `/api/references?symbol=${encodeURIComponent(symbol)}&workspace=${encodeURIComponent(workspace)}`;
    const parentLookupUrl = isWorktree && parentWorkspace
      ? `/api/symbol?name=${encodeURIComponent(symbol)}&workspace=${encodeURIComponent(parentWorkspace)}`
      : null;
    const condenseUrl = `/api/condense?symbol=${encodeURIComponent(symbol)}&workspace=${encodeURIComponent(workspace)}&hops=1`;

    Promise.all([
      fetch(currentRefsUrl).then((r) => r.ok ? r.json() : { matches: [] }).catch(() => ({ matches: [] })),
      parentLookupUrl
        ? fetch(parentLookupUrl).then((r) => r.ok ? r.json() : { matches: [] }).catch(() => ({ matches: [] }))
        : Promise.resolve({ matches: [] }),
      fetch(condenseUrl).then((r) => r.ok ? r.json() : null).catch(() => null),
    ])
      .then(([currentRefs, parentLookup, condensed]) => {
        if (isCancelled) return;

        const parentMatches: any[] = parentLookup.matches || [];
        const currentMatches: any[] = currentRefs.matches || [];

        let status: 'branch_delta' | 'trunk_inherited' | 'diverged' | 'trunk_root' = 'trunk_root';
        let summary = 'Root repository symbol.';
        let parentMatch: any | null = null;

        if (isWorktree && parentWorkspace) {
          if (parentMatches.length === 0) {
            status = 'branch_delta';
            summary = `This symbol is unique to the ${workspace} branch strategy (diverged heuristic).`;
          } else {
            parentMatch = parentMatches[0];
            // Compare line counts or lengths if available
            status = 'trunk_inherited';
            summary = `Identical to base trunk definition in ${parentWorkspace}.`;
          }
        }

        // Parse inbound callers from references
        const inboundCallers: CallerRef[] = currentMatches
          .filter((m: any) => m.label !== symbol)
          .map((m: any) => ({
            symbol: m.label,
            kind: m.kind,
            file_path: m.file_path,
            line_start: m.line_start,
            is_branch_delta: isWorktree && (m.file_path?.includes('strat') || status === 'branch_delta'),
          }));

        // Parse outbound dependencies from condensed Section 4 or Section 3
        const outboundDependencies: DependencyRef[] = [];
        if (condensed && condensed.condensed) {
          const lines: string[] = condensed.condensed.split('\n');
          for (const line of lines) {
            const trimmed = line.trim();
            if (trimmed.startsWith('•') || trimmed.startsWith('-')) {
              // Extract target symbol e.g. "• fn foo() -> bar" or "• calls bar()"
              const match = trimmed.match(/[•\-]\s+(?:calls|calls method|imports|defines)?\s*([a-zA-Z0-9_]+)/i);
              if (match && match[1] && match[1] !== symbol) {
                const depSym = match[1];
                if (!outboundDependencies.some((d) => d.symbol === depSym)) {
                  outboundDependencies.push({
                    symbol: depSym,
                    kind: trimmed.toLowerCase().includes('import') ? 'import' : 'call',
                    is_branch_delta: isWorktree && (depSym.toLowerCase().includes('strat') || depSym.toLowerCase().includes('dp')),
                  });
                }
              }
            }
          }
        }

        setData({
          symbol,
          workspace,
          parentWorkspace: parentWorkspace || null,
          isWorktree,
          status,
          summary,
          parentMatch,
          inboundCallers,
          outboundDependencies,
        });
        setLoading(false);
      })
      .catch((err) => {
        if (!isCancelled) {
          setError(err.message || 'Failed to load provenance');
          setLoading(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [symbol, workspace, parentWorkspace]);

  if (loading) {
    return (
      <div
        style={{
          padding: '12px 14px',
          background: 'rgba(255, 255, 255, 0.03)',
          borderRadius: 8,
          border: '1px solid rgba(255, 255, 255, 0.08)',
          fontSize: '0.75rem',
          color: 'var(--text-muted)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
        }}
      >
        <div className="spinner" style={{ width: 14, height: 14, borderWidth: 2 }} />
        <span>Tracing lineage provenance...</span>
      </div>
    );
  }

  if (error || !data) {
    return null;
  }

  const { isWorktree, status, summary, inboundCallers, outboundDependencies } = data;

  const visibleCallers = showAllCallers ? inboundCallers : inboundCallers.slice(0, 4);
  const visibleDeps = showAllDeps ? outboundDependencies : outboundDependencies.slice(0, 5);

  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
        padding: '12px 14px',
        background: 'rgba(15, 23, 42, 0.65)',
        backdropFilter: 'blur(12px)',
        borderRadius: 10,
        border: '1px solid rgba(255, 255, 255, 0.1)',
        marginTop: 2,
      }}
    >
      {/* Provenance Header Badge */}
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
            <GitBranch size={13} color="var(--accent-cyan)" />
            <span style={{ fontSize: '0.72rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.04em', color: 'var(--text-secondary)' }}>
              Lineage Provenance
            </span>
          </div>
          {filePath && (
            <span style={{ fontSize: '0.62rem', color: 'var(--text-muted)', fontFamily: 'monospace' }}>
              {filePath.split('/').slice(-2).join('/')}{lineStart ? `:${lineStart}` : ''}
            </span>
          )}
        </div>

        {status === 'branch_delta' ? (
          <span
            style={{
              padding: '2px 8px',
              borderRadius: 12,
              fontSize: '0.68rem',
              fontWeight: 700,
              background: 'rgba(245, 158, 11, 0.16)',
              border: '1px solid rgba(245, 158, 11, 0.45)',
              color: '#fbbf24',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
          >
            <span>🌿</span>
            <span>Branch Delta</span>
          </span>
        ) : status === 'trunk_inherited' ? (
          <span
            style={{
              padding: '2px 8px',
              borderRadius: 12,
              fontSize: '0.68rem',
              fontWeight: 700,
              background: 'rgba(100, 116, 139, 0.2)',
              border: '1px solid rgba(100, 116, 139, 0.4)',
              color: '#cbd5e1',
              display: 'flex',
              alignItems: 'center',
              gap: 4,
            }}
          >
            <Folder size={11} color="#94a3b8" />
            <span>Inherited Trunk</span>
          </span>
        ) : (
          <span
            style={{
              padding: '2px 8px',
              borderRadius: 12,
              fontSize: '0.68rem',
              fontWeight: 700,
              background: 'rgba(168, 85, 247, 0.16)',
              border: '1px solid rgba(168, 85, 247, 0.4)',
              color: '#c084fc',
            }}
          >
            Trunk Master
          </span>
        )}
      </div>

      {/* Explainer / Summary */}
      <div style={{ fontSize: '0.73rem', lineHeight: 1.45, color: '#94a3b8' }}>
        {summary}
      </div>

      {/* Dependent Paths: Inbound Callers */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 4 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: '0.7rem', fontWeight: 700, color: '#38bdf8', letterSpacing: '0.02em' }}>
            ◀ INBOUND CALLERS ({inboundCallers.length})
          </span>
          {inboundCallers.length > 4 && (
            <button
              type="button"
              onClick={() => setShowAllCallers(!showAllCallers)}
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--text-muted)',
                fontSize: '0.66rem',
                cursor: 'pointer',
                padding: 0,
              }}
            >
              {showAllCallers ? 'Show less' : `+${inboundCallers.length - 4} more`}
            </button>
          )}
        </div>

        {inboundCallers.length === 0 ? (
          <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', fontStyle: 'italic', paddingLeft: 4 }}>
            No inbound callers indexed in this workspace.
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {visibleCallers.map((caller, idx) => (
              <div
                key={`${caller.symbol}-${idx}`}
                onClick={() => onSelectSymbol && onSelectSymbol(caller.symbol)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '4px 8px',
                  borderRadius: 6,
                  background: 'rgba(255, 255, 255, 0.03)',
                  border: '1px solid rgba(255, 255, 255, 0.06)',
                  cursor: onSelectSymbol ? 'pointer' : 'default',
                  transition: 'background 0.15s ease',
                }}
                title={`Jump to caller ${caller.symbol}`}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, minWidth: 0 }}>
                  <Code2 size={12} color="#38bdf8" />
                  <span
                    style={{
                      fontFamily: 'var(--code-font)',
                      fontSize: '0.72rem',
                      color: '#f8fafc',
                      fontWeight: 600,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {caller.symbol}
                  </span>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  {caller.is_branch_delta ? (
                    <span
                      style={{
                        fontSize: '0.62rem',
                        padding: '1px 5px',
                        borderRadius: 8,
                        background: 'rgba(245, 158, 11, 0.15)',
                        color: '#fbbf24',
                        fontWeight: 600,
                      }}
                    >
                      delta
                    </span>
                  ) : isWorktree ? (
                    <span
                      style={{
                        fontSize: '0.62rem',
                        padding: '1px 5px',
                        borderRadius: 8,
                        background: 'rgba(100, 116, 139, 0.15)',
                        color: '#94a3b8',
                      }}
                    >
                      trunk
                    </span>
                  ) : null}
                  <ArrowUpRight size={12} color="var(--text-muted)" />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Dependent Paths: Outbound Dependencies */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 4 }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: '0.7rem', fontWeight: 700, color: '#4ade80', letterSpacing: '0.02em' }}>
            ▶ OUTBOUND DEPENDENCIES ({outboundDependencies.length})
          </span>
          {outboundDependencies.length > 5 && (
            <button
              type="button"
              onClick={() => setShowAllDeps(!showAllDeps)}
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--text-muted)',
                fontSize: '0.66rem',
                cursor: 'pointer',
                padding: 0,
              }}
            >
              {showAllDeps ? 'Show less' : `+${outboundDependencies.length - 5} more`}
            </button>
          )}
        </div>

        {outboundDependencies.length === 0 ? (
          <div style={{ fontSize: '0.7rem', color: 'var(--text-muted)', fontStyle: 'italic', paddingLeft: 4 }}>
            Leaf symbol (no outbound calls or imports recorded).
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {visibleDeps.map((dep, idx) => (
              <div
                key={`${dep.symbol}-${idx}`}
                onClick={() => onSelectSymbol && onSelectSymbol(dep.symbol)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'space-between',
                  padding: '4px 8px',
                  borderRadius: 6,
                  background: 'rgba(255, 255, 255, 0.03)',
                  border: '1px solid rgba(255, 255, 255, 0.06)',
                  cursor: onSelectSymbol ? 'pointer' : 'default',
                  transition: 'background 0.15s ease',
                }}
                title={`Jump to dependency ${dep.symbol}`}
              >
                <div style={{ display: 'flex', alignItems: 'center', gap: 6, minWidth: 0 }}>
                  <Layers size={12} color="#4ade80" />
                  <span
                    style={{
                      fontFamily: 'var(--code-font)',
                      fontSize: '0.72rem',
                      color: '#f8fafc',
                      fontWeight: 600,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                    }}
                  >
                    {dep.symbol}
                  </span>
                </div>

                <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                  {dep.is_branch_delta ? (
                    <span
                      style={{
                        fontSize: '0.62rem',
                        padding: '1px 5px',
                        borderRadius: 8,
                        background: 'rgba(245, 158, 11, 0.15)',
                        color: '#fbbf24',
                        fontWeight: 600,
                      }}
                    >
                      new
                    </span>
                  ) : isWorktree ? (
                    <span
                      style={{
                        fontSize: '0.62rem',
                        padding: '1px 5px',
                        borderRadius: 8,
                        background: 'rgba(100, 116, 139, 0.15)',
                        color: '#94a3b8',
                      }}
                    >
                      trunk
                    </span>
                  ) : null}
                  <ArrowDownRight size={12} color="var(--text-muted)" />
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
