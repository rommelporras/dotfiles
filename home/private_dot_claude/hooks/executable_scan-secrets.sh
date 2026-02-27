#!/usr/bin/env bash
# PreToolUse hook — blocks Write/Edit if content contains known secret patterns.
# Catches secrets hardcoded into non-sensitive-named files (e.g. constants.ts).
#
# Exit 2 = block the tool call and show error to user.

INPUT=$(cat)
FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // ""')

# Extract content: Write uses 'content', Edit uses 'new_string'
CONTENT=$(echo "$INPUT" | jq -r '.tool_input.content // .tool_input.new_string // ""')

if [[ -z "$CONTENT" ]]; then
  exit 0
fi

BLOCKED=0

check() {
  local name="$1"
  local pattern="$2"
  if printf '%s' "$CONTENT" | grep -qP -- "$pattern"; then
    echo "BLOCKED: Detected potential ${name} in content being written to '${FILE:-file}'." >&2
    BLOCKED=1
  fi
}

# PEM private keys (RSA, EC, OpenSSH)
check "private key" '-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----'

# AWS access key IDs
check "AWS access key" 'AKIA[0-9A-Z]{16}'

# GitHub personal access tokens (classic and fine-grained)
check "GitHub token" 'gh[pousr]_[A-Za-z0-9]{36,255}'

# Anthropic API keys
check "Anthropic API key" 'sk-ant-api\d{2}-[A-Za-z0-9\-_]{80,}'

# OpenAI project keys (new format)
check "OpenAI API key" 'sk-proj-[A-Za-z0-9\-_]{40,}'

if [[ $BLOCKED -eq 1 ]]; then
  echo "Use environment variables or a secret manager (e.g. 1Password op:// references) instead." >&2
  exit 2
fi

exit 0
