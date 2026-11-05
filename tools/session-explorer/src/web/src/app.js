// ==========================================================================
// Session Explorer — Single Page Application Logic
// Implements: R-008..R-016, R-021, R-022, R-025, R-026
// ==========================================================================

import { marked } from 'marked';
import hljs from 'highlight.js';
import katex from 'katex';
import mermaid from 'mermaid';

// Initialize Mermaid with Dark Glass theme matching PolyTuring docs portal
mermaid.initialize({
  startOnLoad: false,
  theme: 'dark',
  securityLevel: 'loose',
  fontFamily: "'JetBrains Mono', 'Inter', monospace",
  themeVariables: {
    darkMode: true,
    background: '#090d16',
    primaryColor: '#00f2fe',
    primaryTextColor: '#f8fafc',
    primaryBorderColor: '#38bdf8',
    lineColor: '#a855f7',
    secondaryColor: '#1e293b',
    tertiaryColor: '#0f172a'
  }
});

// Heuristic to distinguish mathematical expressions ($Y=0$, $40 \times 40$) from currency ($100, $10-$20)
function isLikelyMath(inner) {
  if (!inner) return false;
  // If it contains LaTeX backslash commands (\approx, \text, \implies, etc.), it is math
  if (/\\\w+/.test(inner)) return true;
  // If it contains mathematical operators, subscripts, superscripts, or relations
  if (/[\^_=<>~±×÷≠≤≥]/.test(inner)) return true;
  // If it starts with a digit, exclude pure currency numbers or currency ranges ($100, $25.50, $10 - $20)
  if (/^\d/.test(inner)) {
    if (/^\d+(?:[.,]\d+)?(?:\s*[-–—]\s*\$?\d+(?:[.,]\d+)?)?$/.test(inner)) {
      return false;
    }
  }
  // Single variable (e.g. x, n, i, j, k, T, Y) or algebraic expression (e.g. n + 1)
  if (/^[a-zA-Z]$/.test(inner) || /^[a-zA-Z]\s*[\+\-\*\/]\s*[a-zA-Z0-9]/.test(inner)) return true;
  // Non-digit without currency words
  if (!/^\d/.test(inner) && !/\b(dollars?|cents?|usd|eur|gbp)\b/i.test(inner)) return true;
  return false;
}

// KaTeX Block Math Extension ($$ ... $$ and \[ ... \])
const mathBlockExt = {
  name: 'mathBlock',
  level: 'block',
  start(src) {
    const idx1 = src.indexOf('$$');
    const idx2 = src.indexOf('\\[');
    if (idx1 === -1) return idx2 === -1 ? undefined : idx2;
    if (idx2 === -1) return idx1;
    return Math.min(idx1, idx2);
  },
  tokenizer(src) {
    const match1 = /^\$\$([\s\S]+?)\$\$/.exec(src);
    if (match1) {
      return { type: 'mathBlock', raw: match1[0], text: match1[1].trim() };
    }
    const match2 = /^\\\[([\s\S]+?)\\\]/.exec(src);
    if (match2) {
      return { type: 'mathBlock', raw: match2[0], text: match2[1].trim() };
    }
  },
  renderer(token) {
    try {
      const rendered = katex.renderToString(token.text, { displayMode: true, throwOnError: false, errorColor: '#f87171' });
      return `<div class="katex-math-block">${rendered}</div>`;
    } catch (e) {
      return `<pre class="katex-math-error"><code>${escapeHtml(token.text)}</code></pre>`;
    }
  }
};

// KaTeX Inline Math Extension ($ ... $ and \( ... \))
const inlineMathExt = {
  name: 'inlineMath',
  level: 'inline',
  start(src) {
    const idx1 = src.indexOf('$');
    const idx2 = src.indexOf('\\(');
    if (idx1 === -1) return idx2 === -1 ? undefined : idx2;
    if (idx2 === -1) return idx1;
    return Math.min(idx1, idx2);
  },
  tokenizer(src) {
    const match1 = /^\$([^\$\s](?:[^\$\n]*?[^\$\s])?)\$/.exec(src);
    if (match1) {
      const inner = match1[1].trim();
      if (!isLikelyMath(inner)) {
        return undefined;
      }
      return { type: 'inlineMath', raw: match1[0], text: inner };
    }
    const match2 = /^\\\(([\s\S]+?)\\\)/.exec(src);
    if (match2) {
      return { type: 'inlineMath', raw: match2[0], text: match2[1].trim() };
    }
  },
  renderer(token) {
    try {
      const rendered = katex.renderToString(token.text, { displayMode: false, throwOnError: false, errorColor: '#f87171' });
      return `<span class="katex-math-inline">${rendered}</span>`;
    } catch (e) {
      return `<code>${escapeHtml(token.text)}</code>`;
    }
  }
};

const calloutIcons = {
  NOTE: 'ℹ️',
  TIP: '💡',
  IMPORTANT: '📌',
  WARNING: '⚠️',
  CAUTION: '🛑'
};

// Detect if a code block is a Mermaid diagram by language tag or code signature
function isMermaidCode(text, lang) {
  const cleanLang = (lang || '').trim().toLowerCase();
  if (['mermaid', 'diagram', 'chart', 'flowchart'].includes(cleanLang)) return true;
  if (!text) return false;
  // Strip comments and frontmatter/directives
  const cleaned = text.replace(/%%\{[\s\S]*?\}%%/g, '').replace(/%%.*$/gm, '').trim();
  return /^(graph\s+(?:TD|TB|BT|RL|LR)|flowchart\s+(?:TD|TB|BT|RL|LR)|sequenceDiagram|classDiagram|stateDiagram(?:-v2)?|erDiagram|journey|gantt|pie|quadrantChart|requirementDiagram|gitGraph|C4(?:Context|Container|Component|Dynamic|Deployment)|mindmap|timeline|zenuml|sankey(?:-beta)?|xychart(?:-beta)?|block(?:-beta)?|packet(?:-beta)?|kanban|architecture(?:-beta)?)\b/i.test(cleaned);
}

// Configure marked with highlight.js, rich code blocks, file links, and Mermaid diagrams
const renderer = {
  code({ text, lang }) {
    if (isMermaidCode(text, lang)) {
      const diagramId = 'mermaid-' + Math.random().toString(36).substring(2, 11);
      return `
        <div class="canvas-container mermaid-card" id="${diagramId}-card">
          <div class="canvas-header-bar">
            <span class="canvas-type-tag">📊 Diagram Canvas</span>
            <div class="canvas-controls-group">
              <button class="canvas-tool-btn zoom-out-btn" title="Zoom Out">−</button>
              <span class="zoom-level-badge">100%</span>
              <button class="canvas-tool-btn zoom-in-btn" title="Zoom In">+</button>
              <button class="canvas-tool-btn zoom-reset-btn" title="Reset Zoom">⟲</button>
              <button class="canvas-tool-btn theater-btn" title="Open Fullscreen Theater">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M8 3H5a2 2 0 0 0-2 2v3m18 0V5a2 2 0 0 0-2-2h-3m0 18h3a2 2 0 0 0 2-2v-3M3 16v3a2 2 0 0 0 2 2h3"></path>
                </svg>
                <span>Theater</span>
              </button>
              <button class="canvas-tool-btn toggle-mermaid-src-btn" data-target="${diagramId}-src" title="View Diagram Source">Source</button>
              <button class="canvas-tool-btn copy-mermaid-src-btn" data-code="${escapeHtml(text)}" title="Copy Diagram Code">Copy</button>
            </div>
          </div>
          <div class="canvas-viewport" id="${diagramId}-viewport" title="Double click to open in Theater mode">
            <div class="canvas-content-surface">
              <pre class="mermaid" id="${diagramId}">${escapeHtml(text)}</pre>
            </div>
          </div>
          <pre class="mermaid-source-pre" id="${diagramId}-src" style="display: none;"><code>${escapeHtml(text)}</code></pre>
        </div>
      `;
    }
    const cleanLang = (lang || '').trim().toLowerCase();
    const validLang = !!(cleanLang && hljs.getLanguage(cleanLang));
    const highlighted = validLang
      ? hljs.highlight(text, { language: cleanLang, ignoreIllegals: true }).value
      : escapeHtml(text);
    return `
      <div class="code-block-wrapper">
        <div class="code-block-header">
          <span class="code-lang-badge">${cleanLang ? cleanLang.toUpperCase() : 'TELEMETRY'}</span>
          <button class="btn-ghost-sm copy-code-btn" data-code="${escapeHtml(text)}" title="Copy code to clipboard">Copy</button>
        </div>
        <pre><code class="hljs ${validLang ? 'language-' + cleanLang : ''}">${highlighted}</code></pre>
      </div>
    `;
  },
  link(token) {
    const textHtml = this.parser.parseInline(token.tokens);
    const href = token.href || '';
    if (href.startsWith('file://') || href.startsWith('/')) {
      const cleanPath = href.replace(/^file:\/\//, '');
      const formattedPath = cleanPath.replace(/#L(\d+)(?:-L(\d+))?$/, ':$1');
      return `
        <span class="file-link-chip-group">
          <a href="${href}" class="file-path-link" data-filepath="${cleanPath}" title="Click to view file in browser"><span class="file-icon">📄</span> <span class="file-link-text">${textHtml}</span> <span class="file-chip-action" title="Open file in browser">↗</span></a>
          <button class="file-link-quick-copy-btn" data-path="${cleanPath}" title="Copy file path directly to clipboard">📋</button>
        </span>
      `;
    }
    const titleAttr = token.title ? ` title="${token.title}"` : '';
    return `<a href="${href}" target="_blank" rel="noopener noreferrer"${titleAttr}>${textHtml}</a>`;
  }
};

marked.use({
  extensions: [mathBlockExt, inlineMathExt],
  renderer,
  breaks: true,
  gfm: true,
  hooks: {
    postprocess(html) {
      // Responsive table scrolling wrapper
      html = html.replace(/<table>/g, '<div class="table-responsive"><table>').replace(/<\/table>/g, '</table></div>');
      // GitHub alerts callouts (> [!NOTE], > [!TIP], etc.)
      html = html.replace(/<blockquote>\s*<p>\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\](?:\s*<br\s*\/?>)?([\s\S]*?)<\/blockquote>/gi, (match, type, content) => {
        const uType = type.toUpperCase();
        const icon = calloutIcons[uType] || 'ℹ️';
        return `<div class="callout-card callout-${uType.toLowerCase()}"><div class="callout-header"><span class="callout-icon">${icon}</span> <strong class="callout-title">${uType}</strong></div><div class="callout-body"><p>${content}</div>`;
      });
      return html;
    }
  }
});

// App State
const state = {
  sessions: [],
  stats: null,
  selectedSessionId: null,
  selectedSessionDetail: null,
  currentView: 'dashboard', // 'dashboard' | 'detail'
  groupBy: localStorage.getItem('se_group_by') || 'project', // 'project' | 'date' | 'flat'
  filterWorkspace: '',
  filterRuntime: '', // '' | 'ide' | 'cli' | 'core'
  dateFrom: '',
  dateTo: '',
  sortBy: 'date',
  sortOrder: 'desc',
  messageFilter: 'ALL', // 'ALL' | 'USER_INPUT' | 'PLANNER_RESPONSE' | 'TOOL_CALLS' | 'ERROR'
  searchQuery: '',
  searchResults: [],
  theme: localStorage.getItem('se_theme') || 'dark',
  drilledProject: null, // active project when drilled into a repository
  sessionStepSearch: '', // search query for steps within active session detail
  sessionStepSort: 'step_desc', // sort mode for steps within active session detail ('step_desc' by default, 'step_asc', 'time_desc', 'time_asc', 'tools_desc', 'length_desc')
  projectFilters: {
    search: '',
    sort: 'date_desc',
    dateFrom: '',
    dateTo: '',
    datePreset: 'all'
  },
  liveInterval: parseInt(localStorage.getItem('se_live_interval') || '10000', 10),
  lastSynced: new Date(),
  activeTimelineSessionId: null,
  followStream: true,
  tailEventSource: null,
  tailStreaming: true,
  pendingStepsBuffer: [], // buffered incoming steps when user is detached (followStream: false)
  renderedStepMap: new Map(), // step_index -> HTMLElement
  theaterZoom: 1.0,
  theaterPanX: 0,
  theaterPanY: 0,
  theaterRawSrc: '',
  fileViewerData: null,
  fileViewerMode: 'preview', // 'preview' | 'raw'
  fileViewerWrap: false,
  fileViewerTargetLine: null,
  fileViewerTargetLineEnd: null,
  fileFindQuery: '',
  fileFindCaseSensitive: false,
  fileFindMatches: [],
  fileFindActiveIndex: -1,
  sessionFilesMap: new Map(), // cleanPath -> { path, name, type, tool, stepIndex }
  agentMatrix: {
    scope: 'global', // 'global' | 'workspace' | 'cascade'
    type: 'ALL',     // 'ALL' | 'RULES' | 'SKILLS' | 'PLUGINS' | 'HOOKS'
    search: '',
    targetWorkspace: null,
    globalContext: null,
    workspaceContext: null
  },
  sessionMatrix: {
    lens: 'active', // 'active' | 'cascade' | 'telemetry'
    type: 'ALL',    // 'ALL' | 'SKILLS' | 'MCP' | 'NATIVE' | 'RULES' | 'ORCHESTRATION'
    search: '',
    context: null,
    sessionId: null
  },
  sessionContextData: null
};

// DOM Elements
const el = {
  themeToggleBtn: document.getElementById('themeToggleBtn'),
  insightsBtn: document.getElementById('insightsBtn'),
  refreshBtn: document.getElementById('refreshBtn'),
  liveIntervalSelect: document.getElementById('liveIntervalSelect'),
  liveDot: document.getElementById('liveDot'),
  lastSyncLabel: document.getElementById('lastSyncLabel'),
  globalSearchInput: document.getElementById('globalSearchInput'),
  statsStrip: document.getElementById('statsStrip'),
  workspaceSelect: document.getElementById('workspaceSelect'),
  workspaceStrip: document.getElementById('workspaceStrip'),
  sortBySelect: document.getElementById('sortBySelect'),
  dateFromInput: document.getElementById('dateFromInput'),
  dateToInput: document.getElementById('dateToInput'),
  sessionCountBadge: document.getElementById('sessionCountBadge'),
  sessionGrid: document.getElementById('sessionGrid'),
  viewContainer: document.getElementById('viewContainer'),
  groupByControl: document.getElementById('groupByControl'),
  groupByBtns: document.querySelectorAll('#groupByControl .segment-btn'),
  runtimeControl: document.getElementById('runtimeControl'),
  runtimeBtns: document.querySelectorAll('#runtimeControl .segment-btn'),
  projectHeroBanner: document.getElementById('projectHeroBanner'),
  dashboardView: document.getElementById('dashboardView'),
  detailView: document.getElementById('detailView'),
  detailBackBtn: document.getElementById('detailBackBtn'),
  detailSessionId: document.getElementById('detailSessionId'),
  detailWorkspaceBadge: document.getElementById('detailWorkspaceBadge'),
  detailRuntimeBadge: document.getElementById('detailRuntimeBadge'),
  detailTime: document.getElementById('detailTime'),
  detailFilesBtn: document.getElementById('detailFilesBtn'),
  detailFilesCountBadge: document.getElementById('detailFilesCountBadge'),
  exportMdBtn: document.getElementById('exportMdBtn'),
  exportJsonBtn: document.getElementById('exportJsonBtn'),
  timelineContainer: document.getElementById('timelineContainer'),
  tailStreamControl: document.getElementById('tailStreamControl'),
  toggleStreamBtn: document.getElementById('toggleStreamBtn'),
  detailLiveDot: document.getElementById('detailLiveDot'),
  tailStatusLabel: document.getElementById('tailStatusLabel'),
  toggleFollowStreamBtn: document.getElementById('toggleFollowStreamBtn'),
  refreshTrailBtn: document.getElementById('refreshTrailBtn'),
  newStepsFloatingBtn: document.getElementById('newStepsFloatingBtn'),
  msgFilterChips: document.querySelectorAll('.msg-filter-chip'),
  sessionStepSearchInput: document.getElementById('sessionStepSearchInput'),
  clearSessionStepSearchBtn: document.getElementById('clearSessionStepSearchBtn'),
  sessionStepSortSelect: document.getElementById('sessionStepSortSelect'),
  detailStepCountBadge: document.getElementById('detailStepCountBadge'),
  resetSessionStepFiltersBtn: document.getElementById('resetSessionStepFiltersBtn'),
  datePresets: document.querySelectorAll('.date-preset-chip'),
  searchModal: document.getElementById('searchModal'),
  modalSearchInput: document.getElementById('modalSearchInput'),
  modalSearchResults: document.getElementById('modalSearchResults'),
  closeSearchModalBtn: document.getElementById('closeSearchModalBtn'),
  insightsModal: document.getElementById('insightsModal'),
  closeInsightsModalBtn: document.getElementById('closeInsightsModalBtn'),
  topToolsList: document.getElementById('topToolsList'),
  topPromptsList: document.getElementById('topPromptsList'),
  installSkillBtn: document.getElementById('installSkillBtn'),
  installSkillModal: document.getElementById('installSkillModal'),
  closeInstallModalBtn: document.getElementById('closeInstallModalBtn'),
  doInstallSkillBtn: document.getElementById('doInstallSkillBtn'),
  copyInstallCmdBtn: document.getElementById('copyInstallCmdBtn'),
  installStatusText: document.getElementById('installStatusText'),
  // Theater Modal Elements (R-037)
  diagramTheaterModal: document.getElementById('diagramTheaterModal'),
  diagramTheaterTitleText: document.getElementById('diagramTheaterTitleText'),
  diagramTheaterViewport: document.getElementById('diagramTheaterViewport'),
  diagramTheaterSurface: document.getElementById('diagramTheaterSurface'),
  theaterZoomOutBtn: document.getElementById('theaterZoomOutBtn'),
  theaterZoomBadge: document.getElementById('theaterZoomBadge'),
  theaterZoomInBtn: document.getElementById('theaterZoomInBtn'),
  theaterZoomResetBtn: document.getElementById('theaterZoomResetBtn'),
  theaterFitBtn: document.getElementById('theaterFitBtn'),
  theaterCopySvgBtn: document.getElementById('theaterCopySvgBtn'),
  theaterCopySrcBtn: document.getElementById('theaterCopySrcBtn'),
  closeDiagramTheaterBtn: document.getElementById('closeDiagramTheaterBtn'),
  // File & Artifact Viewer Elements (R-038)
  fileViewerModal: document.getElementById('fileViewerModal'),
  fileViewerIcon: document.getElementById('fileViewerIcon'),
  fileViewerFilename: document.getElementById('fileViewerFilename'),
  fileViewerPath: document.getElementById('fileViewerPath'),
  fileViewerLangBadge: document.getElementById('fileViewerLangBadge'),
  fileViewerLinesBadge: document.getElementById('fileViewerLinesBadge'),
  fileViewerSizeBadge: document.getElementById('fileViewerSizeBadge'),
  fileViewerModeControl: document.getElementById('fileViewerModeControl'),
  fileViewerWrapBtn: document.getElementById('fileViewerWrapBtn'),
  fileViewerCopyContentBtn: document.getElementById('fileViewerCopyContentBtn'),
  fileViewerCopyPathBtn: document.getElementById('fileViewerCopyPathBtn'),
  closeFileViewerBtn: document.getElementById('closeFileViewerBtn'),
  fileViewerBody: document.getElementById('fileViewerBody'),
  fileViewerLoading: document.getElementById('fileViewerLoading'),
  fileViewerError: document.getElementById('fileViewerError'),
  fileViewerMarkdown: document.getElementById('fileViewerMarkdown'),
  fileViewerCodeContainer: document.getElementById('fileViewerCodeContainer'),
  fileViewerLineNumbers: document.getElementById('fileViewerLineNumbers'),
  fileViewerCodeBlock: document.getElementById('fileViewerCodeBlock'),
  fileViewerImageContainer: document.getElementById('fileViewerImageContainer'),
  fileViewerImage: document.getElementById('fileViewerImage'),
  fileViewerOpenFindBtn: document.getElementById('fileViewerOpenFindBtn'),
  fileViewerFindWidget: document.getElementById('fileViewerFindWidget'),
  fileFindInput: document.getElementById('fileFindInput'),
  fileFindCaseBtn: document.getElementById('fileFindCaseBtn'),
  fileFindCountBadge: document.getElementById('fileFindCountBadge'),
  fileFindPrevBtn: document.getElementById('fileFindPrevBtn'),
  fileFindNextBtn: document.getElementById('fileFindNextBtn'),
  fileFindCloseBtn: document.getElementById('fileFindCloseBtn'),
  // Session Files Modal
  sessionFilesModal: document.getElementById('sessionFilesModal'),
  closeSessionFilesModalBtn: document.getElementById('closeSessionFilesModalBtn'),
  sessionFilesTotalCount: document.getElementById('sessionFilesTotalCount'),
  sessionFilesListContainer: document.getElementById('sessionFilesListContainer'),
  // Agent Context Matrix & Brain Elements (R-039)
  agentBrainBtn: document.getElementById('agentBrainBtn'),
  agentMatrixModal: document.getElementById('agentMatrixModal'),
  closeAgentMatrixModalBtn: document.getElementById('closeAgentMatrixModalBtn'),
  agentMatrixScopeControl: document.getElementById('agentMatrixScopeControl'),
  agentMatrixWorkspaceSegment: document.getElementById('agentMatrixWorkspaceSegment'),
  agentMatrixSearchInput: document.getElementById('agentMatrixSearchInput'),
  clearAgentMatrixSearchBtn: document.getElementById('clearAgentMatrixSearchBtn'),
  agentMatrixTotalCountBadge: document.getElementById('agentMatrixTotalCountBadge'),
  agentMatrixRulesCount: document.getElementById('agentMatrixRulesCount'),
  agentMatrixSkillsCount: document.getElementById('agentMatrixSkillsCount'),
  agentMatrixPluginsCount: document.getElementById('agentMatrixPluginsCount'),
  agentMatrixHooksCount: document.getElementById('agentMatrixHooksCount'),
  agentMatrixContent: document.getElementById('agentMatrixContent'),
  detailAgentContextBtn: document.getElementById('detailAgentContextBtn'),
  detailAgentContextCount: document.getElementById('detailAgentContextCount'),
  sessionContextModal: document.getElementById('sessionContextModal'),
  closeSessionContextModalBtn: document.getElementById('closeSessionContextModalBtn'),
  sessionContextIdBadge: document.getElementById('sessionContextIdBadge'),
  sessionContextContent: document.getElementById('sessionContextContent'),
  // Session Matrix Holographic Elements (R-040)
  sessionMatrixTotalCountBadge: document.getElementById('sessionMatrixTotalCountBadge'),
  sessionMatrixLensControl: document.getElementById('sessionMatrixLensControl'),
  sessionMatrixSearchInput: document.getElementById('sessionMatrixSearchInput'),
  clearSessionMatrixSearchBtn: document.getElementById('clearSessionMatrixSearchBtn'),
  sessionMatrixChipAll: document.getElementById('sessionMatrixChipAll'),
  sessionMatrixChipSkills: document.getElementById('sessionMatrixChipSkills'),
  sessionMatrixChipSkillsLabel: document.getElementById('sessionMatrixChipSkillsLabel'),
  sessionMatrixChipMCP: document.getElementById('sessionMatrixChipMCP'),
  sessionMatrixChipNative: document.getElementById('sessionMatrixChipNative'),
  sessionMatrixChipRules: document.getElementById('sessionMatrixChipRules'),
  sessionMatrixChipOrch: document.getElementById('sessionMatrixChipOrch'),
  sessionMatrixInjectedVal: document.getElementById('sessionMatrixInjectedVal'),
  sessionMatrixMCPVal: document.getElementById('sessionMatrixMCPVal'),
  sessionMatrixDiversityVal: document.getElementById('sessionMatrixDiversityVal'),
  sessionMatrixOndemandVal: document.getElementById('sessionMatrixOndemandVal'),
  sessionMatrixReliabilityVal: document.getElementById('sessionMatrixReliabilityVal'),
  toast: document.getElementById('toastNotification')
};

// Initialize App
document.addEventListener('DOMContentLoaded', () => {
  initTheme();
  initEventListeners();
  initLiveWatch();
  initRouter();
  applyDatePreset('all');
});

// Theme Management
function initTheme() {
  document.documentElement.setAttribute('data-theme', state.theme);
  updateThemeIcon();
}

function toggleTheme() {
  state.theme = state.theme === 'dark' ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', state.theme);
  localStorage.setItem('se_theme', state.theme);
  updateThemeIcon();
}

function updateThemeIcon() {
  if (el.themeToggleBtn) {
    el.themeToggleBtn.textContent = state.theme === 'dark' ? '☀️' : '🌙';
  }
}

// Toast Notifications (R-026)
function showToast(message) {
  if (!el.toast) return;
  el.toast.textContent = message;
  el.toast.classList.add('active');
  setTimeout(() => {
    el.toast.classList.remove('active');
  }, 2500);
}

// Live Watch Management
let liveTimer = null;
let lastSyncTimer = null;

function initLiveWatch() {
  if (el.liveIntervalSelect) {
    el.liveIntervalSelect.value = String(state.liveInterval);
    el.liveIntervalSelect.addEventListener('change', (e) => {
      setLiveInterval(parseInt(e.target.value, 10));
    });
  }
  setLiveInterval(state.liveInterval);

  if (lastSyncTimer) clearInterval(lastSyncTimer);
  lastSyncTimer = setInterval(updateLastSyncLabel, 4000);
}

function setLiveInterval(ms) {
  if (liveTimer) {
    clearInterval(liveTimer);
    liveTimer = null;
  }
  state.liveInterval = ms;
  localStorage.setItem('se_live_interval', String(ms));

  // Dynamically synchronize watch interval with Go backend daemon
  const sec = Math.floor(ms / 1000);
  fetch('/api/watch-interval', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ interval_seconds: sec })
  }).catch(() => {});

  if (el.liveDot) {
    if (ms > 0) {
      el.liveDot.classList.add('active', 'pulsing');
      el.liveDot.title = `Live Watch active (every ${ms / 1000}s)`;
    } else {
      el.liveDot.classList.remove('active', 'pulsing');
      el.liveDot.title = 'Live Watch paused (Manual mode)';
    }
  }

  if (ms > 0) {
    liveTimer = setInterval(() => {
      refreshData(true);
    }, ms);
  }
}

function updateLastSyncLabel() {
  if (!el.lastSyncLabel) return;
  const elapsedSec = Math.floor((Date.now() - state.lastSynced.getTime()) / 1000);
  if (elapsedSec < 3) {
    el.lastSyncLabel.textContent = 'Synced just now';
  } else if (elapsedSec < 60) {
    el.lastSyncLabel.textContent = `Synced ${elapsedSec}s ago`;
  } else {
    const mins = Math.floor(elapsedSec / 60);
    el.lastSyncLabel.textContent = `Synced ${mins}m ago`;
  }
}

