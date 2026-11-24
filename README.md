# Hack the Halls

`hth` is the command-line harness for the Hack the Halls LSCC Christmas kata. It logs your team in, runs your program against puzzles, submits the answers, and renders light shows. The full kata spec is in [docs/spec.md](docs/spec.md).

## Install

Download the zip for your machine from the GitHub release (`hth-darwin-arm64.zip`, `hth-darwin-amd64.zip`, `hth-linux-amd64.zip`, `hth-linux-arm64.zip` or `hth-windows-amd64.zip`), unzip it, and put `hth` (`hth.exe` on Windows) on your `PATH`. The zip also holds `examples/`, `stubs/` and this README.

On macOS the binary is unsigned, so clear the quarantine flag once:

```
xattr -d com.apple.quarantine hth
```

Check it works with `hth --version`.

## Commands

```
hth login <team> [--secret S] [--server URL]
hth test <dir> -- <command...>
hth run <puzzle> -- <command...>
hth show <file> -- <command...>
hth --version
```

- `hth login <team>` registers a new team and saves its secret. To join an existing team, pass `--secret`. The config is stored in your user config folder as `hth/config.json`.
- `hth test examples -- <cmd>` runs your program on every example in `examples/` and reports which trees match.
- `hth run <puzzle> -- <cmd>` fetches a live challenge for a puzzle such as `3.2`, runs your program, and submits the answer.
- `hth show <file> -- <cmd>` runs your program on a light show script and publishes the result (see Light shows).

Everything after `--` is your program's command line, for example `hth test examples -- python3 stubs/python/emulator.py`. Starter programs for Python, JavaScript, Java, C# and Go are in `stubs/`; `stubs/README.md` has the run command for each.

## Scoring and unlocks

Each puzzle is scored on the share of bulbs your render gets right. Reaching 80% unlocks the next puzzle. A perfect tree earns 100 points, 80% to 99% earns `40 + 2 x (pct - 80)` points, and anything under 80% earns 0. Your best score per puzzle counts.

## Your program

- The script arrives on stdin.
- Print exactly one answer line on stdout: one character from `RGBWK` per bulb, row by row from the star, left to right. A tree with `n` rows has `n(n+1)/2` bulbs.
- You have 10 seconds per run.
- stderr is shown to the team live, so print debugging there. Extra lines on stdout are rejected.
- A non-zero exit submits nothing.
- Trailing `\r`, spaces and tabs are trimmed before the answer is checked, so Windows line endings are fine.

## Animations

After a run is submitted, `hth` makes up to 120 extra runs of your program on cut-down versions of the script and turns each answer into a frame, then adds the final tree. A slow emulator therefore animates slowly, but the animation never affects your score. A frame run that fails or takes over 10 seconds is skipped. Each frame run is a separate run of your program, at most 8 at a time.

GIFs are written to the current folder, at 600 x 600:

- `hth run`: `hth-run-<puzzle>.gif`
- `hth test`: `hth-test-<example>.gif`
- `hth show`: `hth-show-<file base name>.gif`

A local viewer page also opens in your browser while the animation plays. If the viewer cannot start, you get a one-line warning and the run still submits and writes the GIF. Live puzzles never show an expected tree.

## Light shows

A perfect score on Secret Santa (7.1) unlocks light shows. A show is a script of up to 500 steps, run with `SIZE 16`. Every step becomes one frame, and shows over 500 steps are refused. The latest show per team is kept and shown on the scoreboard. Shows are not scored.

## Troubleshooting

- **Viewer port in use.** The viewer listens on port 47474. Set `HTH_VIEWER_PORT` to use another port.
- **Windows and wrapper commands.** When a run times out, `hth` stops the process it started. If that process is a wrapper such as `npm` or a `.cmd` script, Windows may leave the real program running. Call the interpreter directly (`node emulator.js`, not `npm start`).
- **`403`.** The puzzle is locked. Reach 80% on the previous puzzle first. For shows, get full marks on 7.1.
- **`410`.** The challenge expired (challenges last 5 minutes). Run the command again for a fresh one.
- **`409`.** The challenge was already submitted. Run again for a fresh one.
- **`401`.** You are not logged in, or the secret is wrong. Run `hth login <team> --secret <secret>`.
