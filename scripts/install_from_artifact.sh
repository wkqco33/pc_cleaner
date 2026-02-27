#!/usr/bin/env bash
set -euo pipefail

# Install pc_cleaner from the latest successful GitHub Actions artifact.
# Requirements: gh (GitHub CLI) and an authenticated session.

REPO="${REPO:-seomini/pc_cleaner}"
WORKFLOW="${WORKFLOW:-build.yml}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=""
ARCH=""

case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux) OS="linux" ;;
  *)
    echo "Unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
 esac

case "$(uname -m)" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Unsupported arch: $(uname -m)" >&2
    exit 1
    ;;
 esac

ARTIFACT_NAME="pc_cleaner_${OS}_${ARCH}"

if ! command -v gh >/dev/null 2>&1; then
  echo "gh (GitHub CLI) is required. Install: https://cli.github.com/" >&2
  exit 1
fi

RUN_ID="${RUN_ID:-}"
if [[ -z "${RUN_ID}" ]]; then
  RUN_ID="$(gh run list -R "${REPO}" -w "${WORKFLOW}" -s success -L 1 --json databaseId --jq '.[0].databaseId')"
fi

if [[ -z "${RUN_ID}" || "${RUN_ID}" == "null" ]]; then
  echo "No successful workflow run found for ${WORKFLOW}." >&2
  exit 1
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

gh run download "${RUN_ID}" -R "${REPO}" -n "${ARTIFACT_NAME}" -D "${TMP_DIR}"

BIN_PATH="${TMP_DIR}/pc_cleaner_${OS}_${ARCH}"
if [[ ! -f "${BIN_PATH}" ]]; then
  echo "Artifact did not contain expected binary: ${BIN_PATH}" >&2
  exit 1
fi

chmod +x "${BIN_PATH}"

if [[ ! -d "${INSTALL_DIR}" ]]; then
  echo "Install dir not found: ${INSTALL_DIR}" >&2
  exit 1
fi

if [[ -w "${INSTALL_DIR}" ]]; then
  cp "${BIN_PATH}" "${INSTALL_DIR}/pc_cleaner"
else
  sudo cp "${BIN_PATH}" "${INSTALL_DIR}/pc_cleaner"
fi

"${INSTALL_DIR}/pc_cleaner" --version

echo "Installed to ${INSTALL_DIR}/pc_cleaner"
