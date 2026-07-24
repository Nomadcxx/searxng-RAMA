# SearXNG RAMA — Plan of Record

**Date:** 2026-07-23
**Status:** Authoritative. Supersedes `docs/audit-2026-07-23.md` (opencode) and the Hallmark UI audit.
**Author:** reconciliation of two parallel audits + architecture decisions with the maintainer.

---

## 0. Implementation status (2026-07-23) — on branch `dev`, E2E-tested, not yet pushed

**Done + verified end-to-end on this Arch host:**
- **Theme redesign** (RAMA + Google light/dark) — unified layout, per-theme palettes; 39-check WCAG gate passes.
- **Theme-switch (A1)** — pre-built CSS variant per theme (`gen-variant.py` + `scripts/build-themes.sh`), swapped by the installer. **Works live** (verified: switching flips the served bundle).
- **Cross-distro installer (A2/A3)** — `install.sh` for Debian/Ubuntu + Fedora (distro deps, Go floor, NodeSource Node 20). Theme build verified on Debian 12, Ubuntu 24.04, Fedora 40.
- **Code nits (D1-D4)** — done in `main.go`.
- **Packaging hygiene** — secret key moved to `post_install` (per-machine, verified unique in a real install); `post_upgrade` venv-rebuild guard.

**E2E results:** Layer 1 AUR `makepkg` ✅ (contents verified) · Layer 2 render + live switch + `post_install` key ✅ · Layer 3 cross-distro build ✅ (Debian/Ubuntu/Fedora) · Layer 4 systemd service start — pending (nspawn or maintainer local).

**Bugs E2E caught & fixed:** `gen-variant.py` not copied into `$srcdir` (theme loop copies dirs only); Node floor (Debian ships 18); nodejs.org tarball npm skips rolldown's native binding → use NodeSource.

**Theme distribution per channel (decided 2026-07-24 — supersedes the earlier
"ship `rama-installer` in the AUR package" idea):**
- **AUR** — no TUI in the package. Ships the RAMA variant active + all pre-built
  bundles, plus `searxng-rama-theme` (shell) to swap them. Uninstall = pacman.
- **Docker** — declarative via `RAMA_THEME` env (entrypoint publishes the chosen
  bundle at container start); nothing to run inside the container.
- **Bare metal (Debian/Ubuntu/Fedora)** — `install.sh` + the Go TUI keeps
  Switch-Theme/Uninstall modes.

**Open:** push `dev`, regen `.SRCINFO`, merge to `main`, AUR push.

---

## 1. Locked decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | **Scope = full rethink** | Fix correctness *and* revisit theming architecture + install model, not just patch bugs. |
| D2 | **Three install channels, each with a distinct job** | See §2. Dual bare-metal + container + native, no overlap. |
| D3 | **Theme selection must actually change the served CSS** | Current mechanism cannot work (§3). This is the central engineering problem. |

### The three channels (D2)

| Channel | Audience | Mechanism | Owns |
|---------|----------|-----------|------|
| **AUR / PKGBUILD** | Arch Linux | `yay -S searxng-rama` / `makepkg -si` | clone SearXNG, `npx vite build`, venv, systemd, package to `/opt/searxng-rama` |
| **Docker** | Anyone wanting containers | `docker compose up` | portable, self-contained |
| **`install.sh` + Go installer** | **Non-Arch, non-Docker, bare-metal** — scope limited to **Debian/Ubuntu + Fedora** for now | `curl \| sudo bash` | clone SearXNG, **build CSS**, venv, systemd, install to `/opt/searxng-rama` |

> **Cross-distro dependency handling** reuses the distro-detection pattern already proven in the **sysc-greet installer** (apt/dnf branching for python, node/npm, systemd). Native `.deb`/`.rpm` packaging via **gopacker** is a *future* option — more complex than in-installer distro branching, so **deferred**; the installer covers Debian/Ubuntu/Fedora bare-metal in this cycle.

The Go installer's **Install mode** is the cross-distro installer; its **Switch-Theme / Uninstall modes** are post-install management. Both AUR and the cross-distro installer must produce identical installed layouts so Switch-Theme works regardless of how the system was installed.

---

## 2. How SearXNG theming actually works (ground truth)