async function refreshData(silent = false) {
  if (!silent && el.refreshBtn) {
    el.refreshBtn.classList.add('spinning');
  }
  try {
    // 1. Trigger backend re-scan
    await fetch('/api/refresh', { method: 'POST' }).catch(() => {});

    // 2. Fetch fresh session listing
    await loadSessions();

    // 3. If currently inspecting an active session, refresh its steps too!
    if (state.currentView === 'detail' && state.selectedSessionId) {
      try {
        const res = await fetch(`/api/sessions/${state.selectedSessionId}`);
        if (res.ok) {
          const detail = await res.json();
          state.selectedSessionDetail = detail;
          if (el.detailStepCountBadge && detail.session) {
            el.detailStepCountBadge.textContent = `${detail.session.step_count} steps`;
          }
          renderTimelineMessages();
        }
      } catch (e) {
        // silent detail refresh
      }
    }

    state.lastSynced = new Date();
    updateLastSyncLabel();
    if (!silent) {
      showToast('🔄 Sessions refreshed');
    }
  } catch (err) {
    console.error('Refresh error:', err);
  } finally {
    if (!silent && el.refreshBtn) {
      setTimeout(() => el.refreshBtn.classList.remove('spinning'), 500);
    }
  }
}

// Event Listeners
function initEventListeners() {
  if (el.themeToggleBtn) el.themeToggleBtn.addEventListener('click', toggleTheme);
  if (el.refreshBtn) el.refreshBtn.addEventListener('click', () => refreshData(false));
  if (el.insightsBtn) el.insightsBtn.addEventListener('click', openInsightsModal);
  if (el.closeInsightsModalBtn) el.closeInsightsModalBtn.addEventListener('click', closeInsightsModal);

  // Install Skill Modal (R-033)
  if (el.installSkillBtn) el.installSkillBtn.addEventListener('click', openInstallModal);
  if (el.closeInstallModalBtn) el.closeInstallModalBtn.addEventListener('click', closeInstallModal);
  if (el.copyInstallCmdBtn) {
    el.copyInstallCmdBtn.addEventListener('click', () => {
      navigator.clipboard.writeText('make setup-session-wip');
      showToast('Command copied: make setup-session-wip');
    });
  }
  if (el.doInstallSkillBtn) el.doInstallSkillBtn.addEventListener('click', executeInstallSkill);

  // Global search triggers
  if (el.globalSearchInput) {
    el.globalSearchInput.addEventListener('focus', () => {
      openSearchModal();
      el.globalSearchInput.blur();
    });
  }

  // Search modal
  if (el.closeSearchModalBtn) el.closeSearchModalBtn.addEventListener('click', closeSearchModal);
  if (el.modalSearchInput) {
    let debounceTimer;
    el.modalSearchInput.addEventListener('input', (e) => {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => {
        executeSearch(e.target.value.trim());
      }, 250);
    });
  }

  // Keyboard Shortcuts (R-022)
  window.addEventListener('keydown', (e) => {
    // Cmd+K or Ctrl+K
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault();
      openSearchModal();
    }
    // Cmd+F or Ctrl+F inside File Viewer (Find in File)
    if ((e.metaKey || e.ctrlKey) && (e.key === 'f' || e.key === 'F')) {
      if (el.fileViewerModal && el.fileViewerModal.classList.contains('file-viewer-open')) {
        e.preventDefault();
        openFileFindWidget();
      }
    }
    // / to focus search if not inside input
    if (e.key === '/' && document.activeElement.tagName !== 'INPUT' && document.activeElement.tagName !== 'TEXTAREA') {
      e.preventDefault();
      openSearchModal();
    }
    // 'r' or 'R' refreshes session trail when in detail view (R-044)
    if ((e.key === 'r' || e.key === 'R') && !e.metaKey && !e.ctrlKey && !e.altKey) {
      if (state.currentView === 'detail' &&
          document.activeElement.tagName !== 'INPUT' &&
          document.activeElement.tagName !== 'TEXTAREA') {
        e.preventDefault();
        refreshCurrentSessionTrail(true);
      }
    }
    // Esc closes modals and detail view
    if (e.key === 'Escape') {
      if (el.fileViewerFindWidget && el.fileViewerFindWidget.style.display !== 'none') {
        closeFileFindWidget();
      } else if (el.diagramTheaterModal && el.diagramTheaterModal.classList.contains('theater-open')) {
        closeDiagramTheater();
      } else if (el.fileViewerModal && el.fileViewerModal.classList.contains('file-viewer-open')) {
        closeFileViewer();
      } else if (el.sessionFilesModal && el.sessionFilesModal.classList.contains('active')) {
        closeSessionFilesModal();
      } else if (el.agentMatrixModal && el.agentMatrixModal.classList.contains('active')) {
        closeAgentMatrixModal();
      } else if (el.sessionContextModal && el.sessionContextModal.classList.contains('active')) {
        closeSessionContextModal();
      } else if (el.searchModal && el.searchModal.classList.contains('active')) {
        closeSearchModal();
      } else if (el.insightsModal && el.insightsModal.classList.contains('active')) {
        closeInsightsModal();
      } else if (el.installSkillModal && el.installSkillModal.classList.contains('active')) {
        closeInstallModal();
      } else if (state.currentView === 'detail') {
        showDashboard();
      }
    }
  });

  // Filters & Sorting (Immediate reactivity)
  if (el.workspaceSelect) {
    el.workspaceSelect.addEventListener('change', (e) => {
      const val = e.target.value;
      state.filterWorkspace = val;
      if (state.groupBy === 'project') {
        state.drilledProject = val || null;
      }
      loadSessions();
    });
  }

  if (el.sortBySelect) {
    el.sortBySelect.addEventListener('change', (e) => {
      const val = e.target.value;
      if (val === 'date_asc') {
        state.sortBy = 'date';
        state.sortOrder = 'asc';
      } else {
        state.sortBy = val;
        state.sortOrder = 'desc';
      }
      loadSessions();
    });
  }

  const handleDateChange = () => {
    state.dateFrom = el.dateFromInput ? el.dateFromInput.value : '';
    state.dateTo = el.dateToInput ? el.dateToInput.value : '';
    if (el.datePresets) {
      el.datePresets.forEach(p => p.classList.remove('active'));
      if (!state.dateFrom && !state.dateTo) {
        const allPreset = Array.from(el.datePresets).find(p => p.dataset.preset === 'all');
        if (allPreset) allPreset.classList.add('active');
      }
    }
    loadSessions();
  };

  if (el.dateFromInput) {
    el.dateFromInput.addEventListener('change', handleDateChange);
    el.dateFromInput.addEventListener('input', handleDateChange);
  }

  if (el.dateToInput) {
    el.dateToInput.addEventListener('change', handleDateChange);
    el.dateToInput.addEventListener('input', handleDateChange);
  }

  // Date Presets
  el.datePresets.forEach(preset => {
    preset.addEventListener('click', () => {
      applyDatePreset(preset.dataset.preset);
    });
  });

  // Message Filter Chips
  el.msgFilterChips.forEach(chip => {
    chip.addEventListener('click', () => {
      el.msgFilterChips.forEach(c => c.classList.remove('active'));
      chip.classList.add('active');
      state.messageFilter = chip.dataset.filter;
      renderTimelineMessages();
    });
  });

  // In-session Step Search & Sort (Interrelated)
  if (el.sessionStepSearchInput) {
    el.sessionStepSearchInput.addEventListener('input', (e) => {
      state.sessionStepSearch = e.target.value;
      if (el.clearSessionStepSearchBtn) {
        el.clearSessionStepSearchBtn.style.display = state.sessionStepSearch ? 'inline-flex' : 'none';
      }
      renderTimelineMessages();
    });
  }

  if (el.clearSessionStepSearchBtn) {
    el.clearSessionStepSearchBtn.addEventListener('click', () => {
      state.sessionStepSearch = '';
      if (el.sessionStepSearchInput) {
        el.sessionStepSearchInput.value = '';
        el.sessionStepSearchInput.focus();
      }
      el.clearSessionStepSearchBtn.style.display = 'none';
      renderTimelineMessages();
    });
  }

  if (el.sessionStepSortSelect) {
    el.sessionStepSortSelect.addEventListener('change', (e) => {
      state.sessionStepSort = e.target.value;
      renderTimelineMessages();
    });
  }

  if (el.resetSessionStepFiltersBtn) {
    el.resetSessionStepFiltersBtn.addEventListener('click', () => {
      resetSessionStepFilters();
    });
  }

  // Detail View Back Button
  if (el.detailBackBtn) {
    el.detailBackBtn.addEventListener('click', showDashboard);
  }

  // Export Buttons (R-021)
  if (el.exportMdBtn) {
    el.exportMdBtn.addEventListener('click', () => exportCurrentSession('markdown'));
  }
  if (el.exportJsonBtn) {
    el.exportJsonBtn.addEventListener('click', () => exportCurrentSession('json'));
  }

  // Group By Segmented Switcher (R-028, R-029)
  if (el.groupByBtns) {
    el.groupByBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        el.groupByBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        state.groupBy = btn.dataset.group;
        state.drilledProject = null;
        localStorage.setItem('se_group_by', state.groupBy);
        renderView();
      });
    });
  }

  // Runtime Mode Segmented Switcher (R-036)
  if (el.runtimeBtns) {
    el.runtimeBtns.forEach(btn => {
      btn.addEventListener('click', () => {
        el.runtimeBtns.forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        state.filterRuntime = btn.dataset.runtime || '';
        loadSessions();
      });
    });
  }

  // Live Stream Play/Pause Toggle
  if (el.toggleStreamBtn) {
    el.toggleStreamBtn.addEventListener('click', (e) => {
      e.stopPropagation();
      if (state.tailStreaming) {
        stopSessionLiveTail(true);
      } else if (state.selectedSessionId) {
        startSessionLiveTail(state.selectedSessionId);
        refreshCurrentSessionTrail(false);
      }
    });
  }

  // Live Tail Auto-Follow Toggle
  if (el.toggleFollowStreamBtn) {
    el.toggleFollowStreamBtn.addEventListener('click', () => {
      state.followStream = !state.followStream;
      el.toggleFollowStreamBtn.classList.toggle('active', state.followStream);
      if (state.followStream) {
        flushPendingSteps();
        refreshCurrentSessionTrail(false);
      }
    });
  }

  // Dedicated Session Trail Refresh Button (R-044)
  if (el.refreshTrailBtn) {
    el.refreshTrailBtn.addEventListener('click', () => {
      refreshCurrentSessionTrail(true);
    });
  }

  // Floating "New steps arrived" Click Listener
  if (el.newStepsFloatingBtn) {
    el.newStepsFloatingBtn.addEventListener('click', () => {
      flushPendingSteps();
    });
  }

  // Window scroll listener: auto-detect if user scrolled away from the stream head/tail
  window.addEventListener('scroll', () => {
    if (state.currentView !== 'detail') return;
    const isDesc = (state.sessionStepSort || 'step_desc') === 'step_desc';
    const isAtHead = isDesc
      ? window.scrollY <= 150
      : (window.innerHeight + window.scrollY >= document.body.offsetHeight - 150);

    if (!isAtHead) {
      if (state.followStream) {
        state.followStream = false;
        if (el.toggleFollowStreamBtn) el.toggleFollowStreamBtn.classList.remove('active');
      }
    } else {
      // At stream head: if there are no pending buffered steps, hide the floating button
      if (!state.pendingStepsBuffer || state.pendingStepsBuffer.length === 0) {
        if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';
      }
    }
  }, { passive: true });

  // Window resize debouncer for EC2-Style Bin-Packing Dynamic Re-Pack (R-043)
  let matrixResizeTimer = null;
  window.addEventListener('resize', () => {
    clearTimeout(matrixResizeTimer);
    matrixResizeTimer = setTimeout(() => {
      if (el.sessionContextModal && el.sessionContextModal.classList.contains('active')) {
        renderSessionMatrix();
      }
      if (el.agentMatrixModal && el.agentMatrixModal.classList.contains('active')) {
        renderAgentMatrix();
      }
    }, 120);
  }, { passive: true });

  // Brand button resets drilldown and returns to main view
  const brandBtn = document.getElementById('brandBtn');
  if (brandBtn) {
    brandBtn.addEventListener('click', () => {
      state.drilledProject = null;
      state.filterWorkspace = '';
      if (el.workspaceSelect) el.workspaceSelect.value = '';
      window.location.hash = '#/';
      showDashboard();
      loadSessions();
    });
  }

  // Diagram Theater Listeners (R-037)
  if (el.closeDiagramTheaterBtn) {
    el.closeDiagramTheaterBtn.addEventListener('click', closeDiagramTheater);
  }
  if (el.theaterZoomInBtn) {
    el.theaterZoomInBtn.addEventListener('click', () => {
      state.theaterZoom = Math.min(4.0, +(state.theaterZoom + 0.25).toFixed(2));
      updateDiagramTheaterTransform();
    });
  }
  if (el.theaterZoomOutBtn) {
    el.theaterZoomOutBtn.addEventListener('click', () => {
      state.theaterZoom = Math.max(0.4, +(state.theaterZoom - 0.25).toFixed(2));
      updateDiagramTheaterTransform();
    });
  }
  if (el.theaterZoomResetBtn) {
    el.theaterZoomResetBtn.addEventListener('click', () => {
      state.theaterZoom = 1.0;
      state.theaterPanX = 0;
      state.theaterPanY = 0;
      updateDiagramTheaterTransform();
    });
  }
  if (el.theaterFitBtn) {
    el.theaterFitBtn.addEventListener('click', () => {
      state.theaterZoom = 1.0;
      state.theaterPanX = 0;
      state.theaterPanY = 0;
      updateDiagramTheaterTransform();
    });
  }
  if (el.theaterCopySvgBtn) {
    el.theaterCopySvgBtn.addEventListener('click', () => {
      const svg = el.diagramTheaterSurface?.querySelector('svg');
      if (svg) {
        copyToClipboard(svg.outerHTML, 'Diagram SVG copied to clipboard!');
      } else {
        showToast('No SVG diagram found');
      }
    });
  }
  if (el.theaterCopySrcBtn) {
    el.theaterCopySrcBtn.addEventListener('click', () => {
      copyToClipboard(state.theaterRawSrc, 'Diagram code copied to clipboard!');
    });
  }
  if (el.diagramTheaterViewport) {
    el.diagramTheaterViewport.addEventListener('mousedown', (e) => {
      if (e.target.closest('button')) return;
      isTheaterPanning = true;
      theaterStartX = e.clientX - state.theaterPanX;
      theaterStartY = e.clientY - state.theaterPanY;
      el.diagramTheaterViewport.style.cursor = 'grabbing';
    });
    window.addEventListener('mousemove', (e) => {
      if (!isTheaterPanning) return;
      state.theaterPanX = e.clientX - theaterStartX;
      state.theaterPanY = e.clientY - theaterStartY;
      updateDiagramTheaterTransform();
    });
    window.addEventListener('mouseup', () => {
      if (isTheaterPanning) {
        isTheaterPanning = false;
        if (el.diagramTheaterViewport) el.diagramTheaterViewport.style.cursor = 'grab';
      }
    });
    el.diagramTheaterViewport.addEventListener('wheel', (e) => {
      if (e.ctrlKey || e.metaKey) {
        e.preventDefault();
        const delta = e.deltaY < 0 ? 0.15 : -0.15;
        state.theaterZoom = Math.min(4.0, Math.max(0.4, +(state.theaterZoom + delta).toFixed(2)));
        updateDiagramTheaterTransform();
      }
    }, { passive: false });
  }

  // File Viewer Listeners (R-038)
  if (el.closeFileViewerBtn) {
    el.closeFileViewerBtn.addEventListener('click', closeFileViewer);
  }
  if (el.fileViewerWrapBtn) {
    el.fileViewerWrapBtn.addEventListener('click', () => {
      state.fileViewerWrap = !state.fileViewerWrap;
      el.fileViewerWrapBtn.classList.toggle('active', state.fileViewerWrap);
      el.fileViewerWrapBtn.innerHTML = state.fileViewerWrap ? '✓ Wrapped' : '↩ Wrap';
      const pre = el.fileViewerCodeContainer?.querySelector('.code-content-pre');
      if (pre) pre.classList.toggle('wrap-code', state.fileViewerWrap);
      showToast(state.fileViewerWrap ? 'Line wrap enabled' : 'Line wrap disabled');
    });
  }
  if (el.fileViewerCopyContentBtn) {
    el.fileViewerCopyContentBtn.addEventListener('click', () => {
      if (state.fileViewerData && state.fileViewerData.content) {
        copyToClipboard(state.fileViewerData.content, 'File content copied to clipboard!');
        const origHtml = el.fileViewerCopyContentBtn.innerHTML;
        el.fileViewerCopyContentBtn.innerHTML = '✓ Copied!';
        el.fileViewerCopyContentBtn.classList.add('btn-action-success');
        setTimeout(() => {
          el.fileViewerCopyContentBtn.innerHTML = origHtml;
          el.fileViewerCopyContentBtn.classList.remove('btn-action-success');
        }, 1500);
      }
    });
  }
  if (el.fileViewerCopyPathBtn) {
    el.fileViewerCopyPathBtn.addEventListener('click', () => {
      if (state.fileViewerData && state.fileViewerData.path) {
        copyToClipboard(state.fileViewerData.path, 'File path copied to clipboard!');
        const origHtml = el.fileViewerCopyPathBtn.innerHTML;
        el.fileViewerCopyPathBtn.innerHTML = '✓ Path Copied!';
        el.fileViewerCopyPathBtn.classList.add('btn-action-success');
        setTimeout(() => {
          el.fileViewerCopyPathBtn.innerHTML = origHtml;
          el.fileViewerCopyPathBtn.classList.remove('btn-action-success');
        }, 1500);
      }
    });
  }
  if (el.fileViewerPath) {
    el.fileViewerPath.addEventListener('click', () => {
      if (state.fileViewerData && state.fileViewerData.path) {
        copyToClipboard(state.fileViewerData.path, 'File path copied to clipboard!');
        el.fileViewerPath.classList.add('copied-flash');
        setTimeout(() => {
          el.fileViewerPath.classList.remove('copied-flash');
        }, 1500);
      }
    });
  }
  if (el.fileViewerModeControl) {
    el.fileViewerModeControl.querySelectorAll('.segment-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        state.fileViewerMode = btn.dataset.mode;
        renderFileViewerContent();
      });
    });
  }

  // In-File Find Listeners (R-038)
  if (el.fileViewerOpenFindBtn) {
    el.fileViewerOpenFindBtn.addEventListener('click', openFileFindWidget);
  }
  if (el.fileFindCloseBtn) {
    el.fileFindCloseBtn.addEventListener('click', closeFileFindWidget);
  }
  if (el.fileFindCaseBtn) {
    el.fileFindCaseBtn.addEventListener('click', toggleFileFindCase);
  }
  if (el.fileFindNextBtn) {
    el.fileFindNextBtn.addEventListener('click', goToNextFindMatch);
  }
  if (el.fileFindPrevBtn) {
    el.fileFindPrevBtn.addEventListener('click', goToPrevFindMatch);
  }
  if (el.fileFindInput) {
    el.fileFindInput.addEventListener('input', () => {
      executeFileFind();
    });
    el.fileFindInput.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        if (e.shiftKey) {
          goToPrevFindMatch();
        } else {
          goToNextFindMatch();
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        closeFileFindWidget();
      } else if (e.key === 'ArrowDown') {
        e.preventDefault();
        goToNextFindMatch();
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        goToPrevFindMatch();
      }
    });
  }

  // Session Files Modal Listeners
  if (el.detailFilesBtn) {
    el.detailFilesBtn.addEventListener('click', openSessionFilesModal);
  }
  if (el.closeSessionFilesModalBtn) {
    el.closeSessionFilesModalBtn.addEventListener('click', closeSessionFilesModal);
  }

  // Agent Matrix Listeners (R-039)
  initAgentMatrixListeners();
}

