import React, { useState, useEffect, useMemo } from 'react';
import {
  ArrowLeft,
  Search,
  Download,
  Play,
  Check,
  Copy,
  Terminal,
  Layers,
  ChevronDown,
  ChevronRight,
  RefreshCw,
  AlertCircle,
  Database,
  FileJson,
  Zap,
  Globe,
} from 'lucide-react';

export interface ApiSpecExplorerProps {
  onBackToGraph: () => void;
  selectedWorkspace?: string;
}

interface EndpointDoc {
  path: string;
  method: 'GET' | 'POST' | 'DELETE' | 'PUT';
  tag: string;
  summary: string;
  description: string;
  parameters?: Array<{
    name: string;
    in: 'query' | 'path' | 'header';
    required?: boolean;
    description?: string;
    schema?: {
      type: string;
      default?: any;
      example?: any;
    };
  }>;
  requestBody?: {
    required?: boolean;
    content?: {
      'application/json'?: {
        schema?: {
          type?: string;
          required?: string[];
          properties?: Record<
            string,
            {
              type?: string;
              description?: string;
              default?: any;
              example?: any;
            }
          >;
        };
      };
    };
  };
}

export const ApiSpecExplorer: React.FC<ApiSpecExplorerProps> = ({ onBackToGraph, selectedWorkspace }) => {
  const [openapiSpec, setOpenapiSpec] = useState<any | null>(null);
  const [isLoadingSpec, setIsLoadingSpec] = useState<boolean>(true);
  const [specError, setSpecError] = useState<string | null>(null);

  // Search & Tag Filter
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [activeTag, setActiveTag] = useState<string>('ALL');

  // Expanded card state
  const [expandedEndpoints, setExpandedEndpoints] = useState<Record<string, boolean>>({});

  // Live Testing Playground State
  const [inputValues, setInputValues] = useState<Record<string, Record<string, string>>>({});
  const [bodyInputs, setBodyInputs] = useState<Record<string, string>>({});
  const [testResults, setTestResults] = useState<
    Record<
      string,
      {
        status: number;
        statusText: string;
        data: any;
        latencyMs: number;
        error?: string;
        isLoading?: boolean;
      }
    >
  >({});

  const [copiedId, setCopiedId] = useState<string | null>(null);

  // 1. Fetch OpenAPI Specification on Mount
  useEffect(() => {
    let isMounted = true;
    const fetchSpec = async () => {
      setIsLoadingSpec(true);
      setSpecError(null);
      try {
        const res = await fetch('/api/openapi.json');
        if (!res.ok) {
          throw new Error(`Failed to load OpenAPI spec: HTTP ${res.status}`);
        }
        const json = await res.json();
        if (isMounted) {
          setOpenapiSpec(json);
        }
      } catch (err: any) {
        if (isMounted) {
          setSpecError(err.message || 'Unable to connect to OpenAPI endpoint');
        }
      } finally {
        if (isMounted) {
          setIsLoadingSpec(false);
        }
      }
    };

    fetchSpec();
    return () => {
      isMounted = false;
    };
  }, []);

  // 2. Parse paths into structured list
  const endpointsList = useMemo<EndpointDoc[]>(() => {
    if (!openapiSpec || !openapiSpec.paths) return [];

    const list: EndpointDoc[] = [];
    Object.entries(openapiSpec.paths).forEach(([pathKey, pathItem]: [string, any]) => {
      ['get', 'post', 'delete', 'put'].forEach((methodKey) => {
        if (pathItem[methodKey]) {
          const op = pathItem[methodKey];
          const tag = (op.tags && op.tags[0]) || 'General';
          list.push({
            path: pathKey,
            method: methodKey.toUpperCase() as any,
            tag,
            summary: op.summary || '',
            description: op.description || '',
            parameters: op.parameters,
            requestBody: op.requestBody,
          });
        }
      });
    });

    return list;
  }, [openapiSpec]);

  // Extract available tags
  const tagsList = useMemo<string[]>(() => {
    const set = new Set<string>();
    endpointsList.forEach((ep) => set.add(ep.tag));
    return ['ALL', ...Array.from(set)];
  }, [endpointsList]);

  // Filtered endpoints
  const filteredEndpoints = useMemo(() => {
    const q = searchQuery.toLowerCase().trim();
    return endpointsList.filter((ep) => {
      const matchTag = activeTag === 'ALL' || ep.tag.toLowerCase() === activeTag.toLowerCase();
      if (!matchTag) return false;
      if (!q) return true;
      return (
        ep.path.toLowerCase().includes(q) ||
        ep.summary.toLowerCase().includes(q) ||
        ep.description.toLowerCase().includes(q) ||
        ep.method.toLowerCase().includes(q) ||
        ep.tag.toLowerCase().includes(q)
      );
    });
  }, [endpointsList, activeTag, searchQuery]);

  const toggleEndpoint = (key: string) => {
    setExpandedEndpoints((prev) => ({
      ...prev,
      [key]: !prev[key],
    }));
  };

  const handleInputChange = (endpointKey: string, paramName: string, value: string) => {
    setInputValues((prev) => ({
      ...prev,
      [endpointKey]: {
        ...(prev[endpointKey] || {}),
        [paramName]: value,
      },
    }));
  };

  const handleBodyChange = (endpointKey: string, value: string) => {
    setBodyInputs((prev) => ({
      ...prev,
      [endpointKey]: value,
    }));
  };

  // 3. Execute "Try It Live"
  const handleExecuteLive = async (ep: EndpointDoc) => {
    const key = `${ep.method}-${ep.path}`;
    const params = inputValues[key] || {};
    let url = `/api${ep.path}`;

    // Substitute path params
    Object.entries(params).forEach(([pName, pVal]) => {
      if (url.includes(`{${pName}}`)) {
        url = url.replace(`{${pName}}`, encodeURIComponent(pVal));
      }
    });

    // Append query params
    const queryParams = new URLSearchParams();
    if (ep.parameters) {
      ep.parameters.forEach((p) => {
        if (p.in === 'query') {
          const val = params[p.name] !== undefined ? params[p.name] : (p.name === 'workspace' && selectedWorkspace ? selectedWorkspace : '');
          if (val) {
            queryParams.append(p.name, val);
          }
        }
      });
    }
    const queryString = queryParams.toString();
    if (queryString) {
      url += (url.includes('?') ? '&' : '?') + queryString;
    }

    setTestResults((prev) => ({
      ...prev,
      [key]: {
        status: 0,
        statusText: 'Executing...',
        data: null,
        latencyMs: 0,
        isLoading: true,
      },
    }));

    const startTime = performance.now();

    try {
      const options: RequestInit = {
        method: ep.method,
        headers: {},
      };

      if (ep.method === 'POST' || ep.method === 'PUT') {
        (options.headers as any)['Content-Type'] = 'application/json';
        const bodyContent = bodyInputs[key];
        if (bodyContent && bodyContent.trim()) {
          options.body = bodyContent;
        } else if (ep.requestBody) {
          // Fallback minimal JSON
          options.body = JSON.stringify({});
        }
      }

      const res = await fetch(url, options);
      const latencyMs = Math.round(performance.now() - startTime);

      let parsedData: any = null;
      const contentType = res.headers.get('content-type') || '';
      if (contentType.includes('application/json')) {
        parsedData = await res.json();
      } else {
        parsedData = await res.text();
      }

      setTestResults((prev) => ({
        ...prev,
        [key]: {
          status: res.status,
          statusText: res.statusText || `${res.status}`,
          data: parsedData,
          latencyMs,
          isLoading: false,
        },
      }));
    } catch (err: any) {
      const latencyMs = Math.round(performance.now() - startTime);
      setTestResults((prev) => ({
        ...prev,
        [key]: {
          status: 500,
          statusText: 'Network / Client Error',
          data: null,
          latencyMs,
          error: err.message || 'Connection failed',
          isLoading: false,
        },
      }));
    }
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  const generateCurlCommand = (ep: EndpointDoc) => {
    const key = `${ep.method}-${ep.path}`;
    const params = inputValues[key] || {};
    let url = `http://localhost:8080/api${ep.path}`;

    Object.entries(params).forEach(([pName, pVal]) => {
      if (url.includes(`{${pName}}`)) {
        url = url.replace(`{${pName}}`, encodeURIComponent(pVal));
      }
    });

    const queryParams = new URLSearchParams();
    if (ep.parameters) {
      ep.parameters.forEach((p) => {
        if (p.in === 'query') {
          const val = params[p.name] !== undefined ? params[p.name] : (p.name === 'workspace' && selectedWorkspace ? selectedWorkspace : '');
          if (val) {
            queryParams.append(p.name, val);
          }
        }
      });
    }
    const qStr = queryParams.toString();
    if (qStr) {
      url += (url.includes('?') ? '&' : '?') + qStr;
    }

    if (ep.method === 'GET') {
      return `curl -s "${url}"`;
    }

    const bodyVal = bodyInputs[key] || '{}';
    return `curl -s -X ${ep.method} "${url}" \\\n  -H "Content-Type: application/json" \\\n  -d '${bodyVal}'`;
  };

  const downloadSpecJson = () => {
    if (!openapiSpec) return;
    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(JSON.stringify(openapiSpec, null, 2));
    const dlAnchorElem = document.createElement('a');
    dlAnchorElem.setAttribute('href', dataStr);
    dlAnchorElem.setAttribute('download', 'omni-graph-openapi.json');
    dlAnchorElem.click();
  };

  const getMethodBadgeStyle = (method: string) => {
    switch (method) {
      case 'GET':
        return { background: 'rgba(16, 185, 129, 0.15)', color: '#34d399', border: '1px solid rgba(16, 185, 129, 0.35)' };
      case 'POST':
        return { background: 'rgba(168, 85, 247, 0.15)', color: '#c084fc', border: '1px solid rgba(168, 85, 247, 0.35)' };
      case 'DELETE':
        return { background: 'rgba(239, 68, 68, 0.15)', color: '#f87171', border: '1px solid rgba(239, 68, 68, 0.35)' };
      default:
        return { background: 'rgba(56, 189, 248, 0.15)', color: '#38bdf8', border: '1px solid rgba(56, 189, 248, 0.35)' };
    }
  };

  return (
    <div
      style={{
        width: '100vw',
        height: '100vh',
        overflowY: 'auto',
        background: 'linear-gradient(180deg, #050813 0%, #03040a 100%)',
        color: '#f8fafc',
        fontFamily: 'var(--font-family, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif)',
        padding: '24px 32px 64px',
        boxSizing: 'border-box',
      }}
    >
      <div style={{ maxWidth: 1200, margin: '0 auto', display: 'flex', flexDirection: 'column', gap: 24 }}>
        {/* Top Header Bar */}
        <header
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            paddingBottom: 20,
            borderBottom: '1px solid rgba(255, 255, 255, 0.08)',
            flexWrap: 'wrap',
            gap: 16,
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <button
              onClick={onBackToGraph}
              className="cyber-button-secondary"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '8px 14px',
                borderRadius: 8,
                fontSize: '0.82rem',
                cursor: 'pointer',
              }}
            >
              <ArrowLeft size={16} />
              <span>Back to 3D Graph Studio</span>
            </button>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <h1 style={{ margin: 0, fontSize: '1.45rem', fontWeight: 800, letterSpacing: '-0.02em', color: '#f8fafc' }}>
                  Omni-Graph API Explorer
                </h1>
                <span
                  style={{
                    background: 'rgba(56, 189, 248, 0.15)',
                    border: '1px solid rgba(56, 189, 248, 0.35)',
                    color: '#38bdf8',
                    padding: '2px 8px',
                    borderRadius: 6,
                    fontSize: '0.72rem',
                    fontWeight: 700,
                    fontFamily: 'monospace',
                  }}
                >
                  OpenAPI 3.0.3
                </span>
                <span
                  style={{
                    background: 'rgba(168, 85, 247, 0.15)',
                    border: '1px solid rgba(168, 85, 247, 0.35)',
                    color: '#c084fc',
                    padding: '2px 8px',
                    borderRadius: 6,
                    fontSize: '0.72rem',
                    fontWeight: 700,
                  }}
                >
                  v0.3.0 Contract
                </span>
              </div>
              <p style={{ margin: '4px 0 0', fontSize: '0.82rem', color: 'var(--text-muted, #94a3b8)' }}>
                Interactive REST specification, real-time testing playground, and SDK contract for AI coding agents.
              </p>
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <button
              onClick={downloadSpecJson}
              disabled={!openapiSpec}
              className="cyber-button-secondary"
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 7,
                padding: '8px 14px',
                borderRadius: 8,
                fontSize: '0.8rem',
                cursor: 'pointer',
              }}
              title="Download openapi.json"
            >
              <Download size={15} />
              <span>Download Spec</span>
            </button>
          </div>
        </header>

        {/* Live Architecture Overview & Metrics Banner */}
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
            gap: 14,
          }}
        >
          <div
            style={{
              background: 'rgba(15, 23, 42, 0.65)',
              border: '1px solid rgba(255, 255, 255, 0.08)',
              borderRadius: 12,
              padding: '14px 18px',
              backdropFilter: 'blur(16px)',
              display: 'flex',
              alignItems: 'center',
              gap: 14,
            }}
          >
            <div
              style={{
                width: 40,
                height: 40,
                borderRadius: 10,
                background: 'rgba(56, 189, 248, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#38bdf8',
              }}
            >
              <Globe size={20} />
            </div>
            <div>
              <div style={{ fontSize: '0.72rem', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>BASE URL</div>
              <div style={{ fontSize: '0.92rem', fontWeight: 700, color: '#f8fafc', fontFamily: 'monospace' }}>
                http://localhost:8080/api
              </div>
            </div>
          </div>

          <div
            style={{
              background: 'rgba(15, 23, 42, 0.65)',
              border: '1px solid rgba(255, 255, 255, 0.08)',
              borderRadius: 12,
              padding: '14px 18px',
              backdropFilter: 'blur(16px)',
              display: 'flex',
              alignItems: 'center',
              gap: 14,
            }}
          >
            <div
              style={{
                width: 40,
                height: 40,
                borderRadius: 10,
                background: 'rgba(168, 85, 247, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#c084fc',
              }}
            >
              <Layers size={20} />
            </div>
            <div>
              <div style={{ fontSize: '0.72rem', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>TOTAL ENDPOINTS</div>
              <div style={{ fontSize: '0.92rem', fontWeight: 700, color: '#f8fafc' }}>
                {endpointsList.length} Registered Routes
              </div>
            </div>
          </div>

          <div
            style={{
              background: 'rgba(15, 23, 42, 0.65)',
              border: '1px solid rgba(255, 255, 255, 0.08)',
              borderRadius: 12,
              padding: '14px 18px',
              backdropFilter: 'blur(16px)',
              display: 'flex',
              alignItems: 'center',
              gap: 14,
            }}
          >
            <div
              style={{
                width: 40,
                height: 40,
                borderRadius: 10,
                background: 'rgba(52, 211, 153, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#34d399',
              }}
            >
              <Zap size={20} />
            </div>
            <div>
              <div style={{ fontSize: '0.72rem', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>ORCHESTRATION</div>
              <div style={{ fontSize: '0.92rem', fontWeight: 700, color: '#34d399' }}>
                FIFO Queued • Clean Refresh
              </div>
            </div>
          </div>

          <div
            style={{
              background: 'rgba(15, 23, 42, 0.65)',
              border: '1px solid rgba(255, 255, 255, 0.08)',
              borderRadius: 12,
              padding: '14px 18px',
              backdropFilter: 'blur(16px)',
              display: 'flex',
              alignItems: 'center',
              gap: 14,
            }}
          >
            <div
              style={{
                width: 40,
                height: 40,
                borderRadius: 10,
                background: 'rgba(245, 158, 11, 0.15)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: '#fbbf24',
              }}
            >
              <Database size={20} />
            </div>
            <div>
              <div style={{ fontSize: '0.72rem', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>ACTIVE WORKSPACE</div>
              <div style={{ fontSize: '0.92rem', fontWeight: 700, color: '#fbbf24' }}>
                {selectedWorkspace || 'PolyTuring'}
              </div>
            </div>
          </div>
        </div>

        {/* Filter Bar & Category Chips */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <div
            style={{
              position: 'relative',
              width: '100%',
            }}
          >
            <Search
              size={17}
              style={{
                position: 'absolute',
                left: 14,
                top: '50%',
                transform: 'translateY(-50%)',
                color: 'var(--text-muted, #94a3b8)',
              }}
            />
            <input
              type="text"
              placeholder="Search endpoints by path, description, HTTP method, or tag..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{
                width: '100%',
                padding: '12px 16px 12px 42px',
                background: 'rgba(15, 23, 42, 0.75)',
                border: '1px solid rgba(255, 255, 255, 0.1)',
                borderRadius: 10,
                color: '#fff',
                fontSize: '0.88rem',
                outline: 'none',
                boxSizing: 'border-box',
                backdropFilter: 'blur(16px)',
              }}
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                style={{
                  position: 'absolute',
                  right: 12,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-muted, #94a3b8)',
                  cursor: 'pointer',
                  fontSize: '0.8rem',
                }}
              >
                Clear
              </button>
            )}
          </div>

          {/* Tags Chips Bar */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
            <span style={{ fontSize: '0.72rem', color: 'var(--text-muted, #94a3b8)', fontWeight: 600, marginRight: 4 }}>
              TAG CATEGORY:
            </span>
            {tagsList.map((tag) => (
              <button
                key={tag}
                onClick={() => setActiveTag(tag)}
                style={{
                  background: activeTag === tag ? 'rgba(56, 189, 248, 0.22)' : 'rgba(255, 255, 255, 0.05)',
                  border: activeTag === tag ? '1px solid rgba(56, 189, 248, 0.5)' : '1px solid rgba(255, 255, 255, 0.08)',
                  color: activeTag === tag ? '#38bdf8' : 'var(--text-secondary, #cbd5e1)',
                  borderRadius: 20,
                  padding: '4px 12px',
                  fontSize: '0.74rem',
                  fontWeight: 600,
                  cursor: 'pointer',
                  transition: 'all 0.2s ease',
                }}
              >
                {tag}
              </button>
            ))}
          </div>
        </div>

        {/* Loading / Error State */}
        {isLoadingSpec && (
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', gap: 10, padding: 60, color: '#38bdf8' }}>
            <RefreshCw size={24} className="spinning-icon" />
            <span style={{ fontSize: '0.95rem', fontWeight: 600 }}>Loading OpenAPI Specification...</span>
          </div>
        )}

        {specError && (
          <div
            style={{
              background: 'rgba(239, 68, 68, 0.12)',
              border: '1px solid rgba(239, 68, 68, 0.35)',
              borderRadius: 10,
              padding: 16,
              color: '#fca5a5',
              display: 'flex',
              alignItems: 'center',
              gap: 12,
            }}
          >
            <AlertCircle size={20} color="#ef4444" />
            <div style={{ fontSize: '0.85rem' }}>{specError}</div>
          </div>
        )}

        {/* Endpoints List */}
        {!isLoadingSpec && !specError && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            {filteredEndpoints.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '48px 0', color: 'var(--text-muted, #94a3b8)' }}>
                <p style={{ margin: 0, fontSize: '0.95rem' }}>No API endpoints matching "{searchQuery}"</p>
              </div>
            ) : (
              filteredEndpoints.map((ep) => {
                const key = `${ep.method}-${ep.path}`;
                const isExpanded = !!expandedEndpoints[key];
                const badgeStyle = getMethodBadgeStyle(ep.method);
                const result = testResults[key];
                const curlCmd = generateCurlCommand(ep);

                return (
                  <div
                    key={key}
                    style={{
                      background: 'rgba(15, 23, 42, 0.75)',
                      border: isExpanded ? '1px solid rgba(56, 189, 248, 0.4)' : '1px solid rgba(255, 255, 255, 0.08)',
                      borderRadius: 12,
                      overflow: 'hidden',
                      backdropFilter: 'blur(16px)',
                      transition: 'border-color 0.2s ease, box-shadow 0.2s ease',
                      boxShadow: isExpanded ? '0 8px 30px rgba(0, 0, 0, 0.5), 0 0 15px rgba(56, 189, 248, 0.1)' : 'none',
                    }}
                  >
                    {/* Collapsible Card Header */}
                    <div
                      onClick={() => toggleEndpoint(key)}
                      style={{
                        padding: '14px 18px',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        cursor: 'pointer',
                        userSelect: 'none',
                        gap: 14,
                        background: isExpanded ? 'rgba(255, 255, 255, 0.02)' : 'transparent',
                      }}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: 12, overflow: 'hidden' }}>
                        <span
                          style={{
                            ...badgeStyle,
                            padding: '3px 8px',
                            borderRadius: 6,
                            fontSize: '0.72rem',
                            fontWeight: 800,
                            fontFamily: 'monospace',
                            letterSpacing: '0.5px',
                            flexShrink: 0,
                          }}
                        >
                          {ep.method}
                        </span>
                        <span
                          style={{
                            fontFamily: 'monospace',
                            fontSize: '0.92rem',
                            fontWeight: 700,
                            color: '#fff',
                            whiteSpace: 'nowrap',
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                          }}
                        >
                          /api{ep.path}
                        </span>
                        <span style={{ fontSize: '0.8rem', color: 'var(--text-muted, #94a3b8)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                          — {ep.summary}
                        </span>
                      </div>

                      <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
                        <span
                          style={{
                            background: 'rgba(255, 255, 255, 0.05)',
                            padding: '2px 8px',
                            borderRadius: 6,
                            fontSize: '0.68rem',
                            color: 'var(--text-muted, #94a3b8)',
                          }}
                        >
                          {ep.tag}
                        </span>
                        {isExpanded ? <ChevronDown size={17} color="#94a3b8" /> : <ChevronRight size={17} color="#94a3b8" />}
                      </div>
                    </div>

                    {/* Expanded Drawer Content */}
                    {isExpanded && (
                      <div
                        style={{
                          padding: '18px 20px',
                          borderTop: '1px solid rgba(255, 255, 255, 0.08)',
                          background: 'rgba(0, 0, 0, 0.3)',
                          display: 'flex',
                          flexDirection: 'column',
                          gap: 16,
                        }}
                      >
                        {/* Description */}
                        {ep.description && (
                          <div style={{ fontSize: '0.84rem', color: 'var(--text-secondary, #cbd5e1)', lineHeight: 1.55 }}>
                            {ep.description}
                          </div>
                        )}

                        {/* Parameters Table */}
                        {ep.parameters && ep.parameters.length > 0 && (
                          <div>
                            <div style={{ fontSize: '0.74rem', fontWeight: 700, color: '#38bdf8', letterSpacing: '0.5px', marginBottom: 8 }}>
                              REQUEST PARAMETERS
                            </div>
                            <div
                              style={{
                                background: 'rgba(0, 0, 0, 0.4)',
                                border: '1px solid rgba(255, 255, 255, 0.06)',
                                borderRadius: 8,
                                overflow: 'hidden',
                              }}
                            >
                              <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: '0.78rem' }}>
                                <thead>
                                  <tr style={{ background: 'rgba(255, 255, 255, 0.03)', textAlign: 'left' }}>
                                    <th style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>NAME</th>
                                    <th style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>IN</th>
                                    <th style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>TYPE</th>
                                    <th style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>DESCRIPTION</th>
                                    <th style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)', fontWeight: 600 }}>VALUE</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {ep.parameters.map((param) => {
                                    const currVal =
                                      inputValues[key]?.[param.name] !== undefined
                                        ? inputValues[key][param.name]
                                        : (param.name === 'workspace' && selectedWorkspace ? selectedWorkspace : '');

                                    return (
                                      <tr key={param.name} style={{ borderTop: '1px solid rgba(255, 255, 255, 0.04)' }}>
                                        <td style={{ padding: '8px 12px', fontFamily: 'monospace', fontWeight: 700, color: '#fff' }}>
                                          {param.name}
                                          {param.required && <span style={{ color: '#f87171', marginLeft: 4 }}>*</span>}
                                        </td>
                                        <td style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)' }}>{param.in}</td>
                                        <td style={{ padding: '8px 12px', color: '#c084fc', fontFamily: 'monospace' }}>
                                          {param.schema?.type || 'string'}
                                        </td>
                                        <td style={{ padding: '8px 12px', color: 'var(--text-muted, #94a3b8)' }}>{param.description || '—'}</td>
                                        <td style={{ padding: '6px 12px', width: 220 }}>
                                          <input
                                            type="text"
                                            placeholder={param.schema?.default ? `default: ${param.schema.default}` : 'Enter value...'}
                                            value={currVal}
                                            onChange={(e) => handleInputChange(key, param.name, e.target.value)}
                                            style={{
                                              width: '100%',
                                              padding: '5px 8px',
                                              background: 'rgba(15, 23, 42, 0.8)',
                                              border: '1px solid rgba(255, 255, 255, 0.12)',
                                              borderRadius: 4,
                                              color: '#fff',
                                              fontSize: '0.76rem',
                                              fontFamily: 'monospace',
                                            }}
                                          />
                                        </td>
                                      </tr>
                                    );
                                  })}
                                </tbody>
                              </table>
                            </div>
                          </div>
                        )}

                        {/* Request Body Editor (for POST/PUT) */}
                        {ep.requestBody && (
                          <div>
                            <div style={{ fontSize: '0.74rem', fontWeight: 700, color: '#c084fc', letterSpacing: '0.5px', marginBottom: 8 }}>
                              REQUEST BODY (JSON)
                            </div>
                            <textarea
                              rows={4}
                              placeholder="Enter JSON request payload..."
                              value={
                                bodyInputs[key] !== undefined
                                  ? bodyInputs[key]
                                  : ep.path === '/query'
                                  ? JSON.stringify({ prompt: 'Explain the AST architecture', workspace: selectedWorkspace || 'PolyTuring' }, null, 2)
                                  : ep.path === '/ingest'
                                  ? JSON.stringify({ path: '/workspace', project: 'PolyTuring', async: true }, null, 2)
                                  : '{\n  \n}'
                              }
                              onChange={(e) => handleBodyChange(key, e.target.value)}
                              style={{
                                width: '100%',
                                padding: '10px 12px',
                                background: 'rgba(0, 0, 0, 0.5)',
                                border: '1px solid rgba(255, 255, 255, 0.12)',
                                borderRadius: 6,
                                color: '#e2e8f0',
                                fontFamily: 'monospace',
                                fontSize: '0.78rem',
                                resize: 'vertical',
                                boxSizing: 'border-box',
                              }}
                            />
                          </div>
                        )}

                        {/* Execution Actions & Live Try Button */}
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 12 }}>
                          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                            <button
                              onClick={() => handleExecuteLive(ep)}
                              disabled={result?.isLoading}
                              className="cyber-button"
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 7,
                                padding: '8px 16px',
                                borderRadius: 6,
                                fontSize: '0.8rem',
                                cursor: 'pointer',
                              }}
                            >
                              {result?.isLoading ? <RefreshCw size={14} className="spinning-icon" /> : <Play size={14} fill="#fff" />}
                              <span>{result?.isLoading ? 'Executing...' : 'Try It Live'}</span>
                            </button>

                            <button
                              onClick={() => copyToClipboard(curlCmd, `curl-${key}`)}
                              className="cyber-button-secondary"
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 6,
                                padding: '8px 12px',
                                borderRadius: 6,
                                fontSize: '0.78rem',
                                cursor: 'pointer',
                              }}
                            >
                              {copiedId === `curl-${key}` ? <Check size={14} color="#34d399" /> : <Terminal size={14} />}
                              <span>{copiedId === `curl-${key}` ? 'cURL Copied!' : 'Copy cURL'}</span>
                            </button>
                          </div>

                          {/* Latency and Status Badge */}
                          {result && (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                              <span
                                style={{
                                  background: result.status >= 200 && result.status < 300 ? 'rgba(52, 211, 153, 0.2)' : 'rgba(239, 68, 68, 0.2)',
                                  border:
                                    result.status >= 200 && result.status < 300
                                      ? '1px solid rgba(52, 211, 153, 0.4)'
                                      : '1px solid rgba(239, 68, 68, 0.4)',
                                  color: result.status >= 200 && result.status < 300 ? '#34d399' : '#f87171',
                                  padding: '3px 10px',
                                  borderRadius: 6,
                                  fontSize: '0.74rem',
                                  fontWeight: 700,
                                  fontFamily: 'monospace',
                                }}
                              >
                                {result.status} {result.statusText}
                              </span>
                              <span
                                style={{
                                  background: 'rgba(255, 255, 255, 0.05)',
                                  border: '1px solid rgba(255, 255, 255, 0.1)',
                                  color: 'var(--text-muted, #94a3b8)',
                                  padding: '3px 8px',
                                  borderRadius: 6,
                                  fontSize: '0.72rem',
                                  fontFamily: 'monospace',
                                }}
                              >
                                ⏱️ {result.latencyMs}ms
                              </span>
                            </div>
                          )}
                        </div>

                        {/* Live Response Panel */}
                        {result && (
                          <div
                            style={{
                              background: 'rgba(0, 0, 0, 0.6)',
                              border: '1px solid rgba(255, 255, 255, 0.1)',
                              borderRadius: 8,
                              overflow: 'hidden',
                            }}
                          >
                            <div
                              style={{
                                display: 'flex',
                                alignItems: 'center',
                                justifyContent: 'space-between',
                                padding: '8px 12px',
                                background: 'rgba(255, 255, 255, 0.03)',
                                borderBottom: '1px solid rgba(255, 255, 255, 0.06)',
                              }}
                            >
                              <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: '0.72rem', color: '#94a3b8', fontWeight: 600 }}>
                                <FileJson size={13} color="#38bdf8" />
                                <span>RESPONSE PAYLOAD</span>
                              </div>
                              <button
                                onClick={() =>
                                  copyToClipboard(
                                    typeof result.data === 'object' ? JSON.stringify(result.data, null, 2) : String(result.data || result.error),
                                    `resp-${key}`
                                  )
                                }
                                style={{
                                  background: 'none',
                                  border: 'none',
                                  color: 'var(--text-muted, #94a3b8)',
                                  cursor: 'pointer',
                                  display: 'flex',
                                  alignItems: 'center',
                                  gap: 4,
                                  fontSize: '0.7rem',
                                }}
                              >
                                {copiedId === `resp-${key}` ? <Check size={12} color="#34d399" /> : <Copy size={12} />}
                                <span>{copiedId === `resp-${key}` ? 'Copied' : 'Copy'}</span>
                              </button>
                            </div>
                            <pre
                              style={{
                                margin: 0,
                                padding: '12px 14px',
                                fontSize: '0.76rem',
                                fontFamily: 'monospace',
                                color: result.error ? '#f87171' : '#e2e8f0',
                                maxHeight: 280,
                                overflowY: 'auto',
                                whiteSpace: 'pre-wrap',
                                wordBreak: 'break-word',
                              }}
                            >
                              {result.error
                                ? `Error: ${result.error}`
                                : typeof result.data === 'object'
                                ? JSON.stringify(result.data, null, 2)
                                : String(result.data)}
                            </pre>
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default ApiSpecExplorer;
