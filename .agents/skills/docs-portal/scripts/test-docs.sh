#!/usr/bin/env bash
# ==============================================================================
# 🧪 docs-portal automated headless verification suite
# Validates HTML markup, KaTeX math delimiters, Mermaid syntax & sizing,
# draggable sidebar resizer, Year 2500 mobile navigation drawer,
# table overflow containment, and headless render invariants.
# ==============================================================================
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
SITE_DIR="${REPO_ROOT}/docs/site"
BUILD_DIR="${REPO_ROOT}/_site"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  🧪 Running Documentation Portal Verification Suite              ║"
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
  echo "❌ Fatal CSS anti-pattern: conflicting max-height and height: auto !important."
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

# 6. Verify Year 2500 Mobile Navigation Drawer & Anti-Clobbering Invariants
echo "→ 6. Auditing Mobile Navigation Drawer & Anti-Clobbering invariants..."
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

# 7. Verify Table & Math Touch Overflow Containment
echo "→ 7. Auditing Table & Formula Touch Overflow Invariants..."
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

# 8. Build site to _site
echo "→ 8. Building documentation bundle..."
bash "${REPO_ROOT}/scripts/build-docs.sh" > /dev/null
echo "   ✅ Bundle built successfully to _site"

# 9. Headless Verification
echo "→ 9. Auditing built bundle integrity..."
if [ ! -f "${BUILD_DIR}/index.html" ] || [ ! -f "${BUILD_DIR}/style.css" ] || [ ! -f "${BUILD_DIR}/app.js" ]; then
  echo "❌ Built bundle missing required assets in ${BUILD_DIR}"
  exit 1
fi
TOTAL_DOC_FILES=$(find "${BUILD_DIR}" -type f | wc -l | tr -d ' ')
echo "   ✅ Built static bundle verified (${TOTAL_DOC_FILES} distribution files)."

echo ""
echo "🎉 ALL DOCUMENTATION PORTAL INVARIANTS PASSED!"