// Local date helper to prevent timezone skew (e.g. UTC shifting into tomorrow)
function formatLocalDate(d) {
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

// Date Presets Logic (R-010)
function applyDatePreset(preset) {
  if (el.datePresets) {
    el.datePresets.forEach(p => {
      p.classList.toggle('active', p.dataset.preset === preset);
    });
  }

  const now = new Date();
  let from = '';
  let to = '';

  switch (preset) {
    case 'today':
      from = formatLocalDate(now);
      to = formatLocalDate(now);
      break;
    case '7d': {
      const past = new Date(now);
      past.setDate(past.getDate() - 7);
      from = formatLocalDate(past);
      to = formatLocalDate(now);
      break;
    }
    case '30d': {
      const past = new Date(now);
      past.setDate(past.getDate() - 30);
      from = formatLocalDate(past);
      to = formatLocalDate(now);
      break;
    }
    case 'all':
    default:
      from = '';
      to = '';
      break;
  }

  state.dateFrom = from;
  state.dateTo = to;
  if (el.dateFromInput) el.dateFromInput.value = from;
  if (el.dateToInput) el.dateToInput.value = to;
  loadSessions();
}

// API Calls
async function loadSessions() {
  try {
    const params = new URLSearchParams();
    const effectiveWorkspace = state.filterWorkspace || (state.groupBy === 'project' && state.drilledProject ? state.drilledProject : '');
    if (effectiveWorkspace) params.append('workspace', effectiveWorkspace);
    if (state.filterRuntime) params.append('runtime', state.filterRuntime);
    if (state.sortBy) params.append('sort_by', state.sortBy);
    if (state.sortOrder) params.append('sort_order', state.sortOrder);
    if (state.dateFrom) params.append('from', state.dateFrom);
    if (state.dateTo) params.append('to', state.dateTo);

    const res = await fetch(`/api/sessions?${params.toString()}`);
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    const data = await res.json();

    state.sessions = data.sessions || [];
    state.stats = data.stats || null;

    renderStats();
    renderWorkspaceOptions();
    renderView();
  } catch (err) {
    console.error('Failed to load sessions:', err);
    showToast('❌ Error loading sessions');
  }
}

// ============================================================================
//  Live Reactive Telemetry Sync & Holographic Pulse Animation (R-044)
// ============================================================================
function pulseCounterBadge(badgeEl) {
  if (!badgeEl) return;
  badgeEl.classList.remove('updated');
  // Trigger reflow to restart CSS keyframe animation
  void badgeEl.offsetWidth;
  badgeEl.classList.add('updated');
  setTimeout(() => {
    badgeEl.classList.remove('updated');
  }, 650);
}

function updateAgentContextBadge(ctx, forcePulse = false) {
  if (!ctx || !el.detailAgentContextCount) return;
  const activeRules = ctx.active_rules ? ctx.active_rules.length : 0;
  const activeSkills = ctx.active_skills ? ctx.active_skills.length : 0;
  let activeMCP = 0;
  (ctx.mcp_servers || []).forEach(srv => {
    activeMCP += (srv.active_tools || []).length;
  });
  const nativeTools = ctx.native_tools ? ctx.native_tools.length : 0;
  const subagents = ctx.subagents ? ctx.subagents.length : 0;
  const totalActive = activeRules + activeSkills + activeMCP + nativeTools + subagents;

  const prevText = el.detailAgentContextCount.textContent.trim();
  const prevVal = parseInt(prevText, 10);
  el.detailAgentContextCount.textContent = totalActive.toString();

  if (el.detailAgentContextBtn) {
    el.detailAgentContextBtn.title = `${totalActive} active agent elements (${activeRules} rules • ${activeSkills} skills • ${activeMCP} MCP tools • ${nativeTools} native tools • ${subagents} subagents)`;
  }

  if (forcePulse || (!isNaN(prevVal) && prevVal !== totalActive)) {
    pulseCounterBadge(el.detailAgentContextCount);
  }
}

let liveContextSyncTimeout = null;
function triggerLiveContextSync(sessionId) {
  if (!sessionId) return;
  if (liveContextSyncTimeout) {
    clearTimeout(liveContextSyncTimeout);
  }
  liveContextSyncTimeout = setTimeout(() => {
    fetch(`/api/sessions/${sessionId}/context`)
      .then(r => r.ok ? r.json() : null)
      .then(ctx => {
        if (!ctx) return;
        state.sessionContextData = ctx;
        state.sessionMatrix.context = ctx;
        updateAgentContextBadge(ctx);
        if (el.sessionContextModal && el.sessionContextModal.classList.contains('active')) {
          renderSessionMatrix();
        }
      })
      .catch(e => console.warn('Live context sync error:', e));
  }, 350);
}

async function loadSessionDetail(sessionId) {
  try {
    const res = await fetch(`/api/sessions/${sessionId}`);
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    const data = await res.json();

    state.selectedSessionId = sessionId;
    state.selectedSessionDetail = data;

    showDetailView(data);

    // Fetch session agent context in background (R-039)
    fetch(`/api/sessions/${sessionId}/context`)
      .then(r => r.ok ? r.json() : null)
      .then(ctx => {
        if (!ctx) return;
        state.sessionContextData = ctx;
        state.sessionMatrix.context = ctx;
        updateAgentContextBadge(ctx, false);
      })
      .catch(e => console.warn('Context fetch error:', e));
  } catch (err) {
    console.error('Failed to load session detail:', err);
    showToast(`❌ Error loading session detail: ${err.message}`);
  }
}

async function executeSearch(query) {
  if (!query) {
    el.modalSearchResults.innerHTML = '<div class="empty-state"><div class="empty-state-icon">🔍</div><p>Type keywords to search across all prompt transcripts</p></div>';
    return;
  }

  try {
    const params = new URLSearchParams({ q: query });
    const effectiveWorkspace = state.filterWorkspace || (state.groupBy === 'project' && state.drilledProject ? state.drilledProject : '');
    if (effectiveWorkspace) params.append('workspace', effectiveWorkspace);

    const res = await fetch(`/api/search?${params.toString()}`);
    if (!res.ok) throw new Error(`Search error: ${res.status}`);
    const data = await res.json();

    renderSearchResults(data.results || [], query);
  } catch (err) {
    console.error('Search failed:', err);
    el.modalSearchResults.innerHTML = '<div class="empty-state"><p>Search error occurred.</p></div>';
  }
}

// Stats Strip Rendering (R-017)
function renderStats() {
  if (!state.stats || !el.statsStrip) return;
  const s = state.stats;

  el.statsStrip.innerHTML = `
    <div class="stat-card">
      <div class="stat-icon" style="background: rgba(99, 102, 241, 0.15); color: #818cf8;">📂</div>
      <div class="stat-content">
        <span class="stat-val">${s.total_sessions || 0}</span>
        <span class="stat-lbl">Total Sessions</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon" style="background: rgba(168, 85, 247, 0.15); color: #c084fc;">👤</div>
      <div class="stat-content">
        <span class="stat-val">${s.total_user_messages || 0}</span>
        <span class="stat-lbl">User Prompts</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon" style="background: rgba(6, 182, 212, 0.15); color: #22d3ee;">🤖</div>
      <div class="stat-content">
        <span class="stat-val">${s.total_agent_responses || 0}</span>
        <span class="stat-lbl">Agent Responses</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon" style="background: rgba(16, 185, 129, 0.15); color: #34d399;">⚡</div>
      <div class="stat-content">
        <span class="stat-val">${s.total_tool_calls || 0}</span>
        <span class="stat-lbl">Tool Calls</span>
      </div>
    </div>
    <div class="stat-card">
      <div class="stat-icon" style="background: rgba(245, 158, 11, 0.15); color: #fbbf24;">📁</div>
      <div class="stat-content">
        <span class="stat-val">${(s.projects || []).length || (s.workspaces || []).length}</span>
        <span class="stat-lbl">Repositories</span>
      </div>
    </div>
  `;
}

// Workspace Filter & Strip (R-011, R-028)
function renderWorkspaceOptions() {
  if (!state.stats || !state.stats.workspaces) return;

  // Strict client-side deduplication & validation (defense-in-depth)
  const rawWorkspaces = state.stats.workspaces || [];
  const seen = new Set();
  const workspaces = [];
  rawWorkspaces.forEach(ws => {
    if (!ws) return;
    const clean = String(ws).trim();
    if (clean === '' || clean === 'workspace' || clean === 'aparv') return;
    if (clean.includes('.') || clean.includes('/') || clean.includes('\\')) return;
    if (clean.startsWith('.') || clean.startsWith('_')) return;
    const key = clean.toLowerCase();
    if (!seen.has(key)) {
      seen.add(key);
      workspaces.push(clean);
    }
  });

  workspaces.sort((a, b) => a.localeCompare(b, undefined, { sensitivity: 'base' }));

  const currentVal = state.filterWorkspace || (state.groupBy === 'project' && state.drilledProject ? state.drilledProject : '') || '';

  // Populate Dropdown
  if (el.workspaceSelect) {
    const existingOptions = Array.from(el.workspaceSelect.options).map(o => o.value).filter(Boolean);
    const optionsChanged = existingOptions.length !== workspaces.length ||
      existingOptions.some((val, i) => val !== workspaces[i]);

    if (optionsChanged || el.workspaceSelect.options.length <= 1) {
      let html = '<option value="">All Workspaces / Repositories</option>';
      workspaces.forEach(ws => {
        html += `<option value="${escapeHtml(ws)}">${escapeHtml(ws)}</option>`;
      });
      el.workspaceSelect.innerHTML = html;
    }
    el.workspaceSelect.value = currentVal;
  }

  // Populate Strip Pills
  if (el.workspaceStrip) {
    let stripHtml = `<button class="chip ${currentVal === '' ? 'active' : ''}" data-ws="">All (${state.stats.total_sessions || 0})</button>`;
    workspaces.slice(0, 12).forEach(ws => {
      const active = (currentVal === ws) ? 'active' : '';
      stripHtml += `<button class="chip ${active}" data-ws="${escapeHtml(ws)}">${escapeHtml(ws)}</button>`;
    });
    el.workspaceStrip.innerHTML = stripHtml;

    el.workspaceStrip.querySelectorAll('.chip').forEach(btn => {
      btn.addEventListener('click', () => {
        state.filterWorkspace = btn.dataset.ws;
        if (state.groupBy === 'project') {
          state.drilledProject = btn.dataset.ws || null;
        }
        if (el.workspaceSelect) el.workspaceSelect.value = state.filterWorkspace;
        loadSessions();
      });
    });
  }
}

// Multi-mode View Dispatcher (R-028, R-029)
function renderView() {
  if (state.currentView === 'detail') return;

  // Sync Group By control button state
  if (el.groupByBtns) {
    el.groupByBtns.forEach(b => {
      b.classList.toggle('active', b.dataset.group === state.groupBy);
    });
  }

  // Sync Runtime control button state (R-036)
  if (el.runtimeBtns) {
    el.runtimeBtns.forEach(b => {
      b.classList.toggle('active', (b.dataset.runtime || '') === state.filterRuntime);
    });
  }

  switch (state.groupBy) {
    case 'project':
      renderProjectHub();
      break;
    case 'date':
      renderDateGroupedView();
      break;
    case 'flat':
    default:
      renderFlatGridView();
      break;
  }
}

// Mode 1: Repository / Project Clustered Hub (R-028)
function renderProjectHub() {
  const container = el.viewContainer || el.sessionGrid;
  if (!container) return;

  // Case A: User has drilled down into a specific repository
  if (state.drilledProject) {
    const projName = state.drilledProject;
    const projSessions = state.sessions.filter(s =>
      s.project_name === projName || (s.projects_touched && s.projects_touched.includes(projName))
    );

    const projSummary = (state.stats && state.stats.projects) ?
      state.stats.projects.find(p => p.name === projName) : null;
    const isGitRepo = projSummary ? projSummary.is_git : (projName !== 'Default');
    const projPath = projSummary ? projSummary.path : (projName === 'Default' ? '~' : `/Users/aparv/Library/CloudStorage/OneDrive-Personal/G-Drive/Interviews/knowledge/${projName}`);
    const totalSteps = projSessions.reduce((acc, s) => acc + s.step_count, 0);
    const totalPrompts = projSessions.reduce((acc, s) => acc + s.user_message_count, 0);
    const totalTools = projSessions.reduce((acc, s) => acc + s.tool_call_count, 0);

    // Filter and sort projSessions based on state.projectFilters
    function getFilteredAndSortedProjSessions() {
      let list = [...projSessions];

      // 1. In-Project Search (independent from global search)
      const q = (state.projectFilters.search || '').trim().toLowerCase();
      if (q) {
        list = list.filter(s => {
          const idMatch = (s.id || '').toLowerCase().includes(q);
          const promptMatch = (s.first_user_prompt_preview || '').toLowerCase().includes(q);
          const wsMatch = (s.workspace || '').toLowerCase().includes(q);
          const touchedMatch = (s.projects_touched || []).some(pt => pt.toLowerCase().includes(q));
          let toolMatch = false;
          if (s.tool_calls_map) {
            toolMatch = Object.keys(s.tool_calls_map).some(k => k.toLowerCase().includes(q));
          }
          return idMatch || promptMatch || wsMatch || touchedMatch || toolMatch;
        });
      }

      // 2. In-Project Date Filter (inclusive lifetime overlap)
      if (state.projectFilters.dateFrom || state.projectFilters.dateTo) {
        const fromTime = state.projectFilters.dateFrom ? new Date(state.projectFilters.dateFrom + 'T00:00:00Z').getTime() : 0;
        const toTime = state.projectFilters.dateTo ? new Date(state.projectFilters.dateTo + 'T23:59:59.999Z').getTime() : Infinity;

        list = list.filter(s => {
          const cTime = new Date(s.created_at).getTime();
          const lTime = s.last_message_at ? new Date(s.last_message_at).getTime() : cTime;
          return (lTime >= fromTime) && (cTime <= toTime);
        });
      }

      // 3. In-Project Sort
      const pSort = state.projectFilters.sort || 'date_desc';
      list.sort((a, b) => {
        switch (pSort) {
          case 'date_asc':
            return new Date(a.created_at) - new Date(b.created_at);
          case 'steps_desc':
            return (b.step_count || 0) - (a.step_count || 0);
          case 'messages_desc':
            return (b.user_message_count || 0) - (a.user_message_count || 0);
          case 'tools_desc':
            return (b.tool_call_count || 0) - (a.tool_call_count || 0);
          case 'size_desc':
            return (b.transcript_size_bytes || 0) - (a.transcript_size_bytes || 0);
          case 'date_desc':
          default:
            return new Date(b.created_at) - new Date(a.created_at);
        }
      });

      return list;
    }

    const filtered = getFilteredAndSortedProjSessions();
    const hasActiveFilters = Boolean(
      state.projectFilters.search ||
      state.projectFilters.dateFrom ||
      state.projectFilters.dateTo ||
      state.projectFilters.datePreset !== 'all' ||
      state.projectFilters.sort !== 'date_desc'
    );

    // Show Hero Banner & In-Project Session Stack Toolbar
    if (el.projectHeroBanner) {
      el.projectHeroBanner.style.display = 'flex';
      el.projectHeroBanner.innerHTML = `
        <div class="project-hero-top">
          <div class="project-hero-left">
            <button class="back-btn" id="projectHeroBackBtn"><span>←</span> All Repositories</button>
            <div class="project-hero-info">
              <div class="project-hero-title">
                📁 ${escapeHtml(projName)}
                <span class="git-badge ${isGitRepo ? '' : 'non-git'}">${isGitRepo ? 'git repo' : 'folder'}</span>
              </div>
              <div class="project-hero-path">${escapeHtml(projPath)}</div>
            </div>
          </div>
          <div class="project-hero-stats">
            <span class="chip active">📂 ${projSessions.length} Total Sessions</span>
            <span class="chip">👤 ${totalPrompts} Prompts</span>
            <span class="chip">⚡ ${totalTools} Tool Calls</span>
            <button class="btn-pill" id="projectAgentContextBtn" style="background: rgba(99, 102, 241, 0.2); border-color: rgba(99, 102, 241, 0.4); color: var(--accent-indigo); cursor: pointer;" title="Inspect rules, skills, and instructions for ${escapeHtml(projName)}">
              🧠 Agent Context
            </button>
          </div>
        </div>

        <!-- Dedicated In-Project Session Stack Toolbar (R-029) -->
        <div class="project-session-toolbar">
          <div class="project-search-box">
            <span class="search-icon">🔍</span>
            <input type="text" id="projectSearchInput" class="project-search-input" placeholder="Search sessions in ${escapeHtml(projName)}..." value="${escapeHtml(state.projectFilters.search || '')}">
            <button class="clear-search-btn" id="clearProjectSearchBtn" title="Clear in-project search" style="display: ${state.projectFilters.search ? 'inline-flex' : 'none'};">✕</button>
          </div>

          <div class="filter-group">
            <span class="filter-label">Sort:</span>
            <select id="projectSortSelect" class="select-custom" aria-label="Sort sessions in this project">
              <option value="date_desc" ${state.projectFilters.sort === 'date_desc' ? 'selected' : ''}>Most Recent</option>
              <option value="date_asc" ${state.projectFilters.sort === 'date_asc' ? 'selected' : ''}>Oldest First</option>
              <option value="steps_desc" ${state.projectFilters.sort === 'steps_desc' ? 'selected' : ''}>Most Steps</option>
              <option value="messages_desc" ${state.projectFilters.sort === 'messages_desc' ? 'selected' : ''}>Most Prompts</option>
              <option value="tools_desc" ${state.projectFilters.sort === 'tools_desc' ? 'selected' : ''}>Most Tool Calls</option>
              <option value="size_desc" ${state.projectFilters.sort === 'size_desc' ? 'selected' : ''}>Largest Size</option>
            </select>
          </div>

          <div class="filter-group project-date-group">
            <span class="filter-label">Date:</span>
            <input type="date" id="projectDateFrom" class="date-input" value="${state.projectFilters.dateFrom || ''}" aria-label="Start date">
            <span style="color: var(--text-muted); font-size: 0.8rem;">to</span>
            <input type="date" id="projectDateTo" class="date-input" value="${state.projectFilters.dateTo || ''}" aria-label="End date">
            <div class="project-date-presets">
              <button class="chip date-preset-chip ${state.projectFilters.datePreset === 'all' ? 'active' : ''}" data-preset="all">All Time</button>
              <button class="chip date-preset-chip ${state.projectFilters.datePreset === 'today' ? 'active' : ''}" data-preset="today">Today</button>
              <button class="chip date-preset-chip ${state.projectFilters.datePreset === '7d' ? 'active' : ''}" data-preset="7d">7 Days</button>
              <button class="chip date-preset-chip ${state.projectFilters.datePreset === '30d' ? 'active' : ''}" data-preset="30d">30 Days</button>
            </div>
          </div>

          <div class="project-toolbar-summary">
            <span class="project-count-badge" id="projectSessionCountBadge">
              Showing ${filtered.length} of ${projSessions.length} session${projSessions.length === 1 ? '' : 's'}
            </span>
            <button class="btn-ghost-sm" id="resetProjectFiltersBtn" style="display: ${hasActiveFilters ? 'inline-block' : 'none'};">Reset Filters</button>
          </div>
        </div>
      `;

      // Back Button Listener
      document.getElementById('projectHeroBackBtn')?.addEventListener('click', () => {
        state.drilledProject = null;
        state.filterWorkspace = '';
        state.projectFilters = { search: '', sort: 'date_desc', dateFrom: '', dateTo: '', datePreset: 'all' };
        if (el.workspaceSelect) el.workspaceSelect.value = '';
        window.location.hash = '#/';
      });

      // Project Agent Context Listener
      document.getElementById('projectAgentContextBtn')?.addEventListener('click', () => {
        openAgentBrainModal(projName);
      });

      // In-Project Search Input Listener (Reactive, maintains input focus)
      const pSearchInput = document.getElementById('projectSearchInput');
      const clearSearchBtn = document.getElementById('clearProjectSearchBtn');

      if (pSearchInput) {
        pSearchInput.addEventListener('input', (e) => {
          state.projectFilters.search = e.target.value;
          if (clearSearchBtn) {
            clearSearchBtn.style.display = state.projectFilters.search ? 'inline-flex' : 'none';
          }
          const updated = getFilteredAndSortedProjSessions();
          renderProjectSessionGrid(updated);
          updateProjectToolbarSummary(updated.length, projSessions.length);
        });
      }

      if (clearSearchBtn) {
        clearSearchBtn.addEventListener('click', () => {
          state.projectFilters.search = '';
          if (pSearchInput) {
            pSearchInput.value = '';
            pSearchInput.focus();
          }
          clearSearchBtn.style.display = 'none';
          const updated = getFilteredAndSortedProjSessions();
          renderProjectSessionGrid(updated);
          updateProjectToolbarSummary(updated.length, projSessions.length);
        });
      }

      // In-Project Sort Select Listener
      document.getElementById('projectSortSelect')?.addEventListener('change', (e) => {
        state.projectFilters.sort = e.target.value;
        const updated = getFilteredAndSortedProjSessions();
        renderProjectSessionGrid(updated);
        updateProjectToolbarSummary(updated.length, projSessions.length);
      });

      // In-Project Date Pickers
      const pDateFrom = document.getElementById('projectDateFrom');
      const pDateTo = document.getElementById('projectDateTo');

      const onDateChange = () => {
        state.projectFilters.dateFrom = pDateFrom?.value || '';
        state.projectFilters.dateTo = pDateTo?.value || '';
        state.projectFilters.datePreset = '';
        document.querySelectorAll('.project-date-presets .date-preset-chip').forEach(c => c.classList.remove('active'));
        const updated = getFilteredAndSortedProjSessions();
        renderProjectSessionGrid(updated);
        updateProjectToolbarSummary(updated.length, projSessions.length);
      };

      pDateFrom?.addEventListener('change', onDateChange);
      pDateTo?.addEventListener('change', onDateChange);
      pDateFrom?.addEventListener('input', onDateChange);
      pDateTo?.addEventListener('input', onDateChange);

      // In-Project Date Preset Chips
      document.querySelectorAll('.project-date-presets .date-preset-chip').forEach(btn => {
        btn.addEventListener('click', () => {
          document.querySelectorAll('.project-date-presets .date-preset-chip').forEach(c => c.classList.remove('active'));
          btn.classList.add('active');
          const preset = btn.dataset.preset;
          state.projectFilters.datePreset = preset;
          const today = new Date();
          if (preset === 'all') {
            state.projectFilters.dateFrom = '';
            state.projectFilters.dateTo = '';
          } else if (preset === 'today') {
            const todayStr = formatLocalDate(today);
            state.projectFilters.dateFrom = todayStr;
            state.projectFilters.dateTo = todayStr;
          } else if (preset === '7d') {
            const past = new Date(today);
            past.setDate(past.getDate() - 7);
            state.projectFilters.dateFrom = formatLocalDate(past);
            state.projectFilters.dateTo = formatLocalDate(today);
          } else if (preset === '30d') {
            const past = new Date(today);
            past.setDate(past.getDate() - 30);
            state.projectFilters.dateFrom = formatLocalDate(past);
            state.projectFilters.dateTo = formatLocalDate(today);
          }
          if (pDateFrom) pDateFrom.value = state.projectFilters.dateFrom;
          if (pDateTo) pDateTo.value = state.projectFilters.dateTo;
          const updated = getFilteredAndSortedProjSessions();
          renderProjectSessionGrid(updated);
          updateProjectToolbarSummary(updated.length, projSessions.length);
        });
      });

      // Reset Project Filters Button
      document.getElementById('resetProjectFiltersBtn')?.addEventListener('click', () => {
        state.projectFilters = { search: '', sort: 'date_desc', dateFrom: '', dateTo: '', datePreset: 'all' };
        renderProjectHub();
      });
    }

    function renderProjectSessionGrid(list) {
      if (list.length === 0) {
        container.innerHTML = `
          <div class="empty-state">
            <div class="empty-state-icon">📂</div>
            <h3>No sessions match project filters in ${escapeHtml(projName)}</h3>
            <p>Try adjusting your search query, sort option, or date filter.</p>
            <button class="btn-pill" id="emptyResetBtn" style="margin-top: 1rem;">Reset Project Filters</button>
          </div>
        `;
        document.getElementById('emptyResetBtn')?.addEventListener('click', () => {
          state.projectFilters = { search: '', sort: 'date_desc', dateFrom: '', dateTo: '', datePreset: 'all' };
          renderProjectHub();
        });
      } else {
        container.innerHTML = `<div class="session-grid">${renderSessionCardsHtml(list)}</div>`;
        attachSessionCardClicks(container);
      }
    }

    function updateProjectToolbarSummary(filteredCount, totalCount) {
      const badge = document.getElementById('projectSessionCountBadge');
      if (badge) {
        badge.textContent = `Showing ${filteredCount} of ${totalCount} session${totalCount === 1 ? '' : 's'}`;
      }
      if (el.sessionCountBadge) {
        el.sessionCountBadge.textContent = `${filteredCount} session${filteredCount === 1 ? '' : 's'} in ${projName}`;
      }
      const resetBtn = document.getElementById('resetProjectFiltersBtn');
      const hasFilters = Boolean(
        state.projectFilters.search ||
        state.projectFilters.dateFrom ||
        state.projectFilters.dateTo ||
        state.projectFilters.datePreset !== 'all' ||
        state.projectFilters.sort !== 'date_desc'
      );
      if (resetBtn) {
        resetBtn.style.display = hasFilters ? 'inline-block' : 'none';
      }
    }

    if (el.sessionCountBadge) {
      el.sessionCountBadge.textContent = `${filtered.length} session${filtered.length === 1 ? '' : 's'} in ${projName}`;
    }

    renderProjectSessionGrid(filtered);
    return;
  }

  // Case B: Clustered Repository Hub Grid
  if (el.projectHeroBanner) {
    el.projectHeroBanner.style.display = 'none';
  }

  const rawProjects = (state.stats && state.stats.projects) ? state.stats.projects : [];

  // Dynamically map sessions matching active filters to projects
  const sessionsByProj = new Map();
  state.sessions.forEach(s => {
    const pName = s.project_name || 'Default';
    if (!sessionsByProj.has(pName)) {
      sessionsByProj.set(pName, []);
    }
    sessionsByProj.get(pName).push(s);
  });

  const filterActive = Boolean(state.dateFrom || state.dateTo || state.filterWorkspace);

  // Combine project metadata with active filtered session counts
  let projects = rawProjects.map(p => {
    const pSessions = sessionsByProj.get(p.name) || [];
    const countInFilter = pSessions.length;
    return {
      ...p,
      display_session_count: filterActive ? countInFilter : p.session_count,
      count_in_filter: countInFilter,
      has_active_sessions: countInFilter > 0,
      filtered_sessions: pSessions
    };
  });

  // Filter by workspace if selected
  if (state.filterWorkspace) {
    projects = projects.filter(p => p.name.toLowerCase().includes(state.filterWorkspace.toLowerCase()));
  }

  // Sort projects dynamically according to active sort criteria
  projects.sort((a, b) => {
    // If date/workspace filter is active, prioritize projects with active sessions
    if (filterActive) {
      if (a.has_active_sessions && !b.has_active_sessions) return -1;
      if (!a.has_active_sessions && b.has_active_sessions) return 1;
    }

    switch (state.sortBy) {
      case 'steps':
        return b.total_steps - a.total_steps;
      case 'messages':
        return b.total_user_messages - a.total_user_messages;
      case 'size':
        return (b.total_steps + b.total_tool_calls) - (a.total_steps + a.total_tool_calls);
      case 'date':
      default:
        if (state.sortOrder === 'asc') {
          return new Date(a.last_active_at || 0) - new Date(b.last_active_at || 0);
        }
        return new Date(b.last_active_at || 0) - new Date(a.last_active_at || 0);
    }
  });

  if (el.sessionCountBadge) {
    const activeCount = projects.filter(p => p.display_session_count > 0).length;
    if (filterActive) {
      el.sessionCountBadge.textContent = `${state.sessions.length} sessions in ${activeCount} active repos`;
    } else {
      el.sessionCountBadge.textContent = `${activeCount} Active Repos (${projects.length} Total)`;
    }
  }

  if (projects.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <div class="empty-state-icon">📂</div>
        <h3>No repositories found</h3>
        <p>No project repositories matched your current filter.</p>
      </div>
    `;
    return;
  }

  let html = '<div class="project-grid">';
  projects.forEach(p => {
    const relTime = p.last_active_at ? formatRelativeTime(p.last_active_at) : 'No recorded activity';

    let promptsHtml = '';
    if (p.recent_prompts && p.recent_prompts.length > 0) {
      promptsHtml = p.recent_prompts.map(pr => `
        <div class="project-prompt-chip" title="${escapeHtml(pr)}">
          <span class="prompt-icon">💬</span> ${escapeHtml(pr)}
        </div>
      `).join('');
    } else {
      promptsHtml = `<div class="project-prompt-chip" style="opacity: 0.5;">No user prompts recorded</div>`;
    }

    let sessionBadge = '';
    if (filterActive) {
      if (p.has_active_sessions) {
        sessionBadge = `<span class="badge" style="background: rgba(99, 102, 241, 0.2); color: #818cf8; border: 1px solid rgba(99, 102, 241, 0.4); padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.72rem;">✨ ${p.count_in_filter} of ${p.session_count} in filter</span>`;
      } else {
        sessionBadge = `<span class="badge" style="background: rgba(148, 163, 184, 0.12); color: #94a3b8; border: 1px solid rgba(148, 163, 184, 0.25); padding: 0.2rem 0.5rem; border-radius: 4px; font-size: 0.72rem;">0 in filter</span>`;
      }
    }

    html += `
      <div class="project-card ${p.has_active_sessions ? 'active-filter-card' : ''}" data-project="${escapeHtml(p.name)}">
        <div>
          <div class="project-card-top">
            <div class="project-card-title">
              📁 ${escapeHtml(p.name)}
            </div>
            <div style="display: flex; gap: 0.4rem; align-items: center;">
              ${sessionBadge}
              <span class="git-badge ${p.is_git ? '' : 'non-git'}">${p.is_git ? 'git repo' : 'folder'}</span>
            </div>
          </div>
          <div class="project-card-path" title="${escapeHtml(p.path)}">${escapeHtml(p.path)}</div>

          <div class="project-prompts-preview" style="margin-top: 0.85rem;">
            ${promptsHtml}
          </div>
        </div>

        <div>
          <div class="card-metrics" style="margin-bottom: 0.75rem;">
            <span class="metric-item" title="Total conversation sessions in this repository">📂 ${p.session_count} session${p.session_count === 1 ? '' : 's'}</span>
            <span class="metric-item" title="User prompts in this repository">👤 ${p.total_user_messages} prompts</span>
            <span class="metric-item" title="Tool calls invoked in this repository">⚡ ${p.total_tool_calls} tools</span>
            <span class="metric-item" title="Total steps executed">🪜 ${p.total_steps} steps</span>
          </div>
          <div class="project-card-footer">
            <span>🕒 ${relTime}</span>
            <span class="project-explore-btn">Explore Sessions & Prompts →</span>
          </div>
        </div>
      </div>
    `;
  });
  html += '</div>';

  container.innerHTML = html;

  // Project Card Click Handlers
  container.querySelectorAll('.project-card').forEach(card => {
    card.addEventListener('click', () => {
      const proj = card.dataset.project;
      window.location.hash = `#/project/${encodeURIComponent(proj)}`;
    });
  });
}

// Mode 2: Date Grouped View (R-029)
function renderDateGroupedView() {
  if (el.projectHeroBanner) {
    el.projectHeroBanner.style.display = 'none';
  }
  const container = el.viewContainer || el.sessionGrid;
  if (!container) return;

  const sessions = state.sessions;
  if (el.sessionCountBadge) {
    el.sessionCountBadge.textContent = `${sessions.length} session${sessions.length === 1 ? '' : 's'}`;
  }

  if (sessions.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <div class="empty-state-icon">📂</div>
        <h3>No sessions found</h3>
        <p>No Antigravity IDE sessions matched your filters.</p>
      </div>
    `;
    return;
  }

  const now = new Date();
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
  const startOfYesterday = new Date(startOfToday);
  startOfYesterday.setDate(startOfYesterday.getDate() - 1);
  const startOfWeek = new Date(startOfToday);
  startOfWeek.setDate(startOfWeek.getDate() - 7);

  const buckets = {
    today: { title: '🌟 Today', sessions: [] },
    yesterday: { title: '⚡ Yesterday', sessions: [] },
    thisWeek: { title: '📅 Past 7 Days', sessions: [] },
    older: { title: '🗓️ Older', sessions: [] }
  };

  sessions.forEach(s => {
    const d = new Date(s.created_at);
    if (d >= startOfToday) {
      buckets.today.sessions.push(s);
    } else if (d >= startOfYesterday) {
      buckets.yesterday.sessions.push(s);
    } else if (d >= startOfWeek) {
      buckets.thisWeek.sessions.push(s);
    } else {
      buckets.older.sessions.push(s);
    }
  });

  let html = '';
  for (const key of ['today', 'yesterday', 'thisWeek', 'older']) {
    const b = buckets[key];
    if (b.sessions.length === 0) continue;

    html += `
      <div class="date-group-section">
        <div class="date-group-header">
          <span>${b.title}</span>
          <span class="date-group-badge">${b.sessions.length} session${b.sessions.length === 1 ? '' : 's'}</span>
        </div>
        <div class="session-grid">
          ${renderSessionCardsHtml(b.sessions)}
        </div>
      </div>
    `;
  }

  container.innerHTML = html;
  attachSessionCardClicks(container);
}

// Mode 3: Flat Grid View (R-008, R-018)
function renderFlatGridView() {
  if (el.projectHeroBanner) {
    el.projectHeroBanner.style.display = 'none';
  }
  const container = el.viewContainer || el.sessionGrid;
  if (!container) return;

  const sessions = state.sessions;
  if (el.sessionCountBadge) {
    el.sessionCountBadge.textContent = `${sessions.length} session${sessions.length === 1 ? '' : 's'}`;
  }

  if (sessions.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <div class="empty-state-icon">📂</div>
        <h3>No sessions found</h3>
        <p>No Antigravity IDE sessions matched your filters.</p>
      </div>
    `;
    return;
  }

  container.innerHTML = `<div class="session-grid">${renderSessionCardsHtml(sessions)}</div>`;
  attachSessionCardClicks(container);
}

// Helper: Get runtime badge HTML for a session (R-036)
function getRuntimeBadgeHtml(s) {
  const rm = (s.runtime_mode || 'ide').toLowerCase();
  const prov = (s.provider || 'antigravity').toLowerCase();
  if (rm === 'cli') {
    return `<span class="runtime-badge runtime-cli" title="Antigravity CLI">⚡ CLI</span>`;
  }
  if (rm === 'core') {
    return `<span class="runtime-badge runtime-core" title="Antigravity Core">⚙️ Core</span>`;
  }
  if (prov === 'claude') {
    return `<span class="runtime-badge runtime-claude" title="Claude Code">🤖 Claude</span>`;
  }
  if (prov === 'copilot') {
    return `<span class="runtime-badge runtime-copilot" title="GitHub Copilot">🐙 Copilot</span>`;
  }
  return `<span class="runtime-badge runtime-ide" title="Antigravity IDE">💻 IDE</span>`;
}

