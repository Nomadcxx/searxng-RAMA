# SearXNG RAMA — Theme Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the approved RAMA UI redesign into SearXNG's `simple` theme — a modern, dark, accessible metasearch UI delivered via a LESS override layer plus two minimal template forks — and make the second ("google") theme complete and compilable.

**Architecture:** SearXNG's `simple` theme is authored in LESS under `client/simple/src/less/` and compiled **once at build time** by vite into `sxng-ltr.min.css`. There is **no runtime LESS compilation.** The redesign is therefore an *appended LESS override layer* (`themes/rama/rama.less`, imported last so it wins the cascade) driven by a new token palette in `themes/rama/definitions.less`, plus self-hosted fonts and two forked Jinja templates (homepage hero + results header). Everything ports from the approved visual contract at [`docs/redesign/prototype.html`](../redesign/prototype.html).

**Tech Stack:** LESS · vite (build) · Jinja2 (templates) · self-hosted woff2 fonts · Python 3 (contrast gate). No new runtime dependencies.

## Working rules (READ FIRST — non-negotiable)

1. **Do the tasks in order, one at a time.** Do not skip ahead. Finish a task's every step (including its build + commit) before starting the next.
2. **Copy, don't invent.** Where a step says "copy from the prototype" or gives a code block, reproduce it **verbatim**. Do not paraphrase CSS, rename tokens, or "improve" it. The prototype at `docs/redesign/prototype.html` is the exact target.
3. **If a step fails, STOP and report the exact error. Never fabricate success, never guess a workaround, never mark a checkbox you did not verify.** A failed build or a failing gate is a reason to stop and ask — not to continue. It is always correct to stop and say "Task N Step M failed with: <output>".
4. **Every verification command's expected output is written in the step.** Run it, compare, and only proceed if it matches. Paste the real output in your report.
5. **Touch only the files named in the current task.** If you think another file needs changing, stop and report it instead of changing it.
6. **Never edit anything under `.git/`, never run `git push`, never delete files.** Commit locally only.

## Global Constraints

*(Every task implicitly includes these. Values are exact.)*

- **Ground truth:** SearXNG serves precompiled `sxng-ltr.min.css`; copying a `.less` file at runtime changes nothing. All theme changes take effect only through a vite rebuild.
- **Visual contract:** [`docs/redesign/prototype.html`](../redesign/prototype.html) is the approved design. Port from it exactly — same tokens, same component treatment, same SearXNG DOM classes it already uses.
- **RAMA is dark-only.** The "google" theme keeps light + dark.
- **Token completeness:** every theme's `definitions.less` MUST define the **full 105-token colour set** (the canonical list is `theme/rama/definitions.less`). A theme that wholesale-replaces upstream `definitions.less` with a partial set leaves the UI unstyled — this is a shipping bug.
- **No external requests (privacy):** zero CDN/webfont/analytics requests. Fonts are self-hosted woff2. Remove any `@import url('https://fonts.googleapis.com/...')`.
- **Accessibility gate:** `python3 docs/redesign/check-contrast.py` must exit 0. Add every new coloured text/surface pair to it; it must pass before any commit that changes colours.
- **Do NOT touch:** routes, Python, engine logic, search behaviour, preferences *behaviour*, or the privacy/referrer/CSP response headers. Redesign the visual + interaction layer only.
- **Non-destructive:** additive LESS + minimal template forks. Template forks must stay **semantic/class-based** (both themes style the same markup). Never delete upstream template logic — fork = copy + restyle the wrapper, keep every Jinja block/variable.
- **Locked palette (dark):** `--ground #1e2030` · `--ground-2 #191b28` · `--surface #282a3b` · `--surface-2 #313349` · `--surface-3 #3a3c54` · `--ink #eef2f6` · `--muted #aab4c5` · `--faint #8d99ae` · `--accent #ff7282` (interactive text/hover) · `--accent-fill #e11235` / `--accent-fill-2 #c30a29` (CTA/pill fills) · `--ok #5fd08a` · `--warn #e6c15a`. Type: JetBrains Mono (search vernacular) + Inter (titles/body), both self-hosted.
- **Work on a branch**, never `main`. Commit after every task. Conventional-commit messages.