Correcting the misconception in the opencode audit:

1. SearXNG's `simple` theme is authored in **LESS** under `client/simple/src/less/` (`definitions.less` is the token source).
2. CSS is compiled **once, at build time**, by **vite** (`npx vite build` — *not* grunt). The served bundles are **`sxng-ltr.min.css` / `sxng-rtl.min.css`** (referenced in `base.html`), plus `sxng-core.min.js`. *(The A1 swap targets `sxng-ltr.min.css`.)*
3. At **runtime** the Python app serves the precompiled `.min.css`. **There is no runtime LESS→CSS compilation.** The venv contains only Python deps.
4. `.min.css` is a **build artifact** — upstream does **not** commit it.

**Consequences:**
- Copying `definitions.less` anywhere at runtime (to `css/` *or* `src/less/`) changes nothing a user sees. The served `.min.css` is never regenerated.
- Any install path that copies SearXNG source **without running the vite build** yields an **unstyled** site.

---

## 3. The central problem: theme switching is non-functional as built

The recent 8-commit feature ("switch theme without reinstall") copies a `.less` source file into the installed tree. Per §2 that has **zero effect** on the served CSS. `strings.Replace(...,1)` (BUG-6) mangles that file, but the file was never going to be served anyway.

### Chosen solution: **pre-build all theme CSS variants; swap the active file**

At build time (both PKGBUILD and the cross-distro installer), for **each** theme variant (`rama`, `google-light`, `google-dark`):
1. Copy that variant's `definitions.less` into `client/simple/src/less/definitions.less`.
2. Run `npx vite build`.
3. Save the compiled output as `css/searxng.<variant>.min.css` (keep all variants).

**Switch-Theme** then becomes: swap which precompiled file the templates load — either copy `searxng.<variant>.min.css` → the served filename, or repoint a symlink, then restart the service. No compiler needed at switch time. `sourcePath == installPath == /opt/searxng-rama` is **correct** for this design.

*(Rejected alternatives: (b) ship node + recompile on switch — heavy, needs a toolchain in prod; (c) build-time-only theme — loses the "switch without reinstall" feature the maintainer wants.)*

---

## 4. Reconciled authoritative issue list

### A. Architecture (the real work)
- **A1** — Theme-switch mechanism is non-functional (§3). Implement pre-built-variant swap.
- **A2** — Go installer **Install mode has no CSS build step**; upstream ships no compiled CSS → cross-distro install renders unstyled. Add node/npm + `vite build` (mirror PKGBUILD). *Neither prior audit caught this.*
- **A3** — Cross-distro installer needs **distro-aware dependency handling** (python, node/npm, systemd) since it now targets non-Arch. Scope: **Debian/Ubuntu + Fedora**. Reuse the sysc-greet apt/dnf branching pattern. gopacker native packages deferred.

### B. Release blockers
- **B1** — `theme/google/` is **untracked** → missing from clones *and* AUR builds. Commit it. *(opencode BUG-3 — valid; matches Hallmark audit.)*
- **B2** — Google theme **cannot compile**: imports non-existent `./components.less` (`theme/google/definitions.less:239`). *(Hallmark audit.)*
- **B3** — Google theme defines **46 of 105 color tokens** → 56% of the UI unstyled when it replaces upstream `definitions.less`. Must reach full token coverage, light + dark. *(Hallmark audit.)*
- **B4** — `.SRCINFO` ↔ PKGBUILD **version drift** (`.SRCINFO` r9135/rel3 vs PKGBUILD r9176/rel1). Regenerate with `makepkg --printsrcinfo` before AUR push.

### C. UI correctness / accessibility
- **C1** — **Result-title links fail WCAG**: red `#ef233c` on result bg `#323447` = **2.90:1** (below 3.0 large-text floor; titles are normal text). The single most-used element. Lighten accent for surfaces or use bright-neutral titles + accent on interaction. *(Hallmark audit.)*
- **C2** — **Google Fonts CDN import** (`theme/rama/definitions.less:17`) on a privacy engine, directly contradicting the file's own "no external CDN" comment. Self-host Inter + JetBrains Mono, or drop to system stack. *(Hallmark audit.)*
- **C3** — Body-text links `#ef233c`/`#2b2d42` = 3.20:1 and button labels `#edf2f4`/`#ef233c` = 3.74:1 both fail AA-normal (4.5). *(Hallmark audit.)*
- **C4** — No real type pairing (`--font-display` == `--font-body` == Inter); neon success/warning clash with the muted palette; 100+ lines of verbatim `.dark-themes()` duplication. *(Hallmark audit — folds into the redesign, §7.)*
- **C5** — Google theme `theme-auto` `@media` blocks are stubs (`/* Copy all … here */`) → half-applied auto scheme. Mono stack leads with macOS/Windows faces for a Linux audience. *(Hallmark audit.)*