// Helper: Render session cards HTML
function renderSessionCardsHtml(sessions) {
  return sessions.map(s => {
    const timeStr = formatDateTime(s.created_at);
    const relTime = formatRelativeTime(s.created_at);
    const preview = s.first_user_prompt_preview || 'No explicit user prompt';
    const sizeKB = Math.round((s.transcript_size_bytes || 0) / 1024);

    return `
      <div class="session-card" data-id="${s.id}">
        <div>
          <div class="card-top">
            <div style="display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap;">
              <span class="workspace-badge" title="${escapeHtml(s.workspace || s.project_name)}">
                📁 ${escapeHtml(s.project_name || 'Default')}
              </span>
              ${getRuntimeBadgeHtml(s)}
            </div>
            <span class="session-time" title="${timeStr}">${relTime}</span>
          </div>
          <div class="prompt-snippet" title="${escapeHtml(preview)}">
            "${escapeHtml(preview)}"
          </div>
        </div>
        <div class="card-metrics">
          <span class="metric-item" title="Total steps in session">🪜 ${s.step_count}</span>
          <span class="metric-item" title="User prompts">👤 ${s.user_message_count}</span>
          <span class="metric-item" title="Agent responses">🤖 ${s.agent_response_count}</span>
          <span class="metric-item" title="Tool calls">⚡ ${s.tool_call_count}</span>
          <span class="metric-item" style="margin-left: auto;" title="Transcript file size">💾 ${sizeKB}KB</span>
        </div>
      </div>
    `;
  }).join('');
}

// Helper: Attach card clicks
function attachSessionCardClicks(container) {
  container.querySelectorAll('.session-card').forEach(card => {
    card.addEventListener('click', () => {
      window.location.hash = `#/session/${card.dataset.id}`;
    });
  });
}

// Detail View (R-009, R-015, R-016)
function showDetailView(detail) {
  state.currentView = 'detail';
  if (el.dashboardView) el.dashboardView.style.display = 'none';
  if (el.detailView) el.detailView.classList.add('active');

  const s = detail.session;
  state.selectedSessionId = s.id;
  state.selectedSessionDetail = detail;
  state.sessionStepSearch = '';
  state.sessionStepSort = 'step_desc';
  state.messageFilter = 'ALL';
  state.followStream = true;
  state.pendingStepsBuffer = [];
  if (el.toggleFollowStreamBtn) el.toggleFollowStreamBtn.classList.add('active');
  if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';

  if (el.sessionStepSearchInput) el.sessionStepSearchInput.value = '';
  if (el.clearSessionStepSearchBtn) el.clearSessionStepSearchBtn.style.display = 'none';
  if (el.sessionStepSortSelect) el.sessionStepSortSelect.value = 'step_desc';
  if (el.resetSessionStepFiltersBtn) el.resetSessionStepFiltersBtn.style.display = 'none';

  document.querySelectorAll('.msg-filter-chip').forEach(c => {
    if (c.dataset.filter === 'ALL') c.classList.add('active');
    else c.classList.remove('active');
  });

  if (el.detailSessionId) el.detailSessionId.textContent = s.id;
  if (el.detailWorkspaceBadge) el.detailWorkspaceBadge.textContent = `📁 ${s.project_name || 'Default'}`;
  if (el.detailRuntimeBadge) {
    const rm = (s.runtime_mode || 'ide').toLowerCase();
    const prov = (s.provider || 'antigravity').toLowerCase();
    el.detailRuntimeBadge.style.display = 'inline-flex';
    el.detailRuntimeBadge.className = 'runtime-badge';
    if (rm === 'cli') {
      el.detailRuntimeBadge.classList.add('runtime-cli');
      el.detailRuntimeBadge.textContent = '⚡ CLI';
      el.detailRuntimeBadge.title = `Antigravity CLI (${s.storage_dir || ''})`;
    } else if (rm === 'core') {
      el.detailRuntimeBadge.classList.add('runtime-core');
      el.detailRuntimeBadge.textContent = '⚙️ Core';
      el.detailRuntimeBadge.title = `Antigravity Core (${s.storage_dir || ''})`;
    } else if (prov === 'claude') {
      el.detailRuntimeBadge.classList.add('runtime-claude');
      el.detailRuntimeBadge.textContent = '🤖 Claude';
      el.detailRuntimeBadge.title = 'Claude Code';
    } else if (prov === 'copilot') {
      el.detailRuntimeBadge.classList.add('runtime-copilot');
      el.detailRuntimeBadge.textContent = '🐙 Copilot';
      el.detailRuntimeBadge.title = 'GitHub Copilot';
    } else {
      el.detailRuntimeBadge.classList.add('runtime-ide');
      el.detailRuntimeBadge.textContent = '💻 IDE';
      el.detailRuntimeBadge.title = `Antigravity IDE (${s.storage_dir || ''})`;
    }
  }
  if (el.detailTime) el.detailTime.textContent = formatDateTime(s.created_at, true);

  if (!window.location.hash.includes(s.id)) {
    window.location.hash = `#/session/${s.id}`;
  }

  if (state.activeTimelineSessionId !== s.id) {
    state.renderedStepMap.clear();
    state.activeTimelineSessionId = s.id;
    if (el.timelineContainer) el.timelineContainer.innerHTML = '';
  }

  updateSessionFilesBadge(detail);
  renderTimelineMessages();

  startSessionLiveTail(s.id);

  const hash = window.location.hash;
  const stepMatch = hash.match(/step=(\d+)/);
  if (stepMatch) {
    const targetStep = parseInt(stepMatch[1], 10);
    const node = state.renderedStepMap.get(targetStep);
    if (node) {
      node.scrollIntoView({ behavior: 'smooth', block: 'center' });
      return;
    }
  }
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function showDashboard(updateHash = true) {
  stopSessionLiveTail();
  state.currentView = 'dashboard';
  state.selectedSessionId = null;
  state.activeTimelineSessionId = null;
  state.renderedStepMap.clear();
  state.pendingStepsBuffer = [];
  if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';

  if (el.detailView) el.detailView.classList.remove('active');
  if (el.dashboardView) el.dashboardView.style.display = 'block';

  if (updateHash && window.location.hash.startsWith('#/session')) {
    window.location.hash = '#/';
  }
}

// Hash-based Client-Side Router
function initRouter() {
  window.addEventListener('hashchange', handleRoute);
  handleRoute();
}

function handleRoute() {
  const hash = window.location.hash || '#/';
  const clean = hash.replace(/^#\/?/, '');

  if (clean.startsWith('session/')) {
    const rawTarget = clean.slice('session/'.length);
    const [sessionId] = rawTarget.split('?');
    if (sessionId) {
      if (state.selectedSessionId !== sessionId || state.currentView !== 'detail') {
        loadSessionDetail(sessionId);
      }
      return;
    }
  }

  if (clean.startsWith('project/')) {
    const projName = decodeURIComponent(clean.slice('project/'.length));
    if (projName) {
      if (state.currentView === 'detail') {
        showDashboard(false);
      }
      state.drilledProject = projName;
      state.filterWorkspace = projName;
      if (el.workspaceSelect) el.workspaceSelect.value = projName;
      // Reset global date filter so all sessions in repository are loaded for in-project exploration
      state.dateFrom = '';
      state.dateTo = '';
      state.datePreset = 'all';
      if (el.dateFromInput) el.dateFromInput.value = '';
      if (el.dateToInput) el.dateToInput.value = '';
      if (el.datePresets) {
        el.datePresets.forEach(p => p.classList.remove('active'));
        const allBtn = Array.from(el.datePresets).find(p => p.dataset.preset === 'all');
        if (allBtn) allBtn.classList.add('active');
      }
      loadSessions();
      return;
    }
  }

  // Dashboard route
  if (state.currentView === 'detail') {
    showDashboard(false);
  }
  if (!clean.startsWith('project/') && !clean.startsWith('session/')) {
    if (state.drilledProject) {
      state.drilledProject = null;
      state.filterWorkspace = '';
      if (el.workspaceSelect) el.workspaceSelect.value = '';
      loadSessions();
    }
  }
}

function resetSessionStepFilters() {
  state.sessionStepSearch = '';
  state.sessionStepSort = 'step_desc';
  state.messageFilter = 'ALL';
  state.pendingStepsBuffer = [];
  if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';

  if (el.sessionStepSearchInput) el.sessionStepSearchInput.value = '';
  if (el.clearSessionStepSearchBtn) el.clearSessionStepSearchBtn.style.display = 'none';
  if (el.sessionStepSortSelect) el.sessionStepSortSelect.value = 'step_desc';
  if (el.resetSessionStepFiltersBtn) el.resetSessionStepFiltersBtn.style.display = 'none';

  document.querySelectorAll('.msg-filter-chip').forEach(c => {
    if (c.dataset.filter === 'ALL') c.classList.add('active');
    else c.classList.remove('active');
  });

  renderTimelineMessages();
}

function getFilteredAndSortedMessages() {
  if (!state.selectedSessionDetail) return [];
  const messages = state.selectedSessionDetail.messages || [];

  // 1. Filter by Message Type (R-013, R-035)
  let list = messages.filter(m => {
    const isUser = m.type === 'USER_INPUT';
    const isAgent = m.type === 'PLANNER_RESPONSE';
    const hasTools = Boolean(m.tool_calls && m.tool_calls.length > 0);
    const isError = m.status === 'ERROR' || m.type === 'ERROR_MESSAGE';
    const hasContent = Boolean(m.content && m.content.trim());

    if (state.messageFilter === 'USER_INPUT') return isUser;
    if (state.messageFilter === 'AGENT_RESPONSES') {
      return isAgent && !hasTools && hasContent;
    }
    if (state.messageFilter === 'DIALOGUE') {
      if (isUser) return true;
      if (isAgent && !hasTools && hasContent) return true;
      return false;
    }
    if (state.messageFilter === 'PLANNER_RESPONSE') return isAgent;
    if (state.messageFilter === 'TOOL_CALLS') return hasTools;
    if (state.messageFilter === 'ERROR') return isError;
    return true;
  });

  // 2. In-Session Step Search (Interrelated with message filter and sort)
  const stepQuery = (state.sessionStepSearch || '').trim().toLowerCase();
  if (stepQuery) {
    list = list.filter(m => {
      const contentMatch = (m.content || '').toLowerCase().includes(stepQuery);
      const thinkingMatch = (m.thinking || '').toLowerCase().includes(stepQuery);
      const stepMatch = String(m.step_index).includes(stepQuery);
      let toolMatch = false;
      if (m.tool_calls && Array.isArray(m.tool_calls)) {
        toolMatch = m.tool_calls.some(tc =>
          (tc.name || '').toLowerCase().includes(stepQuery) ||
          (tc.args_preview || '').toLowerCase().includes(stepQuery) ||
          (typeof tc.args === 'object' && JSON.stringify(tc.args).toLowerCase().includes(stepQuery))
        );
      }
      return contentMatch || thinkingMatch || stepMatch || toolMatch;
    });
  }

  // 3. In-Session Step Sort (Interrelated with search results and type filters)
  const sortMode = state.sessionStepSort || 'step_desc';
  list.sort((a, b) => {
    const stepA = typeof a.step_index === 'number' ? a.step_index : (parseInt(a.step_index, 10) || 0);
    const stepB = typeof b.step_index === 'number' ? b.step_index : (parseInt(b.step_index, 10) || 0);

    switch (sortMode) {
      case 'step_asc':
        return stepA - stepB;
      case 'time_desc': {
        const timeA = a.created_at ? (new Date(a.created_at).getTime() || 0) : 0;
        const timeB = b.created_at ? (new Date(b.created_at).getTime() || 0) : 0;
        return (timeB - timeA) || (stepB - stepA);
      }
      case 'time_asc': {
        const timeA = a.created_at ? (new Date(a.created_at).getTime() || 0) : 0;
        const timeB = b.created_at ? (new Date(b.created_at).getTime() || 0) : 0;
        return (timeA - timeB) || (stepA - stepB);
      }
      case 'tools_desc': {
        const toolsA = (a.tool_calls || []).length;
        const toolsB = (b.tool_calls || []).length;
        return (toolsB - toolsA) || (stepB - stepA);
      }
      case 'length_desc': {
        const lenA = (a.content || '').length + (a.thinking || '').length;
        const lenB = (b.content || '').length + (b.thinking || '').length;
        return (lenB - lenA) || (stepB - stepA);
      }
      case 'step_desc':
      default:
        return stepB - stepA;
    }
  });

  return list;
}

function renderTimelineMessages() {
  if (!el.timelineContainer || !state.selectedSessionDetail) return;
  const allMessages = state.selectedSessionDetail.messages || [];
  const filteredAndSorted = getFilteredAndSortedMessages();

  const stepQuery = (state.sessionStepSearch || '').trim();
  const sortMode = state.sessionStepSort || 'step_desc';
  const hasActiveStepFilters = Boolean(
    stepQuery ||
    sortMode !== 'step_desc' ||
    state.messageFilter !== 'ALL'
  );

  // Update summary count and reset button visibility
  const countBadge = el.detailStepCountBadge || document.getElementById('detailStepCountBadge');
  if (countBadge) {
    if (filteredAndSorted.length === allMessages.length) {
      countBadge.textContent = `${allMessages.length} step${allMessages.length === 1 ? '' : 's'}`;
    } else {
      countBadge.textContent = `Showing ${filteredAndSorted.length} of ${allMessages.length} steps`;
    }
  }

  const resetBtn = el.resetSessionStepFiltersBtn || document.getElementById('resetSessionStepFiltersBtn');
  if (resetBtn) {
    resetBtn.style.display = hasActiveStepFilters ? 'inline-flex' : 'none';
  }

  const clearStepBtn = el.clearSessionStepSearchBtn || document.getElementById('clearSessionStepSearchBtn');
  if (clearStepBtn) {
    clearStepBtn.style.display = stepQuery ? 'inline-flex' : 'none';
  }

  const stepSortSelect = el.sessionStepSortSelect || document.getElementById('sessionStepSortSelect');
  if (stepSortSelect && stepSortSelect.value !== sortMode) {
    stepSortSelect.value = sortMode;
  }

  if (filteredAndSorted.length === 0) {
    el.timelineContainer.innerHTML = `
      <div class="empty-state">
        <div class="empty-state-icon">🔍</div>
        <h3>No matching steps found</h3>
        <p>${stepQuery ? `No steps match "${escapeHtml(stepQuery)}"` : 'No steps match the current filter selection.'}</p>
        <button class="btn-pill" id="clearStepSearchEmptyBtn" style="margin-top: 1rem;">Reset Step Filters</button>
      </div>
    `;
    document.getElementById('clearStepSearchEmptyBtn')?.addEventListener('click', () => {
      resetSessionStepFilters();
    });
    return;
  }

  // Keyed DOM Mounting:
  // Mount nodes in the EXACT sequence of filteredAndSorted.
  // We reuse existing cached DOM elements from state.renderedStepMap to preserve
  // interactive states and avoid redundant Markdown/KaTeX/Highlight.js parsing.
  el.timelineContainer.innerHTML = '';
  state.pendingStepsBuffer = [];
  if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';

  const fragment = document.createDocumentFragment();
  for (const msg of filteredAndSorted) {
    let node;
    if (state.renderedStepMap.has(msg.step_index)) {
      const existingNode = state.renderedStepMap.get(msg.step_index);
      if (existingNode.dataset.status !== msg.status) {
        existingNode.dataset.status = msg.status;
        const temp = document.createElement('div');
        temp.innerHTML = buildStepHtml(msg);
        node = temp.firstElementChild;
        state.renderedStepMap.set(msg.step_index, node);
        attachStepInteractivity(node);
      } else {
        node = existingNode;
      }
    } else {
      const temp = document.createElement('div');
      temp.innerHTML = buildStepHtml(msg);
      node = temp.firstElementChild;
      state.renderedStepMap.set(msg.step_index, node);
      attachStepInteractivity(node);
    }
    if (node) {
      node.id = `step-${msg.step_index}`;
    }
    fragment.appendChild(node);
  }

  el.timelineContainer.appendChild(fragment);
  renderAllMermaidDiagrams(el.timelineContainer);
}

// Construct HTML string for a single message step (O(1) Markdown & syntax processing)
function buildStepHtml(msg) {
  const isUser = msg.type === 'USER_INPUT';
  const isAgent = msg.type === 'PLANNER_RESPONSE';
  const isToolOutput = !isUser && !isAgent && msg.type !== 'ERROR_MESSAGE';
  const isError = msg.type === 'ERROR_MESSAGE' || msg.status === 'ERROR';

  const authorName = isUser ? 'User' : isAgent ? 'Agent' : isError ? 'Error' : msg.type;
  const authorIcon = isUser ? '👤' : isAgent ? '🤖' : isError ? '⚠️' : '⚙️';
  const timeStr = formatDateTime(msg.created_at, true);
  const relTimeStr = formatRelativeTime(msg.created_at);

  let bodyHtml = '';
  let toolOutputHtml = '';

  if (isUser || isAgent) {
    if (msg.content && msg.content.trim()) {
      try {
        bodyHtml = marked.parse(msg.content);
      } catch (e) {
        bodyHtml = `<p style="white-space: pre-wrap;">${escapeHtml(msg.content)}</p>`;
      }
    }
  } else if (isToolOutput) {
    const contentLen = msg.content ? msg.content.length : 0;
    const sizeStr = contentLen > 1024 ? `${Math.round(contentLen / 1024)}KB` : `${contentLen}B`;
    toolOutputHtml = `
      <details class="tool-output-accordion">
        <summary class="tool-output-summary">
          <span>⚙️ <strong>${escapeHtml(msg.type)}</strong> output (${sizeStr})</span>
          <span style="font-size: 0.72rem; opacity: 0.7;">Click to toggle output</span>
        </summary>
        <pre class="tool-output-pre"><code>${escapeHtml(msg.content || '')}</code></pre>
      </details>
    `;
  } else if (isError) {
    bodyHtml = `<div class="error-banner" style="padding: 0.75rem; background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); border-radius: var(--radius-sm); color: #f87171;">⚠️ ${escapeHtml(msg.content || 'Unknown error')}</div>`;
  }

  let thinkingHtml = '';
  if (msg.has_thinking && msg.thinking) {
    let thinkingParsed = '';
    try {
      thinkingParsed = marked.parse(msg.thinking);
    } catch (_) {
      thinkingParsed = escapeHtml(msg.thinking);
    }
    thinkingHtml = `
      <details class="thinking-accordion">
        <summary class="thinking-summary">🧠 Reasoning / Inner Monologue (${msg.thinking.length} chars)</summary>
        <div class="thinking-content">${thinkingParsed}</div>
      </details>
    `;
  }

  let toolCallsHtml = '';
  if (msg.tool_calls && msg.tool_calls.length > 0) {
    toolCallsHtml = `
      <div class="tool-calls-container">
        <div style="font-size: 0.78rem; font-weight: 600; color: var(--text-secondary); margin-bottom: 0.5rem;">
          ⚡ Invoked Tools (${msg.tool_calls.length})
        </div>
        ${msg.tool_calls.map(tc => {
          const args = tc.args || {};
          const argsPretty = tc.args ? JSON.stringify(tc.args, null, 2) : (tc.args_preview || '{}');

          const rawFilePath = args.AbsolutePath || args.TargetFile || args.path || args.file_path || (typeof args.file === 'string' ? args.file : '');
          let fileChipHtml = '';
          if (rawFilePath && typeof rawFilePath === 'string') {
            const cleanP = rawFilePath.replace(/^file:\/\//, '');
            const fileName = cleanP.split('/').pop() || cleanP;
            const lineInfo = args.StartLine ? `#L${args.StartLine}${args.EndLine ? `-L${args.EndLine}` : ''}` : '';
            const fullTarget = cleanP + lineInfo;
            fileChipHtml = `
              <div class="tool-file-chip-group">
                <button class="tool-file-pill" data-filepath="${escapeHtml(fullTarget)}" title="Open ${escapeHtml(fileName)} in browser">📄 ${escapeHtml(fileName)} ↗</button>
                <button class="tool-file-copy-btn" data-path="${escapeHtml(fullTarget)}" title="Copy file path directly to clipboard">📋</button>
              </div>
            `;
          }

          let artifactHtml = '';
          if (args.ArtifactMetadata && typeof args.ArtifactMetadata === 'object') {
            const summary = args.ArtifactMetadata.Summary || 'Artifact';
            const targetFile = args.TargetFile || rawFilePath;
            artifactHtml = `
              <div class="tool-artifact-badge">
                <span>✨ Artifact: ${escapeHtml(summary)}</span>
                ${targetFile ? `<button class="open-artifact-btn" data-filepath="${escapeHtml(targetFile)}">View Artifact ↗</button>` : ''}
              </div>
            `;
          }

          return `
            <details class="tool-call-card">
              <summary class="tool-call-header">
                <div style="display: flex; align-items: center; gap: 6px; min-width: 0; flex: 1; overflow: hidden;">
                  <span class="tool-call-name">🛠️ ${escapeHtml(tc.name)}</span>
                  ${fileChipHtml}
                </div>
                <button class="copy-tool-args-btn btn-icon" style="width: 26px; height: 26px; font-size: 0.75rem; flex-shrink: 0;" title="Copy Args" data-args="${escapeHtml(argsPretty)}">📋</button>
              </summary>
              ${artifactHtml}
              <pre class="tool-call-args"><code>${escapeHtml(argsPretty)}</code></pre>
            </details>
          `;
        }).join('')}
      </div>
    `;
  }

  const formattedStep = typeof msg.step_index === 'number' ? msg.step_index.toLocaleString() : (msg.step_index ?? '0');

  return `
    <div class="message-item ${isUser ? 'user' : isAgent ? 'agent' : 'tool-result'}" id="step-${msg.step_index}" data-step="${msg.step_index}" data-status="${escapeHtml(msg.status || '')}">
      <div class="timeline-node" title="Step ${formattedStep}">${formattedStep}</div>
      <div class="message-card">
        <div class="message-header">
          <span class="message-author">${authorIcon} ${authorName}</span>
          <div class="message-meta">
            <span>Step ${formattedStep}</span>
            <span>•</span>
            <span class="step-timestamp" title="${timeStr}">
              <span class="exact-time">📅 ${timeStr}</span>
              ${relTimeStr ? `<span class="relative-time">(${relTimeStr})</span>` : ''}
            </span>
            ${msg.content ? `<button class="copy-msg-btn btn-icon" style="width: 26px; height: 26px; font-size: 0.75rem;" title="Copy Content" data-content="${escapeHtml(msg.content)}">📋</button>` : ''}
          </div>
        </div>
        ${thinkingHtml}
        ${bodyHtml ? `<div class="message-body">${bodyHtml}</div>` : ''}
        ${toolOutputHtml}
        ${toolCallsHtml}
      </div>
    </div>
  `;
}

// Attach interactive listeners specifically to given container or step node
function attachStepInteractivity(container) {
  if (!container) return;

  container.querySelectorAll('.copy-msg-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.content, 'Message copied to clipboard!');
    };
  });

  container.querySelectorAll('.copy-tool-args-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.args, 'Tool arguments copied to clipboard!');
    };
  });

  container.querySelectorAll('.toggle-mermaid-src-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      const targetId = btn.dataset.target;
      const srcEl = document.getElementById(targetId);
      if (srcEl) {
        const isHidden = srcEl.style.display === 'none';
        srcEl.style.display = isHidden ? 'block' : 'none';
        btn.textContent = isHidden ? 'Hide Source' : 'Source';
        btn.classList.toggle('active', isHidden);
      }
    };
  });

  container.querySelectorAll('.copy-mermaid-src-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.code, 'Diagram code copied to clipboard!');
    };
  });

  container.querySelectorAll('.copy-code-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.code, 'Code copied to clipboard!');
    };
  });

  container.querySelectorAll('.file-path-link').forEach(link => {
    link.onclick = (e) => {
      e.preventDefault();
      e.stopPropagation();
      const path = link.dataset.filepath || link.getAttribute('href');
      openFileViewer(path);
    };
  });

  container.querySelectorAll('.file-link-quick-copy-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.preventDefault();
      e.stopPropagation();
      const path = btn.dataset.path;
      copyToClipboard(path, 'File path copied to clipboard!');
      const origText = btn.textContent;
      btn.textContent = '✓';
      btn.classList.add('copy-success');
      setTimeout(() => {
        btn.textContent = origText;
        btn.classList.remove('copy-success');
      }, 1500);
    };
  });

  container.querySelectorAll('.tool-file-pill').forEach(btn => {
    btn.onclick = (e) => {
      e.preventDefault();
      e.stopPropagation();
      openFileViewer(btn.dataset.filepath);
    };
  });

  container.querySelectorAll('.tool-file-copy-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.preventDefault();
      e.stopPropagation();
      const path = btn.dataset.path;
      copyToClipboard(path, 'File path copied to clipboard!');
      const origText = btn.textContent;
      btn.textContent = '✓';
      btn.classList.add('copy-success');
      setTimeout(() => {
        btn.textContent = origText;
        btn.classList.remove('copy-success');
      }, 1500);
    };
  });

  container.querySelectorAll('.open-artifact-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.preventDefault();
      e.stopPropagation();
      openFileViewer(btn.dataset.filepath);
    };
  });

  renderAllMermaidDiagrams(container);
}

// Test if an incoming step matches currently selected filters (Implements: R-013, R-035)
function stepMatchesActiveFilters(step) {
  const isUser = step.type === 'USER_INPUT';
  const isAgent = step.type === 'PLANNER_RESPONSE';
  const hasTools = Boolean(step.tool_calls && step.tool_calls.length > 0);
  const isError = step.type === 'ERROR_MESSAGE' || step.status === 'ERROR';
  const hasContent = Boolean(step.content && step.content.trim());

  if (state.messageFilter === 'USER_INPUT' && !isUser) return false;
  if (state.messageFilter === 'AGENT_RESPONSES') {
    if (!isAgent || hasTools || !hasContent) return false;
  }
  if (state.messageFilter === 'DIALOGUE') {
    const isCleanAgent = isAgent && !hasTools && hasContent;
    if (!isUser && !isCleanAgent) return false;
  }
  if (state.messageFilter === 'PLANNER_RESPONSE' && !isAgent) return false;
  if (state.messageFilter === 'TOOL_CALLS' && !hasTools) return false;
  if (state.messageFilter === 'ERROR' && !isError) return false;

  const q = (state.sessionStepSearch || '').trim().toLowerCase();
  if (!q) return true;

  if (step.content && step.content.toLowerCase().includes(q)) return true;
  if (step.thinking && step.thinking.toLowerCase().includes(q)) return true;
  if (step.type && step.type.toLowerCase().includes(q)) return true;
  if (step.tool_calls) {
    for (const tc of step.tool_calls) {
      if (tc.name && tc.name.toLowerCase().includes(q)) return true;
      if (tc.args_preview && tc.args_preview.toLowerCase().includes(q)) return true;
    }
  }
  return false;
}

// Atomically mount buffered incoming steps and glide smoothly to stream head
function flushPendingSteps() {
  const buffered = state.pendingStepsBuffer || [];
  state.pendingStepsBuffer = [];

  if (el.newStepsFloatingBtn) {
    el.newStepsFloatingBtn.style.display = 'none';
  }

  state.followStream = true;
  if (el.toggleFollowStreamBtn) {
    el.toggleFollowStreamBtn.classList.add('active');
  }

  const isDesc = (state.sessionStepSort || 'step_desc') === 'step_desc';

  if (buffered.length > 0) {
    // Sort in ascending order of step_index so prepend/append maintains chronological integrity
    buffered.sort((a, b) => a.step_index - b.step_index);
    for (const step of buffered) {
      if (!stepMatchesActiveFilters(step)) continue;
      if (state.renderedStepMap.has(step.step_index)) continue;

      const temp = document.createElement('div');
      temp.innerHTML = buildStepHtml(step);
      const newNode = temp.firstElementChild;
      newNode.classList.add('step-enter');
      state.renderedStepMap.set(step.step_index, newNode);

      if (isDesc) {
        el.timelineContainer.prepend(newNode);
      } else {
        el.timelineContainer.appendChild(newNode);
      }
      attachStepInteractivity(newNode);
    }
  }

  if (isDesc) {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  } else {
    window.scrollTo({ top: document.body.scrollHeight, behavior: 'smooth' });
  }
}

// Real-Time Incremental Append & Patch (O(Δ) Keyed DOM Reconciliation)
function appendOrPatchStep(step) {
  if (!state.selectedSessionDetail) return;
  if (!state.selectedSessionDetail.messages) {
    state.selectedSessionDetail.messages = [];
  }

  const existingIdx = state.selectedSessionDetail.messages.findIndex(m => m.step_index === step.step_index);
  if (existingIdx !== -1) {
    state.selectedSessionDetail.messages[existingIdx] = step;
  } else {
    state.selectedSessionDetail.messages.push(step);
  }

  const allMessages = state.selectedSessionDetail.messages;
  if (el.detailStepCountBadge) {
    el.detailStepCountBadge.textContent = `${allMessages.length} step${allMessages.length === 1 ? '' : 's'}`;
  }

  // Live Reactive Telemetry Sync (R-044): Dynamic Files & Agent Context Updates
  updateSessionFilesBadge(state.selectedSessionDetail);
  if (el.sessionFilesModal && el.sessionFilesModal.classList.contains('active')) {
    renderSessionFilesList();
  }
  triggerLiveContextSync(state.selectedSessionId || (state.selectedSessionDetail && state.selectedSessionDetail.id));

  // Keyed DOM: check if node already exists in DOM (e.g. status transition RUNNING -> DONE)
  if (state.renderedStepMap.has(step.step_index)) {
    const existingNode = state.renderedStepMap.get(step.step_index);
    if (existingNode.dataset.status !== step.status) {
      existingNode.dataset.status = step.status;
      const temp = document.createElement('div');
      temp.innerHTML = buildStepHtml(step);
      const updatedNode = temp.firstElementChild;
      existingNode.replaceWith(updatedNode);
      state.renderedStepMap.set(step.step_index, updatedNode);
      attachStepInteractivity(updatedNode);
    }
    return;
  }

  // Check if step matches active search/filter
  if (!stepMatchesActiveFilters(step)) {
    return;
  }

  // Determine if user is currently watching the live stream head
  const isDesc = (state.sessionStepSort || 'step_desc') === 'step_desc';
  const isAtStreamHead = isDesc
    ? window.scrollY <= 150
    : (window.innerHeight + window.scrollY >= document.body.offsetHeight - 150);

  // If detached from stream head or follow is disabled:
  // Buffer the step into memory to prevent viewport shifts / layout jumping while reading.
  if (!state.followStream || !isAtStreamHead) {
    const bufIdx = state.pendingStepsBuffer.findIndex(s => s.step_index === step.step_index);
    if (bufIdx !== -1) {
      state.pendingStepsBuffer[bufIdx] = step;
    } else {
      state.pendingStepsBuffer.push(step);
    }

    const count = state.pendingStepsBuffer.length;
    if (el.newStepsFloatingBtn) {
      el.newStepsFloatingBtn.innerHTML = isDesc
        ? `<span class="new-steps-count-badge">${count}</span> new step${count > 1 ? 's' : ''} arrived <span>↑</span>`
        : `<span class="new-steps-count-badge">${count}</span> new step${count > 1 ? 's' : ''} arrived <span>↓</span>`;
      el.newStepsFloatingBtn.style.display = 'inline-flex';
    }
    return;
  }

  // User is actively following stream at head: mount directly
  const temp = document.createElement('div');
  temp.innerHTML = buildStepHtml(step);
  const newNode = temp.firstElementChild;
  newNode.classList.add('step-enter');
  state.renderedStepMap.set(step.step_index, newNode);

  if (isDesc) {
    el.timelineContainer.prepend(newNode);
  } else {
    el.timelineContainer.appendChild(newNode);
  }
  attachStepInteractivity(newNode);

  if (isDesc) {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  } else {
    newNode.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }
  if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';
}

// Real-Time SSE Log Tailer Management
function startSessionLiveTail(sessionId) {
  stopSessionLiveTail(false);
  state.tailStreaming = true;

  let highestStep = -1;
  const messages = (state.selectedSessionDetail && state.selectedSessionDetail.messages) || [];
  messages.forEach(m => {
    if (m.step_index > highestStep) highestStep = m.step_index;
  });

  if (el.tailStreamControl) {
    el.tailStreamControl.classList.remove('paused');
    el.tailStreamControl.classList.add('streaming');
  }
  if (el.tailStatusLabel) el.tailStatusLabel.textContent = 'Live Stream';
  if (el.detailLiveDot) el.detailLiveDot.classList.add('active', 'pulsing');
  if (el.toggleStreamBtn) el.toggleStreamBtn.title = 'Click to Pause Live Streaming';

  const sseUrl = `/api/sessions/${encodeURIComponent(sessionId)}/tail?last_step=${highestStep}`;
  const es = new EventSource(sseUrl);
  state.tailEventSource = es;

  es.addEventListener('connected', () => {
    if (el.tailStatusLabel) el.tailStatusLabel.textContent = 'Live Stream ⚡';
  });

  es.addEventListener('step', (event) => {
    try {
      const step = JSON.parse(event.data);
      appendOrPatchStep(step);
    } catch (err) {
      console.warn('Failed to parse SSE step event:', err);
    }
  });

  es.onerror = () => {
    if (state.tailStreaming) {
      if (el.tailStatusLabel) el.tailStatusLabel.textContent = 'Reconnecting...';
      if (el.detailLiveDot) el.detailLiveDot.classList.remove('pulsing');
    }
  };
}

function stopSessionLiveTail(userInitiated = false) {
  if (state.tailEventSource) {
    state.tailEventSource.close();
    state.tailEventSource = null;
  }
  if (userInitiated) {
    state.tailStreaming = false;
    if (el.tailStreamControl) {
      el.tailStreamControl.classList.remove('streaming');
      el.tailStreamControl.classList.add('paused');
    }
    if (el.tailStatusLabel) el.tailStatusLabel.textContent = 'Stream Paused';
    if (el.detailLiveDot) el.detailLiveDot.classList.remove('pulsing');
    if (el.toggleStreamBtn) el.toggleStreamBtn.title = 'Click to Resume Live Streaming';
  } else {
    state.tailStreaming = false;
    if (el.tailStreamControl) el.tailStreamControl.classList.remove('streaming', 'paused');
    if (el.detailLiveDot) el.detailLiveDot.classList.remove('pulsing');
  }
}

// Dedicated Non-Destructive Session Trail & Telemetry Refresh (R-044)
async function refreshCurrentSessionTrail(manual = true) {
  const sessionId = state.selectedSessionId || (state.selectedSessionDetail && state.selectedSessionDetail.session && state.selectedSessionDetail.session.id);
  if (!sessionId) return;

  if (el.refreshTrailBtn) {
    el.refreshTrailBtn.classList.add('refreshing');
    el.refreshTrailBtn.disabled = true;
  }

  try {
    const [detailRes, contextRes] = await Promise.all([
      fetch(`/api/sessions/${sessionId}`),
      fetch(`/api/sessions/${sessionId}/context`)
    ]);

    if (detailRes.ok) {
      const detail = await detailRes.json();
      state.selectedSessionDetail = detail;

      // Update step badge
      const allMessages = detail.messages || [];
      if (el.detailStepCountBadge) {
        el.detailStepCountBadge.textContent = `${allMessages.length} step${allMessages.length === 1 ? '' : 's'}`;
      }

      // Re-evaluate and pulse files & artifacts
      updateSessionFilesBadge(detail);
      if (el.sessionFilesModal && el.sessionFilesModal.classList.contains('active')) {
        renderSessionFilesList();
      }

      // Re-render timeline messages to incorporate newest steps and commentary
      if (manual) {
        state.pendingStepsBuffer = [];
        if (el.newStepsFloatingBtn) el.newStepsFloatingBtn.style.display = 'none';
        renderTimelineMessages();
      }
    }

    if (contextRes.ok) {
      const ctx = await contextRes.json();
      state.sessionContextData = ctx;
      state.sessionMatrix.context = ctx;
      updateAgentContextBadge(ctx, manual);
      if (el.sessionContextModal && el.sessionContextModal.classList.contains('active')) {
        renderSessionMatrix();
      }
    }

    if (manual) {
      showToast('✨ Session trail, agent context & files synchronized');
    }
  } catch (err) {
    console.error('Session trail refresh failed:', err);
    if (manual) {
      showToast(`❌ Refresh failed: ${err.message}`);
    }
  } finally {
    if (el.refreshTrailBtn) {
      setTimeout(() => {
        el.refreshTrailBtn.classList.remove('refreshing');
        el.refreshTrailBtn.disabled = false;
      }, 400);
    }
  }
}


// Asynchronously render all pending Mermaid diagrams using Mermaid native runner
async function renderAllMermaidDiagrams(container = document) {
  const unrendered = container.querySelectorAll('pre.mermaid:not([data-processed="true"])');
  if (!unrendered.length) return;

  try {
    await mermaid.run({
      nodes: Array.from(unrendered),
      suppressErrors: true
    });
  } catch (err) {
    console.warn('Mermaid batch render warning:', err);
  }

  // Ensure zoom controls are wired for all canvas containers
  initCanvasZoomInteractions(container);
}

// Wire interactive zoom controls and theater triggers for diagram canvases
function initCanvasZoomInteractions(container = document) {
  container.querySelectorAll('.canvas-container:not([data-zoom-initialized="true"])').forEach(card => {
    card.setAttribute('data-zoom-initialized', 'true');
    const surface = card.querySelector('.canvas-content-surface');
    const viewport = card.querySelector('.canvas-viewport');
    const zoomInBtn = card.querySelector('.zoom-in-btn');
    const zoomOutBtn = card.querySelector('.zoom-out-btn');
    const zoomResetBtn = card.querySelector('.zoom-reset-btn');
    const theaterBtn = card.querySelector('.theater-btn');
    const badge = card.querySelector('.zoom-level-badge');
    const titleTag = card.querySelector('.canvas-type-tag')?.textContent || 'Diagram Canvas';
    const srcCode = card.querySelector('.mermaid-source-pre code')?.textContent || '';
    if (!surface) return;

    let zoom = 1.0;
    const update = () => {
      surface.style.transform = `scale(${zoom})`;
      if (badge) badge.textContent = `${Math.round(zoom * 100)}%`;
    };

    if (zoomInBtn) {
      zoomInBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        zoom = Math.min(3.0, +(zoom + 0.2).toFixed(2));
        update();
      });
    }
    if (zoomOutBtn) {
      zoomOutBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        zoom = Math.max(0.4, +(zoom - 0.2).toFixed(2));
        update();
      });
    }
    if (zoomResetBtn) {
      zoomResetBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        zoom = 1.0;
        update();
      });
    }
    if (theaterBtn) {
      theaterBtn.addEventListener('click', (e) => {
        e.stopPropagation();
        openDiagramTheater(titleTag, surface.innerHTML, srcCode);
      });
    }
    if (viewport) {
      viewport.addEventListener('dblclick', (e) => {
        if (e.target.closest('button')) return;
        openDiagramTheater(titleTag, surface.innerHTML, srcCode);
      });
    }
  });
}

// Export Session (R-021)
function exportCurrentSession(format) {
  if (!state.selectedSessionDetail) return;
  const detail = state.selectedSessionDetail;
  const s = detail.session;
  const messages = detail.messages || [];

  if (format === 'json') {
    const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(detail, null, 2));
    const dlAnchor = document.createElement('a');
    dlAnchor.setAttribute("href", dataStr);
    dlAnchor.setAttribute("download", `session-${s.id}.json`);
    dlAnchor.click();
    showToast('Downloaded session JSON!');
  } else if (format === 'markdown') {
    let md = `# Conversation Session: ${s.id}\n\n`;
    md += `- **Date**: ${formatDateTime(s.created_at)}\n`;
    md += `- **Workspace**: ${s.workspace || s.project_name}\n`;
    md += `- **Total Steps**: ${s.step_count}\n`;
    md += `- **User Messages**: ${s.user_message_count}\n`;
    md += `- **Agent Responses**: ${s.agent_response_count}\n\n---\n\n`;

    messages.forEach(m => {
      const isUser = m.type === 'USER_INPUT';
      const author = isUser ? '👤 **User**' : '🤖 **Agent**';
      md += `### ${author} (Step ${m.step_index} • ${formatDateTime(m.created_at)})\n\n`;
      if (m.content) md += `${m.content}\n\n`;
      if (m.tool_calls && m.tool_calls.length > 0) {
        md += `#### Tools Used:\n`;
        m.tool_calls.forEach(tc => {
          md += `- \`${tc.name}\`: \`\`\`json\n${JSON.stringify(tc.args, null, 2)}\n\`\`\`\n`;
        });
        md += `\n`;
      }
      md += `---\n\n`;
    });

    const dataStr = "data:text/markdown;charset=utf-8," + encodeURIComponent(md);
    const dlAnchor = document.createElement('a');
    dlAnchor.setAttribute("href", dataStr);
    dlAnchor.setAttribute("download", `session-${s.id}.md`);
    dlAnchor.click();
    showToast('Downloaded session Markdown!');
  }
}

