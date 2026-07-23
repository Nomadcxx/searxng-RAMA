# Design — SearXNG RAMA

A locked design system for this app. Every page redesign reads this file before
emitting code. Do not regenerate per page — extend or amend this file when the
system needs to grow.

Values are **hex**, not OKLCH: this project's tokens are hex end-to-end (LESS +
`docs/redesign/check-contrast.py`), and inventing a parallel OKLCH set would be a
second source of truth. Contrast is gated by that script — every coloured
text/surface pair must be in it and must pass.

## Genre

**Atmospheric dark tool.** Restrained. The home page carries the only atmosphere
(a faint top glow + index grain); app pages are utility surfaces where function
carries the page.

## Macrostructure families

- **Entry page** (`index.html`) — *Atmospheric search hero*. Lockup, search field,
  category chips, corner settings cog. Built.
- **App pages** (`results.html`, `preferences.html`, `stats.html`) — **Workbench
  shell**: top-anchored app bar → content rail → optional side panel. Variation
  knobs: side panel present/absent, rail width, tab strip present/absent.
- **Content pages** (`info.html`, `404.html`, `new_issue.html`) — **Long Document**:
  single contained prose column, no side panel.

## Theme (RAMA — the default variant)

| Token | Value | Role |
|---|---|---|
| `--ground` / `--ground-2` | `#1e2030` / `#191b28` | page ground |
| `--surface` / `--surface-2` / `--surface-3` | `#282a3b` / `#313349` / `#3a3c54` | raised surfaces, hover, chips |
| `--ink` / `--muted` / `--faint` | `#eef2f6` / `#aab4c5` / `#8d99ae` | text ramp |
| `--accent` | `#ff7282` | **interactive text only** — hover, active, focus ring |
| `--accent-fill` / `--accent-fill-2` | `#e11235` / `#c30a29` | CTA + current-state fills (white label) |
| `--ok` / `--warn` | `#5fd08a` / `#e6c15a` | semantic, never used as accent |
| `--line` / `--line-2` | `rgba(255,255,255,.08)` / `.14` | hairlines |

`google-light` and `google-dark` are **palette variants of this same layout** —
they define the same token names in `theme/google/definitions.less`. Layout never
forks per theme.

## Typography

- **Mono — JetBrains Mono** (self-hosted): the *search vernacular*. Wordmark, URLs
  and breadcrumbs, engine badges, result counts, kbd hints, form labels.
- **Body/UI — Inter** (self-hosted): result titles, snippets, prose, controls.
- No third face. Headings are **roman** — never italic.
- Result title `1.12rem/1.35` weight 600. Meta `.72–.78rem`. Prose max `64ch`.

## Spacing

4-pt named scale, already in `theme/rama/rama.less` (`--sp-1`…`--sp-10`,
`--r-sm/md/lg/pill`). Pages use named tokens, never raw values.

## Motion

- Easing `--ease: cubic-bezier(.16,1,.3,1)`, duration `--dur: .22s`.
- **Reveal pattern: none.** Content must paint immediately — never gate content
  behind an entrance animation (a `backwards`-filled fade once made every search
  result invisible). Hover-lift and focus rings only.
- `prefers-reduced-motion: reduce` collapses transitions to ≤150 ms.
- Max three motion primitives across the app.

## Microinteractions stance

- Silent success. No celebratory toasts.
- Focus ring shows **instantly**, never animated; ≥3:1 contrast.
- Hover tooltips 800 ms delay; focus tooltips 0 ms.

## CTA voice

- **Primary** — filled `--accent-fill`→`--accent-fill-2` gradient, white label,
  pill radius, icon-only buttons are perfect squares (never padded for a label
  that isn't rendered).
- **Secondary** — `--surface` fill, `--line-2` hairline, `--muted` label; hover
  raises to `--surface-2` + `--ink`.

## Per-page allowances

- Entry page MAY use the atmospheric glow + grain. Nothing else may.
- App pages MUST NOT use enrichment, illustration, or decorative gradient.
- Content pages: typography only.

## What pages MUST share

- The `searxng/rama` mono lockup + glyph — **never** the legacy pixel `SEARXNG` PNG.
- One **app bar** (lockup ▸ context ▸ actions). Upstream `#links_on_top` stays hidden.
- Accent reserved for interaction; ≤5% of any viewport.
- The Inter + JetBrains Mono pairing and the type scale above.
- Button/pill voice, surface layering, hairline language, focus ring.

## What pages MAY differ on

- Macrostructure **within** the family (App pages may drop the side panel).
- Tab strip present/absent; rail width.
- Density (results are denser than preferences).

## Non-negotiables (learned the hard way)

1. **Never hide content behind animation.** See Motion.
2. **One left edge.** In any result/list row, URL, title and snippet share a single
   left edge; media is fixed-size and never wrapped around by text.
3. **Style the real DOM.** Upstream ships ion-icons, checkbox-categories and
   `hide_if_nojs` / `show_if_nojs` spans. Style what SearXNG actually renders, not
   an idealised markup.
4. **Every new coloured pair goes into `docs/redesign/check-contrast.py`** and must
   pass before commit.
5. Changes to the shared layer land on **all three variants** — verify each.

## Exports

The consumable form of this system is LESS custom properties, not Tailwind/DTCG.

- **Tokens** — `theme/rama/definitions.less` (`:root`, palette + upstream
  `--color-*`) and the structural block at the top of `theme/rama/rama.less`
  (fonts, spacing, radii, easing).
- **Layout layer** — `theme/rama/rama.less`, imported last so it wins the cascade.
- **Variant palettes** — `theme/google/definitions.less`, flattened per variant by
  `theme/gen-variant.py`.
