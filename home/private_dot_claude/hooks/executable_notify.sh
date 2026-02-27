#!/usr/bin/env bash
# Notification hook — fires when Claude needs user attention.
# Auto-detects: macOS | WSL2 (wsl-notify-send > notify-send > PowerShell) | native Linux

INPUT=$(cat)
TYPE=$(echo "$INPUT" | jq -r '.notification_type // ""' 2>/dev/null)

case "$TYPE" in
  permission_prompt)  MSG="Permission approval needed" ;;
  idle_prompt)        MSG="Waiting for your input" ;;
  elicitation_dialog) MSG="Input required" ;;
  *)                  MSG="Claude needs your attention" ;;
esac

TITLE="Claude Code"

if [[ "$OSTYPE" == "darwin"* ]]; then
  # macOS — built-in, no install required
  osascript -e "display notification \"$MSG\" with title \"$TITLE\"" 2>/dev/null &

elif grep -qi microsoft /proc/version 2>/dev/null; then
  # WSL2 — try best available method in order of preference
  if command -v wsl-notify-send &>/dev/null; then
    # Best: proper Windows toast with "Claude Code" as source
    wsl-notify-send --category "$TITLE" "$MSG" 2>/dev/null &
  elif command -v notify-send &>/dev/null; then
    # WSLg: works if WSLg display is running
    notify-send "$TITLE" "$MSG" 2>/dev/null &
  elif command -v powershell.exe &>/dev/null; then
    # Fallback: balloon notification (shows "Windows PowerShell" as source)
    powershell.exe -NonInteractive -NoProfile -Command "
      Add-Type -AssemblyName System.Windows.Forms
      \$n = New-Object System.Windows.Forms.NotifyIcon
      \$n.Icon = [System.Drawing.SystemIcons]::Information
      \$n.Visible = \$true
      \$n.BalloonTipTitle = '$TITLE'
      \$n.BalloonTipText = '$MSG'
      \$n.ShowBalloonTip(4000)
      Start-Sleep -Milliseconds 4100
      \$n.Dispose()
    " 2>/dev/null &
  fi

else
  # Native Linux — notify-send (libnotify)
  if command -v notify-send &>/dev/null; then
    notify-send "$TITLE" "$MSG" 2>/dev/null &
  fi
fi

# Terminal bell — guaranteed fallback on every platform
printf '\a'

exit 0
