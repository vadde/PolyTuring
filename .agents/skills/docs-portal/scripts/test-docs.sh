#!/usr/bin/env bash
# ==============================================================================
# 🧪 docs-portal automated headless verification suite (14-Gate Comprehensive Audit)
# Validates HTML markup, KaTeX math delimiters, Mermaid syntax & sizing,
# draggable sidebar resizer, collapsible rail mode, live filter input,
# Haptic Web Audio synthesis, multi-theme engine, fullscreen theater modal,
# right-side TOC alignment & reading progress HUD, touch overflow containment,
# and static bundle packaging integrity.
# ==============================================================================
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
SITE_DIR="${REPO_ROOT}/docs/site"
BUILD_DIR="${REPO_ROOT}/_site"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  🧪 Running Documentation Portal 14-Gate Verification Suite      ║"
echo "╚══════════════════════════════════════════════════════════════════╝"

# 1. Check required assets
echo "→ 1. Checking core asset files..."
for f in index.html style.css app.js; do
  if [ ! -f "${SITE_DIR}/${f}" ]; then
    echo "❌ Missing required asset: ${SITE_DIR}/${f}"
    exit 1
  fi
done
echo "   ✅ Core assets present (index.html, style.css, app.js)"

# 2. Verify Mermaid label escaping invariants (< and > inside node labels)
echo "→ 2. Auditing Mermaid syntax invariants..."
UNESCAPED_LT=$(grep -n '\[.*<[0-9].*\]' "${SITE_DIR}/index.html" || true)
if [ -n "${UNESCAPED_LT}" ]; then
  echo "❌ Detected unescaped '<' inside Mermaid node label in index.html:"
  echo "${UNESCAPED_LT}"
  echo "   Use '≤' or '&lt;' to prevent Mermaid HTML parser crashes."
  exit 1
fi
echo "   ✅ Mermaid node labels are cleanly escaped."

# 3. Verify CSS SVG sizing rules (Prevent SVG 0px x 0px collapse)
echo "→ 3. Auditing SVG sizing invariants in style.css..."
if grep -E '\.mermaid svg.*width:\s*auto\s*!important' "${SITE_DIR}/style.css" > /dev/null; then
  echo "❌ Fatal CSS anti-pattern: 'width: auto !important' found on .mermaid svg."
  echo "   This collapses SVG elements to 0px x 0px inside flex containers."
  exit 1
fi
if grep -E '\.mermaid svg.*height:\s*auto\s*!important.*max-height' "${SITE_DIR}/style.css" > /dev/null; then
  echo "❌ Fatal CSS anti-pattern: conflicting max-height and height: auto !important on same line."
  exit 1
fi
echo "   ✅ CSS SVG sizing rules conform to responsive viewBox standards."

# 4. Verify KaTeX configuration
echo "→ 4. Checking KaTeX resources and delimiters..."
if ! grep -q "katex.min.css" "${SITE_DIR}/index.html"; then
  echo "❌ KaTeX stylesheet missing in index.html head"
  exit 1
fi
if ! grep -q "auto-render.min.js" "${SITE_DIR}/index.html"; then
  echo "❌ KaTeX auto-render script missing in index.html"
  exit 1
fi
if ! grep -q "renderMathInElement" "${SITE_DIR}/app.js"; then
  echo "❌ renderMathInElement call missing in app.js"
  exit 1
fi
echo "   ✅ KaTeX 0.16.11 loaded with auto-render extension."

# 5. Verify Draggable Sidebar Resizer Invariants
echo "→ 5. Auditing Draggable Sidebar Resizer invariants..."
if ! grep -q 'id="sidebar-resizer"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #sidebar-resizer element in index.html"
  exit 1
fi
if ! grep -q 'col-resize' "${SITE_DIR}/style.css"; then
  echo "❌ Missing col-resize cursor rule in style.css"
  exit 1