## Dev/build/preview environment

Two locations, distinct roles:

| Location | Role |
|----------|------|
| `/home/nomadx/searxng-custom` | Full SearXNG checkout (has `client/simple` + `node_modules` + vite). **Do all LESS/template dev + build + preview here.** |
| `/home/nomadx/searxng-RAMA` | The packaging repo (this repo). Theme source of truth is `theme/<name>/`. **Final files are mirrored here + PKGBUILD wired (Phase 5).** |

Build + preview commands (run in `searxng-custom`):

```bash
cd /home/nomadx/searxng-custom/client/simple
npm install                 # once, if node_modules is stale
npm run build:vite          # compile LESS -> sxng-ltr.min.css (skips icon build, which needs sharp)
cd /home/nomadx/searxng-custom
make run                     # ./manage webapp.run — dev instance to preview in a browser
```

Expected `npm run build:vite` output: `✓ built in …`, and `searx/static/themes/simple/sxng-ltr.min.css` exists afterward.

---

## File Structure

**In `searxng-custom` (dev) — mirrored to `searxng-RAMA/theme/` in Phase 5:**

- `client/simple/src/less/themes/rama/definitions.less` — MODIFY: new locked dark token palette + type vars; remove the Google-Fonts `@import`.
- `client/simple/src/less/themes/rama/rama.less` — CREATE: layout/component override layer (ported from the prototype). One file, imported last.
- `client/simple/src/less/themes/rama/fonts.less` — CREATE: `@font-face` rules for the self-hosted faces.
- `client/simple/src/less/style.less` — MODIFY (build-appended): add `@import "themes/rama/rama.less";` as the **last** line so overrides win.
- `client/simple/src/less/themes/google/definitions.less` — CREATE/MODIFY: full 105-token light+dark theme (fix the partial/broken current file).
- `searx/static/themes/simple/fonts/` — ADD: `JetBrainsMono-Regular.woff2`, `JetBrainsMono-Medium.woff2`, `Inter-Regular.woff2`, `Inter-Medium.woff2`, `Inter-SemiBold.woff2` (subset, latin).
- `searx/templates/simple/index.html` — FORK: redesigned homepage hero.
- `searx/templates/simple/simple_search.html` — reference only (keep classes); the search box is styled, not restructured.
- `searx/templates/simple/results.html` — FORK (minimal): results header wrapper for the sticky/condensing bar. Keep every existing block/include.

**In `searxng-RAMA` (Phase 5):** mirror the four theme files into `theme/rama/` and `theme/google/`, add `theme/rama/templates/`, and wire the PKGBUILD `build()` to copy them + append the import.

---

## Phase 1 — Foundation: tokens, fonts, and a compiling build

### Task 1: RAMA token palette + remove the CDN font

**Files:**
- Modify: `client/simple/src/less/themes/rama/definitions.less`

**Interfaces:**
- Produces: the full CSS-custom-property token set consumed by `rama.less` (Task 4) and every SearXNG component. Token names are the upstream SearXNG names (`--color-base-background`, `--color-result-link-font`, …) mapped to RAMA values, **plus** the redesign's own semantic tokens listed in Global Constraints.

