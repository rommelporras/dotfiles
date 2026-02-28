# Code Reviewer Agent Memory -- dotfiles repo

## Project Type
chezmoi-managed dotfiles repo. Going public on GitHub.

## Sensitive Files to Watch
- `home/dot_zshrc.tmpl` -- contains environment-specific secrets (IPs, hostnames, employer info)
- `home/run_once_before_bootstrap.sh.tmpl` -- contains self-hosted GitLab hostname
- `home/dot_gitconfig.tmpl` -- contains real name and employer directory refs
- `bin/ai-sandbox` -- references employer directory name
- `CLAUDE.md` -- references private repo paths and internal account names
- `home/.chezmoi.toml.tmpl` -- prompt text reveals account names

## Known Issues (as of 2026-02-28)
- Private IP `10.10.30.22` hardcoded in zshrc (OTel endpoint) -- needs templating
- Self-hosted GitLab hostname `gitlab.k8s.rommelporras.com` -- needs templating
- Employer abbreviation `eam` in aliases, gitconfig, ai-sandbox -- needs templating
- Git history contains all of the above -- needs filter-repo cleanup before public push

## Conventions
- chezmoi templating with Go text/template
- Secrets must go through chezmoi data prompts (stored locally in ~/.config/chezmoi/chezmoi.toml)
- gitleaks pre-commit hook with custom rules in `.gitleaks.toml`
- Conventional commits (feat:, fix:, docs:, refactor:, chore:, infra:)
- No AI attribution in commits or code