// Search Modal (R-012, R-022)
function openSearchModal() {
  if (!el.searchModal) return;
  el.searchModal.classList.add('active');
  if (el.modalSearchInput) {
    el.modalSearchInput.value = '';
    setTimeout(() => el.modalSearchInput.focus(), 100);
  }
}

function closeSearchModal() {
  if (!el.searchModal) return;
  el.searchModal.classList.remove('active');
}

function renderSearchResults(results, query) {
  if (!el.modalSearchResults) return;

  if (results.length === 0) {
    el.modalSearchResults.innerHTML = `
      <div class="empty-state">
        <p>No results found for "<strong>${escapeHtml(query)}</strong>"</p>
      </div>
    `;
    return;
  }

  el.modalSearchResults.innerHTML = `
    <div style="font-size: 0.8rem; color: var(--text-muted); margin-bottom: 0.75rem;">
      Found ${results.length} match${results.length === 1 ? '' : 'es'}:
    </div>
    <div class="search-results-list">
      ${results.map(r => {
        return `
          <div class="search-result-item" data-session-id="${r.session_id}">
            <div style="display: flex; justify-content: space-between; font-size: 0.75rem; color: var(--text-muted); margin-bottom: 0.35rem;">
              <span class="workspace-badge" style="font-size: 0.7rem; padding: 0.1rem 0.4rem;">📁 ${escapeHtml(r.workspace || 'Workspace')}</span>
              <span>Step ${r.step_index} • ${formatRelativeTime(r.created_at)}</span>
            </div>
            <div style="font-size: 0.85rem; color: var(--text-primary); line-height: 1.4;">
              ${highlightSearchMatch(r.content_preview, query)}
            </div>
          </div>
        `;
      }).join('')}
    </div>
  `;

  el.modalSearchResults.querySelectorAll('.search-result-item').forEach(item => {
    item.addEventListener('click', () => {
      closeSearchModal();
      loadSessionDetail(item.dataset.sessionId);
    });
  });
}

function highlightSearchMatch(text, query) {
  if (!text || !query) return escapeHtml(text || '');
  const escapedText = escapeHtml(text);
  const regex = new RegExp(`(${escapeRegExp(query)})`, 'gi');
  return escapedText.replace(regex, '<mark style="background: rgba(99, 102, 241, 0.4); color: #fff; padding: 0.1rem 0.25rem; border-radius: 3px;">$1</mark>');
}

// Insights Modal (R-025)
function openInsightsModal() {
  if (!el.insightsModal || !state.stats) return;
  const s = state.stats;

  // Render Top Tools
  if (el.topToolsList) {
    const tools = s.top_tool_calls || [];
    if (tools.length === 0) {
      el.topToolsList.innerHTML = '<p style="color: var(--text-muted);">No tool call data available.</p>';
    } else {
      const maxCount = Math.max(...tools.map(t => t.count), 1);
      el.topToolsList.innerHTML = tools.map(t => {
        const pct = Math.round((t.count / maxCount) * 100);
        return `
          <div style="margin-bottom: 0.85rem;">
            <div style="display: flex; justify-content: space-between; font-size: 0.82rem; margin-bottom: 0.25rem;">
              <span style="font-family: var(--font-mono); color: #38bdf8;">🛠️ ${escapeHtml(t.name)}</span>
              <span style="font-weight: 600;">${t.count} call${t.count === 1 ? '' : 's'}</span>
            </div>
            <div style="background: var(--bg-glass-active); height: 6px; border-radius: 999px; overflow: hidden;">
              <div style="background: var(--accent-gradient); width: ${pct}%; height: 100%;"></div>
            </div>
          </div>
        `;
      }).join('');
    }
  }

  // Render Top Prompts (R-025)
  if (el.topPromptsList) {
    const prompts = s.top_prompts || [];
    if (prompts.length === 0) {
      el.topPromptsList.innerHTML = '<p style="color: var(--text-muted);">No prompt patterns analyzed yet.</p>';
    } else {
      el.topPromptsList.innerHTML = prompts.map(p => {
        return `
          <div style="background: var(--bg-glass); border: 1px solid var(--border-glass); border-radius: var(--radius-sm); padding: 0.75rem 1rem; margin-bottom: 0.75rem; display: flex; align-items: center; justify-content: space-between; gap: 0.75rem;">
            <div style="font-size: 0.85rem; color: var(--text-primary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
              "${escapeHtml(p.prompt)}"
            </div>
            <div style="display: flex; align-items: center; gap: 0.5rem; flex-shrink: 0;">
              <span class="version-pill">${p.count}x</span>
              <button class="btn-icon" style="width: 26px; height: 26px; font-size: 0.75rem;" title="Copy Prompt" onclick="copyToClipboard('${escapeHtml(p.prompt).replace(/'/g, "\\'")}', 'Prompt copied!')">📋</button>
            </div>
          </div>
        `;
      }).join('');
    }
  }

  el.insightsModal.classList.add('active');
}

function closeInsightsModal() {
  if (!el.insightsModal) return;
  el.insightsModal.classList.remove('active');
}

// Install Skill Modal (R-033)
function openInstallModal() {
  if (!el.installSkillModal) return;
  el.installSkillModal.classList.add('active');
  if (el.installStatusText) el.installStatusText.textContent = '';
}

function closeInstallModal() {
  if (!el.installSkillModal) return;
  el.installSkillModal.classList.remove('active');
}

async function executeInstallSkill() {
  const selectedTarget = document.querySelector('input[name="skillTarget"]:checked')?.value || 'workspace';
  if (el.installStatusText) el.installStatusText.textContent = '⏳ Installing skill...';
  if (el.doInstallSkillBtn) el.doInstallSkillBtn.disabled = true;

  try {
    const res = await fetch('/api/skill/install', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ target: selectedTarget })
    });
    const data = await res.json();
    if (res.ok && data.success) {
      if (el.installStatusText) el.installStatusText.textContent = '✅ ' + data.message;
      showToast('✅ Skill installed: ' + selectedTarget);
    } else {
      if (el.installStatusText) el.installStatusText.textContent = '❌ ' + (data.message || 'Installation failed');
      showToast('❌ Installation error');
    }
  } catch (err) {
    if (el.installStatusText) el.installStatusText.textContent = '❌ ' + err.message;
    showToast('❌ ' + err.message);
  } finally {
    if (el.doInstallSkillBtn) el.doInstallSkillBtn.disabled = false;
  }
}

// Helpers
window.copyToClipboard = function(text, successMsg = 'Copied to clipboard!') {
  if (!navigator.clipboard) {
    // Fallback for non-https/legacy
    const ta = document.createElement('textarea');
    ta.value = text;
    document.body.appendChild(ta);
    ta.select();
    document.execCommand('copy');
    document.body.removeChild(ta);
    showToast(successMsg);
    return;
  }
  navigator.clipboard.writeText(text).then(() => {
    showToast(successMsg);
  }).catch(err => {
    console.error('Clipboard copy failed:', err);
  });
};

function formatDateTime(iso, includeSeconds = false) {
  if (!iso) return '—';
  try {
    const d = new Date(iso);
    const opts = {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    };
    if (includeSeconds) {
      opts.second = '2-digit';
    }
    return d.toLocaleString(undefined, opts);
  } catch (_) {
    return iso;
  }
}