- [ ] **Step 1: Map the locked palette onto the upstream token names.** Keep all 105 upstream `--color-*` names (canonical list = current `theme/rama/definitions.less`). Apply the locked palette:
  - `--color-base-background: #1e2030;` · `--color-base-font: #eef2f6;`
  - result surface `--color-result-background: #282a3b;` · `--color-result-border: rgba(255,255,255,.08);`
  - **`--color-result-link-font: #eef2f6;`** (bright-neutral titles — the fix for the 2.90:1 failure) with hover handled in `rama.less` via `--accent #ff7282`.
  - `--color-url-font: #ff7282;` · secondary/meta (`--color-result-publishdate-font`, `--color-result-engines-font`): `#aab4c5`.
  - buttons `--color-btn-background: #e11235;` · `--color-btn-font: #ffffff;`
  - toned semantics: `--color-success: #5fd08a;` · `--color-warning: #e6c15a;` · `--color-error: #ff7282;`

  Do **not** add any other new tokens here. (The prototype's own tokens — `--ground`, `--surface`, `--accent`, etc. — go into `rama.less` in Task 4, not here.) `definitions.less` only sets the 105 upstream `--color-*` names so nothing is left unstyled; `rama.less` restyles components on top.
- [ ] **Step 2: DELETE the CDN import.** Find and remove the entire line that starts `@import url('https://fonts.googleapis.com/` (currently line 17). Set the three font vars to the self-hosted stacks (copy verbatim):
  ```less
  --font-display: "Inter", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  --font-body:    "Inter", system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
  --font-mono:    "JetBrains Mono", ui-monospace, "SF Mono", Menlo, monospace;
  ```
- [ ] **Step 3 (leave the mixins alone):** Do **not** refactor the `.dark-themes()` / `.black-themes()` mixins. They work as-is. Just make sure any colour values inside `.dark-themes()` that you changed in `:root` are changed identically there (RAMA is dark-only, so the two blocks hold the same values). If unsure, leave a block untouched rather than risk breaking it.
- [ ] **Step 4: Verify contrast.** Run `python3 /home/nomadx/searxng-RAMA/docs/redesign/check-contrast.py` — expected: last line `All contrast checks pass.` and exit code 0 (`echo $?` → `0`). If any row says `FAIL!`, STOP and report it.
- [ ] **Step 5: Verify it compiles.** In `client/simple/` run `npm run build:vite` — expected: build succeeds, `sxng-ltr.min.css` regenerated, and it contains NO `googleapis.com`:
  ```bash
  grep -c googleapis ../../searx/static/themes/simple/sxng-ltr.min.css   # expected: 0
  ```
- [ ] **Step 6: Commit.**
  ```bash
  git add client/simple/src/less/themes/rama/definitions.less
  git commit -m "feat(rama): new dark token palette, self-hosted fonts, no CDN"
  ```

### Task 2: Self-hosted fonts

**Note:** Fonts are an **enhancement**. The font stacks in Task 1 already fall back to `system-ui` / `ui-monospace`, so the redesign works without these files. If the downloads fail, do Step 4b (empty fonts.less) and keep going — do not get stuck.

**Files:**
- Create: `client/simple/src/less/themes/rama/fonts.less`
- Add (binary): five `.woff2` files under `searx/static/themes/simple/fonts/`

- [ ] **Step 1: Download the five woff2 files** (exact commands — run from the checkout root):
  ```bash
  cd /home/nomadx/searxng-custom/searx/static/themes/simple/fonts   # mkdir -p this dir first if missing
  curl -fLO https://cdn.jsdelivr.net/npm/@fontsource/inter@5/files/inter-latin-400-normal.woff2
  curl -fLO https://cdn.jsdelivr.net/npm/@fontsource/inter@5/files/inter-latin-500-normal.woff2
  curl -fLO https://cdn.jsdelivr.net/npm/@fontsource/inter@5/files/inter-latin-600-normal.woff2
  curl -fLO https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5/files/jetbrains-mono-latin-400-normal.woff2
  curl -fLO https://cdn.jsdelivr.net/npm/@fontsource/jetbrains-mono@5/files/jetbrains-mono-latin-500-normal.woff2
  ls -1   # expected: the 5 .woff2 files listed
  ```
- [ ] **Step 2: Create `fonts.less`** with exactly these five `@font-face` blocks (absolute `/static/…` URLs — do not use relative paths):
  ```less
  /* Self-hosted (OFL) — no external requests */
  @font-face{font-family:"Inter";font-style:normal;font-weight:400;font-display:swap;src:url("/static/themes/simple/fonts/inter-latin-400-normal.woff2") format("woff2");}
  @font-face{font-family:"Inter";font-style:normal;font-weight:500;font-display:swap;src:url("/static/themes/simple/fonts/inter-latin-500-normal.woff2") format("woff2");}
  @font-face{font-family:"Inter";font-style:normal;font-weight:600;font-display:swap;src:url("/static/themes/simple/fonts/inter-latin-600-normal.woff2") format("woff2");}
  @font-face{font-family:"JetBrains Mono";font-style:normal;font-weight:400;font-display:swap;src:url("/static/themes/simple/fonts/jetbrains-mono-latin-400-normal.woff2") format("woff2");}
  @font-face{font-family:"JetBrains Mono";font-style:normal;font-weight:500;font-display:swap;src:url("/static/themes/simple/fonts/jetbrains-mono-latin-500-normal.woff2") format("woff2");}
  ```
- [ ] **Step 3 (only if Step 1 downloads failed):** instead of Step 2, create `fonts.less` containing a single line `/* fonts skipped — falling back to system stack; maintainer to add woff2 */` and note this in your report. Then skip Step 4's font check.
- [ ] **Step 4: Verify (after Task 3 wires fonts.less in and you build).** In a browser via `make run`, open devtools → Network → filter `Font`: expected the `.woff2` load from `/static/themes/simple/fonts/…`, **zero** requests to any external host. (If they 404, the instance's static path differs — report it; do not guess.)
- [ ] **Step 5: Commit.**
  ```bash
  git add client/simple/src/less/themes/rama/fonts.less searx/static/themes/simple/fonts/
  git commit -m "feat(rama): self-host Inter + JetBrains Mono (woff2)"
  ```

---

## Phase 2 — RAMA redesign override layer

### Task 3: Create rama.less with the prototype's tokens + wire it in

The redesign CSS already exists, complete, in [`docs/redesign/prototype.html`](../redesign/prototype.html) inside the `<style>…</style>` block. This task and the next **copy that CSS in, verbatim,** removing only the demo-only pieces. You are not writing new CSS.

**Files:**
- Create: `client/simple/src/less/themes/rama/rama.less`
- Modify: `client/simple/src/less/style.less` (append one line at the very end)

- [ ] **Step 1: Open** `docs/redesign/prototype.html`. Find `<style>` near the top. Everything between `<style>` and `</style>` is the source CSS.
- [ ] **Step 2: Create `rama.less`** starting with these two lines, then paste the prototype's **`:root { … }` block verbatim** as the third piece:
  ```less
  /* RAMA redesign override layer — imported last so it wins the cascade */
  @import "themes/rama/fonts.less";
  ```
  Then paste the whole `:root{ … }` block from the prototype (the one that begins `color-scheme: dark;` and defines `--ground`, `--surface`, `--accent`, `--font-mono`, `--sp-*`, `--r-*`, etc.) **exactly as written.** Do not rename or drop any token. (These are the prototype's own tokens; they coexist with the `--color-*` tokens from Task 1.)
- [ ] **Step 3: Append the import to `style.less`.** Add this as the **last line** of `client/simple/src/less/style.less` (the path is already correct relative to that file):
  ```less
  @import "themes/rama/rama.less";
  ```
- [ ] **Step 4: Build.** In `client/simple/` run `npm run build:vite` — expected: `✓ built in …`, exit 0. If it errors, STOP and report the error.
- [ ] **Step 5: Commit.**
  ```bash
  git add client/simple/src/less/themes/rama/rama.less client/simple/src/less/style.less
  git commit -m "build(rama): add rama.less override layer, import last"
  ```

### Task 4: Copy the component CSS in, verbatim (minus demo-only rules)

**Files:**
- Modify: `client/simple/src/less/themes/rama/rama.less`

Continue pasting the prototype's CSS into `rama.less`, **after** the `:root` block from Task 3.

- [ ] **Step 1: COPY these rule groups verbatim** from the prototype `<style>` into `rama.less` (they target SearXNG's real classes, so they just work):
  - the base rules: `*{box-sizing…}`, `html,body{…}`, `a{…}`, `button{…}`, `:focus-visible{…}`
  - search box: `#search`, `.search_box`, `.search_box:focus-within`, `.search_box .lead`, `#q`, `#q::placeholder`, `#clear_search`(+`:hover`), `#send_search`(+`:hover`,`:active`), `#send_search svg`, `#send_search .label`
  - categories: `.search_categories`, `.category`(+`:hover`,`.selected`), `.category svg`, `.category_name`, `.hint`, `.hint kbd`
  - home hero: `.home`, `.home::before`, `.home-inner`, `.mark`, `.mark .glyph`(+`svg`), `.wordmark`(+`b`,`span`), `.edition`, `.home-settings`(+`:hover`,`svg`)
  - results header: `.rheader` and every `.rheader …` rule
  - results body: `.results-shell`, `.rtabs`(+children), `.rgrid`, `#urls`, `.result`(+`:hover`), `.url_header`, `.favicon`, `.url_wrapper`(+children), `.result h3`(+`a`, hover), `.result .content`, `.result .engines`(+`span`), `.result .cache_link`(+`:hover`)
  - sidebar: `#sidebar`, `.card`, `.card .title`, `.infobox …`, `.suggestions …`
  - pagination: `#pagination`, `#pagination .page_number`(+`:hover`), `#pagination .page_number_current`
  - the responsive block `@media (max-width:820px){ … }` and the `@media (prefers-reduced-motion:reduce){ … }` block
- [ ] **Step 2: DROP these demo-only rules** — do NOT copy them (they belong to the preview page, not the product): `.meta`, `.meta *`, `.switch`, `.switch button`, `.stage`, `.stage.on` (the `display:none/block` toggle).
- [ ] **Step 3: FIX two selectors after pasting** (small, mechanical):
  - The load-reveal rules in the prototype read `.stage.on .result{…}` and `.stage.on #urls .result:nth-child(n){…}`. Delete the `.stage.on ` prefix from each so they read `.result{ animation: rise … }` and `#urls .result:nth-child(n){ … }`.
  - The footer rule reads `footer.site{…}` (and `footer.site a`). Change `footer.site` → `footer` and `footer.site a` → `footer a` (SearXNG's page footer is a plain `<footer>`).
- [ ] **Step 4: Build.** `npm run build:vite` — expected success. STOP and report if it errors.
- [ ] **Step 5: Contrast gate.** `python3 /home/nomadx/searxng-RAMA/docs/redesign/check-contrast.py` → `All contrast checks pass.`, exit 0.
- [ ] **Step 6: Visual check.** `make run`; open `/search?q=privacy` in a browser. Compare against the prototype **results** view (toggle "results" in the prototype). Verify at a wide window AND a narrow (~375px) window: result titles are near-white and turn red on hover; rows lift slightly on hover; engine badges are small mono chips; no sideways scrolling on mobile; the sidebar sits above the results on mobile. If it does not match, STOP and report what differs — do not tweak values freely.
- [ ] **Step 7: Commit.**
  ```bash
  git add client/simple/src/less/themes/rama/rama.less
  git commit -m "feat(rama): redesigned search, results, sidebar, pagination, hero"
  ```

---

## Phase 3 — Template forks (minimal, class-based)

### Task 5: Homepage hero fork

**Files:**
- Modify (fork): `searx/templates/simple/index.html`

Current file is 8 lines (`.index > .title > h1 + simple_search`). The fork adds the redesigned hero markup while **keeping the `{% extends %}`, `{% block content %}`, and the `simple_search.html` include** (do not restructure the search form — it is styled by `rama.less`).

- [ ] **Step 1:** Replace the inner markup of `{% block content %}` with the prototype's `.home-inner` structure: logomark (`.mark .glyph` + `.wordmark` with `searxng/rama` + `.edition`), the existing `{% include 'simple/simple_search.html' %}`, the categories include, the `.hint`, and the floating `.home-settings` cog linking to `{{ url_for('preferences') }}`. Wrap in `<div class="index home">` so both the upstream `.index` hooks and the new `.home` styles apply. Use `{{ url_for('preferences') }}` for the cog and `{{ instance_name }}` for any brand text — no hard-coded copy.
- [ ] **Step 2: Build + preview.** `make run`, open `/` — compare against the prototype **home** view. Glyph + wordmark lockup, subtle top glow, search field focus ring, category chips, settings cog bottom-right that navigates to Preferences.
- [ ] **Step 3: Note the shared-fork caveat.** This fork is shared across themes (templates aren't theme-scoped). Verify `/` renders correctly under **RAMA** now. Google-theme rendering of this markup is verified later in Task 7's preview and the handback checklist — do not block on it here.
- [ ] **Step 4: Commit.** `git commit -am "feat(rama): redesigned homepage hero (index.html fork)"`

### Task 6: Results sticky header fork

**Files:**
- Modify (fork): `searx/templates/simple/results.html`

- [ ] **Step 1:** Wrap the existing top-of-results search area in a `<header class="rheader">` carrying the compact logomark (link to `/`), the existing `{% include 'simple/search.html' %}` search form, and the About/Preferences actions (`.nav-actions`, using `url_for('info', pagename='about')` and `url_for('preferences')`). **Keep every existing include, block, and the entire `#results`/`#sidebar`/`#urls`/`#pagination` structure unchanged.** Only the header wrapper is new.
- [ ] **Step 2: Build + preview.** Confirm the header is sticky and blurs on scroll, the search field stays inline, and nothing below `#results` shifted or broke.
- [ ] **Step 3: Mobile check** at 375px: header wraps to a second row for the search field (per `rama.less` `@media(max-width:820px)`); no horizontal scroll.
- [ ] **Step 4: Commit.** `git commit -am "feat(rama): sticky condensing results header (results.html fork)"`

---

## Phase 4 — Google theme correctness

### Task 7: Complete + fix the google theme

**Files:**
- Modify: `client/simple/src/less/themes/google/definitions.less`

- [ ] **Step 1: Remove the broken import (unblocks the build).** Delete the line `@import "./components.less";` (last line of the file). That file does not exist and breaks compilation.
- [ ] **Step 2: Fix the mono stack order** for the Linux audience — set `--font-mono` to `"JetBrains Mono", ui-monospace, "SF Mono", Menlo, monospace;` (JetBrains Mono first).
- [ ] **Step 3: Fill the auto-scheme stubs.** Find the two blocks `@media (prefers-color-scheme: dark){ :root.theme-auto{ /* Copy all … here */ } }` and the `light` one. Replace each `/* Copy … */` comment by **copying the full body** of the matching `:root.theme-dark{…}` / `:root.theme-light{…}` block into it (same token lines). No new values — a literal copy.
- [ ] **Step 4: Add the 61 missing tokens to BOTH the `:root.theme-light` and `:root.theme-dark` blocks**, using this exact bucket table (light value / dark value). This is a lookup, not a judgement call — assign every listed token its bucket's value:

  | Bucket | Light | Dark | Tokens (add each with the bucket value) |
  |---|---|---|---|
  | Panel/surface | `#f1f3f4` | `#303134` | `answer-background` `backtotop-background` `result-detail-background` `result-image-background` `result-keyvalue-col-table` `result-keyvalue-odd` `settings-table-group-background` `toolkit-dialog-background` `toolkit-engine-tooltip-background` `toolkit-select-background` `toolkit-checkbox-onoff-off-background` `toolkit-checkbox-onoff-on-background` `toolkit-checkbox-label-background` |
  | Alt row | `#ffffff` | `#28292c` | `result-keyvalue-even` |
  | Primary text | `#202124` | `#e8eaed` | `answer-font` `backtotop-font` `doc-code` `result-detail-font` `result-description-highlight-font` `result-image-span-font` `result-search-url-font` `toolkit-input-text-font` `toolkit-badge-font` `loading-indicator-gap` |
  | Secondary text | `#5f6368` | `#9aa0a6` | `result-detail-label-font` `settings-engine-description-font` |
  | Inverted text | `#ffffff` | `#202124` | `toolkit-kbd-font` `toolkit-checkbox-onoff-off-mark-color` `result-image-span-font-selected` |
  | Border | `#dadce0` | `#5f6368` | `backtotop-border` `result-detail-hr` `result-search-url-border` `toolkit-dialog-border` `toolkit-engine-tooltip-border` `toolkit-select-border` `toolkit-tabs-section-border` `toolkit-checkbox-label-border` |
  | Subtle border | `#e8eaed` | `#3d3f54` | `toolkit-tabs-label-border` |
  | Red accent | `#e11235` | `#e11235` | `result-vim-arrow` `bar-chart-primary` `toolkit-checkbox-input-border` `toolkit-checkbox-onoff-on-mark-background` |
  | Red-2 | `#c30a29` | `#c30a29` | `bar-chart-secondary` |
  | Link | `#1a0dab` | `#8ab4f8` | `result-detail-link` |
  | On-red mark | `#ffffff` | `#ffffff` | `toolkit-checkbox-onoff-on-mark-color` |
  | Chip fill | `#5f6368` | `#9aa0a6` | `toolkit-badge-background` `toolkit-checkbox-onoff-off-mark-background` |
  | Kbd fill | `#202124` | `#e8eaed` | `toolkit-kbd-background` |
  | Hover/selected | `#f1f3f4` | `#3d3f54` | `result-vim-selected` `settings-tr-hover` `toolkit-select-background-hover` |
  | Code bg | `#f1f3f4` | `#171717` | `doc-code-background` |
  | Loader border | `rgba(0,0,0,.1)` | `rgba(255,255,255,.2)` | `result-detail-loader-border` `toolkit-loader-border` |
  | Loader left | `rgba(0,0,0,0)` | `rgba(0,0,0,0)` | `result-detail-loader-borderleft` `toolkit-loader-borderleft` |
  | Loading ind. | `rgba(0,0,0,.2)` | `rgba(255,255,255,.2)` | `loading-indicator` |
  | Image overlay bg | `rgba(0,0,0,.5)` | `rgba(0,0,0,.5)` | `image-resolution-background` |
  | On-overlay text | `#ffffff` | `#ffffff` | `image-resolution-font` |
  | Success bg | `#e6f4ea` | `#1e3a28` | `success-background` |
  | Warning bg | `#fef7e0` | `#4d4225` | `warning-background` |

  Each token name gets the `--color-` prefix (e.g. `--color-answer-background: #f1f3f4;`). That is exactly 61 tokens per block.
- [ ] **Step 5: Verify count.** For each of the light and dark blocks:
  ```bash
  awk '/theme-light|theme-dark/{f=1} f&&/--color-/{print}' client/simple/src/less/themes/google/definitions.less | grep -oE '\-\-color-[a-z-]+' | sort -u | wc -l
  ```
  Expected: `105`. If not 105, a token is missing — STOP and report which.
- [ ] **Step 6: Build under the google theme.** Copy google's file over the entry the build reads (`cp client/simple/src/less/themes/google/definitions.less client/simple/src/less/definitions.less`) then `npm run build:vite`. Expected: success, and preview shows no unstyled areas (detail modal, toolkit, settings). Then restore RAMA: `git checkout client/simple/src/less/definitions.less`. If the build errors, STOP and report.
- [ ] **Step 7: Commit.** `git add client/simple/src/less/themes/google/definitions.less && git commit -m "fix(google): complete 105 tokens, drop broken import, fix auto+mono"`

> **Flag for audit:** the google light/dark values above are a mechanical first pass to guarantee completeness + compilation. The maintainer will refine specific hues in the audit — do not spend effort tuning them.

---

## Phase 5 — Integrate into the RAMA packaging repo

### Task 8: Mirror finalized files + wire PKGBUILD

**Files (in `/home/nomadx/searxng-RAMA`):**
- Modify: `theme/rama/definitions.less`, `theme/google/definitions.less`
- Create: `theme/rama/rama.less`, `theme/rama/fonts.less`, `theme/rama/fonts/*.woff2`, `theme/rama/templates/index.html`, `theme/rama/templates/results.html`
- Modify: `PKGBUILD`

- [ ] **Step 1:** Copy the finalized `definitions.less`, `rama.less`, `fonts.less`, `fonts/*.woff2` from `searxng-custom` into `theme/rama/`; copy `index.html` + `results.html` into `theme/rama/templates/`; copy google's `definitions.less` into `theme/google/`.
- [ ] **Step 2: Extend PKGBUILD `build()`** (the RAMA-customization section) to, after copying `definitions.less`:
  - copy `rama.less` + `fonts.less` into `client/simple/src/less/themes/rama/`,
  - copy `fonts/*.woff2` into `searx/static/themes/simple/fonts/`,
  - append `@import "themes/rama/rama.less";` to `client/simple/src/less/style.less`,
  - copy `theme/rama/templates/*.html` over `searx/templates/simple/`,
  - then the existing `npx vite build`.
- [ ] **Step 3: Local package smoke test.** `makepkg -f` in a clean dir; confirm the built `sxng-ltr.min.css` contains the RAMA overrides and the package installs the fonts + forked templates. (If a full `makepkg` is too heavy in the agent's environment, stop here and flag for the joint audit — do not fake the result.)
- [ ] **Step 4: Commit.** `git commit -am "build: package RAMA redesign (LESS layer, fonts, template forks)"`

---

## Out of scope for this plan (separate plan, done with the maintainer)

Do **not** attempt these here — they are a different subsystem and higher-risk:
- **Multi-variant CSS build + runtime theme-switch** (pre-build `sxng-<variant>.min.css`, swap the served file). This plan ships RAMA as the built-in default only.
- **Cross-distro installer** (Debian/Ubuntu/Fedora dep branching, CSS build step in the Go installer, `install.sh` SearXNG clone).
- **Go installer code nits** (`strings.ReplaceAll`, `os.RemoveAll`, drop `.git` copy, pip timeout, backup cleanup).
- **`.SRCINFO` regen + AUR push** (release step — after audit).
- **Deep Preferences-page layout redesign.** The preferences page inherits the new dark tokens automatically (via `--color-toolkit-*` in `definitions.less`), so it will look on-brand, but a bespoke layout pass on `preferences.less` / `toolkit.less` is deferred to a later redesign round.

---

## Handback checklist (what the maintainer audits afterward)

Report status against each; attach the failing output if a gate did not pass.

- [ ] `npm run build:vite` succeeds for RAMA and for google.
- [ ] `grep -c googleapis sxng-ltr.min.css` → `0`.
- [ ] `python3 docs/redesign/check-contrast.py` → exit 0 (RAMA + google pairs).
- [ ] Both themes define 105 `--color-*` tokens; no unstyled surfaces (detail modal, toolkit, settings, image results, answers, key-value, code, back-to-top).
- [ ] Home + results render at 1280 / 768 / 375 px with no horizontal scroll; match the prototype.
- [ ] Result titles bright-neutral → `#ff7282` on hover; hover-lift rows; sticky results header; settings reachable on home + results.
- [ ] `prefers-reduced-motion` collapses spatial motion; `:focus-visible` rings visible on search, buttons, results, pagination.
- [ ] Template forks kept every upstream Jinja block/include/variable; no route/engine/Python changes.
- [ ] **Open for maintainer decision:** the `#e11235`/`#c30a29` fill deepening (vs original `#ef233c`/`#d90429`) — confirm or veto. Any residual google-theme roughness under the shared template forks.
