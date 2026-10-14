#!/usr/bin/env bash
# ==============================================================================
# setup-skill.sh — Session WIP Skill & Hook Installer
#
# Deploys the session-wip recovery skill and companion scripts so AI coding
# assistants can autonomously recover session state, detect WIP changes,
# and respect the Multi-Session Disambiguation Invariant.
#
# Supported Targets:
#   1. Current Workspace:  ./scripts/setup-skill.sh [workspace]
#   2. Global System:      ./scripts/setup-skill.sh global
#   3. Target Directory:   ./scripts/setup-skill.sh /path/to/any/repo
# ==============================================================================

set -euo pipefail

TARGET="${1:-workspace}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TOOL_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${TOOL_DIR}/../.." && pwd)"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║  ⚡ Installing Session WIP Recovery Skill & Agent Hooks          ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo "  → Target mode: ${TARGET}"

# Ensure source scripts are executable
chmod +x "${TOOL_DIR}/scripts/wip.sh"

if [[ "${TARGET}" == "global" || "${TARGET}" == "--global" ]]; then
  DEST_DIR="${HOME}/.gemini/config"
  SKILL_DEST="${DEST_DIR}/skills/session-wip"
  echo "  → Target directory: ${DEST_DIR}"

  mkdir -p "${SKILL_DEST}/scripts"

  # Copy skill definition and script
  cp "${REPO_ROOT}/.agents/skills/session-wip/SKILL.md" "${SKILL_DEST}/SKILL.md"
  cp "${TOOL_DIR}/scripts/wip.sh" "${SKILL_DEST}/scripts/wip.sh"
  chmod +x "${SKILL_DEST}/scripts/wip.sh"

  echo "  ✅ Installed global skill: ~/.gemini/config/skills/session-wip/"
  echo "  ✅ Installed companion script: ~/.gemini/config/skills/session-wip/scripts/wip.sh"

elif [[ "${TARGET}" == "workspace" || "${TARGET}" == "." ]]; then
  DEST_DIR="${REPO_ROOT}/.agents"
  SKILL_DEST="${DEST_DIR}/skills/session-wip"
  echo "  → Target directory: ${DEST_DIR}"

  mkdir -p "${SKILL_DEST}/scripts"

  # Copy skill definition and script
  cp "${REPO_ROOT}/.agents/skills/session-wip/SKILL.md" "${SKILL_DEST}/SKILL.md"
  cp "${TOOL_DIR}/scripts/wip.sh" "${SKILL_DEST}/scripts/wip.sh"
  chmod +x "${SKILL_DEST}/scripts/wip.sh"

  echo "  ✅ Installed workspace skill: .agents/skills/session-wip/"
  echo "  ✅ Installed companion script: .agents/skills/session-wip/scripts/wip.sh"

else
  # Custom directory path
  TARGET_PATH="$(cd "${TARGET}" && pwd)"
  SKILL_DEST="${TARGET_PATH}/.agents/skills/session-wip"
  echo "  → Target directory: ${TARGET_PATH}"

  mkdir -p "${SKILL_DEST}/scripts"

  cp "${REPO_ROOT}/.agents/skills/session-wip/SKILL.md" "${SKILL_DEST}/SKILL.md"
  cp "${TOOL_DIR}/scripts/wip.sh" "${SKILL_DEST}/scripts/wip.sh"
  chmod +x "${SKILL_DEST}/scripts/wip.sh"

  echo "  ✅ Installed skill to target: ${SKILL_DEST}"
fi

echo ""
echo "✨ Installation complete! Coding assistants can now run '/most-recent-wip' or execute:"
echo "   ./tools/session-explorer/scripts/wip.sh"
