#!/usr/bin/env bash
# PreToolUse hook — blocks writes to sensitive files on ALL projects.
# Project-level hooks in .claude/settings.json add project-specific patterns on top.
#
# Claude Code calls this before Write and Edit tool uses.
# Exit 2 = block the tool call and show error to user.

INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""')

# Bail early if no file path (non-file tool calls)
if [[ -z "$FILE" ]]; then
  exit 0
fi

PROTECTED=(
  ".env"
  ".env.local"
  ".env.production"
  ".env.development"
  ".env.staging"
  ".env.test"
  ".pem"
  "credentials.json"
  ".credentials.json"
  "id_rsa"
  "id_ed25519"
  "id_ecdsa"
)

for pattern in "${PROTECTED[@]}"; do
  if [[ "$FILE" == *"$pattern"* ]]; then
    echo "BLOCKED: $FILE matches sensitive file pattern '$pattern'." >&2
    echo "Edit this file manually in your terminal." >&2
    exit 2
  fi
done

exit 0
