# Viewer art assets

Precomputed art for the live viewer and the GIF writer in `internal/art`.
The server copies this whole folder verbatim, so keep the file names and
formats exactly as they are.

| File | What it is |
| --- | --- |
| `tree-<n>.png` | 1200 x 1200 background for a tree of `<n>` rows: tree, trunk and stand on `#0A0E14`, no bulbs. |
| `bulbs-<n>.json` | `[{"x":600,"y":120,"r":21.3}, ...]`: one entry per bulb, in answer order (row by row from the star, left to right). Numbers may have decimals. |

Sizes with assets: 3, 4, 16 and 48. Any other size falls back to a plain
`#0A0E14` background with bulbs from the render layout.

## Editing rules

- Keep the PNG exactly 1200 x 1200.
- Do not use any colour within squared RGB distance 3000 of the bulb
  colours `#FF0000`, `#00FF00`, `#0000FF`, `#FFFFFF`, `#303030`, so a lit
  bulb never blends into the decoration. `go test ./internal/art` checks this.
- Use at most 240 distinct colours per PNG (no anti-aliasing gradients
  across the whole picture), so the GIF needs no dithering. The same test
  checks this.
- Leave the bulb areas empty. Bulbs are drawn over the background at the
  positions in `bulbs-<n>.json`; keep that table in answer order, with `y`
  never decreasing and `x` increasing within a row.

## Regenerating

`go run ./cmd/genart` writes a first draft of all eight files. It refuses to
overwrite existing assets, to protect hand-finished art; pass `-force` to
replace them.