> **Note:** C1-C5 are *correctness* fixes (make the existing re-skin accessible + honest). They are **table stakes**, not the goal. The goal is §7 — the Hallmark redesign that modernizes the UI/UX beyond the re-skin.

### D. Code correctness (valid opencode findings — keep)
- **D1** — `strings.Replace(…,1)` misses repeats in google transform (`main.go:437-445`). Use `ReplaceAll`. *(BUG-6 — real, but downstream of A1.)*
- **D2** — `.git` copied into `/opt/searxng-rama` (installer + PKGBUILD). Drop it. *(SEC-1.)*
- **D3** — `exec rm -rf` → `os.RemoveAll` (`main.go:535`). *(SEC-2.)*
- **D4** — pip has no timeout; timestamped `.bak` files accumulate; swallowed `id -u` error; `.git` mis-categorized as file. *(MIN-1..4, MOD-2 — small.)*

### E. What the opencode audit got wrong (do not action)
- Its three headline **BLOCKERs (BUG-1/2/4)** rest on the assumption that `install.sh` is the primary path and the installer must self-source themes. Corrected by D2: BUG-1/2 are **real but reframed** as "the cross-distro installer needs source **and a CSS build**" (see A2); **BUG-4 is wrong** — PKGBUILD *does* populate `searx/static/themes/simple/themes/` (lines 112-115).
- **BUG-5's fix is wrong** — writing to `src/less/` doesn't work either (§2); the answer is pre-built variants (A1).
- **"PKGBUILD builds CSS via grunt"** is factually wrong — it's vite.

---

## 5. Execution sequence (proposed)

1. **Commit hygiene** — commit the 8-commit feature branch state + the untracked work (`theme/google/`, `AGENTS.md`, `docs/`, go.mod migration), reconcile with origin's 1-commit README lead, so we start clean. *(B1 rides here.)* Also gitignore the stray `searxng/` + `searxng-RAMA/` bare-clone build artifacts in the repo root.
2. **Theme correctness (table stakes)** — B2, B3, C1, C2, C3, C5: make **both** themes complete, compilable, accessible, CDN-free. Prerequisite to A1 (can't pre-build variants that don't compile).
3. **Hallmark redesign (primary deliverable)** — §7. The modern UI/UX pass, extending the override boundary beyond `definitions.less`.
4. **CSS-variant build** — A1 + A2: teach both the PKGBUILD and the Go installer to build every theme variant's CSS; implement swap-based Switch-Theme.
5. **Cross-distro installer** — A2/A3: source clone + vite build + Debian/Ubuntu/Fedora dep branching (sysc-greet pattern) in `install.sh` / Go installer Install mode. (BUG-1/2 resolved here.)
6. **Code nits** — D1-D4.
7. **Release** — B4: regenerate `.SRCINFO`, bump `pkgrel`, push repo + AUR.

Every step verified (build + render check) before release. No AUR push until both themes compile + render and RAMA passes contrast.

---

## 6. Resolved decisions (was: open items)
- **Cross-distro scope:** Debian/Ubuntu + Fedora only, via in-installer dep branching (sysc-greet pattern). gopacker native packages deferred. ✓
- **RAMA light/dark:** **dark-only** confirmed. Switch-Theme offers RAMA (dark) + Google (light/dark). ✓

---

## 7. Hallmark UI/UX redesign — the primary deliverable

The point of this cycle is **not** to ship a corrected re-skin — it's to make SearXNG *look and feel modern*, well beyond the token swap searxng-RAMA does today. Genre: **atmospheric / modern-minimal** dark tool. Boundary and scope to be confirmed with the maintainer (see the redesign brief), but the working plan:

### Override boundary (the key architecture choice)
- **Today:** override `definitions.less` only (tokens) → pure re-skin.
- **Redesign (recommended): LESS-deep, no template forks.** Add RAMA override LESS (imported *after* the base) that restyles layout + components — spacing rhythm, a real type scale, result-card treatment, sticky/condensing search, result density, focus-visible rings, hover/transition microinteractions, empty/loading states. Achieves major modernization in **CSS/LESS alone** — no coupling to upstream template DOM, so upstream drift stays cheap.
- **Reserved:** fork specific `templates/simple/*.html` only where a genuine structural win *requires* DOM changes (e.g. homepage hero, results header). Case-by-case, minimized.

### Redesign surface (from the upstream `simple` theme LESS)
`style.less` (1180 — results/header/footer/general) · `search.less` (382 — search box) · `index.less` (homepage) · `toolkit.less` (640 — form controls) · `detail.less` (258 — image/detail modal) · `autocomplete.less` · `result_templates.less` · `preferences.less`.

### Redesign targets (draft — confirm in brief)
1. **Homepage** — a considered search-first hero (not the stock centered box): brand, generous type, refined input with proper focus state.
2. **Results page** — modern density + rhythm: card/row treatment, clear title/URL/snippet hierarchy, bright-neutral titles with accent on interaction (fixes C1), calmer engine/meta chips, better sidebar.
3. **Type system** — real display/body scale (fixes C4), self-hosted fonts (fixes C2).
4. **Interaction** — focus-visible rings, hover transitions, loading/empty states — within Hallmark's motion discipline (≤3 primitives, transform/opacity only, reduced-motion).
5. **Chrome** — header/footer/preferences brought up to the same standard.

Runs through the Hallmark `redesign` flow (design-context gate → macrostructure/component picks scoped to what a search UI allows → preview → build → slop test). Dark-only.

### Locked design tokens (approved 2026-07-23 via prototype)
Visual direction approved by the maintainer from the interactive prototype (home + results, real DOM classes). Ported into the RAMA LESS layer as the token source.

| Token | Value | Role |
|-------|-------|------|
| `--ground` | `#1e2030` | base (space-cadet, deepened — chosen blue-purple neutral) |
| `--surface` / `--surface-2` / `--surface-3` | `#282a3b` / `#313349` / `#3a3c54` | layered raised surfaces (cards, hover) |
| `--ink` / `--muted` / `--faint` | `#eef2f6` / `#aab4c5` / `#8d99ae` | text ramp |
| `--accent` | `#ff7282` | **interactive text / title-hover / underline** (5.14:1 on surface — fixes C1) |
| `--accent-fill` / `--accent-fill-2` | `#e11235` / `#c30a29` | brand-red fills (buttons, current page, logomark gradient). *Deepened from #ef233c/#d90429 so white labels clear WCAG AA (4.84–6.18:1); maintainer to confirm in audit.* |
| `--ok` / `--warn` | `#5fd08a` / `#e6c15a` | semantic, kept separate from accent (fixes neon C4) |

Contrast is gated by a runnable script — `docs/redesign/check-contrast.py` (exit 0 = all pass). Every new coloured text/surface pair must be added to it and pass before merge.

**Type roles:** **monospace** (JetBrains Mono, self-hosted) carries the search vernacular — wordmark, URLs, engine badges, result counts, kbd; **humanist sans** (Inter, self-hosted) carries titles + snippets + body. This pairing *is* the identity.

**Signature moves:** split accent (bright text-red vs. brand fill-red) · airy hover-lift result rows (no flat bordered boxes) · bright-neutral titles → accent on hover · atmospheric hero (accent radial glow + faint index-grain) · sticky blurred results header · numbered pagination pills · one restrained load reveal. Motion ≤3 primitives, transform/opacity only, `prefers-reduced-motion` honored.

**Approved prototype:** `https://claude.ai/code/artifact/d16bcd1e-a0ba-4edd-8970-530be6507eb4` (static mock — the visual contract to port).

---

## 8. Remaining open item for the maintainer
- **Confirm the override boundary** (LESS-deep vs. LESS-deep + selective template forks) and react to the redesign brief before build. This is the one decision gating §7.
