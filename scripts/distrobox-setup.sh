#!/bin/bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec uv run --project "$(dirname "$SCRIPT_DIR")" python "$SCRIPT_DIR/distrobox_setup.py" "$@"
