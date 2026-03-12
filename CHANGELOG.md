# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-03-12

### Added

- chezmoi-managed dotfiles for WSL2 (Ubuntu 24.04), Aurora DX (immutable Fedora),
  and Distrobox containers
- Two-variable environment model: platform (auto-detected) + context (user-selected)
- Distrobox container lifecycle automation (`scripts/distrobox_setup.py`)
- AI sandbox — Podman container for AI agents with tiered credential access (`bin/ai-sandbox`)
- `dotctl` Go CLI — collect dotfiles status, push metrics/logs to OTel Collector,
  query Prometheus + Loki terminal dashboard
- Systemd timer for periodic collection (`dotctl/deploy/`)
- setup-creds — automated credential seeding for Distrobox containers (Claude Code
  plugins, Context7 MCP, Atuin, glab, kubeconfig)
- 84 integration test assertions across 4 container types