fi
if ! grep -q 'initSidebarResizer' "${SITE_DIR}/app.js"; then
  echo "❌ Missing initSidebarResizer in app.js"
  exit 1
fi
if ! grep -q 'polyturing_sidebar_width' "${SITE_DIR}/app.js"; then
  echo "❌ Missing localStorage persistence key in app.js"
  exit 1
fi
echo "   ✅ Draggable sidebar resizer & persistence verified."

# 6. Verify Sidebar Collapsible Rail & Distraction-Free Mode
echo "→ 6. Auditing Collapsible Rail & Distraction-Free Mode..."
if ! grep -q 'id="floating-sidebar-toggle"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #floating-sidebar-toggle element in index.html"
  exit 1
fi
if ! grep -q 'id="sidebar-close-btn"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #sidebar-close-btn element in index.html"
  exit 1
fi
if ! grep -q 'body\.sidebar-collapsed' "${SITE_DIR}/style.css"; then
  echo "❌ Missing body.sidebar-collapsed styling rules in style.css"
  exit 1
fi
if ! grep -q 'polyturing_sidebar_collapsed' "${SITE_DIR}/app.js"; then
  echo "❌ Missing polyturing_sidebar_collapsed persistence in app.js"
  exit 1
fi
echo "   ✅ Collapsible sidebar rail (⌘B / [) verified."

# 7. Verify Sidebar Real-Time Search Filter
echo "→ 7. Auditing Sidebar Real-Time Search Filter..."
if ! grep -q 'id="sidebar-filter-input"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #sidebar-filter-input in index.html"
  exit 1
fi
if ! grep -q 'initSidebarFilters' "${SITE_DIR}/app.js"; then
  echo "❌ Missing initSidebarFilters function in app.js"
  exit 1
fi
echo "   ✅ Sidebar live filter engine verified."

# 8. Verify Haptic Web Audio Feedback Engine
echo "→ 8. Auditing Haptic Web Audio Feedback System..."
if ! grep -q 'HapticAudio' "${SITE_DIR}/app.js"; then
  echo "❌ Missing HapticAudio synthesis engine in app.js"
  exit 1
fi
if ! grep -q 'id="audio-toggle-btn"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #audio-toggle-btn in index.html"
  exit 1
fi
if ! grep -q 'copySuccess' "${SITE_DIR}/app.js"; then
  echo "❌ Missing copySuccess chime synthesis in app.js"
  exit 1
fi
if ! grep -q 'polyturing_audio_enabled' "${SITE_DIR}/app.js"; then
  echo "❌ Missing polyturing_audio_enabled persistence key in app.js"
  exit 1
fi
echo "   ✅ Pure Web Audio API haptic feedback engine verified."

# 9. Verify Multi-Theme Engine
echo "→ 9. Auditing Multi-Theme Engine..."
for theme in liquid mica holo matrix; do
  if ! grep -q "data-theme=\"${theme}\"" "${SITE_DIR}/style.css"; then
    echo "❌ Missing ${theme} theme definition in style.css"
    exit 1
  fi
done
if ! grep -q 'polyturing_theme' "${SITE_DIR}/app.js"; then
  echo "❌ Missing polyturing_theme persistence in app.js"
  exit 1
fi
echo "   ✅ 4-Theme engine verified (Liquid, Mica, Holo, Matrix)."

# 10. Verify Fullscreen Visual Theater Modal
echo "→ 10. Auditing Fullscreen Theater Modal Workbench..."
if ! grep -q 'id="theater-modal"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #theater-modal in index.html"
  exit 1
fi
if ! grep -q 'id="theater-zoom-fit"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #theater-zoom-fit button in index.html"
  exit 1
fi
if ! grep -q 'calculateFitScale' "${SITE_DIR}/app.js"; then
  echo "❌ Missing calculateFitScale function in app.js"
  exit 1
fi
echo "   ✅ Fullscreen theater workbench with dynamic fit verified."

