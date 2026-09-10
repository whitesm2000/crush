# Local Crush features (branch `mine`)

This clone tracks upstream charmbracelet/crush. `main` stays pristine
(only `git pull` touches it). Every local change lives on its own
feature branch and is merged into `mine`, which is the branch we build
the daily-driver `crush.exe` from.

## Workflow

- New change? Create a `feat/<name>` branch from `main`, commit there
  (small, clean commits), then run `update-crush.ps1` (it rebuilds
  `mine` from all feature branches). Never commit on `main`, and never
  commit on `mine` directly - it gets rebuilt from scratch on every
  update, which would wipe unmerged work.
- Repo meta files (this file, `update-crush.ps1`) live on the
  `local-meta` branch, which the update script treats like a feature.
- Drop a feature: `git branch -D feat/<name>` and re-run the script.

## Active features

| Branch | Description | Upstream PR |
|---|---|---|
| `feat/agent-local-tweaks`
| `feat/tool-result-adjacency` | Re-emit tool results immediately after their tool calls (fixes permanently-locked sessions on strict providers like Kimi/DeepSeek when messages interleave). Recovered from fork branch `fix/kimi-400` (authored Sep 8) | none; consider upstream PR | | Other-session work: Ctrl+B pushes foreground bash to background job, `llm-max-retries` option, Windows console title + notify handling. Source: `Documents/VS Code/crush-build/crush` (kept in sync via the `crush-build` remote; a ui.go-only WIP remains in that repo's stash) | private |
| `feat/mcp-ref-inlining` | Inline local `$ref`/`$defs` definitions when forwarding MCP tool schemas; fixes Moonshot 400 "not a valid moonshot flavored json schema" on tools like TickTick that use `$ref: #/$defs/...` | [charmbracelet/crush#3782](https://github.com/charmbracelet/crush/pull/3782) |

## Pending WIP (not merged)

- Window-title/busy-marker session title: stashed in the
  `crush-build` clone (stash@{0}) and exported to
  `Documents/VS Code/crush-build/crush/window-title-wip.patch`.
  Untested; confirm with the other session before merging.

## Retired features

(empty - when upstream merges a PR for one of ours, delete the branch
and move its row here with the release that contains it)

## Update ritual (what `update-crush.ps1` does)

1. `git checkout main && git pull` - fetch official updates
2. Show the keep-list: `git log --cherry-pick --right-only main...mine`
   (your commits upstream lacks; anything upstream already merged
   disappears from this list, so you know it is safe to drop)
3. Rebase each `feat/*` + `local-meta` branch onto new `main`
4. Rebuild `mine` from scratch: reset to `main`, merge every feature
5. Build `crush-patched.exe`; optionally deploy over your installed
   `crush.exe` via `-DeployTo`