function formatRelativeTime(iso) {
  if (!iso) return '';
  try {
    const d = new Date(iso);
    const diffSec = Math.floor((Date.now() - d.getTime()) / 1000);
    if (diffSec <= 0 || diffSec < 60) return 'just now';
    if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
    if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
    const days = Math.floor(diffSec / 86400);
    if (days < 30) return `${days}d ago`;
    const months = Math.floor(days / 30);
    if (months < 12) return `${months}mo ago`;
    return `${Math.floor(months / 12)}y ago`;
  } catch (_) {
    return '';
  }
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

function escapeRegExp(string) {
  return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}

// ============================================================================
//  Diagram Theater Engine (R-037)
// ============================================================================
let isTheaterPanning = false;
let theaterStartX = 0;
let theaterStartY = 0;

function openDiagramTheater(title, surfaceHtml, rawSrc = '') {
  if (!el.diagramTheaterModal || !el.diagramTheaterSurface) return;
  if (el.diagramTheaterTitleText) el.diagramTheaterTitleText.textContent = title || 'Diagram Theater';
  el.diagramTheaterSurface.innerHTML = surfaceHtml;
  state.theaterRawSrc = rawSrc;
  state.theaterZoom = 1.15;
  state.theaterPanX = 0;
  state.theaterPanY = 0;
  updateDiagramTheaterTransform();

  el.diagramTheaterModal.classList.add('theater-open');
  document.body.style.overflow = 'hidden';
}

function closeDiagramTheater() {
  if (!el.diagramTheaterModal) return;
  el.diagramTheaterModal.classList.remove('theater-open');
  document.body.style.overflow = '';
}

function updateDiagramTheaterTransform() {
  if (!el.diagramTheaterSurface) return;
  el.diagramTheaterSurface.style.transform = `scale(${state.theaterZoom}) translate(${state.theaterPanX}px, ${state.theaterPanY}px)`;
  if (el.theaterZoomBadge) el.theaterZoomBadge.textContent = `${Math.round(state.theaterZoom * 100)}%`;
}

// ============================================================================
//  Native File & Artifact Viewer Engine (R-038)
// ============================================================================
function getFileIcon(path) {
  const ext = (path || '').split('.').pop().toLowerCase();
  switch (ext) {
    case 'md': case 'markdown': return '📝';
    case 'go': return '🐹';
    case 'js': case 'ts': case 'jsx': case 'tsx': return '⚡';
    case 'py': return '🐍';
    case 'rs': return '🦀';
    case 'json': case 'yaml': case 'yml': case 'toml': return '⚙️';
    case 'sh': case 'bash': case 'zsh': return '💻';
    case 'html': case 'css': case 'scss': return '🎨';
    case 'png': case 'jpg': case 'jpeg': case 'svg': case 'webp': case 'gif': return '🖼️';
    default: return '📄';
  }
}

async function openFileViewer(filePathWithAnchor) {
  if (!filePathWithAnchor || !el.fileViewerModal) return;

  // Parse optional line anchor e.g. /path/to/file.go#L45-L60 or /path/to/file.go:45
  let cleanPath = filePathWithAnchor.trim().replace(/^file:\/\//, '');
  let lineStart = null;
  let lineEnd = null;

  const hashIdx = cleanPath.indexOf('#');
  if (hashIdx !== -1) {
    const anchor = cleanPath.substring(hashIdx + 1);
    cleanPath = cleanPath.substring(0, hashIdx);
    const m = anchor.match(/L(\d+)(?:-L(\d+))?/i);
    if (m) {
      lineStart = parseInt(m[1], 10);
      lineEnd = m[2] ? parseInt(m[2], 10) : lineStart;
    }
  } else {
    const colonMatch = cleanPath.match(/:(\d+)(?::(\d+))?$/);
    if (colonMatch) {
      cleanPath = cleanPath.substring(0, colonMatch.index);
      lineStart = parseInt(colonMatch[1], 10);
      lineEnd = colonMatch[2] ? parseInt(colonMatch[2], 10) : lineStart;
    }
  }

  state.fileViewerTargetLine = lineStart;
  state.fileViewerTargetLineEnd = lineEnd;
  state.fileViewerMode = 'preview';
  state.fileViewerWrap = false;

  el.fileViewerModal.classList.add('file-viewer-open');
  document.body.style.overflow = 'hidden';

  if (el.fileViewerLoading) el.fileViewerLoading.style.display = 'flex';
  if (el.fileViewerError) el.fileViewerError.style.display = 'none';
  if (el.fileViewerMarkdown) el.fileViewerMarkdown.style.display = 'none';
  if (el.fileViewerCodeContainer) el.fileViewerCodeContainer.style.display = 'none';
  if (el.fileViewerImageContainer) el.fileViewerImageContainer.style.display = 'none';
  if (el.fileViewerModeControl) el.fileViewerModeControl.style.display = 'none';

  const fileName = cleanPath.split('/').pop() || cleanPath;
  if (el.fileViewerFilename) el.fileViewerFilename.textContent = fileName;
  if (el.fileViewerPath) el.fileViewerPath.textContent = cleanPath;
  if (el.fileViewerIcon) el.fileViewerIcon.textContent = getFileIcon(cleanPath);

  try {
    const res = await fetch(`/api/file?path=${encodeURIComponent(cleanPath)}`);
    if (!res.ok) {
      const errText = await res.text();
      throw new Error(errText || `HTTP ${res.status}`);
    }
    const fileData = await res.json();
    state.fileViewerData = fileData;
    renderFileViewerContent();
  } catch (err) {
    if (el.fileViewerLoading) el.fileViewerLoading.style.display = 'none';
    if (el.fileViewerError) {
      el.fileViewerError.style.display = 'block';
      el.fileViewerError.innerHTML = `
        <div style="font-size: 1.25rem; font-weight: 700; margin-bottom: 0.5rem; color: var(--accent-rose, #f43f5e);">⚠️ File Read Error</div>
        <p style="margin-bottom: 0.75rem; font-size: 0.92rem;">${escapeHtml(err.message)}</p>
        <code style="font-size: 0.78rem; opacity: 0.85; word-break: break-all; background: rgba(0,0,0,0.3); padding: 4px 8px; border-radius: 4px;">${escapeHtml(cleanPath)}</code>
      `;
    }
  }
}

function closeFileViewer() {
  if (!el.fileViewerModal) return;
  closeFileFindWidget();
  el.fileViewerModal.classList.remove('file-viewer-open');
  document.body.style.overflow = '';
  state.fileViewerData = null;
}

function renderFileViewerContent() {
  const data = state.fileViewerData;
  if (!data) return;

  if (el.fileViewerLoading) el.fileViewerLoading.style.display = 'none';
  if (el.fileViewerError) el.fileViewerError.style.display = 'none';
  if (el.fileViewerMarkdown) el.fileViewerMarkdown.style.display = 'none';
  if (el.fileViewerCodeContainer) el.fileViewerCodeContainer.style.display = 'none';
  if (el.fileViewerImageContainer) el.fileViewerImageContainer.style.display = 'none';

  const ext = (data.extension || '').replace('.', '').toUpperCase() || 'FILE';
  if (el.fileViewerLangBadge) el.fileViewerLangBadge.textContent = ext;
  if (el.fileViewerLinesBadge) el.fileViewerLinesBadge.textContent = `${(data.line_count || 0).toLocaleString()} lines`;
  if (el.fileViewerSizeBadge) {
    const kb = (data.size / 1024).toFixed(1);
    el.fileViewerSizeBadge.textContent = data.size > 1048576 ? `${(data.size / 1048576).toFixed(2)} MB` : `${kb} KB`;
  }

  const isMarkdown = data.extension === '.md' || data.extension === '.markdown';
  const isImage = data.is_binary && (data.content.startsWith('data:image/') || ['PNG', 'JPG', 'JPEG', 'GIF', 'WEBP', 'ICO'].includes(ext));

  if (isImage) {
    if (el.fileViewerImageContainer && el.fileViewerImage) {
      el.fileViewerImage.src = data.content;
      el.fileViewerImageContainer.style.display = 'flex';
    }
    return;
  }

  if (isMarkdown) {
    if (el.fileViewerModeControl) {
      el.fileViewerModeControl.style.display = 'inline-flex';
      el.fileViewerModeControl.querySelectorAll('.segment-btn').forEach(b => {
        b.classList.toggle('active', b.dataset.mode === state.fileViewerMode);
      });
    }

    if (state.fileViewerMode === 'preview') {
      if (el.fileViewerMarkdown) {
        el.fileViewerMarkdown.innerHTML = marked.parse(data.content);
        attachStepInteractivity(el.fileViewerMarkdown);
        el.fileViewerMarkdown.style.display = 'block';
        renderAllMermaidDiagrams(el.fileViewerMarkdown);
      }
      return;
    }
  }

  // Code or Raw Display Mode
  if (el.fileViewerWrapBtn) {
    el.fileViewerWrapBtn.classList.toggle('active', state.fileViewerWrap);
    el.fileViewerWrapBtn.innerHTML = state.fileViewerWrap ? '✓ Wrapped' : '↩ Wrap';
  }
  if (el.fileViewerCodeContainer && el.fileViewerCodeBlock && el.fileViewerLineNumbers) {
    el.fileViewerCodeContainer.style.display = 'flex';
    const pre = el.fileViewerCodeContainer.querySelector('.code-content-pre');
    if (pre) {
      pre.classList.toggle('wrap-code', state.fileViewerWrap);
    }

    const lines = data.content.split('\n');
    const totalLines = lines.length;

    let gutterHtml = '';
    const startL = state.fileViewerTargetLine;
    const endL = state.fileViewerTargetLineEnd || startL;

    for (let i = 1; i <= totalLines; i++) {
      const isTarget = startL && i >= startL && i <= endL;
      gutterHtml += `<span class="${isTarget ? 'line-highlight' : ''}" data-line="${i}">${i}</span>`;
    }
    el.fileViewerLineNumbers.innerHTML = gutterHtml;

    const lang = (data.extension || '').replace('.', '').toLowerCase();
    const validLang = lang && hljs.getLanguage(lang);
    let highlighted = validLang
      ? hljs.highlight(data.content, { language: lang, ignoreIllegals: true }).value
      : escapeHtml(data.content);

    if (startL) {
      const highlightedLines = highlighted.split('\n');
      highlighted = highlightedLines.map((line, idx) => {
        const lineNum = idx + 1;
        if (lineNum >= startL && lineNum <= endL) {
          return `<mark class="line-highlight" id="file-line-${lineNum}">${line}</mark>`;
        }
        return line;
      }).join('\n');
    }

    el.fileViewerCodeBlock.className = `hljs ${validLang ? 'language-' + lang : ''}`;
    el.fileViewerCodeBlock.innerHTML = highlighted;

    if (startL) {
      setTimeout(() => {
        const targetEl = document.getElementById(`file-line-${startL}`) || el.fileViewerLineNumbers.querySelector(`[data-line="${startL}"]`);
        if (targetEl) {
          targetEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }, 60);
    }
  }

  // If Find widget is open, re-evaluate find matches in the current view
  if (el.fileViewerFindWidget && el.fileViewerFindWidget.style.display !== 'none' && state.fileFindQuery) {
    setTimeout(() => executeFileFind(), 40);
  }
}

// ============================================================================
//  In-File Find & Navigation Engine (R-038)
// ============================================================================
function openFileFindWidget() {
  if (!el.fileViewerFindWidget || !el.fileFindInput) return;
  el.fileViewerFindWidget.style.display = 'flex';
  el.fileFindInput.focus();
  el.fileFindInput.select();
  if (el.fileFindInput.value.trim()) {
    executeFileFind();
  }
}

function closeFileFindWidget() {
  if (!el.fileViewerFindWidget) return;
  el.fileViewerFindWidget.style.display = 'none';
  clearFileFindHighlights();
  state.fileFindQuery = '';
  state.fileFindMatches = [];
  state.fileFindActiveIndex = -1;
  if (el.fileFindCountBadge) {
    el.fileFindCountBadge.textContent = '0 of 0';
    el.fileFindCountBadge.className = 'find-count-badge';
  }
}

function toggleFileFindCase() {
  state.fileFindCaseSensitive = !state.fileFindCaseSensitive;
  if (el.fileFindCaseBtn) {
    el.fileFindCaseBtn.classList.toggle('active', state.fileFindCaseSensitive);
  }
  executeFileFind();
}

function clearFileFindHighlights() {
  const containers = [el.fileViewerMarkdown, el.fileViewerCodeBlock];
  containers.forEach(container => {
    if (!container) return;
    const marks = container.querySelectorAll('mark.find-match');
    marks.forEach(mark => {
      const parent = mark.parentNode;
      if (!parent) return;
      while (mark.firstChild) {
        parent.insertBefore(mark.firstChild, mark);
      }
      parent.removeChild(mark);
      parent.normalize();
    });
  });
}

function executeFileFind() {
  clearFileFindHighlights();
  state.fileFindMatches = [];
  state.fileFindActiveIndex = -1;

  const query = el.fileFindInput ? el.fileFindInput.value : '';
  state.fileFindQuery = query;

  if (!query) {
    if (el.fileFindCountBadge) {
      el.fileFindCountBadge.textContent = '0 of 0';
      el.fileFindCountBadge.className = 'find-count-badge';
    }
    return;
  }

  const flags = state.fileFindCaseSensitive ? 'g' : 'gi';
  let regex;
  try {
    regex = new RegExp(escapeRegExp(query), flags);
  } catch (_) {
    return;
  }

  // Determine active view container
  const activeContainer = (state.fileViewerMode === 'preview' && el.fileViewerMarkdown && el.fileViewerMarkdown.style.display !== 'none')
    ? el.fileViewerMarkdown
    : el.fileViewerCodeBlock;

  if (!activeContainer) return;

  // Walk text nodes and wrap matches in <mark class="find-match">
  const walker = document.createTreeWalker(
    activeContainer,
    NodeFilter.SHOW_TEXT,
    {
      acceptNode(node) {
        if (!node.nodeValue.trim()) return NodeFilter.FILTER_SKIP;
        const parent = node.parentNode;
        if (parent && (parent.nodeName === 'SCRIPT' || parent.nodeName === 'STYLE' || parent.classList?.contains('find-match'))) {
          return NodeFilter.FILTER_REJECT;
        }
        return NodeFilter.FILTER_ACCEPT;
      }
    }
  );

  const textNodes = [];
  let currentNode;
  while ((currentNode = walker.nextNode())) {
    textNodes.push(currentNode);
  }

  let matchIndex = 0;
  textNodes.forEach(node => {
    const text = node.nodeValue;
    regex.lastIndex = 0;
    if (!regex.test(text)) return;

    regex.lastIndex = 0;
    const frag = document.createDocumentFragment();
    let lastIdx = 0;
    let match;

    while ((match = regex.exec(text)) !== null) {
      if (match.index > lastIdx) {
        frag.appendChild(document.createTextNode(text.substring(lastIdx, match.index)));
      }
      const mark = document.createElement('mark');
      mark.className = 'find-match';
      mark.dataset.findIdx = matchIndex++;
      mark.textContent = match[0];
      frag.appendChild(mark);
      state.fileFindMatches.push(mark);

      lastIdx = regex.lastIndex;
    }

    if (lastIdx < text.length) {
      frag.appendChild(document.createTextNode(text.substring(lastIdx)));
    }

    if (node.parentNode) {
      node.parentNode.replaceChild(frag, node);
    }
  });

  const total = state.fileFindMatches.length;
  if (total === 0) {
    if (el.fileFindCountBadge) {
      el.fileFindCountBadge.textContent = '0 of 0';
      el.fileFindCountBadge.className = 'find-count-badge no-matches';
    }
  } else {
    state.fileFindActiveIndex = 0;
    highlightActiveFindMatch();
  }
}

function highlightActiveFindMatch() {
  const total = state.fileFindMatches.length;
  if (total === 0 || state.fileFindActiveIndex < 0) return;

  state.fileFindMatches.forEach((mark, idx) => {
    mark.classList.toggle('find-match-current', idx === state.fileFindActiveIndex);
  });

  if (el.fileFindCountBadge) {
    el.fileFindCountBadge.textContent = `${state.fileFindActiveIndex + 1} of ${total}`;
    el.fileFindCountBadge.className = 'find-count-badge has-matches';
  }

  const currentMark = state.fileFindMatches[state.fileFindActiveIndex];
  if (currentMark) {
    currentMark.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
}

function goToNextFindMatch() {
  if (state.fileFindMatches.length === 0) return;
  state.fileFindActiveIndex = (state.fileFindActiveIndex + 1) % state.fileFindMatches.length;
  highlightActiveFindMatch();
}

function goToPrevFindMatch() {
  if (state.fileFindMatches.length === 0) return;
  state.fileFindActiveIndex = (state.fileFindActiveIndex - 1 + state.fileFindMatches.length) % state.fileFindMatches.length;
  highlightActiveFindMatch();
}

// ============================================================================
//  Session Files & Artifacts Aggregator
// ============================================================================
function updateSessionFilesBadge(detail) {
  state.sessionFilesMap.clear();
  if (!detail || !detail.messages) {
    if (el.detailFilesCountBadge) el.detailFilesCountBadge.textContent = '0';
    return;
  }

  detail.messages.forEach(msg => {
    if (msg.tool_calls) {
      msg.tool_calls.forEach(tc => {
        const args = tc.args || {};
        let p = args.Path || args.path || args.AbsolutePath || args.TargetFile || args.file_path || args.FilePath;
        if (p && typeof p === 'string') {
          p = p.trim().replace(/^["']|["']$/g, '');
          const cleanP = p.replace(/^file:\/\//, '');
          if (cleanP && !cleanP.startsWith('{')) {
            const name = cleanP.split('/').pop() || cleanP;
            let fileType = 'read';
            if (tc.name === 'write_file' || tc.name === 'edit_file' || tc.name === 'write_to_file' || tc.name === 'replace_file_content') {
              fileType = 'modified';
            }
            if (args.ArtifactMetadata || cleanP.includes('/brain/') || cleanP.endsWith('.md')) {
              if (args.ArtifactMetadata || cleanP.includes('/brain/')) {
                fileType = 'artifact';
              }
            }
            if (!state.sessionFilesMap.has(cleanP) || fileType === 'artifact') {
              state.sessionFilesMap.set(cleanP, {
                path: cleanP,
                name: name,
                type: fileType,
                tool: tc.name,
                stepIndex: msg.step_index
              });
            }
          }
        }
      });
    }

    if (msg.content) {
      const linkRegex = /\[([^\]]+)\]\((?:file:\/\/)?([^)]+)\)/g;
      let match;
      while ((match = linkRegex.exec(msg.content)) !== null) {
        const rawHref = match[2];
        if (rawHref.startsWith('/') || rawHref.startsWith('file://')) {
          const cleanP = rawHref.replace(/^file:\/\//, '').split('#')[0];
          const name = match[1] || cleanP.split('/').pop() || cleanP;
          if (!state.sessionFilesMap.has(cleanP)) {
            state.sessionFilesMap.set(cleanP, {
              path: cleanP,
              name: name,
              type: 'reference',
              tool: 'markdown',
              stepIndex: msg.step_index
            });
          }
        }
      }
    }
  });

  if (el.detailFilesCountBadge) {
    const prevText = el.detailFilesCountBadge.textContent.trim();
    const prevCount = parseInt(prevText, 10);
    const count = state.sessionFilesMap.size;
    el.detailFilesCountBadge.textContent = count.toString();
    if (el.detailFilesBtn) {
      el.detailFilesBtn.title = `${count} files and artifacts touched in this session (click to inspect drawer)`;
    }
    if (!isNaN(prevCount) && prevCount !== count) {
      pulseCounterBadge(el.detailFilesCountBadge);
    }
  }
}

function renderSessionFilesList() {
  if (!el.sessionFilesModal || !el.sessionFilesListContainer) return;
  const files = Array.from(state.sessionFilesMap.values());
  if (el.sessionFilesTotalCount) el.sessionFilesTotalCount.textContent = files.length.toString();

  if (files.length === 0) {
    el.sessionFilesListContainer.innerHTML = `
      <div class="empty-state" style="padding: 2.5rem; text-align: center;">
        <div style="font-size: 2.2rem; margin-bottom: 0.5rem;">📄</div>
        <p style="color: var(--text-muted);">No file operations or artifacts detected in this session.</p>
      </div>
    `;
    return;
  }

  const typeOrder = { artifact: 0, modified: 1, read: 2, reference: 3 };
  files.sort((a, b) => (typeOrder[a.type] ?? 9) - (typeOrder[b.type] ?? 9));

  el.sessionFilesListContainer.innerHTML = files.map(f => {
    const typeBadge = f.type === 'artifact'
      ? '<span class="file-meta-pill" style="border-color: rgba(168, 85, 247, 0.5); color: #c084fc; background: rgba(168, 85, 247, 0.15);">✨ Artifact</span>'
      : f.type === 'modified'
      ? '<span class="file-meta-pill" style="border-color: rgba(245, 158, 11, 0.5); color: #fbbf24; background: rgba(245, 158, 11, 0.15);">✏️ Modified</span>'
      : '<span class="file-meta-pill">👁️ Read</span>';

    return `
      <div class="session-file-item" data-path="${escapeHtml(f.path)}">
        <div class="session-file-item-left">
          <span style="font-size: 1.3rem;">${getFileIcon(f.path)}</span>
          <div class="session-file-item-info">
            <div style="display: flex; align-items: center; gap: 8px;">
              <span class="session-file-name">${escapeHtml(f.name)}</span>
              ${typeBadge}
              <span style="font-size: 0.72rem; color: var(--text-dim);">Step ${f.stepIndex}</span>
            </div>
            <div class="session-file-path" title="${escapeHtml(f.path)}">${escapeHtml(f.path)}</div>
          </div>
        </div>
        <div style="display: flex; gap: 8px; align-items: center;">
          <button class="btn-ghost-sm open-file-item-btn" data-path="${escapeHtml(f.path)}" title="Open in Browser">Open ↗</button>
          <button class="btn-icon copy-file-path-btn" style="width: 28px; height: 28px; font-size: 0.75rem;" data-path="${escapeHtml(f.path)}" title="Copy Path">📋</button>
        </div>
      </div>
    `;
  }).join('');

  el.sessionFilesListContainer.querySelectorAll('.open-file-item-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      closeSessionFilesModal();
      openFileViewer(btn.dataset.path);
    };
  });

  el.sessionFilesListContainer.querySelectorAll('.session-file-item').forEach(item => {
    item.onclick = () => {
      closeSessionFilesModal();
      openFileViewer(item.dataset.path);
    };
  });

  el.sessionFilesListContainer.querySelectorAll('.copy-file-path-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.path, 'File path copied to clipboard!');
    };
  });
}

function openSessionFilesModal() {
  if (!el.sessionFilesModal) return;
  renderSessionFilesList();
  el.sessionFilesModal.classList.add('active');
}

function closeSessionFilesModal() {
  if (el.sessionFilesModal) el.sessionFilesModal.classList.remove('active');
}

// ==========================================================================
// Agent Brain & Behavioral Context Matrix Controller (R-039)
// ==========================================================================

function initAgentMatrixListeners() {
  if (el.agentBrainBtn) {
    el.agentBrainBtn.addEventListener('click', () => {
      openAgentBrainModal();
    });
  }

  if (el.closeAgentMatrixModalBtn) {
    el.closeAgentMatrixModalBtn.addEventListener('click', closeAgentMatrixModal);
  }

  if (el.detailAgentContextBtn) {
    el.detailAgentContextBtn.addEventListener('click', () => {
      if (state.activeTimelineSessionId) {
        openSessionContextModal(state.activeTimelineSessionId);
      }
    });
  }

  if (el.closeSessionContextModalBtn) {
    el.closeSessionContextModalBtn.addEventListener('click', closeSessionContextModal);
  }

  // Session Matrix Lens Switcher (R-040)
  if (el.sessionMatrixLensControl) {
    el.sessionMatrixLensControl.querySelectorAll('.segment-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        el.sessionMatrixLensControl.querySelectorAll('.segment-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        state.sessionMatrix.lens = btn.dataset.lens;
        renderSessionMatrix();
      });
    });
  }

  // Session Matrix Filter Chips (R-040)
  const sessionTypeChips = document.querySelectorAll('.session-matrix-filter-chips .matrix-filter-chip, .session-matrix-filter-chips .chip');
  sessionTypeChips.forEach(chip => {
    chip.addEventListener('click', () => {
      sessionTypeChips.forEach(c => c.classList.remove('active'));
      chip.classList.add('active');
      state.sessionMatrix.type = chip.dataset.type;
      renderSessionMatrix();
    });
  });

  // Session Matrix Search Input (R-040)
  if (el.sessionMatrixSearchInput) {
    el.sessionMatrixSearchInput.addEventListener('input', (e) => {
      state.sessionMatrix.search = e.target.value.trim().toLowerCase();
      if (el.clearSessionMatrixSearchBtn) {
        el.clearSessionMatrixSearchBtn.style.display = state.sessionMatrix.search ? 'inline-flex' : 'none';
      }
      renderSessionMatrix();
    });
  }

  if (el.clearSessionMatrixSearchBtn) {
    el.clearSessionMatrixSearchBtn.addEventListener('click', () => {
      state.sessionMatrix.search = '';
      if (el.sessionMatrixSearchInput) el.sessionMatrixSearchInput.value = '';
      el.clearSessionMatrixSearchBtn.style.display = 'none';
      renderSessionMatrix();
    });
  }

  // Scope switcher
  if (el.agentMatrixScopeControl) {
    el.agentMatrixScopeControl.querySelectorAll('.segment-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        el.agentMatrixScopeControl.querySelectorAll('.segment-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        state.agentMatrix.scope = btn.dataset.scope;
        renderAgentMatrix();
      });
    });
  }

  // Type filter chips
  const typeChips = document.querySelectorAll('.agent-matrix-filter-chips .matrix-filter-chip, .agent-matrix-filter-chips .chip');
  typeChips.forEach(chip => {
    chip.addEventListener('click', () => {
      typeChips.forEach(c => c.classList.remove('active'));
      chip.classList.add('active');
      state.agentMatrix.type = chip.dataset.type;
      renderAgentMatrix();
    });
  });

  // Search input
  if (el.agentMatrixSearchInput) {
    el.agentMatrixSearchInput.addEventListener('input', (e) => {
      state.agentMatrix.search = e.target.value.trim().toLowerCase();
      if (el.clearAgentMatrixSearchBtn) {
        el.clearAgentMatrixSearchBtn.style.display = state.agentMatrix.search ? 'inline-flex' : 'none';
      }
      renderAgentMatrix();
    });
  }

  if (el.clearAgentMatrixSearchBtn) {
    el.clearAgentMatrixSearchBtn.addEventListener('click', () => {
      state.agentMatrix.search = '';
      if (el.agentMatrixSearchInput) el.agentMatrixSearchInput.value = '';
      el.clearAgentMatrixSearchBtn.style.display = 'none';
      renderAgentMatrix();
    });
  }
}

async function openAgentBrainModal(targetWorkspace) {
  if (!el.agentMatrixModal) return;

  const ws = targetWorkspace || state.drilledProject || state.filterWorkspace || null;
  state.agentMatrix.targetWorkspace = ws;

  if (el.agentMatrixWorkspaceSegment) {
    if (ws) {
      el.agentMatrixWorkspaceSegment.textContent = `📁 ${ws}`;
      el.agentMatrixWorkspaceSegment.style.display = 'inline-block';
    } else {
      el.agentMatrixWorkspaceSegment.textContent = '📁 Workspace';
      el.agentMatrixWorkspaceSegment.style.display = 'inline-block';
    }
  }

  // If a workspace is specified, default to workspace or cascade scope
  if (ws && state.agentMatrix.scope === 'global') {
    state.agentMatrix.scope = 'workspace';
    if (el.agentMatrixScopeControl) {
      el.agentMatrixScopeControl.querySelectorAll('.segment-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.scope === 'workspace');
      });
    }
  }

  el.agentMatrixModal.classList.add('active');
  if (el.agentMatrixContent) {
    el.agentMatrixContent.innerHTML = '<div class="file-viewer-loading"><div class="loading-spinner"></div><span>Discovering agent rules, skills, and constitutions...</span></div>';
  }

  try {
    const promises = [
      fetch('/api/agent/context/global').then(r => r.json())
    ];
    if (ws) {
      promises.push(fetch(`/api/agent/context/workspace?workspace=${encodeURIComponent(ws)}`).then(r => r.json()));
    }

    const [globalCtx, wsCtx] = await Promise.all(promises);
    state.agentMatrix.globalContext = globalCtx;
    state.agentMatrix.workspaceContext = wsCtx || null;

    renderAgentMatrix();
  } catch (err) {
    console.error('Failed to load agent context:', err);
    if (el.agentMatrixContent) {
      el.agentMatrixContent.innerHTML = `<div class="file-viewer-error">Failed to discover agent context: ${escapeHtml(err.message)}</div>`;
    }
  }
}

function closeAgentMatrixModal() {
  if (el.agentMatrixModal) el.agentMatrixModal.classList.remove('active');
}

// ============================================================================
//  EC2-Style Greedy Bin-Packing Spatial Placement Engine (R-043)
// ============================================================================
function binPackCards(cardItems, containerWidth) {
  let numCols = 3;
  if (containerWidth < 768) {
    numCols = 1;
  } else if (containerWidth < 1200) {
    numCols = 2;
  }

  const columns = Array.from({ length: numCols }, () => []);
  const heights = new Array(numCols).fill(0);

  cardItems.forEach(item => {
    // Greedy placement: find the column bin with minimal total weight/height load
    let target = 0;
    let minH = heights[0];
    for (let c = 1; c < numCols; c++) {
      if (heights[c] < minH) {
        minH = heights[c];
        target = c;
      }
    }
    columns[target].push(item.html);
    heights[target] += (item.weight || 150);
  });

  return `
    <div class="agent-matrix-binpack" data-columns="${numCols}">
      ${columns.map(col => `<div class="binpack-column">${col.join('')}</div>`).join('')}
    </div>
  `;
}

