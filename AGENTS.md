# Agent Development Guide - SearXNG RAMA Edition

## Repository Overview

Custom theme + packaging/distribution for SearXNG. **This repo contains no
SearXNG Python source** — it themes and packages upstream
(github.com/searxng/searxng), cloned at build time.

- **Design system**: `design.md` at the repo root is **locked and
  authoritative**. Read it before any UI work; extend it rather than
  regenerating it.
- **Theme model**: LESS is compiled **once at build time by vite** — there is
  no runtime LESS compilation. Every switchable variant (rama, google-light,
  google-dark) is pre-built into a self-contained bundle
  (`sxng-{ltr,rtl}.<variant>.min.css`); "switching theme" always means
  publishing one of those bundles as the served `sxng-{ltr,rtl}.min.css` and
  restarting the service.
- **Canonical theme build**: `scripts/build-themes.sh <searxng-checkout>
  <rama-repo>`. The PKGBUILD inlines the same steps — keep them in sync.

## Distribution channels (decided; do not re-litigate silently)

| Channel | Install | Theme switching |
|---|---|---|
| AUR (`searxng-rama`) | PKGBUILD, pre-built bundles in the package | `searxng-rama-theme` shell script (no TUI is shipped) |
| Docker | multi-stage Dockerfile (classic-builder compatible — **no BuildKit-only syntax** like `COPY --chmod`) | `RAMA_THEME` env, applied by `scripts/docker-entrypoint.sh` |
| Debian/Ubuntu/Fedora | `install.sh` → Go TUI (`cmd/rama-installer`) | TUI Switch-Theme mode |

## Hard rules (enforced or learned the hard way)

1. **Commit messages must not contain AI/agent attribution** — a repo hook
   rejects them (no `Co-Authored-By: Claude`, etc.).
2. **Every new coloured text/surface pair** goes into
   `docs/redesign/check-contrast.py` and must pass (WCAG AA) before commit.
3. **Style the real DOM** — upstream ships ion-icons, checkbox-categories,
   `hide_if_nojs`/`show_if_nojs` spans. Never style idealised markup.
4. **Node must come from a distro/NodeSource package.** The nodejs.org
   tarball's npm silently skips rolldown's native binding and the web build
   fails at runtime. Node ≥ 20 required.
5. **Never generate the secret key at package/image build time** — per-machine
   in `post_install` (AUR), per-container in the entrypoint (Docker).
6. The systemd unit needs `PrivateTmp=true` — SearXNG caches SQLite DBs in
   /tmp and collides with other users' files without it.
7. Changes to the shared theme layer (`theme/rama/rama.less`) land on **all
   three variants** — verify each.

## Build & test commands

```bash
# Theme (all variants) into a searxng checkout
bash scripts/build-themes.sh /path/to/searxng /path/to/searxng-RAMA

# Contrast gate (must pass before committing colour changes)
python3 docs/redesign/check-contrast.py

# Go TUI installer
go build -o rama-installer ./cmd/rama-installer/

# AUR package gate — build from a clean dir before any AUR push
mkdir /tmp/pkgtest && cp PKGBUILD searxng-rama.install /tmp/pkgtest && \
  cd /tmp/pkgtest && makepkg -f && makepkg --printsrcinfo > .SRCINFO

# Docker (verify BOTH builders — users run the classic one too)
docker build -t searxng-rama:test .
DOCKER_BUILDKIT=0 docker build -t searxng-rama:classic .

# Quality
pre-commit run --all-files
shellcheck install.sh scripts/*.sh
go vet ./...
```

## File organization

```
/
├── design.md               # LOCKED design system — read before UI work
├── theme/rama/             # palette (definitions.less) + shared layout layer (rama.less)
│   └── templates/          # forked simple-theme templates (index, search, results, page_with_header)
├── theme/google/           # google-light/dark palette variants
├── theme/gen-variant.py    # flattens a variant palette for pre-building
├── scripts/build-themes.sh # canonical theme build (all variants)
├── scripts/searxng-rama-theme    # AUR/bare-metal theme switcher (ships in package)
├── scripts/docker-entrypoint.sh  # container start: RAMA_THEME + secret key
├── cmd/rama-installer/     # Go TUI (bare-metal channel only)
├── PKGBUILD + searxng-rama-*.install  # AUR packaging
├── Dockerfile + docker-compose.yaml   # container channel
├── install.sh              # cross-distro bootstrap (Debian/Ubuntu/Fedora)
└── docs/                   # plan-of-record, redesign spec, contrast gate
```

## Go code style (`cmd/rama-installer`)

- Standard library imports first, then external (bubbletea/lipgloss).
- Exported PascalCase, private camelCase; wrap errors with context
  (`fmt.Errorf("create install directory: %w", err)`).
- Each `installTask` is atomic and idempotent; keep the TUI state transitions
  (welcome → installing → complete) intact.

## Shell style (`install.sh`, `scripts/*`)

`set -euo pipefail`, snake_case functions, double-quoted expansions. Scripts
must not assume this maintainer's paths — use env overrides
(`RAMA_INSTALL_PATH`, `RAMA_SOURCE_PATH`) with sane defaults.

## Release checklist

1. Work lands on `dev`; E2E-verify locally (theme build, live render,
   contrast gate).
2. Merge `dev` → `main`, push both.
3. **Gate**: clean-dir `makepkg` from the pushed `main` must succeed; inspect
   package contents (variant bundles, served CSS == rama bundle, placeholder
   secret intact, `/usr/bin/searxng-rama-theme` present).
4. Regenerate `.SRCINFO` from the gate build; sync `pkgver` back to the repo.
5. Tag (`vX.Y.Z`), push tag. Bump the frozen `X.Y.Z-RAMA` version strings in
   PKGBUILD + Dockerfile with it.
6. Copy `PKGBUILD`, `.SRCINFO`, `searxng-rama.install` into the AUR clone
   (`~/aur-searxng-rama`), commit, push.

## Security

- Placeholder secrets only in the repo (`ultrasecretkey`).
- `scripts/check-secrets.sh` and `scripts/check-ai-content.sh` run in
  pre-commit.
- Service runs as the `searxng` user; container runs non-root.
