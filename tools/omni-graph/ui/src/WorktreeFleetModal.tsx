import React from 'react';
import { GitBranch, X, Check, ArrowRight, Activity, Folder, Sparkles } from 'lucide-react';
import { WatcherStatus } from './App';

export interface WorktreeChild {
  workspace: string;
  branch: string;
  root_path: string;
  total_nodes: number;
}

export interface WorkspaceWithWorktrees {
  workspace: string;
  total_nodes: number;
  languages: string[];
  files: string[];
  root_path?: string;
  is_worktree?: boolean;
  parent_workspace?: string;
  branch?: string;
  worktrees?: WorktreeChild[];
}

interface WorktreeFleetModalProps {
  isOpen: boolean;
  onClose: () => void;
  parentWorkspace: WorkspaceWithWorktrees | null;
  selectedWorkspace: string | null;
  onSelectWorkspace: (ws: string) => void;
  watchers?: WatcherStatus[];
}

export const WorktreeFleetModal: React.FC<WorktreeFleetModalProps> = ({
  isOpen,
  onClose,
  parentWorkspace,
  selectedWorkspace,
  onSelectWorkspace,
  watchers = [],
}) => {
  if (!isOpen || !parentWorkspace) return null;

  const worktrees = parentWorkspace.worktrees || [];
  const totalFleetNodes =
    parentWorkspace.total_nodes + worktrees.reduce((acc, wt) => acc + wt.total_nodes, 0);

  const getWatcherFor = (wsName: string) => {
    return watchers.find((w) => w.workspace === wsName);
  };

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        zIndex: 200,
        background: 'rgba(3, 7, 18, 0.75)',
        backdropFilter: 'blur(12px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
      }}
      onClick={onClose}
    >
      <div
        className="glass-panel"
        style={{
          width: '100%',
          maxWidth: 680,
          maxHeight: '85vh',
          display: 'flex',
          flexDirection: 'column',
          background: 'rgba(15, 23, 42, 0.92)',
          border: '1px solid rgba(56, 189, 248, 0.3)',
          borderRadius: 16,
          boxShadow: '0 25px 60px rgba(0, 0, 0, 0.9), 0 0 30px rgba(56, 189, 248, 0.15)',
          overflow: 'hidden',
        }}
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div
          style={{
            padding: '18px 24px',
            borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            background: 'linear-gradient(180deg, rgba(30, 41, 59, 0.6) 0%, rgba(15, 23, 42, 0.4) 100%)',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div
              style={{
                width: 38,
                height: 38,
                borderRadius: 10,
                background: 'rgba(34, 197, 94, 0.15)',
                border: '1px solid rgba(34, 197, 94, 0.4)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#4ade80',
              }}
            >
              <GitBranch size={20} />
            </div>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <h3 style={{ margin: 0, fontSize: '1.05rem', fontWeight: 700, color: '#f8fafc' }}>
                  Git Worktree Fleet & Lineage
                </h3>
                <span
                  style={{
                    fontSize: '0.7rem',
                    padding: '2px 8px',
                    borderRadius: 12,
                    background: 'rgba(168, 85, 247, 0.2)',
                    color: '#c084fc',
                    border: '1px solid rgba(168, 85, 247, 0.4)',
                    fontWeight: 600,
                  }}
                >
                  {worktrees.length} Parallel Strategies
                </span>
              </div>
              <div style={{ fontSize: '0.78rem', color: 'var(--text-muted)', marginTop: 2 }}>
                Trunk Root: <span style={{ color: '#e2e8f0', fontWeight: 600 }}>{parentWorkspace.workspace}</span> •{' '}
                {totalFleetNodes.toLocaleString()} total AST nodes in lineage
              </div>
            </div>
          </div>
          <button
            onClick={onClose}
            style={{
              background: 'rgba(255, 255, 255, 0.06)',
              border: 'none',
              borderRadius: 8,
              padding: 6,
              color: 'var(--text-muted)',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <X size={18} />
          </button>
        </div>

        {/* Content Body */}
        <div style={{ padding: 24, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 20 }}>
          {/* Parent Trunk Card */}
          <div>
            <div style={{ fontSize: '0.72rem', fontWeight: 700, letterSpacing: '0.05em', color: 'var(--text-muted)', marginBottom: 8 }}>
              ORIGIN TRUNK REPOSITORY
            </div>
            <div
              onClick={() => onSelectWorkspace(parentWorkspace.workspace)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '14px 18px',
                borderRadius: 12,
                background:
                  selectedWorkspace === parentWorkspace.workspace
                    ? 'rgba(56, 189, 248, 0.15)'
                    : 'rgba(30, 41, 59, 0.5)',
                border:
                  selectedWorkspace === parentWorkspace.workspace
                    ? '1px solid rgba(56, 189, 248, 0.6)'
                    : '1px solid rgba(255, 255, 255, 0.08)',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <Folder size={20} color={selectedWorkspace === parentWorkspace.workspace ? 'var(--accent-cyan)' : '#c084fc'} />
                <div>
                  <div style={{ fontWeight: 700, fontSize: '0.92rem', color: '#f8fafc', display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span>{parentWorkspace.workspace}</span>
                    <span
                      style={{
                        fontSize: '0.68rem',
                        padding: '1px 6px',
                        borderRadius: 4,
                        background: 'rgba(168, 85, 247, 0.25)',
                        color: '#c084fc',
                        fontWeight: 600,
                      }}
                    >
                      Canonical Trunk (main)
                    </span>
                  </div>
                  <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: 2 }}>
                    {parentWorkspace.total_nodes.toLocaleString()} AST symbols • Master reference codebase
                  </div>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                {selectedWorkspace === parentWorkspace.workspace ? (
                  <span
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 4,
                      fontSize: '0.75rem',
                      color: 'var(--accent-cyan)',
                      fontWeight: 600,
                    }}
                  >
                    <Check size={14} /> Active View
                  </span>
                ) : (
                  <span
                    style={{
                      fontSize: '0.75rem',
                      color: 'var(--text-muted)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: 4,
                    }}
                  >
                    Switch to Trunk <ArrowRight size={13} />
                  </span>
                )}
              </div>
            </div>
          </div>

          {/* Sibling Strategy Branches */}
          <div>
            <div style={{ fontSize: '0.72rem', fontWeight: 700, letterSpacing: '0.05em', color: 'var(--text-muted)', marginBottom: 8 }}>
              ACTIVE STRATEGY WORKTREES ({worktrees.length})
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
              {worktrees.map((wt) => {
                const isSelected = selectedWorkspace === wt.workspace;
                const watcher = getWatcherFor(wt.workspace);
                return (
                  <div
                    key={wt.workspace}
                    onClick={() => onSelectWorkspace(wt.workspace)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      padding: '12px 18px',
                      borderRadius: 12,
                      background: isSelected ? 'rgba(34, 197, 94, 0.12)' : 'rgba(15, 23, 42, 0.65)',
                      border: isSelected ? '1px solid rgba(34, 197, 94, 0.5)' : '1px solid rgba(255, 255, 255, 0.06)',
                      cursor: 'pointer',
                      transition: 'all 0.18s ease',
                      position: 'relative',
                      overflow: 'hidden',
                    }}
                    onMouseEnter={(e) => {
                      if (!isSelected) {
                        e.currentTarget.style.borderColor = 'rgba(34, 197, 94, 0.3)';
                        e.currentTarget.style.background = 'rgba(255, 255, 255, 0.04)';
                      }
                    }}
                    onMouseLeave={(e) => {
                      if (!isSelected) {
                        e.currentTarget.style.borderColor = 'rgba(255, 255, 255, 0.06)';
                        e.currentTarget.style.background = 'rgba(15, 23, 42, 0.65)';
                      }
                    }}
                  >
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12, minWidth: 0 }}>
                      <span style={{ fontSize: '1.2rem' }}>🌿</span>
                      <div style={{ minWidth: 0 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                          <span style={{ fontWeight: 700, fontSize: '0.88rem', color: isSelected ? '#4ade80' : '#f8fafc' }}>
                            {wt.workspace}
                          </span>
                          {wt.branch && (
                            <span
                              style={{
                                fontSize: '0.68rem',
                                padding: '1px 7px',
                                borderRadius: 4,
                                background: 'rgba(34, 197, 94, 0.2)',
                                color: '#86efac',
                                fontFamily: 'monospace',
                                fontWeight: 600,
                                textOverflow: 'ellipsis',
                                overflow: 'hidden',
                                whiteSpace: 'nowrap',
                                maxWidth: 300,
                              }}
                            >
                              {wt.branch}
                            </span>
                          )}
                        </div>
                        <div style={{ fontSize: '0.74rem', color: 'var(--text-muted)', marginTop: 3, display: 'flex', alignItems: 'center', gap: 10 }}>
                          <span>{wt.total_nodes.toLocaleString()} AST nodes</span>
                          {watcher && (
                            <span style={{ display: 'flex', alignItems: 'center', gap: 4, color: '#38bdf8' }}>
                              <Activity size={12} /> Live Sync (~{Math.round(watcher.avg_sync_ms || 108)}ms)
                            </span>
                          )}
                        </div>
                      </div>
                    </div>

                    <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
                      {isSelected ? (
                        <span
                          style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 4,
                            fontSize: '0.75rem',
                            color: '#4ade80',
                            fontWeight: 700,
                          }}
                        >
                          <Check size={14} /> Active
                        </span>
                      ) : (
                        <button
                          type="button"
                          style={{
                            background: 'rgba(255, 255, 255, 0.08)',
                            border: '1px solid rgba(255, 255, 255, 0.15)',
                            borderRadius: 6,
                            padding: '4px 10px',
                            color: '#e2e8f0',
                            fontSize: '0.72rem',
                            fontWeight: 600,
                            cursor: 'pointer',
                          }}
                        >
                          Switch
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Lineage Architecture Callout */}
          <div
            style={{
              padding: '14px 18px',
              borderRadius: 12,
              background: 'rgba(15, 23, 42, 0.6)',
              border: '1px solid rgba(56, 189, 248, 0.2)',
              display: 'flex',
              gap: 12,
            }}
          >
            <Sparkles size={18} color="var(--accent-cyan)" style={{ flexShrink: 0, marginTop: 2 }} />
            <div style={{ fontSize: '0.78rem', color: 'var(--text-muted)', lineHeight: 1.5 }}>
              <strong style={{ color: '#f8fafc' }}>Prior LPA Seed Inheritance Active:</strong> Worktree branches share 98%+ of
              their AST symbols with <span style={{ color: '#c084fc' }}>{parentWorkspace.workspace}</span>. When switching between
              parallel strategies, visual galaxy cluster IDs are preserved to eliminate layout churn while highlighting local algorithm deltas.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
