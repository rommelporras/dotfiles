---
name: audit-cleanup-march-2026
description: Sandbox removal audit and cleanup pass -- what was fixed and what remains deferred
type: project
---

Comprehensive audit cleanup completed 2026-03-19. Removed sandbox/personal-fintrack distrobox contexts, has_op_cli variable, and all sandbox template guards. All 47 checked items verified correct.

**Why:** Sandbox isolation moved to ai-sandbox (Podman). Distrobox containers reduced from 4 to 2 (work-eam, personal).

**How to apply:**
- No more `has_op_cli` in any template or Python code
- No more `ne .context "sandbox"` guards anywhere
- Deferred items to watch: unpinned installer URLs in Containerfile (host-spawn uses /latest/), NVM v0.40.1 and Nerd Font v3.3.0 slightly stale, credentials.md still says "non-sandbox"