function renderAgentMatrix() {
  if (!el.agentMatrixContent) return;

  const { scope, type, search, globalContext, workspaceContext } = state.agentMatrix;
  if (!globalContext) return;

  let rules = [];
  let skills = [];
  let plugins = [];
  let mcpServers = [];
  let hooks = [];

  if (scope === 'global' || !workspaceContext) {
    rules = globalContext.rules || [];
    skills = globalContext.skills || [];
    plugins = globalContext.plugins || [];
    mcpServers = globalContext.mcp_servers || [];
    hooks = globalContext.hooks || [];
  } else if (scope === 'workspace') {
    rules = (workspaceContext.rules || []).filter(r => r.scope === 'workspace');
    skills = (workspaceContext.skills || []).filter(s => s.scope === 'workspace');
    plugins = [];
    mcpServers = (workspaceContext.mcp_servers || []).filter(m => m.scope === 'workspace');
    hooks = (workspaceContext.hooks || []).filter(h => h.scope === 'workspace');
  } else {
    // 'cascade': workspace overrides + inherited global items
    rules = workspaceContext.rules || [];
    skills = workspaceContext.skills || [];
    plugins = globalContext.plugins || [];
    mcpServers = workspaceContext.mcp_servers || globalContext.mcp_servers || [];
    hooks = workspaceContext.hooks || globalContext.hooks || [];
  }

  // Update badge counts in toolbar
  if (el.agentMatrixRulesCount) el.agentMatrixRulesCount.textContent = rules.length;
  if (el.agentMatrixSkillsCount) el.agentMatrixSkillsCount.textContent = skills.length;
  if (el.agentMatrixPluginsCount) el.agentMatrixPluginsCount.textContent = (plugins.length + mcpServers.length);
  if (el.agentMatrixHooksCount) el.agentMatrixHooksCount.textContent = hooks.length;

  // Update GenAI context bloat metrics
  const preloadedToks = (scope === 'workspace' && workspaceContext) ? (workspaceContext.preloaded_tokens_est || 0) : (globalContext.preloaded_tokens_est || 0);
  const ondemandToks = (scope === 'workspace' && workspaceContext) ? (workspaceContext.on_demand_tokens_est || 0) : (globalContext.on_demand_tokens_est || 0);
  const totalWords = (scope === 'workspace' && workspaceContext) ? (workspaceContext.total_words || 0) : (globalContext.total_words || 0);
  const totalTokens = preloadedToks + ondemandToks;

  const preloadedEl = document.getElementById('agentMatrixPreloadedVal');
  const ondemandEl = document.getElementById('agentMatrixOndemandVal');
  const totalWordsEl = document.getElementById('agentMatrixTotalWordsVal');

  if (preloadedEl) preloadedEl.textContent = `~${preloadedToks.toLocaleString()} tokens`;
  if (ondemandEl) ondemandEl.textContent = `~${ondemandToks.toLocaleString()} tokens`;
  if (totalWordsEl) totalWordsEl.textContent = `${totalWords.toLocaleString()} words (~${totalTokens.toLocaleString()} tokens)`;

  // Filter items by type
  let items = [];
  if (type === 'ALL' || type === 'RULES') {
    rules.forEach(r => items.push({ kind: 'rule', data: r }));
  }
  if (type === 'ALL' || type === 'SKILLS') {
    skills.forEach(s => items.push({ kind: 'skill', data: s }));
  }
  if (type === 'ALL' || type === 'PLUGINS') {
    plugins.forEach(p => items.push({ kind: 'plugin', data: p }));
    mcpServers.forEach(m => items.push({ kind: 'mcp', data: m }));
  }
  if (type === 'ALL' || type === 'HOOKS') {
    hooks.forEach(h => items.push({ kind: 'hook', data: h }));
  }

  // Filter by search query
  if (search) {
    items = items.filter(it => {
      const d = it.data;
      const textToSearch = [
        d.name,
        d.title,
        d.description,
        d.file_path || d.path || d.command,
        d.provider,
        d.priority,
        d.trigger
      ].filter(Boolean).join(' ').toLowerCase();
      return textToSearch.includes(search);
    });
  }

  if (el.agentMatrixTotalCountBadge) {
    el.agentMatrixTotalCountBadge.textContent = `${items.length} item${items.length === 1 ? '' : 's'}`;
  }

  if (items.length === 0) {
    el.agentMatrixContent.innerHTML = `
      <div class="empty-state" style="padding: 3rem 1.5rem; text-align: center;">
        <div style="font-size: 2.2rem; margin-bottom: 0.5rem;">🧠</div>
        <p style="color: var(--text-muted);">No rules, skills, or instructions matching the active filter.</p>
      </div>
    `;
    return;
  }

  const renderedCards = [];
  items.forEach(it => {
    let cardHtml = '';
    let weight = 160;

    if (it.kind === 'rule') {
      const r = it.data;
      const isConst = (r.priority || '').includes('Constitutional') || r.name.includes('constitution') || r.name.includes('101');
      weight = 160;
      cardHtml = `
        <div class="agent-card">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">📜 ${escapeHtml(r.title || r.name)}</span>
              <div style="display: flex; gap: 0.35rem; align-items: center; flex-wrap: wrap;">
                <span class="agent-card-scope-badge ${r.scope}">${r.scope}</span>
                <span class="agent-card-priority-pill ${isConst ? 'constitutional' : ''}">${escapeHtml(r.priority || 'Standard')}</span>
                ${r.trigger && r.trigger !== 'always_on' ? `<span class="agent-card-priority-pill" style="color: #67e8f9;">⚡ ${escapeHtml(r.trigger)}</span>` : ''}
              </div>
            </div>
            <span style="font-size: 0.72rem; color: var(--text-dim); text-transform: uppercase;">${escapeHtml(r.provider || 'antigravity')}</span>
          </div>
          <p class="agent-card-desc">${escapeHtml(r.description || 'No summary description available.')}</p>
          <div class="agent-card-meta">
            <span class="agent-token-badge">⚡ ~${(r.tokens_est || 0).toLocaleString()} tok</span>
            <span class="agent-word-badge">🔤 ${(r.word_count || 0).toLocaleString()} words</span>
            <span>📏 ${r.line_count} lines</span>
            <span>💾 ${Math.max(1, Math.round(r.size_bytes / 1024))} KB</span>
            ${r.workspace ? `<span>📁 ${escapeHtml(r.workspace)}</span>` : ''}
          </div>
          <div class="agent-card-actions">
            <span class="agent-card-path" title="${escapeHtml(r.file_path)}">${escapeHtml(r.file_path)}</span>
            <div style="display: flex; gap: 0.4rem;">
              <button class="btn-ghost-sm inspect-agent-file-btn" data-path="${escapeHtml(r.file_path)}" title="Open file in embedded File Viewer">📖 Inspect</button>
              <button class="btn-icon copy-agent-path-btn" style="width: 26px; height: 26px; font-size: 0.75rem;" data-path="${escapeHtml(r.file_path)}" title="Copy Path">📋</button>
            </div>
          </div>
        </div>
      `;
    } else if (it.kind === 'skill') {
      const s = it.data;
      weight = 165;
      cardHtml = `
        <div class="agent-card">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">⚡ ${escapeHtml(s.name)}</span>
              <div style="display: flex; gap: 0.35rem; align-items: center; flex-wrap: wrap;">
                <span class="agent-card-scope-badge ${s.scope}">${s.scope}</span>
                ${s.category ? `<span class="agent-card-priority-pill">${escapeHtml(s.category)}</span>` : ''}
                ${s.plugin_name ? `<span class="agent-card-priority-pill" style="color: #facc15;">🧩 ${escapeHtml(s.plugin_name)}</span>` : ''}
              </div>
            </div>
            <span style="font-size: 0.72rem; color: var(--text-dim); text-transform: uppercase;">${escapeHtml(s.provider || 'antigravity')}</span>
          </div>
          <p class="agent-card-desc">${escapeHtml(s.description || 'On-demand procedural runbook for agentic execution.')}</p>
          <div class="agent-card-meta">
            <span class="agent-token-badge" style="color: #34d399; border-color: rgba(16, 185, 129, 0.3); background: rgba(16, 185, 129, 0.08);">⚡ ~${(s.tokens_est || 0).toLocaleString()} tok</span>
            <span class="agent-word-badge">🔤 ${(s.word_count || 0).toLocaleString()} words</span>
            <span>📏 ${s.line_count} lines</span>
            ${s.has_scripts ? '<span>🛠️ Scripts</span>' : ''}
            ${s.has_examples ? '<span>💡 Examples</span>' : ''}
            ${s.has_references ? '<span>📚 References</span>' : ''}
            ${s.workspace ? `<span>📁 ${escapeHtml(s.workspace)}</span>` : ''}
          </div>
          <div class="agent-card-actions">
            <span class="agent-card-path" title="${escapeHtml(s.file_path)}">${escapeHtml(s.file_path)}</span>
            <div style="display: flex; gap: 0.4rem;">
              <button class="btn-ghost-sm inspect-agent-file-btn" data-path="${escapeHtml(s.file_path)}" title="Open skill in embedded File Viewer">📖 Inspect</button>
              <button class="btn-icon copy-agent-path-btn" style="width: 26px; height: 26px; font-size: 0.75rem;" data-path="${escapeHtml(s.file_path)}" title="Copy Path">📋</button>
            </div>
          </div>
        </div>
      `;
    } else if (it.kind === 'plugin') {
      const p = it.data;
      weight = 140;
      cardHtml = `
        <div class="agent-card">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">🧩 ${escapeHtml(p.name)}</span>
              <span class="agent-card-scope-badge plugin">PLUGIN</span>
            </div>
            <span style="font-size: 0.75rem; color: #a5b4fc;">v${escapeHtml(p.version || '1.0.0')}</span>
          </div>
          <p class="agent-card-desc">${escapeHtml(p.description || 'Packaged agent customization bundle.')}</p>
          <div class="agent-card-meta">
            <span>⚡ ${(p.skills || []).length} skills</span>
            <span>📋 ${(p.policies || []).length} policies</span>
          </div>
          <div class="agent-card-actions">
            <span class="agent-card-path" title="${escapeHtml(p.path)}">${escapeHtml(p.path)}</span>
            <button class="btn-icon copy-agent-path-btn" style="width: 26px; height: 26px; font-size: 0.75rem;" data-path="${escapeHtml(p.path)}" title="Copy Path">📋</button>
          </div>
        </div>
      `;
    } else if (it.kind === 'mcp') {
      const m = it.data;
      weight = 130;
      cardHtml = `
        <div class="agent-card">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">🔌 ${escapeHtml(m.name)}</span>
              <span class="agent-card-scope-badge ${m.scope}">MCP SERVER</span>
            </div>
          </div>
          <p class="agent-card-desc" style="font-family: monospace; font-size: 0.78rem;">Command: ${escapeHtml(m.command || 'custom')}</p>
          <div class="agent-card-meta">
            <span>🔧 ${(m.args || []).join(' ')}</span>
          </div>
        </div>
      `;
    } else if (it.kind === 'hook') {
      const h = it.data;
      weight = 130;
      cardHtml = `
        <div class="agent-card">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">🪝 ${escapeHtml(h.name || h.event)}</span>
              <span class="agent-card-scope-badge ${h.scope}">${escapeHtml(h.event)}</span>
            </div>
          </div>
          <p class="agent-card-desc" style="font-family: monospace; font-size: 0.78rem;">Matcher: ${escapeHtml(h.matcher || 'all')}</p>
          <div class="agent-card-actions">
            <span class="agent-card-path" title="${escapeHtml(h.command)}">${escapeHtml(h.command)}</span>
            <button class="btn-ghost-sm inspect-agent-file-btn" data-path="${escapeHtml(h.command)}" title="Open hook in File Viewer">📖 Inspect</button>
          </div>
        </div>
      `;
    }

    if (cardHtml) {
      renderedCards.push({ html: cardHtml, weight });
    }
  });

  const containerW = el.agentMatrixContent.clientWidth || window.innerWidth;
  el.agentMatrixContent.innerHTML = binPackCards(renderedCards, containerW);

  // Attach 1-click Inspect to openFileViewer
  el.agentMatrixContent.querySelectorAll('.inspect-agent-file-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      openFileViewer(btn.dataset.path);
    };
  });

  // Attach Copy Path
  el.agentMatrixContent.querySelectorAll('.copy-agent-path-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.path, 'Path copied to clipboard!');
    };
  });
}

