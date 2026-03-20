# Code Reviewer Agent Memory -- dotfiles repo

## Project Type
chezmoi-managed dotfiles repo. Going public on GitHub.

## Sensitive Files to Watch
- `home/dot_zshrc.tmpl` -- contains environment-specific values (IPs, hostnames)
- `home/run_once_before_bootstrap.sh.tmpl` -- contains self-hosted service hostnames
- `home/dot_gitconfig.tmpl` -- contains real name and directory refs
- `CLAUDE.md` -- references private repo paths and account names
- `home/.chezmoi.toml.tmpl` -- prompt text reveals account names

## Known Issues
- Private IP hardcoded in zshrc (OTel endpoint) -- needs templating
- Self-hosted service hostnames in templates -- acceptable for personal dotfiles
- Git history contains sensitive data -- needs filter-repo cleanup before public push

## Environment Model
- Two-variable model: `.platform` (auto-detected: wsl/aurora/distrobox) + `.context` (user-selected: personal/personal-<project>/work-<name>)
- Sandbox context removed March 2026 -- sandbox isolation now via ai-sandbox (Podman)
- `has_op_cli` variable fully removed -- no longer in templates, Python, or config
- `hasPrefix "prefix" .context` is correct sprig argument order (verified)
- Platform detection order in `.chezmoi.toml.tmpl`: distrobox → aurora → wsl (elif chain)

## Recent Audit
- [Audit cleanup details](project_audit_cleanup.md) -- sandbox removal, has_op_cli removal, deferred items

## Conventions
- chezmoi templating with Go text/template (sprig functions available)
- Secrets must go through chezmoi data prompts (stored locally in ~/.config/chezmoi/chezmoi.toml)
- gitleaks pre-commit hook with custom rules in `.gitleaks.toml`
- Conventional commits (feat:, fix:, docs:, refactor:, chore:, infra:)
- No AI attribution in commits or code