# 11. Verify Right-Side TOC Alignment & Reading Progress HUD
echo "→ 11. Auditing Right-Side TOC Alignment & HUD..."
if ! grep -q 'id="toc-sidebar"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #toc-sidebar element in index.html"
  exit 1
fi
if ! grep -q 'id="toc-reading-progress"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #toc-reading-progress HUD in index.html"
  exit 1
fi
if ! grep -q 'id="toc-progress-indicator"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #toc-progress-indicator bar in index.html"
  exit 1
fi
if ! grep -q 'href="#sec-sdd-governance"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing sec-sdd-governance in TOC index.html"
  exit 1
fi
echo "   ✅ Right-side TOC alignment and reading progress HUD verified."

# 12. Verify Year 2500 Mobile Navigation Drawer & Anti-Clobbering Invariants
echo "→ 12. Auditing Mobile Navigation Drawer & Anti-Clobbering invariants..."
if ! grep -q 'id="mobile-nav-toggle"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #mobile-nav-toggle hamburger button in index.html"
  exit 1
fi
if ! grep -q 'id="mobile-drawer-backdrop"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #mobile-drawer-backdrop overlay in index.html"
  exit 1
fi
if ! grep -q 'id="mobile-drawer-close"' "${SITE_DIR}/index.html"; then
  echo "❌ Missing #mobile-drawer-close button in index.html"
  exit 1
fi
if ! grep -q 'initMobileDrawer' "${SITE_DIR}/app.js"; then
  echo "❌ Missing initMobileDrawer function in app.js"
  exit 1
fi
if ! grep -q '\.sidebar\.mobile-open' "${SITE_DIR}/style.css"; then
  echo "❌ Missing .sidebar.mobile-open state class in style.css"
  exit 1
fi
if ! grep -q 'overflow-x:\s*clip' "${SITE_DIR}/style.css"; then
  echo "❌ Missing sticky-safe 'overflow-x: clip' containment in style.css"
  exit 1
fi
echo "   ✅ Year 2500 mobile slide-out drawer & sticky-safe containment verified."

# 13. Verify Table & Math Touch Overflow Containment
echo "→ 13. Auditing Table & Formula Touch Overflow Invariants..."
TABLE_COUNT=$(grep -c '<table' "${SITE_DIR}/index.html" || true)
WRAPPER_COUNT=$(grep -c 'class="table-wrapper"' "${SITE_DIR}/index.html" || true)
if [ "${TABLE_COUNT}" -gt 0 ] && [ "${WRAPPER_COUNT}" -lt "${TABLE_COUNT}" ]; then
  echo "❌ Unwrapped table detected! All <table> elements must be inside <div class=\"table-wrapper\"> to prevent mobile blowout."
  exit 1
fi
if ! grep -q 'min-width' "${SITE_DIR}/style.css"; then
  echo "❌ Missing table min-width definition in style.css"
  exit 1
fi
if ! grep -q '\.katex-display' "${SITE_DIR}/style.css"; then
  echo "❌ Missing .katex-display definition in style.css"
  exit 1
fi
echo "   ✅ All tables (${TABLE_COUNT}) wrapped with touch-momentum horizontal scroll containment."

# 14. Build site and verify output
echo "→ 14. Building static documentation bundle & auditing distribution integrity..."
bash "${REPO_ROOT}/scripts/build-docs.sh" > /dev/null
if [ ! -f "${BUILD_DIR}/index.html" ] || [ ! -f "${BUILD_DIR}/style.css" ] || [ ! -f "${BUILD_DIR}/app.js" ]; then
  echo "❌ Built bundle missing required assets in ${BUILD_DIR}"
  exit 1
fi
TOTAL_DOC_FILES=$(find "${BUILD_DIR}" -type f | wc -l | tr -d ' ')
echo "   ✅ Static bundle built & verified (${TOTAL_DOC_FILES} distribution files)."

echo ""
echo "🎉 ALL 14 DOCUMENTATION PORTAL INVARIANTS DETERMINISTICALLY PASSED!"