// In-Session Holographic Agent Matrix Modal (R-039, R-040)
async function openSessionContextModal(sessionId) {
  if (!el.sessionContextModal || !el.sessionContextContent) return;

  const sid = sessionId || state.activeTimelineSessionId || state.sessionMatrix.sessionId;
  if (!sid) return;

  state.sessionMatrix.sessionId = sid;
  if (el.sessionContextIdBadge) {
    el.sessionContextIdBadge.textContent = sid.length > 8 ? sid.slice(0, 8) : sid;
    el.sessionContextIdBadge.title = `Full Session UUID: ${sid}`;
  }
  el.sessionContextModal.classList.add('active');
  el.sessionContextContent.innerHTML = '<div class="file-viewer-loading"><div class="loading-spinner"></div><span>Synthesizing holographic session matrix &amp; tool telemetry...</span></div>';

  try {
    const res = await fetch(`/api/sessions/${sid}/context`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const ctx = await res.json();
    state.sessionMatrix.context = ctx;
    renderSessionMatrix();
  } catch (err) {
    console.error('Failed to load session context:', err);
    el.sessionContextContent.innerHTML = `<div class="file-viewer-error">Error loading session matrix: ${escapeHtml(err.message)}</div>`;
  }
}

function closeSessionContextModal() {
  if (el.sessionContextModal) el.sessionContextModal.classList.remove('active');
}

function jumpToTimelineStep(stepIndex) {
  closeSessionContextModal();

  let targetNode = document.getElementById(`step-${stepIndex}`) || document.querySelector(`.message-item[data-step="${stepIndex}"]`);

  // If node is currently hidden by step filter or search, reset them to reveal the turn
  if (!targetNode) {
    if (el.sessionStepSearchInput && el.sessionStepSearchInput.value) {
      el.sessionStepSearchInput.value = '';
      state.sessionStepSearch = '';
    }
    if (state.messageFilter !== 'ALL') {
      state.messageFilter = 'ALL';
      if (el.msgFilterChips) {
        el.msgFilterChips.forEach(c => {
          c.classList.toggle('active', c.dataset.type === 'ALL');
        });
      }
    }
    if (typeof renderTimelineMessages === 'function') {
      renderTimelineMessages();
    }
    targetNode = document.getElementById(`step-${stepIndex}`) || document.querySelector(`.message-item[data-step="${stepIndex}"]`);
  }

  if (targetNode) {
    targetNode.scrollIntoView({ behavior: 'smooth', block: 'center' });
    targetNode.classList.remove('step-target-highlight');
    const card = targetNode.querySelector('.message-card');
    if (card) card.classList.remove('step-target-highlight');

    // Force reflow for smooth re-triggering of animation
    void targetNode.offsetWidth;
    targetNode.classList.add('step-target-highlight');
    if (card) card.classList.add('step-target-highlight');

    setTimeout(() => {
      targetNode.classList.remove('step-target-highlight');
      if (card) card.classList.remove('step-target-highlight');
    }, 2800);

    showToast(`⚡ Navigated to Step #${stepIndex}`);
  } else {
    showToast(`Step #${stepIndex} recorded in session telemetry`);
  }
}

// Category color and icon helpers for Holographic UI (R-041)
function getCategoryColor(cat) {
  switch ((cat || '').toLowerCase()) {
    case 'vcs': return '#a855f7'; // Neon Violet
    case 'build': return '#06b6d4'; // Cyber Cyan
    case 'test': return '#f59e0b'; // Solar Amber
    case 'process': return '#ec4899'; // Hot Magenta
    case 'inspect': return '#10b981'; // Electric Emerald
    case 'network': return '#38bdf8'; // Ice Blue
    case 'runtime': return '#6366f1'; // Indigo Neon
    case 'filesystem': return '#14b8a6'; // Teal
    default: return '#94a3b8'; // Slate
  }
}

function getCategoryIcon(cat) {
  switch ((cat || '').toLowerCase()) {
    case 'vcs': return '🌿';
    case 'build': return '⚙️';
    case 'test': return '🧪';
    case 'process': return '⚡';
    case 'inspect': return '🔍';
    case 'network': return '🌐';
    case 'runtime': return '💻';
    case 'filesystem': return '📁';
    default: return '🛠️';
  }
}

function renderSessionMatrix() {
  if (!el.sessionContextContent || !state.sessionMatrix.context) return;

  const ctx = state.sessionMatrix.context;
  const { lens, type, search } = state.sessionMatrix;

  const activeRules = ctx.active_rules || [];
  const activeSkills = ctx.active_skills || [];
  const availableSkills = ctx.available_skills || [];
  const mcpServers = ctx.mcp_servers || [];
  const nativeTools = ctx.native_tools || [];
  const analytics = ctx.tool_analytics || {
    total_tool_calls: 0,
    total_mcp_calls: 0,
    total_native_calls: 0,
    unique_tools_used: 0,
    tool_diversity_index: 0,
    total_errors: 0
  };
  const subagents = ctx.subagents || [];
  const slashCommands = ctx.slash_commands || [];

  const activeTokens = ctx.active_tokens_est || 0;
  const ondemandTokens = ctx.on_demand_tokens_est || 0;
  const totalWordsInj = ctx.total_words_injected || 0;

  // Compute active vs dormant MCP tools
  let activeMCPTools = [];
  let dormantMCPTools = [];
  let activeMCPServers = 0;

  mcpServers.forEach(srv => {
    if (srv.invoked_count > 0 || (srv.active_tools && srv.active_tools.length > 0)) {
      activeMCPServers++;
    }
    (srv.active_tools || []).forEach(t => activeMCPTools.push(t));
    (srv.dormant_tools || []).forEach(dtName => {
      dormantMCPTools.push({
        name: dtName,
        server_name: srv.name,
        invoked_count: 0
      });
    });
  });

  const totalToolCalls = analytics.total_tool_calls || 0;
  const totalErrors = analytics.total_errors || 0;
  const reliabilityPct = totalToolCalls > 0
    ? Math.max(0, Math.round(((totalToolCalls - totalErrors) / totalToolCalls) * 100))
    : 100;
  const diversityPct = (analytics.tool_diversity_index * 100).toFixed(1);

  // Update Holographic Telemetry Ribbon
  if (el.sessionMatrixInjectedVal) el.sessionMatrixInjectedVal.textContent = `~${activeTokens.toLocaleString()} tok`;
  if (el.sessionMatrixMCPVal) el.sessionMatrixMCPVal.textContent = `${analytics.total_mcp_calls} calls (${activeMCPTools.length} tools)`;
  if (el.sessionMatrixDiversityVal) el.sessionMatrixDiversityVal.textContent = `${diversityPct}% (${analytics.unique_tools_used} distinct)`;
  if (el.sessionMatrixOndemandVal) el.sessionMatrixOndemandVal.textContent = `~${ondemandTokens.toLocaleString()} tok`;
  if (el.sessionMatrixReliabilityVal) el.sessionMatrixReliabilityVal.textContent = `${reliabilityPct}% (${totalErrors} err)`;

  // Calculate element counts for filter chips
  const countSkills = activeSkills.length + (lens === 'cascade' ? availableSkills.length : 0);
  const countMCP = activeMCPTools.length + (lens === 'cascade' ? dormantMCPTools.length : 0);
  const countNative = nativeTools.length;
  const countRules = activeRules.length;
  const countOrch = subagents.length + slashCommands.length;
  const countTotal = countSkills + countMCP + countNative + countRules + countOrch;

  if (el.sessionMatrixChipSkillsLabel) {
    el.sessionMatrixChipSkillsLabel.textContent = lens === 'cascade' ? 'Skills (Active & Standby)' : 'Invoked Skills';
  }
  if (el.sessionMatrixChipAll) el.sessionMatrixChipAll.textContent = countTotal;
  if (el.sessionMatrixChipSkills) {
    el.sessionMatrixChipSkills.textContent = lens === 'cascade' && activeSkills.length > 0
      ? `${activeSkills.length} / ${countSkills}`
      : countSkills;
  }
  if (el.sessionMatrixChipMCP) el.sessionMatrixChipMCP.textContent = countMCP;
  if (el.sessionMatrixChipNative) el.sessionMatrixChipNative.textContent = countNative;
  if (el.sessionMatrixChipRules) el.sessionMatrixChipRules.textContent = countRules;
  if (el.sessionMatrixChipOrch) el.sessionMatrixChipOrch.textContent = countOrch;

  // LENS 3: Holographic Telemetry & Execution Audit Dashboard
  if (lens === 'telemetry') {
    if (el.sessionMatrixTotalCountBadge) el.sessionMatrixTotalCountBadge.textContent = 'Telemetry Mode';

    let html = `
      <div class="holo-audit-dashboard">
        <!-- High-Precision Metric Cards -->
        <div class="holo-metrics-grid">
          <div class="holo-metric-card" style="border-color: rgba(56, 189, 248, 0.4);">
            <div class="holo-metric-lbl">Total Tool Invocations</div>
            <div class="holo-metric-num" style="color: #38bdf8;">${totalToolCalls.toLocaleString()}</div>
            <div style="font-size: 0.76rem; color: var(--text-dim);">
              <span>🔌 ${analytics.total_mcp_calls} MCP</span> • <span>🛠️ ${analytics.total_native_calls} Native</span>
            </div>
          </div>

          <div class="holo-metric-card" style="border-color: rgba(168, 85, 247, 0.4);">
            <div class="holo-metric-lbl">Tool Diversity Index</div>
            <div class="holo-metric-num" style="color: #c084fc;">${diversityPct}%</div>
            <div style="font-size: 0.76rem; color: var(--text-dim);">${analytics.unique_tools_used} distinct tools utilized</div>
            <div class="holo-meter-wrapper">
              <div class="holo-meter-fill diversity" style="width: ${Math.min(100, Math.max(8, analytics.tool_diversity_index * 100))}%;"></div>
            </div>
          </div>

          <div class="holo-metric-card" style="border-color: rgba(16, 185, 129, 0.4);">
            <div class="holo-metric-lbl">Reliability Quotient</div>
            <div class="holo-metric-num" style="color: #34d399;">${reliabilityPct}%</div>
            <div style="font-size: 0.76rem; color: var(--text-dim);">${totalErrors} execution error${totalErrors === 1 ? '' : 's'} recorded</div>
            <div class="holo-meter-wrapper">
              <div class="holo-meter-fill reliability" style="width: ${reliabilityPct}%;"></div>
            </div>
          </div>

          <div class="holo-metric-card">
            <div class="holo-metric-lbl">Active Injected Context</div>
            <div class="holo-metric-num" style="color: #7dd3fc;">~${activeTokens.toLocaleString()}</div>
            <div style="font-size: 0.76rem; color: var(--text-dim);">${totalWordsInj.toLocaleString()} words injected across turns</div>
          </div>

          <div class="holo-metric-card">
            <div class="holo-metric-lbl">Dormant Bloat Capacity</div>
            <div class="holo-metric-num" style="color: #fbbf24;">~${ondemandTokens.toLocaleString()}</div>
            <div style="font-size: 0.76rem; color: var(--text-dim);">${availableSkills.length} skills on standby</div>
          </div>

          <div class="holo-metric-card">
            <div class="holo-metric-lbl">Active Tool Ecosystem</div>
            <div class="holo-metric-num" style="color: #a5b4fc;">${activeMCPServers} / ${mcpServers.length}</div>
            <div style="font-size: 0.76rem; color: var(--text-dim);">${activeMCPServers} active MCP servers • ${nativeTools.length} native tools</div>
          </div>
        </div>

        <!-- Tool Invocations Heatmap Timeline -->
        <div class="holo-timeline-heatmap">
          <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap;">
            <span style="font-size: 0.84rem; font-weight: 600; color: #e2e8f0;">⚡ Execution Velocity &amp; Distribution</span>
            <span style="font-size: 0.74rem; color: var(--text-muted); font-family: var(--font-mono, monospace);">
              Trajectory: ${activeMCPTools.length} MCP tools • ${nativeTools.length} Native tools
            </span>
          </div>

          <div class="holo-heatmap-bar">
            ${analytics.total_mcp_calls > 0 ? `<div class="holo-heatmap-segment" style="width: ${(analytics.total_mcp_calls / totalToolCalls) * 100}%; background: linear-gradient(90deg, #0284c7, #06b6d4);" title="MCP Tools: ${analytics.total_mcp_calls} calls">🔌 MCP (${analytics.total_mcp_calls})</div>` : ''}
            ${analytics.total_native_calls > 0 ? `<div class="holo-heatmap-segment" style="width: ${(analytics.total_native_calls / totalToolCalls) * 100}%; background: linear-gradient(90deg, #7c3aed, #a855f7);" title="Native Tools: ${analytics.total_native_calls} calls">🛠️ Native (${analytics.total_native_calls})</div>` : ''}
          </div>
        </div>

        <!-- Tool Invocations Leaderboard -->
        <div style="background: rgba(15, 23, 42, 0.6); border: 1px solid var(--border-glass); border-radius: 12px; padding: 1.25rem;">
          <div style="font-size: 0.86rem; font-weight: 600; color: #e2e8f0; margin-bottom: 0.85rem; display: flex; align-items: center; justify-content: space-between;">
            <span>🏆 Tool Frequency Leaderboard</span>
            <span style="font-size: 0.74rem; color: var(--text-muted);">Click any step pill to jump straight to that turn</span>
          </div>
          <div style="display: flex; flex-direction: column; gap: 0.6rem;">
    `;

    // Combine all tools into ranked list
    const allRankedTools = [
      ...activeMCPTools.map(t => ({ ...t, kind: 'mcp' })),
      ...nativeTools.map(t => ({ ...t, kind: 'native' }))
    ].sort((a, b) => b.invoked_count - a.invoked_count);

    if (allRankedTools.length === 0) {
      html += `<div style="text-align: center; color: var(--text-muted); padding: 1.5rem;">No tool calls recorded in this session.</div>`;
    } else {
      allRankedTools.forEach(t => {
        const isMCP = t.kind === 'mcp';
        const steps = t.step_indices || [];
        const displaySteps = steps.slice(0, 6);
        const hasMore = steps.length > 6;

        html += `
          <div class="mcp-tool-row" style="background: rgba(15, 23, 42, 0.45);">
            <div class="mcp-tool-row-header">
              <div style="display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap;">
                <span class="${isMCP ? 'mcp-tool-name' : 'tool-name'}" style="font-size: 0.86rem;">${isMCP ? '🔌' : '🛠️'} ${escapeHtml(t.name)}</span>
                ${isMCP ? `<span class="agent-card-scope-badge" style="background: rgba(56, 189, 248, 0.15); color: #38bdf8; font-size: 0.68rem;">${escapeHtml(t.server_name)}</span>` : `<span class="native-category-badge ${t.category || 'Agentic'}">${escapeHtml(t.category || 'Agentic')}</span>`}
              </div>
              <div style="display: flex; align-items: center; gap: 0.4rem;">
                <span class="mcp-tool-count-badge">${t.invoked_count}x</span>
                ${t.error_count > 0 ? `<span class="mcp-tool-count-badge" style="background: rgba(239, 68, 68, 0.2); color: #f87171; border-color: rgba(239, 68, 68, 0.4);">${t.error_count} err</span>` : ''}
              </div>
            </div>
            <div class="steps-row" style="margin-top: 2px;">
              <span class="steps-label" style="margin-right: 2px;">Steps:</span>
              ${displaySteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
              ${hasMore ? `<span class="step-marker-pill more" title="${steps.length} total executions">+${steps.length - 6} more</span>` : ''}
              ${t.recent_args_summary ? `<span class="mcp-tool-args-preview" title="${escapeHtml(t.recent_args_summary)}">${escapeHtml(t.recent_args_summary)}</span>` : ''}
            </div>
          </div>
        `;
      });
    }

    html += `</div></div>`;

    // Global Terminal Command Execution Spectrum (R-041)
    const cmdBreakdown = analytics.command_breakdown || [];
    if (cmdBreakdown.length > 0) {
      html += `
        <div style="background: rgba(15, 23, 42, 0.6); border: 1px solid var(--border-glass); border-radius: 12px; padding: 1.25rem;">
          <div style="font-size: 0.86rem; font-weight: 600; color: #e2e8f0; margin-bottom: 0.85rem; display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem;">
            <span>⚡ Terminal Command Execution Spectrum (${cmdBreakdown.length} binaries)</span>
            <span style="font-size: 0.74rem; color: var(--text-muted);">Decomposed by binary &amp; operational category</span>
          </div>

          <div class="command-spectrum-bar" style="height: 16px; margin-bottom: 1rem;">
            ${cmdBreakdown.map(cb => `
              <div class="spectrum-segment" style="width: ${cb.percentage}%; background-color: ${getCategoryColor(cb.category)};" title="${escapeHtml(cb.label)}: ${cb.count}x (${cb.percentage}%)"></div>
            `).join('')}
          </div>

          <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 0.75rem;">
            ${cmdBreakdown.map(cb => {
              const cbSteps = cb.step_indices || [];
              const dispSteps = cbSteps.slice(0, 5);
              const hasMoreSteps = cbSteps.length > 5;
              const catColor = getCategoryColor(cb.category);
              return `
                <div class="sub-command-row">
                  <div class="sub-command-header">
                    <div class="sub-command-label-group">
                      <span class="sub-command-icon">${getCategoryIcon(cb.category)}</span>
                      <span class="sub-command-key">${escapeHtml(cb.label || cb.key)}</span>
                      <span class="native-category-badge" style="background: ${catColor}20; color: ${catColor}; border: 1px solid ${catColor}40; font-size: 0.64rem;">${escapeHtml(cb.category || 'System')}</span>
                    </div>
                    <div class="sub-command-stats">
                      <span class="sub-command-count">${cb.count}x</span>
                      <span class="sub-command-pct">(${cb.percentage}%)</span>
                    </div>
                  </div>
                  ${cb.recent_items && cb.recent_items.length ? `
                    <div class="sub-command-recent" title="Recent: ${escapeHtml(cb.recent_items.join(' • '))}">
                      <span class="recent-tag">Recent:</span> ${escapeHtml(cb.recent_items[0])}
                    </div>
                  ` : ''}
                  <div class="sub-command-steps">
                    <span class="sub-steps-label">Steps:</span>
                    ${dispSteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
                    ${hasMoreSteps ? `<span class="step-marker-pill more" title="${cbSteps.length} total calls">+${cbSteps.length - 5}</span>` : ''}
                  </div>
                </div>
              `;
            }).join('')}
          </div>
        </div>
      `;
    }

    // Global Filesystem Operations Breakdown (R-041)
    const fileBreakdown = analytics.file_breakdown || [];
    if (fileBreakdown.length > 0) {
      html += `
        <div style="background: rgba(15, 23, 42, 0.6); border: 1px solid var(--border-glass); border-radius: 12px; padding: 1.25rem;">
          <div style="font-size: 0.86rem; font-weight: 600; color: #e2e8f0; margin-bottom: 0.85rem; display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 0.5rem;">
            <span>📁 Filesystem Operations &amp; Extensions (${fileBreakdown.length} types)</span>
            <span style="font-size: 0.74rem; color: var(--text-muted);">Decomposed by target file extension &amp; language</span>
          </div>
          <div class="file-ext-matrix">
            ${fileBreakdown.map(fb => {
              const fbSteps = fb.step_indices || [];
              const dispSteps = fbSteps.slice(0, 4);
              return `
                <div class="file-ext-chip">
                  <div style="display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; width: 100%;">
                    <span class="ext-name">${escapeHtml(fb.key)}</span>
                    <span class="ext-count">${fb.count}x (${fb.percentage}%)</span>
                  </div>
                  ${fb.recent_items && fb.recent_items.length ? `
                    <div class="ext-recent" title="${escapeHtml(fb.recent_items.join(', '))}">
                      ${escapeHtml(fb.recent_items[0])}
                    </div>
                  ` : ''}
                  <div class="steps-row" style="margin-top: 0.25rem;">
                    <span class="steps-label" style="font-size: 0.65rem;">Steps:</span>
                    ${dispSteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
                    ${fbSteps.length > 4 ? `<span class="step-marker-pill more" title="${fbSteps.length} total calls">+${fbSteps.length - 4}</span>` : ''}
                  </div>
                </div>
              `;
            }).join('')}
          </div>
        </div>
      `;
    }

    html += `</div>`;
    el.sessionContextContent.innerHTML = html;

    // Attach step click jump handlers
    el.sessionContextContent.querySelectorAll('.step-marker-pill[data-step]').forEach(btn => {
      btn.onclick = (e) => {
        e.stopPropagation();
        jumpToTimelineStep(parseInt(btn.dataset.step, 10));
      };
    });
    return;
  }

  // LENS 1 & 2: Active Runtime Matrix or Full Context Cascade
  let items = [];

  // 1. SKILLS
  if (type === 'ALL' || type === 'SKILLS') {
    activeSkills.forEach(s => items.push({ kind: 'skill_active', data: s }));
    if (lens === 'cascade') {
      availableSkills.forEach(s => items.push({ kind: 'skill_dormant', data: s }));
    }
  }

  // 2. MCP SERVERS & TOOLS
  if (type === 'ALL' || type === 'MCP') {
    mcpServers.forEach(srv => {
      // In active lens, only show servers that had tool calls
      if (lens === 'active') {
        if (srv.invoked_count > 0 || (srv.active_tools && srv.active_tools.length > 0)) {
          items.push({ kind: 'mcp_server', data: srv });
        }
      } else {
        // In cascade lens, show all configured servers
        items.push({ kind: 'mcp_server', data: srv });
      }
    });
  }

  // 3. NATIVE TOOLS
  if (type === 'ALL' || type === 'NATIVE') {
    nativeTools.forEach(nt => items.push({ kind: 'native_tool', data: nt }));
  }

  // 4. RULES
  if (type === 'ALL' || type === 'RULES') {
    activeRules.forEach(r => items.push({ kind: 'rule', data: r }));
  }

  // 5. ORCHESTRATION (Subagents & Slash Commands)
  if (type === 'ALL' || type === 'ORCHESTRATION') {
    subagents.forEach(sa => items.push({ kind: 'subagent', data: sa }));
    slashCommands.forEach(cmd => items.push({ kind: 'command', data: cmd }));
  }

  // Apply search query filter
  if (search) {
    items = items.filter(it => {
      const d = it.data;
      let text = '';
      if (typeof d === 'string') {
        text = d;
      } else if (it.kind === 'mcp_server') {
        const activeNames = (d.active_tools || []).map(t => `${t.name} ${t.recent_args_summary || ''}`).join(' ');
        const dormantNames = (d.dormant_tools || []).join(' ');
        text = `${d.name} ${d.command || ''} ${(d.args || []).join(' ')} ${activeNames} ${dormantNames}`;
      } else {
        text = [
          d.name,
          d.title,
          d.description,
          d.file_path || d.path || d.command,
          d.category,
          d.recent_args_summary
        ].filter(Boolean).join(' ');
      }
      return text.toLowerCase().includes(search);
    });
  }

  if (el.sessionMatrixTotalCountBadge) {
    el.sessionMatrixTotalCountBadge.textContent = `${items.length} item${items.length === 1 ? '' : 's'}`;
  }

  if (items.length === 0) {
    el.sessionContextContent.innerHTML = `
      <div class="empty-state" style="padding: 3rem 1.5rem; text-align: center;">
        <div style="font-size: 2.2rem; margin-bottom: 0.5rem;">🧠</div>
        <p style="color: var(--text-muted);">No session elements matching "${escapeHtml(search || type)}".</p>
      </div>
    `;
    return;
  }

  const renderedCards = [];

  items.forEach(it => {
    let cardHtml = '';
    let weight = 160;

    // ─── CARD: Skill (Active or Standby / Dormant) (R-042) ───
    if (it.kind === 'skill_active' || it.kind === 'skill_dormant') {
      const s = it.data;
      const isActive = it.kind === 'skill_active';
      const steps = s.step_indices || [];
      const dispSteps = steps.slice(0, 6);
      const hasMoreSteps = steps.length > 6;
      const triggers = s.invocation_triggers || [];
      weight = isActive ? (210 + (steps.length > 0 ? 55 : 0) + (triggers.length > 0 ? 45 : 0)) : 145;

      cardHtml = `
        <div class="agent-card ${isActive ? 'agent-card-highlight' : ''}" style="${isActive ? 'border-color: rgba(16, 185, 129, 0.45); background: rgba(16, 185, 129, 0.04);' : 'opacity: 0.88;'}">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">⚡ ${escapeHtml(s.name)}</span>
              <div style="display: flex; gap: 0.35rem; align-items: center; flex-wrap: wrap;">
                <span class="agent-card-scope-badge ${s.scope}">${s.scope}</span>
                ${s.category ? `<span class="agent-card-priority-pill">${escapeHtml(s.category)}</span>` : ''}
                ${s.plugin_name ? `<span class="agent-card-priority-pill" style="color: #facc15;">🧩 ${escapeHtml(s.plugin_name)}</span>` : ''}
                ${isActive
                  ? `<span class="session-invoked-skill-pill" style="background: rgba(16, 185, 129, 0.15); color: #34d399; border-color: rgba(16, 185, 129, 0.35);">⚡ ${s.invoked_count}x Active</span>`
                  : `<span class="agent-card-priority-pill" style="opacity: 0.65;">💤 Standby</span>`}
              </div>
            </div>
            <span style="font-size: 0.72rem; color: var(--text-dim); text-transform: uppercase;">${escapeHtml(s.provider || 'antigravity')}</span>
          </div>
          <p class="agent-card-desc">${escapeHtml(s.description || 'On-demand procedural runbook for agentic execution.')}</p>
          <div class="agent-card-meta">
            <span class="agent-token-badge" style="color: #34d399; border-color: rgba(16, 185, 129, 0.3); background: rgba(16, 185, 129, 0.08);">⚡ ~${(s.tokens_est || 0).toLocaleString()} tok</span>
            <span class="agent-word-badge">🔤 ${(s.word_count || 0).toLocaleString()} words</span>
            <span>📏 ${s.line_count || 0} lines</span>
            ${s.has_scripts ? '<span>🛠️ Scripts</span>' : ''}
            ${s.has_examples ? '<span>💡 Examples</span>' : ''}
            ${s.has_references ? '<span>📚 References</span>' : ''}
            ${s.workspace ? `<span>📁 ${escapeHtml(s.workspace)}</span>` : ''}
          </div>

          <!-- Active Invocation Steps Timeline (R-042) -->
          ${isActive && steps.length > 0 ? `
            <div style="margin-top: 0.6rem; display: flex; flex-direction: column; gap: 0.35rem;">
              <div style="font-size: 0.72rem; color: #34d399; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; display: flex; justify-content: space-between; align-items: center;">
                <span>⚡ Execution Steps (${steps.length})</span>
                <span style="color: var(--text-dim); font-size: 0.7rem; font-family: var(--font-mono, monospace);">Span: #${s.first_step_idx} ➔ #${s.last_step_idx}</span>
              </div>
              <div class="steps-row">
                <span class="steps-label">Steps:</span>
                ${dispSteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
                ${hasMoreSteps ? `<span class="step-marker-pill more" title="${steps.length} total executions">+${steps.length - 6} more</span>` : ''}
              </div>
            </div>
          ` : ''}

          <!-- Invocation Triggers Breakdown (R-042) -->
          ${isActive && triggers.length > 0 ? `
            <div style="margin-top: 0.5rem; display: flex; flex-direction: column; gap: 0.25rem;">
              <div style="font-size: 0.68rem; color: var(--text-muted); font-weight: 600; text-transform: uppercase;">
                🎯 Invocation Triggers
              </div>
              <div style="display: flex; gap: 0.35rem; flex-wrap: wrap;">
                ${triggers.map(tr => `<span class="sub-command-recent" style="margin: 0; font-size: 0.7rem; padding: 2px 8px;">${escapeHtml(tr)}</span>`).join('')}
              </div>
            </div>
          ` : ''}

          <div class="agent-card-actions" style="margin-top: 0.8rem;">
            <span class="agent-card-path" title="${escapeHtml(s.file_path)}">${escapeHtml(s.file_path)}</span>
            <div style="display: flex; gap: 0.4rem; align-items: center;">
              <button class="btn-ghost-sm inspect-agent-file-btn" data-path="${escapeHtml(s.file_path)}" title="Open skill in embedded File Viewer">📖 Inspect</button>
              <button class="btn-icon copy-agent-path-btn" style="width: 26px; height: 26px; font-size: 0.75rem;" data-path="${escapeHtml(s.file_path)}" title="Copy Path">📋</button>
              ${isActive ? `
                <button class="btn-ghost-sm step-marker-pill" data-step="${s.first_step_idx}" title="Jump to First Execution">⏮ #${s.first_step_idx}</button>
                <button class="btn-ghost-sm step-marker-pill" data-step="${s.last_step_idx}" title="Jump to Last Execution">⏭ #${s.last_step_idx}</button>
              ` : ''}
            </div>
          </div>
        </div>
      `;
    }

    // ─── CARD: MCP Server & Tools ───
    else if (it.kind === 'mcp_server') {
      const srv = it.data;
      const isActive = srv.invoked_count > 0;
      const tools = srv.active_tools || [];
      const dormant = srv.dormant_tools || [];
      weight = 170 + (tools.length * 75) + (lens === 'cascade' && dormant.length > 0 ? 50 : 0);

      cardHtml = `
        <div class="agent-card" style="border-color: ${isActive ? 'rgba(56, 189, 248, 0.45)' : 'rgba(255, 255, 255, 0.08)'}; background: ${isActive ? 'rgba(56, 189, 248, 0.04)' : 'rgba(15, 23, 42, 0.45)'};">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">🔌 ${escapeHtml(srv.name)}</span>
              <div style="display: flex; gap: 0.35rem; align-items: center; flex-wrap: wrap;">
                <span class="agent-card-scope-badge ${srv.scope}">${srv.scope}</span>
                ${isActive
                  ? `<span class="session-invoked-skill-pill" style="background: rgba(56, 189, 248, 0.15); color: #38bdf8; border-color: rgba(56, 189, 248, 0.35);">⚡ ${srv.invoked_count} calls</span>`
                  : `<span class="agent-card-priority-pill" style="opacity: 0.65;">💤 Dormant</span>`}
              </div>
            </div>
            <span style="font-size: 0.72rem; color: #7dd3fc; font-family: var(--font-mono, monospace);">${tools.length} active tool${tools.length === 1 ? '' : 's'}</span>
          </div>
          <p class="agent-card-desc" style="font-family: monospace; font-size: 0.75rem;">
            Command: ${escapeHtml(srv.command || 'runtime dynamic')} ${(srv.args || []).map(a => escapeHtml(a)).join(' ')}
          </p>

          <!-- Active Tools Section -->
          ${tools.length > 0 ? `
            <div style="margin-top: 0.6rem; display: flex; flex-direction: column; gap: 0.4rem;">
              <div style="font-size: 0.72rem; color: var(--text-muted); font-weight: 600; text-transform: uppercase; letter-spacing: 0.05em;">
                ⚡ Active Tools (${tools.length})
              </div>
              ${tools.map(t => {
                const steps = t.step_indices || [];
                const displaySteps = steps.slice(0, 5);
                const hasMore = steps.length > 5;
                return `
                  <div class="mcp-tool-row">
                    <div class="mcp-tool-row-header">
                      <span class="mcp-tool-name">⚡ ${escapeHtml(t.name)}</span>
                      <span class="mcp-tool-count-badge">${t.invoked_count}x</span>
                    </div>
                    <div class="steps-row">
                      <span class="steps-label">Steps:</span>
                      ${displaySteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
                      ${hasMore ? `<span class="step-marker-pill more" title="${steps.length} total calls">+${steps.length - 5}</span>` : ''}
                    </div>
                    ${t.recent_args_summary ? `<div class="mcp-tool-args-preview" title="${escapeHtml(t.recent_args_summary)}">args: ${escapeHtml(t.recent_args_summary)}</div>` : ''}
                  </div>
                `;
              }).join('')}
            </div>
          ` : ''}

          <!-- Dormant Tools Section (Cascade mode) -->
          ${lens === 'cascade' && dormant.length > 0 ? `
            <div style="margin-top: 0.6rem;">
              <div style="font-size: 0.72rem; color: var(--text-muted); margin-bottom: 0.3rem;">💤 Available Uncalled Tools (${dormant.length}):</div>
              <div style="display: flex; flex-wrap: wrap; gap: 0.35rem;">
                ${dormant.map(dName => `<span class="tool-file-pill" style="opacity: 0.6; font-size: 0.72rem;">${escapeHtml(dName)}</span>`).join('')}
              </div>
            </div>
          ` : ''}
        </div>
      `;
    }

    // ─── CARD: Native Agentic Tool (R-040, R-041) ───
    else if (it.kind === 'native_tool') {
      const nt = it.data;
      const steps = nt.step_indices || [];
      const displaySteps = steps.slice(0, 6);
      const hasMore = steps.length > 6;
      const subBreakdown = nt.sub_breakdown || [];
      const isRunCmd = nt.name === 'run_command' || nt.name === 'bash' || nt.name === 'exec' || nt.name === 'terminal';
      const isFileOp = nt.name === 'read_file' || nt.name === 'write_file' || nt.name === 'edit_file' || nt.name === 'view_file';

      if (isRunCmd) {
        weight = 240 + (Math.min(subBreakdown.length, 5) * 75);
      } else if (isFileOp) {
        weight = 180 + (subBreakdown.length * 50);
      } else {
        weight = 160 + (steps.length > 0 ? 40 : 0);
      }

      cardHtml = `
        <div class="agent-card ${isRunCmd ? 'agent-card-highlight' : ''}" style="${isRunCmd ? 'border-color: rgba(6, 182, 212, 0.45); background: rgba(6, 182, 212, 0.03);' : ''}">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">${isRunCmd ? '💻' : (isFileOp ? '📁' : '🛠️')} ${escapeHtml(nt.name)}</span>
              <span class="native-category-badge ${nt.category || 'Agentic'}">${escapeHtml(nt.category || 'Agentic')}</span>
            </div>
            <span class="mcp-tool-count-badge">${nt.invoked_count}x</span>
          </div>
          <p class="agent-card-desc" style="font-size: 0.8rem; color: var(--text-muted);">
            Native agent capability executing across ${steps.length} turns in the session timeline.
          </p>
      `;

      // Year 2126 Holographic Command Spectrum Bar for run_command
      if (isRunCmd && subBreakdown.length > 0) {
        const primaryBreakdown = subBreakdown.slice(0, 5);
        const overflowBreakdown = subBreakdown.slice(5);

        const renderCommandRow = (sb) => {
          const sbSteps = sb.step_indices || [];
          const dispSbSteps = sbSteps.slice(0, 5);
          const sbMore = sbSteps.length > 5;
          const catColor = getCategoryColor(sb.category);
          return `
            <div class="sub-command-row">
              <div class="sub-command-header">
                <div class="sub-command-label-group">
                  <span class="sub-command-icon">${getCategoryIcon(sb.category)}</span>
                  <span class="sub-command-key" title="${escapeHtml(sb.label)}">${escapeHtml(sb.label)}</span>
                  <span class="native-category-badge" style="background: ${catColor}20; color: ${catColor}; border: 1px solid ${catColor}40; font-size: 0.64rem;">${escapeHtml(sb.category || 'System')}</span>
                </div>
                <div class="sub-command-stats">
                  <span class="sub-command-count">${sb.count}x</span>
                  <span class="sub-command-pct">(${sb.percentage}%)</span>
                </div>
              </div>
              ${sb.recent_items && sb.recent_items.length ? `
                <div class="sub-command-recent" title="Recent: ${escapeHtml(sb.recent_items.join(' • '))}">
                  <span class="recent-tag">Recent:</span> ${escapeHtml(sb.recent_items[0])}
                </div>
              ` : ''}
              <div class="sub-command-steps">
                <span class="sub-steps-label">Steps:</span>
                ${dispSbSteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
                ${sbMore ? `<span class="step-marker-pill more" title="${sbSteps.length} total calls">+${sbSteps.length - 5}</span>` : ''}
              </div>
            </div>
          `;
        };

        cardHtml += `
          <div style="margin-top: 0.75rem;">
            <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 0.35rem;">
              <span style="font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: #38bdf8;">
                ⚡ Command Spectrum (${subBreakdown.length} types)
              </span>
              <span style="font-size: 0.7rem; color: var(--text-dim); font-family: var(--font-mono, monospace);">
                ${nt.invoked_count} total executions
              </span>
            </div>
            <div class="command-spectrum-bar" title="Execution distribution across ${subBreakdown.length} command types">
              ${subBreakdown.map(sb => `
                <div class="spectrum-segment" style="width: ${sb.percentage}%; background-color: ${getCategoryColor(sb.category)};" title="${escapeHtml(sb.label)}: ${sb.count}x (${sb.percentage}%)"></div>
              `).join('')}
            </div>
          </div>

          <!-- Dynamic Adaptive Sub-Command Grid -->
          <div class="sub-command-matrix">
            ${primaryBreakdown.map(renderCommandRow).join('')}
            ${overflowBreakdown.length > 0 ? `
              <div class="sub-command-overflow" style="display: none; flex-direction: column; gap: 0.5rem;">
                ${overflowBreakdown.map(renderCommandRow).join('')}
              </div>
              <button class="sub-command-toggle-btn" data-count="${subBreakdown.length}" data-overflow="${overflowBreakdown.length}">
                <span>✨ Show all ${subBreakdown.length} command types (+${overflowBreakdown.length} more)</span>
              </button>
            ` : ''}
          </div>
        `;
      } else if (isFileOp && subBreakdown.length > 0) {
        // Filesystem tools: extension breakdown
        cardHtml += `
          <div style="margin-top: 0.75rem;">
            <div style="font-size: 0.72rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em; color: #34d399; margin-bottom: 0.4rem;">
              📁 File Types Touched (${subBreakdown.length})
            </div>
            <div class="file-ext-matrix">
              ${subBreakdown.map(sb => {
                const sbSteps = sb.step_indices || [];
                const dispSbSteps = sbSteps.slice(0, 3);
                return `
                  <div class="file-ext-chip">
                    <div style="display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; width: 100%;">
                      <span class="ext-name">${escapeHtml(sb.key)}</span>
                      <span class="ext-count">${sb.count}x (${sb.percentage}%)</span>
                    </div>
                    ${sb.recent_items && sb.recent_items.length ? `
                      <div class="ext-recent" title="${escapeHtml(sb.recent_items.join(', '))}">
                        ${escapeHtml(sb.recent_items[0])}
                      </div>
                    ` : ''}
                    <div class="steps-row" style="margin-top: 0.25rem;">
                      <span class="steps-label" style="font-size: 0.65rem;">Steps:</span>
                      ${dispSbSteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
                      ${sbSteps.length > 3 ? `<span class="step-marker-pill more" title="${sbSteps.length} total calls">+${sbSteps.length - 3}</span>` : ''}
                    </div>
                  </div>
                `;
              }).join('')}
            </div>
          </div>
        `;
      } else {
        // Default timeline steps list
        cardHtml += `
          <div class="steps-row" style="margin-top: 0.4rem;">
            <span class="steps-label">Steps:</span>
            ${displaySteps.map(st => `<button class="step-marker-pill" data-step="${st}" title="Jump to Step #${st}">#${st}</button>`).join('')}
            ${hasMore ? `<span class="step-marker-pill more" title="${steps.length} total calls">+${steps.length - 6} more</span>` : ''}
          </div>
        `;
      }

      cardHtml += `
          <div class="agent-card-actions" style="margin-top: 0.8rem;">
            <div style="display: flex; gap: 0.4rem;">
              <button class="btn-ghost-sm step-marker-pill" data-step="${nt.first_step_idx}" title="Jump to First Execution">⏮ #${nt.first_step_idx}</button>
              <button class="btn-ghost-sm step-marker-pill" data-step="${nt.last_step_idx}" title="Jump to Last Execution">⏭ #${nt.last_step_idx}</button>
            </div>
          </div>
        </div>
      `;
    }

    // ─── CARD: Active Rule ───
    else if (it.kind === 'rule') {
      const r = it.data;
      const isConst = (r.priority || '').includes('Constitutional') || r.name.includes('constitution') || r.name.includes('101');
      weight = 165;
      cardHtml = `
        <div class="agent-card">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">📜 ${escapeHtml(r.title || r.name)}</span>
              <div style="display: flex; gap: 0.35rem; align-items: center; flex-wrap: wrap;">
                <span class="agent-card-scope-badge ${r.scope}">${r.scope}</span>
                <span class="agent-card-priority-pill ${isConst ? 'constitutional' : ''}">${escapeHtml(r.priority || 'Standard')}</span>
              </div>
            </div>
          </div>
          <p class="agent-card-desc">${escapeHtml(r.description || 'Governing guideline for agent actions.')}</p>
          <div class="agent-card-meta">
            <span class="agent-token-badge">⚡ ~${(r.tokens_est || 0).toLocaleString()} tok</span>
            <span class="agent-word-badge">🔤 ${(r.word_count || 0).toLocaleString()} words</span>
            <span>📏 ${r.line_count || 0} lines</span>
          </div>
          <div class="agent-card-actions">
            <span class="agent-card-path" title="${escapeHtml(r.file_path)}">${escapeHtml(r.file_path)}</span>
            <div style="display: flex; gap: 0.4rem;">
              <button class="btn-ghost-sm inspect-agent-file-btn" data-path="${escapeHtml(r.file_path)}" title="Open rule in File Viewer">📖 Inspect</button>
              <button class="btn-icon copy-agent-path-btn" style="width: 26px; height: 26px; font-size: 0.75rem;" data-path="${escapeHtml(r.file_path)}" title="Copy Path">📋</button>
            </div>
          </div>
        </div>
      `;
    }

    // ─── CARD: Subagent ───
    else if (it.kind === 'subagent') {
      const sa = it.data;
      weight = 130;
      cardHtml = `
        <div class="agent-card" style="border-color: rgba(99, 102, 241, 0.4);">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">🤖 Subagent Persona</span>
              <span class="agent-card-scope-badge" style="background: rgba(99, 102, 241, 0.2); color: #818cf8;">DELEGATE</span>
            </div>
          </div>
          <p class="agent-card-desc" style="font-size: 0.95rem; font-weight: 600; color: #e0e7ff;">
            ${escapeHtml(sa)}
          </p>
          <div style="font-size: 0.76rem; color: var(--text-muted); margin-top: 0.4rem;">
            Autonomous worker thread spawned with specialized instructions.
          </div>
        </div>
      `;
    }

    // ─── CARD: Slash Command ───
    else if (it.kind === 'command') {
      const cmd = it.data;
      weight = 130;
      cardHtml = `
        <div class="agent-card" style="border-color: rgba(6, 182, 212, 0.4);">
          <div class="agent-card-header">
            <div class="agent-card-title-group">
              <span class="agent-card-title">⚡ Slash Command</span>
              <span class="agent-card-scope-badge" style="background: rgba(6, 182, 212, 0.2); color: #22d3ee;">WORKFLOW</span>
            </div>
          </div>
          <p class="agent-card-desc" style="font-size: 1.05rem; font-family: var(--font-mono, monospace); font-weight: 700; color: #67e8f9;">
            ${escapeHtml(cmd)}
          </p>
          <div style="font-size: 0.76rem; color: var(--text-muted); margin-top: 0.4rem;">
            Interactive slash workflow shortcut executed during session.
          </div>
        </div>
      `;
    }

    if (cardHtml) {
      renderedCards.push({ html: cardHtml, weight });
    }
  });

  const containerW = el.sessionContextContent.clientWidth || window.innerWidth;
  el.sessionContextContent.innerHTML = binPackCards(renderedCards, containerW);

  // Attach Step Jump Navigation to all step marker pills
  el.sessionContextContent.querySelectorAll('.step-marker-pill[data-step]').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      jumpToTimelineStep(parseInt(btn.dataset.step, 10));
    };
  });

  // Attach 1-click Inspect to openFileViewer
  el.sessionContextContent.querySelectorAll('.inspect-agent-file-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      openFileViewer(btn.dataset.path);
    };
  });

  // Attach Copy Path
  el.sessionContextContent.querySelectorAll('.copy-agent-path-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      copyToClipboard(btn.dataset.path, 'Path copied to clipboard!');
    };
  });

  // Attach Sub-Command Overflow Toggle (R-041 Dynamic Density)
  el.sessionContextContent.querySelectorAll('.sub-command-toggle-btn').forEach(btn => {
    btn.onclick = (e) => {
      e.stopPropagation();
      const parent = btn.closest('.sub-command-matrix');
      const overflow = parent?.querySelector('.sub-command-overflow');
      if (overflow) {
        const isHidden = overflow.style.display === 'none';
        overflow.style.display = isHidden ? 'flex' : 'none';
        const total = btn.dataset.count;
        const extra = btn.dataset.overflow;
        btn.innerHTML = isHidden
          ? `<span>🔼 Collapse to top 5 commands</span>`
          : `<span>✨ Show all ${total} command types (+${extra} more)</span>`;
      }
    };
  });
}

// Expose modal functions to window for testing, extensions, and automated verification
window.openDiagramTheater = openDiagramTheater;
window.closeDiagramTheater = closeDiagramTheater;
window.openFileViewer = openFileViewer;
window.closeFileViewer = closeFileViewer;
window.openFileFindWidget = openFileFindWidget;
window.closeFileFindWidget = closeFileFindWidget;
window.executeFileFind = executeFileFind;
window.goToNextFindMatch = goToNextFindMatch;
window.goToPrevFindMatch = goToPrevFindMatch;
window.openSessionFilesModal = openSessionFilesModal;
window.closeSessionFilesModal = closeSessionFilesModal;
window.openAgentBrainModal = openAgentBrainModal;
window.closeAgentMatrixModal = closeAgentMatrixModal;
window.openSessionContextModal = openSessionContextModal;
window.closeSessionContextModal = closeSessionContextModal;
window.renderSessionMatrix = renderSessionMatrix;
window.jumpToTimelineStep = jumpToTimelineStep;
window.loadSessionDetail = loadSessionDetail;


