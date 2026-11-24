# Viewer art assets

Precomputed art for the live viewer and the GIF writer in `internal/art`.
The server copies this whole folder verbatim, so keep the file names and
formats exactly as they are.

| File | What it is |
| --- | --- |
| `tree-<n>.png` | 1200 x 1200 background for the `<n>`-row grid: a night scene around `#0A0E14` with a layered fir, trunk, pot and snow, a dark light-string wire in wiring order and a small brass socket at every bulb position but the star's. No bulbs. |
| `bulbs-<n>.json` | `[{"x":600,"y":120,"r":21.3}, ...]`: one entry per bulb, in answer order (row by row from the star, left to right). Numbers may have decimals. |

Only two grids have assets: 16 and 48. Every tree is drawn on one of them:
a script of 1 to 16 rows on the 16 grid, and 17 to 48 rows on the 48 grid.
Bulb `i` of the answer sits on entry `i` of the grid's table, so a 4-row tree
lights the top four rows of the 16 grid. Grid positions with no bulb of the
answer are drawn as unlit sockets. Any other size falls back to a plain
`#0A0E14` background with bulbs from its own render layout.

## Editing rules

- Keep the PNG exactly 1200 x 1200.
- Do not use any colour within squared RGB distance 3000 of the bulb
  colours `#FF0000`, `#00FF00`, `#0000FF`, `#FFFFFF`, `#303030`, so a lit
  bulb never blends into the decoration. `go test ./internal/art` checks this.
- Use at most 240 distinct colours per PNG (no anti-aliasing gradients
  across the whole picture), so the GIF needs no dithering. The same test
  checks this.
- Keep bulb positions free of anything but the socket and wire. Bulbs are
  drawn over the background at the positions in `bulbs-<n>.json`; keep that
  table in answer order, with `y` never decreasing and `x` increasing within
  a row.

## Regenerating

`go run ./cmd/genart` writes all four files. It refuses to
overwrite existing assets, to protect hand-finished art; pass `-force` to
replace them.
