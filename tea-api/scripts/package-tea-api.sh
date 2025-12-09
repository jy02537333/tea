#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/package-tea-api.sh [options]

Build the tea-api server binary and bundle runtime assets into a tarball.

Options:
  --os <value>       Target GOOS (default: current go env GOOS)
  --arch <value>     Target GOARCH (default: current go env GOARCH)
  --version <value>  Version string for the artifact name (default: git describe or timestamp)
  --entry <path>     Main package path for go build (default: ./cmd/main.go)
  --output <dir>     Directory to place build artifacts (default: dist)
  -h, --help         Show this help and exit
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DEFAULT_OS="$(go env GOOS)"
DEFAULT_ARCH="$(go env GOARCH)"
DEFAULT_VERSION="$(git -C "$PROJECT_ROOT" describe --tags --always 2>/dev/null || date +%Y%m%d%H%M%S)"
DEFAULT_ENTRY="./cmd/main.go"
DEFAULT_OUTPUT_DIR="$PROJECT_ROOT/dist"

TARGET_OS="$DEFAULT_OS"
TARGET_ARCH="$DEFAULT_ARCH"
VERSION="$DEFAULT_VERSION"
ENTRYPOINT="$DEFAULT_ENTRY"
OUTPUT_DIR="$DEFAULT_OUTPUT_DIR"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --os)
      TARGET_OS="$2"
      shift 2
      ;;
    --arch)
      TARGET_ARCH="$2"
      shift 2
      ;;
    --version)
      VERSION="$2"
      shift 2
      ;;
    --entry)
      ENTRYPOINT="$2"
      shift 2
      ;;
    --output)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 1
      ;;
  esac
done

mkdir -p "$OUTPUT_DIR"
STAGING_NAME="tea-api-${VERSION}-${TARGET_OS}-${TARGET_ARCH}"
STAGING_DIR="$OUTPUT_DIR/$STAGING_NAME"
rm -rf "$STAGING_DIR"
mkdir -p "$STAGING_DIR"

BIN_NAME="tea-api"
if [[ "$TARGET_OS" == "windows" ]]; then
  BIN_NAME+=".exe"
fi

echo "[package] building $BIN_NAME for ${TARGET_OS}/${TARGET_ARCH} (entry: $ENTRYPOINT)"
GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" go build -o "$STAGING_DIR/$BIN_NAME" "$PROJECT_ROOT/$ENTRYPOINT"

declare -a COPY_ITEMS=(
  "configs"
  "migrations"
  "doc"
  "docs"
)

for item in "${COPY_ITEMS[@]}"; do
  SRC_PATH="$PROJECT_ROOT/$item"
  if [[ -e "$SRC_PATH" ]]; then
    echo "[package] copying $item"
    cp -R "$SRC_PATH" "$STAGING_DIR/"
  fi
done

if [[ -f "$PROJECT_ROOT/scripts/init.sql" ]]; then
  mkdir -p "$STAGING_DIR/scripts"
  cp "$PROJECT_ROOT/scripts/init.sql" "$STAGING_DIR/scripts/"
fi

cp "$PROJECT_ROOT/README.md" "$STAGING_DIR/README.md"

ARCHIVE_PATH="$OUTPUT_DIR/${STAGING_NAME}.tar.gz"
(
  cd "$OUTPUT_DIR"
  tar -czf "${STAGING_NAME}.tar.gz" "$STAGING_NAME"
)

echo "[package] artifact created: $ARCHIVE_PATH"
