# hth Harness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the 2025 JavaScript/Python clients with `hth`. `hth` is a single static Go binary that:
- runs a team's emulator as a subprocess, then renders and submits its answer;
- runs the worked examples locally;
- animates each run in a browser viewer, using frames built from the team's own emulator, and saves the animation as a GIF;
- uploads light shows.

The repo also ships the worked examples and starter stubs.

**Architecture:** One Go module, standard library only. `cmd/hth` dispatches subcommands to small `internal/` packages:
- `tree`: answer validation and the plain PNG render.
- `program`: one subprocess run.
- `api`: HTTP client and config.
- `examples`: worked-example loading.
- `script`: script structure and frame planning (cut-down scripts).
- `art`: embedded viewer art, frame drawing, and GIF output.
- `viewer`: a detached local daemon that serves the page and streams frames.
- `animate`: the parallel frame runner that ties `script`, `program`, `viewer`, and `art` together.

**Tech Stack:** Go 1.25+ (not installed on this machine; install before starting), `image/png`, `image/gif`, `net/http`, `embed`, `os/exec`. Vanilla JS + canvas in the viewer. GitHub Actions for release builds.

**Spec:** `C:\Users\edfar\source\repos\codurance\lscc\christmas\Hack the Halls LSCC Christmas kata v1.1.md` (copy into `docs/spec.md` in Task 1, without the "Organiser notes" section, which stays private).

**Companion plan:** `ChristmasKataServer/docs/superpowers/plans/2026-11-24-hack-the-halls-server.md`. That plan uses three things from this one:
- Server Task 9 uses the PNG fixtures from this plan's Task 2.
- Server Task 13 copies `internal/art/assets/` from this plan's Task 8.
- Server Task 13 copies `tree.js` from this plan's Task 10.

## Global Constraints

- Commits land on `main`. The 2025 code is already archived on another branch.
- Binary name `hth`, with targets `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, built with `CGO_ENABLED=0`.
- No Go dependencies beyond the standard library.
- Default server `https://lscc.lifeatcodurance.com/`. Config at `os.UserConfigDir()/hth/config.json` as `{"server","team","secret"}`, file mode `0600`.
- Tree: row r holds r + 1 bulbs. Bulb count `n(n+1)/2`. Answer chars `RGBWK` only, one per bulb, in answer order (row by row from the star, left to right).
- Program contract: script on stdin, exactly one answer line on stdout, 10 s time limit. stderr is shown to the team live. A non-zero exit submits nothing. Trailing `\r`, spaces, and tabs are trimmed before validation.
- Render layout (shared with server): canvas 1200 × 1200, background `#000000`, `h = 1000 / (n − 1)`, `v = 0.866 × h`, centre `x = 600 + (p − r/2) × h`, `y = 120 + r × v`, bulb radius `0.32 × h`. Colours: R `#FF0000`, G `#00FF00`, B `#0000FF`, W `#FFFFFF`, K `#303030`. The submitted PNG is plain, with no decoration.
- Viewer art:
  - The canvas is 1200 × 1200, with a deep-green layered tree and a trunk in a stand, on a plain dark background (`#0A0E14`).
  - The decoration uses no palette colour.
  - Precomputed assets exist for sizes 3, 4, 16, and 48. Any other size falls back to the plain dark background with bulbs from the render layout.
- Frames:
  - `hth run` and `hth test` take up to 120 sampled points, plus the final tree.
  - `hth show` keeps every step, and shows over 500 steps are refused.
  - Each frame is one run of the team's program on a cut-down script. At most 8 run in parallel. A failed or timed-out frame run is skipped.
  - For `hth run`, frame runs start only after submission.
- GIFs go to the current working folder:
  - `hth run`: `hth-run-<puzzle>.gif`.
  - `hth test`: `hth-test-<example>.gif`.
  - `hth show`: `hth-show-<file base name>.gif`.
  - Each is 600 × 600. Frames play at 100 ms. The final tree holds for 2 s.
- Live puzzles never show an expected tree.

### API contract (identical in server plan)

Auth headers `X-Team-Name` and `X-Unique-Key` are required on every route except those marked *public*. Errors return `{"error": "<message>"}`.

| Route | Request | Success | Failures |
| --- | --- | --- | --- |
| `GET /register/{teamName}` *public* | — | `200` JSON string: the team secret | `400` invalid or taken name |
| `POST /puzzles/{puzzle}/challenges` | — | `201` `{"challengeId":"<guid>","puzzle":"3.2","size":16,"script":"SIZE 16\n…","expiresAt":"<ISO-8601>"}` | `404` unknown puzzle, `403` locked |
| `POST /challenges/{challengeId}/submission` | body `image/png`, ≤ 1 MiB | `200` `{"puzzle":"3.2","correct":131,"bulbCount":136,"points":72,"bestPoints":72,"unlocked":"3.3"}`. `unlocked` is the puzzle this submission opened, else `null` | `404` unknown or another team's, `409` already submitted, `410` expired, `400` not a 1200 × 1200 PNG, `413` too large |
| `GET /shows/permission` | — | `204` | `403` `Santa comes first: get full marks on 7.1 to unlock light shows` |
| `POST /shows` | body `{"frames":["<136 chars>",…]}`, ≤ 256 KiB | `201` `{"frames":412}` | `403` as above, `400` no frames, over 500 frames, or a frame not 136 `RGBWK` chars, `413` too large |
| `GET /shows` *public* | — | `200` `[{"teamName","frames":[…],"submittedAt"}]`, the latest show per team, oldest submission first | — |
| `GET /leaderboard` *public* | — | `200` `[{"teamName":"team-07","total":1234,"puzzles":{"1.1":100}}]` (points), total descending. | — |
| `GET /feed` *public* | — | `200` `{"tiles":[Tile…]}`, newest first, max 8 | — |
| `GET /feed/events` *public* | — | `text/event-stream`: event `tile` (data Tile), event `leaderboard` (data leaderboard), and event `show` (data `{"teamName","first":bool}`; the page then reloads `/shows`) | — |

`hth` uses only the first six routes. The leaderboard and feed routes serve the scoreboard.

## Review Focus

1. A program built on Windows prints `CRLF`, or a trailing space after the answer. Expected: the answer is accepted. Test in Task 3.
2. A program prints debug lines to stdout. Expected: it is rejected with `expected 1 answer line, got N (print debugging to stderr)`, and nothing is submitted. Test in Task 3.
3. A program hangs, or is slow on every frame run. Expected: the submission still happens first. Each frame run is killed at 10 s and skipped, and the animation shows the frames that worked. Tests in Tasks 3 and 11.
4. The viewer port is taken, or no browser can open (SSH session, locked-down laptop). Expected: the run still submits, prints its score, and writes the GIF. The viewer problem is a one-line warning. Test in Task 11.
5. The team is not logged in, the network is down, the challenge expired (`410`), or the puzzle is locked (`403`). Expected: a one-line actionable message and a non-zero exit, never a Go stack trace. Test in Task 6.

---

### Task 1: Clear the 2025 clients and scaffold the CLI

**Files:**
- Delete: `javascript/`, `python/`
- Create: `go.mod` (`module hth`), `cmd/hth/main.go`, `cmd/hth/main_test.go`, `.gitignore` (`/dist/`, `/hth`, `/hth.exe`, `hth-*.gif`), `docs/spec.md`
- Modify: `README.md` (placeholder: one paragraph plus "see Task 14")

**Interfaces:**
- Produces: `func run(args []string, stdout, stderr io.Writer) int` in `cmd/hth/main.go`. `main` calls `os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))`. Subcommands `login`, `run`, `test`, `show`, and the hidden `viewer` are dispatched by a `switch`. The `--` separator splits the program command from hth's own arguments. `var version = "dev"` is set by ldflags.

- [ ] **Step 1: Write failing tests:**
  - `TestNoArgsPrintsUsage`: `run(nil, …) == 2`, and stderr contains `hth login <team>`, `hth run <puzzle> -- <command…>`, `hth test <dir> -- <command…>`, and `hth show <file> -- <command…>`.
  - `TestUnknownCommand`: `run([]string{"frobnicate"}) == 2`.
  - `TestRunWithoutSeparator`: `run([]string{"run","3.2","python"}) == 2`, and stderr mentions `--`.
  - `TestVersion`: `run([]string{"--version"})` prints `dev`.
- [ ] **Step 2:** Run `go test ./...`. Expected: FAIL.
- [ ] **Step 3:** Implement the dispatch, with subcommand bodies returning "not implemented" (exit 1) until later tasks.
- [ ] **Step 4:** Run `go test ./... && go vet ./...`. Expected: PASS.
- [ ] **Step 5:** Commit: `chore: replace 2025 clients with hth go scaffold`.

### Task 2: Answer validation and PNG rendering

**Files:**
- Create: `internal/tree/tree.go`, `internal/tree/render.go`, `internal/tree/tree_test.go`, `internal/tree/render_test.go`, `internal/tree/testdata/` (fixtures)

**Interfaces:**
- Produces:
  - `func BulbCount(size int) int`.
  - `func Validate(line string, size int) error`. Messages: `answer has L characters, expected E for SIZE n` and `answer has invalid character 'x' at position i (allowed: RGBWK)`.
  - `func Centre(size, r, p int) (x, y float64)` and `func Radius(size int) float64`.
  - `func Diff(expected, actual string, size int) [][2]int`.
  - `func Render(line string, size int) ([]byte, error)`, which returns PNG bytes.

- [ ] **Step 1: Write failing tests:**
  - `TestBulbCount`: 3 → 6, 16 → 136, 48 → 1176.
  - `TestValidate`: `"RRGRGW"`/3 passes; `"RRGRG"`/3 fails with `answer has 5 characters, expected 6 for SIZE 3`; `"RRGRGx"` fails naming `'x'` at position 5.
  - `TestCentre`: on 16 rows, `Centre(16,0,0) == (600,120)` and `Centre(16,15,0) == (100, 120+15*0.866*1000/15)`.
  - `TestDiff`: `Diff("RRGRGW", "RRGRGK", 3) == [(2,2)]`.
  - `TestRenderPixels`: render `"RGBWKR"` at size 3, decode, and check the pixel at each bulb centre is its palette colour, at (0,0) is `#000000`, and the image is 1200 × 1200.
  - `TestFixturesUpToDate`: re-render each fixture and compare bytes. With `-update`, the test writes them instead.
- [ ] **Step 2:** Run `go test ./internal/tree`. Expected: FAIL.
- [ ] **Step 3:** Implement. Fill pixels whose centre `(x+0.5, y+0.5)` lies within the bulb radius. No anti-aliasing, so server sampling is exact. Encode with `png.Encoder{CompressionLevel: png.BestCompression}`.
- [ ] **Step 4:** Generate the fixtures with `go test ./internal/tree -update`. Each one is a `.png` plus a `.txt` holding the line:
  - `tree-3-mixed` (`RGBWKR`)
  - `tree-16-cycle` (136 chars repeating `RGBWK`)
  - `tree-48-cycle` (1176 chars repeating `RGBWK`)
  - `blank-16` (a 1200 × 1200 all-black PNG, written directly; its `.txt` is 136 `.`)
- [ ] **Step 5:** Run `go test ./internal/tree`. Expected: PASS. Then check the PNG size: `tree-48-cycle.png` must be < 1 MiB, the server's upload limit.
- [ ] **Step 6:** Commit: `feat: answer validation and png render with shared fixtures`.

### Task 3: Running the team's program

**Files:**
- Create: `internal/program/program.go`, `internal/program/program_test.go`, `internal/program/programtest/programtest.go`, `internal/program/testdata/fake/main.go` (a test helper program)

**Interfaces:**
- Produces:
  - `func Run(ctx context.Context, argv []string, script string, stderr io.Writer) (answer string, err error)`. The timeout comes from package variable `Timeout = 10 * time.Second`, which tests may lower. stderr is streamed to the writer. Callers pass `io.Discard` for frame runs.
  - Errors: `program exited with code N; nothing submitted`, `program timed out after 10s; nothing submitted`, `no answer line on stdout`, `expected 1 answer line, got N (print debugging to stderr)`, and `could not start "<argv0>": <err>`.
  - `programtest.FakePath(t)`: builds the helper once per test binary and returns its path, for tests in other packages.
- Test helper `fake`: its behaviour is picked by its first argument:
  - `echo-k`: prints `n(n+1)/2` Ks for the script's `SIZE`.
  - `count`: prints `K`s, with `R` at the index equal to the number of non-comment instruction lines, mod the bulb count. Tasks 11 and 12 use it to tell frames apart.
  - `crlf`.
  - `debug-stdout`.
  - `sleep 20`.
  - `exit 3`.
  - `stderr`.
  - `no-read`.
  - `slow-after N`: sleeps 20 s once the script has more than N lines.
  - `sleep-ms N`: sleeps N ms, then behaves as `echo-k`.
  - `log PATH`: appends the current time to PATH, then behaves as `count`.

- [ ] **Step 1: Write failing tests:**
  - `TestAnswerFromStdout`: `echo-k` with `SIZE 3` gives `"KKKKKK"`.
  - `TestCRLFAndTrailingSpaceTrimmed`.
  - `TestDebugOnStdoutRejected`.
  - `TestTimeout`: `sleep 20` errors within 11 s with the timeout message.
  - `TestNonZeroExit`: the message has `code 3`.
  - `TestStderrStreamed`: the stderr buffer contains the helper's text.
  - `TestMissingCommand`: `could not start`.
  - `TestProgramIgnoringStdin`: `no-read` exits 0 with its answer. Writing stdin must not deadlock or fail on a broken pipe.
- [ ] **Step 2:** Run `go test ./internal/program`. Expected: FAIL.
- [ ] **Step 3:** Implement with `exec.CommandContext` and `cmd.WaitDelay = time.Second`. Write stdin from a goroutine and ignore `EPIPE`. Exactly one non-blank stdout line is allowed, checked after `TrimRight(" \t\r")`. Answer validation against `SIZE` stays in `tree.Validate`.
- [ ] **Step 4:** Run the tests. Expected: PASS on Linux and Windows. Note: on Windows, killing a wrapper such as `npm` may leave its child running. The README documents this; it is not fixed.
- [ ] **Step 5:** Commit: `feat: run team program with time limit`.

### Task 4: Config, API client, and `hth login`

**Files:**
- Create: `internal/api/config.go`, `internal/api/client.go`, `internal/api/client_test.go`, `internal/api/config_test.go`
- Modify: `cmd/hth/main.go`

**Interfaces:**
- Produces:
  - `type Config struct { Server, Team, Secret string }`, with `func LoadConfig() (Config, error)` and `func SaveConfig(Config) error`. The path comes from `os.UserConfigDir`, overridable by env `HTH_CONFIG` for tests.
  - `type Client struct { Config; HTTP *http.Client }`, with methods:
    - `func (c Client) Register(team string) (secret string, err error)`.
    - `func (c Client) FetchChallenge(puzzle string) (Challenge, error)`, with `type Challenge struct { ID, Puzzle string; Size int; Script string; ExpiresAt time.Time }`.
    - `func (c Client) Submit(id string, png []byte) (Score, error)`, with `type Score struct { Puzzle string; Correct, BulbCount, Points, BestPoints int; Unlocked *string }`.
    - `func (c Client) ShowPermission() error`.
    - `func (c Client) UploadShow(frames []string) error`.
  - `type APIError struct { Status int; Message string }`. `Error()` maps status codes to actionable text:
    - `401`: `not logged in or wrong secret; run hth login <team> --secret <secret>`
    - `403`: the server message (locked puzzle, or Santa comes first)
    - `404`: the server message
    - `409`: `challenge already submitted; run again for a fresh one`
    - `410`: `challenge expired (5 minutes); run again for a fresh one`
    - other: `server error <status>: <message>`
  - Network errors are wrapped as `could not reach <server>: <err>`.
  - CLI `hth login <team> [--secret S] [--server URL]`. With `--secret`, it saves without calling the server. Without it, it registers, saves, and prints `Logged in as <team>. Secret: <s>` and `Share this secret with your pair: hth login <team> --secret <s>`.

- [ ] **Step 1: Write failing tests** against `httptest.Server`:
  - `TestRegisterReturnsSecret`.
  - `TestFetchSendsAuthHeaders`, and it decodes the challenge.
  - `TestSubmitPostsPNG`: `Content-Type: image/png`, the body is the bytes, and the response decodes `unlocked` both as a string and as `null`.
  - `TestShowPermission`: `204` gives nil; `403` gives an `APIError` whose text is the server's Santa message.
  - `TestUploadShowPostsFrames`: the JSON body has `frames`.
  - `TestErrorMapping`: one case per status above.
  - `TestConfigRoundTrip`: on Unix, the file mode is `0600`.
  - `TestLoginWithSecretSkipsServer`.
- [ ] **Step 2:** Run `go test ./internal/api ./cmd/hth`. Expected: FAIL.
- [ ] **Step 3:** Implement. Use a 30 s HTTP timeout. Trim the trailing `/` from the server URL.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: api client, config and hth login`.

### Task 5: Worked examples and `hth test`

**Files:**
- Create: `examples/*.txt` (table below), `internal/examples/examples.go`, `internal/examples/examples_test.go`
- Modify: `cmd/hth/main.go`

**Interfaces:**
- Produces:
  - `type Example struct { Name, Script string; Size int; Expected string }`.
  - `func Load(dir string) ([]Example, error)`. It reads `*.txt` sorted by name. `Size` comes from the `SIZE` line, and `Expected` from the line `# expect: <answer>`. The emulator ignores that line as a comment. A file without one is an error naming the file.
  - CLI `hth test <dir> -- <command…>`. For each example it prints `PASS 1-lights-on` or `FAIL 2-order-order: wrong bulbs at (2,1) (3,0)`, then `N/M passed`. It exits 1 if any fail. It never contacts the server. Task 11 adds the animation.

Example files. Each starts with `# <round name> — <what it shows>`, then `# expect: …`. The answers were computed with a reference emulator, and the first and fifth match the spec:

| File | Script (one instruction per line) | Expected |
| --- | --- | --- |
| `1-lights-on.txt` | `SIZE 3`, `FILL G`, `SET 0 0 W`, `ROW 2 R`, `ROTATE` | `RRGRGW` |
| `2-order-order.txt` | `SIZE 4`, `SET 3 0 R`, `SET 2 1 G`, `SET 0 0 B`, `MIRROR`, `ROTATE` | `KKKKGKRKKB` |
| `3-spreading-cheer.txt` | `SIZE 4`, `SET 0 0 R`, `SET 3 0 G`, `SET 3 3 B`, `REPEAT 2`, `CYCLE`, `GROW R`, `ROTATE`, `END` | `RRRKKKBKKW` |
| `4-light-show.txt` | `SIZE 4`, `ROW 0 W`, `REPEAT 2`, `FALL`, `SEQUENCE G`, `END`, `ROTATE`, `REPEAT 3`, `CHASE`, `END` | `KKGKKKKGKK` |
| `4-falling-star.txt` | `SIZE 4`, `SET 0 0 R`, `SET 1 0 G`, `SET 1 1 B`, `FALL` | `KKKKRKKGBK` |
| `5-long-haul.txt` | `SIZE 4`, `# comments and blank lines are ignored, even inside blocks`, `SET 0 0 R`, `ROW 3 B`, *(blank)*, `REPEAT 0`, `FILL W`, `END`, `REPEAT 1`, `REPEAT 2`, `# nothing here`, `END`, `REPEAT 3`, `CYCLE`, `REPEAT 2`, `CHASE`, `END`, `END`, `END`, `SET 3 1 G` | `KKGGGGWGKK` |
| `6-silent-night.txt` | `SIZE 4`, `SET 2 1 R`, `SET 3 0 G`, `REPEAT 1000000000`, `CYCLE`, `GROW R`, `ROTATE`, `END` | `GGGRRRRRRR` |
| `7-secret-santa.txt` | `SIZE 4`, `SET 0 0 R`, `SET 2 0 G`, `SET 3 2 B`, `SET 1 1 W`, `REPEAT 1000000000`, `ROTATE`, `CHASE`, `CYCLE`, `END`, `REPEAT 999999999`, `MIRROR`, `CHASE`, `END` | `GKKKBKRWKK` |

Indent block bodies by two spaces in the files, as the spec examples do. Server Task 3 pins the same scripts and answers as tests, so both repos agree.

- [ ] **Step 1: Write failing tests:**
  - `TestLoadExamples`: `Load("../../examples")` gives 8 examples, with `[0].Name == "1-lights-on"`, `[0].Size == 3`, and `[0].Expected == "RRGRGW"`.
  - `TestMissingExpectIsError`.
  - `TestHthTestReportsFailures` in `cmd/hth`: with the fake helper (`echo-k`), `run(["test","examples","--",fake,"echo-k"]) == 1`, and stdout ends `0/8 passed`.
- [ ] **Step 2:** Run the tests. Expected: FAIL.
- [ ] **Step 3:** Implement.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: worked examples and hth test`.

### Task 6: `hth run`

**Files:**
- Modify: `cmd/hth/main.go`
- Create: `cmd/hth/run_test.go`

**Interfaces:**
- Consumes: `api.Client`, `program.Run`, `tree.Validate`, `tree.Render`.
- Steps:
  1. Load the config. If it is missing, print `not logged in; run hth login <team>` and exit 1.
  2. Fetch the challenge and run the program with the script.
  3. Validate the answer against `challenge.Size`, render it, and submit.
  4. Print `Puzzle 3.2: 131/136 correct (96%), 72 points (best 72)`, plus `Unlocked 3.3` when `unlocked` is set. The percentage is `floor(100 × correct / bulbCount)`, as the server computes it.

  Any error prints `hth: <message>` and exits 1, with no stack trace. A validation or program failure submits nothing. Task 11 adds the animation after this point.

- [ ] **Step 1: Write failing tests** with an `httptest.Server` faking the API, and the fake helper:
  - `TestRunHappyPath`: prints `Puzzle 1.1: 136/136 correct (100%), 100 points (best 100)` and `Unlocked 1.2`. The server received one PNG that decodes to 1200 × 1200.
  - `TestRunWrongLengthSubmitsNothing`: the fake prints `KKK` for `SIZE 16`. Exit 1, and the server got no submission.
  - `TestRunNotLoggedIn`.
  - `TestRunLocked`: the fetch returns `403` `puzzle 1.2 is locked: reach 80% on 1.1 first`. Exit 1, and that message is printed.
  - `TestRunExpired`: the submit returns `410`. Exit 1 with the expired message.
  - `TestRunServerDown`: the URL points at a closed port. Exit 1, and the message contains `could not reach`.
- [ ] **Step 2:** Run the tests. Expected: FAIL.
- [ ] **Step 3:** Implement.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: hth run fetches, runs, renders, submits`.

### Task 7: Script structure and frame planning

**Files:**
- Create: `internal/script/script.go`, `internal/script/plan.go`, `internal/script/script_test.go`, `internal/script/plan_test.go`

The harness never interprets instructions. It only knows `SIZE`, `REPEAT n`/`END` nesting, comments, and blank lines. Every other line is an opaque step.

**Interfaces:**
- Produces:
  - `type Node struct { Line string; Count int64; Body []Node; Loop bool }`. A loop has `Loop == true`, its `REPEAT` count in `Count`, and its body in `Body`. Any other node is one step, held in `Line`.
  - `type Script struct { Size int; Body []Node }`.
  - `func Parse(text string) (Script, error)`. Errors: `line N: missing SIZE`, `line N: unmatched END`, `line N: unclosed REPEAT`, `line N: bad REPEAT count`.
  - `func (s Script) Steps() *big.Int`: the total executed steps, counted arithmetically (no loop is walked).
  - `type Frame struct { Script string; SkippedBefore int64 }`. `Script` is a cut-down script that stops after one step. `SkippedBefore` is the number of loop passes the animation jumps over just before this frame, or 0.
  - `type Plan struct { Frames []Frame; SkippedAtEnd int64 }`.
  - `func Sample(s Script, max int) Plan`, for `hth run` and `hth test`, with `max` 120.
  - `func Every(s Script) Plan`, for `hth show`. It keeps every step.

Rules:
- **Cut-down script for a point:** keep `SIZE` and everything before the point unchanged. For each loop the point is inside, at pass k (0-based), emit `REPEAT k` with the full body when `k > 0`, then the body up to the point, cut down the same way. Then stop.
- **Sample:** walk execution in order and record each step as a candidate. Each loop execution walks at most its first 100 passes. The remaining passes become one skip of `count − 100`. The skip attaches to the next frame, or to `SkippedAtEnd` if nothing follows. Stop recording after 100,000 candidates; any later skips still count. Pick `max` candidates spread evenly by index. Always include the first and last candidates, and include the first frame after every skip. The final tree is the real answer line, so it is not a frame.
- **Every:** walk all steps with no pass cap. The caller refuses scripts where `Steps() > 500` before calling it.

- [ ] **Step 1: Write failing tests:**
  - `TestParseNested`: the `5-long-haul` example parses with the right nesting, and comments and blank lines are dropped.
  - `TestParseErrors`: one case per message.
  - `TestStepsCountsArithmetically`: `6-silent-night` has `2 + 3 × 1,000,000,000` steps. A 1000 × 1e9 × 3 nest returns `3,000,000,000,000` without a hang.
  - `TestCutDownInsideLoop`: for `SIZE 4 / SET 0 0 R / REPEAT 5 / CHASE / CYCLE / END`, the point at pass 2 after `CHASE` gives `SIZE 4\nSET 0 0 R\nREPEAT 2\nCHASE\nCYCLE\nEND\nCHASE\n`.
  - `TestCutDownNested`: a point inside an inner loop cuts both loops, outer first.
  - `TestSampleBillionLoop`: on `6-silent-night`, at most 120 frames, all within the first 100 passes, and `SkippedAtEnd == 999,999,900`.
  - `TestSampleSkipThenMore`: two billion-pass loops in sequence. A frame after the first loop has `SkippedBefore == 999,999,900`, and its cut-down script keeps the first loop's full `REPEAT 1000000000`.
  - `TestSampleShortScript`: a 5-step script gives 5 frames and no skips.
  - `TestEveryKeepsAllSteps`: a 40-step script with a `REPEAT 10` gives 40 frames.
  - `TestCutDownRoundTrips`: every frame script re-parses, and its `Steps()` equals the frame's position (1-based) for `Every`.
- [ ] **Step 2:** Run `go test ./internal/script`. Expected: FAIL.
- [ ] **Step 3:** Implement. Represent a point as a path of `(node index, pass)` pairs. Build the cut-down text from the path.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: script structure and frame planning`.

### Task 8: Viewer art and GIF output

**Files:**
- Create:
  - `internal/art/art.go`, `internal/art/draw.go`, `internal/art/gif.go`, `internal/art/art_test.go`.
  - `internal/art/assets/tree-{3,4,16,48}.png` and `internal/art/assets/bulbs-{3,4,16,48}.json`, generated and then hand-finishable.
  - `internal/art/assets/README.md`, which explains the editing rules and that the server copies this folder.
  - `cmd/genart/main.go`, which writes the first-draft assets.

**Interfaces:**
- Produces:
  - Asset formats:
    - `bulbs-<n>.json` is `[{"x":600,"y":120,"r":21.3}, …]`, in answer order.
    - `tree-<n>.png` is the 1200 × 1200 background. It holds the tree, trunk, and stand on `#0A0E14`, with no bulbs.
  - `type Bulb struct { X, Y, R float64 }`.
  - `func Load(size int) (bg image.Image, bulbs []Bulb)`. It returns the embedded assets, or for other sizes a plain `#0A0E14` background with bulbs from `tree.Centre`/`tree.Radius`.
  - `func DrawFrame(dst *image.Paletted, size int, line string, scale float64)`. It copies the background, then draws each bulb in its state colour. Bulb 0 is a five-point star, and the others are discs.
  - `func WriteGIF(path string, size int, frames []string, captions map[int]string, final string) error`. It writes 600 × 600 frames at 100 ms, a 2 s hold on `final`, and a caption bar for frames with a caption.
  - Palette rule: the five bulb colours, plus white caption text, plus the background's own colours. The background holds at most 240 distinct colours, so the GIF needs no dithering.
- `cmd/genart` draws the first draft:
  - Bulb table: from the render layout.
  - Tree: three or four stacked green triangles (`#0B3D1E`, `#145A2E`, `#1E6B38`) behind the bulb area.
  - Trunk: brown `#5A3A1E`.
  - Stand: grey `#3A3F47`.
  - Background: `#0A0E14`.

  The organisers' pixel-art pass edits the PNGs by hand afterwards, so `genart` refuses to overwrite assets without `-force`.

- [ ] **Step 1: Write failing tests:**
  - `TestAssetsPresent`: for 3, 4, 16, and 48, the background is 1200 × 1200, and the bulb table has `BulbCount(n)` entries in answer order (`y` never decreases, and `x` increases within a row).
  - `TestBackgroundAvoidsPalette`: no background pixel is within squared RGB distance 3000 of any of the five palette colours.
  - `TestBackgroundColourBudget`: each background has ≤ 240 distinct colours.
  - `TestFallbackSize`: `Load(7)` gives a plain background and 28 bulbs at the render-layout centres.
  - `TestDrawFrameColoursBulbs`: on 16 rows at `scale` 0.5, the pixel at each bulb centre / 2 is its state colour.
  - `TestWriteGIF`: three frames, a caption on frame 2, and a final tree. Decode with `gif.DecodeAll`: 4 images, 600 × 600, delays `[10,10,10,200]`.
- [ ] **Step 2:** Run `go test ./internal/art`. Expected: FAIL.
- [ ] **Step 3:** Write `cmd/genart`, run `go run ./cmd/genart`, check the PNGs by eye, and commit them. Implement `art`. Downscale with nearest-neighbour so the background palette is kept. Draw the caption with a small embedded 5 × 7 bitmap font for digits, commas, spaces, and the letters in `passes later`.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: precomputed viewer art and gif writer`.

### Task 9: Viewer daemon

**Files:**
- Create:
  - `internal/viewer/daemon.go` and `internal/viewer/publish.go`.
  - `internal/viewer/spawn_unix.go` (`//go:build !windows`, `Setsid`) and `internal/viewer/spawn_windows.go` (`CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS`).
  - `internal/viewer/browser.go`.
  - `internal/viewer/static/index.html`, a placeholder until Task 10.
  - `internal/viewer/viewer_test.go`.
- Modify: `cmd/hth/main.go`. The `viewer` subcommand runs `viewer.Serve`.

**Interfaces:**
- Produces:
  - `type Card struct { ID, Title string; Size int; Answer string; Expected string /* test only */; Summary string; FrameCount int; IntervalMs int }`. One card is one tree, with any expected tree beside it.
  - `type Publication struct { ID, Kind string /* "run"|"test"|"show" */; Cards []Card }`.
  - `type FrameEvent struct { Publication, Card string; Index int; Line string; SkippedBefore int64 }`.
  - `type Session struct { … }`, with `func (s *Session) Frame(FrameEvent)` and `func (s *Session) Done(card string, skippedAtEnd int64)`.
  - `func Open(p Publication) (*Session, error)`. It starts the daemon if needed and opens the browser when no tab is connected.
  - `func Serve(addr string) error` and `func Handler() http.Handler`.
  - `func Addr() string`: `127.0.0.1:47474`, or the env `HTH_VIEWER_PORT`.
- Daemon routes:
  - `GET /` and `GET /static/*`: the embedded page.
  - `GET /art/*`: `internal/art/assets`.
  - `GET /latest`: the last publication with the frames so far, or `204`.
  - `GET /events`: SSE, with events `publication`, `frame`, and `done`.
  - `GET /health`: `{"app":"hth","clients":<n>}`.
  - `POST /publish`, `POST /frame`, and `POST /done`: each requires header `X-HTH: 1` and a JSON body ≤ 1 MiB, otherwise `403`/`413`. The custom header blocks cross-site form posts from web pages.
- `Open`:
  1. `GET /health`. If unreachable, spawn `os.Executable() viewer` detached and poll health for up to 3 s.
  2. If something answers that is not hth, return `port 47474 is in use; set HTH_VIEWER_PORT`.
  3. `POST /publish`.
  4. If `clients == 0`, open the browser (`open`, `xdg-open`, or `rundll32 url.dll,FileProtocolHandler`). A browser failure returns `open http://127.0.0.1:47474 in your browser`.

  `Session` methods post in the background and drop errors after the first, so a dead viewer never slows the frame runs.

- [ ] **Step 1: Write failing tests** against `httptest.NewServer(viewer.Handler())`:
  - `TestPostsRequireHeader`.
  - `TestLatestIncludesFrames`: publish, then two frames. `/latest` has both, in index order.
  - `TestEventsStream`: an SSE client receives `publication`, `frame`, and `done` within 1 s each.
  - `TestHealthCountsClients`.
  - `TestBodyLimit`.
  - `TestForeignPortError`: the health endpoint returns HTML.
  - `TestArtServed`: `/art/bulbs-16.json` returns `200`.
- [ ] **Step 2:** Run the tests. Expected: FAIL.
- [ ] **Step 3:** Implement. `Serve` wraps `Handler()` with `http.Server` on `Addr()`.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: viewer daemon with streamed frames`.

### Task 10: Viewer page

**Files:**
- Create: `internal/viewer/static/index.html`, `viewer.css`, `viewer.js`, `tree.js`

**Interfaces:**
- Produces: `tree.js` (an ES module). The server's scoreboard copies it verbatim, so it has no viewer-specific imports:
  - `export async function loadArt(baseUrl, size)` returns `{ background: HTMLImageElement | null, bulbs: [{x,y,r}] }`. Sizes without assets get `background: null`, and bulbs come from the render layout (formula copied from Global Constraints).
  - `export function drawTree(ctx, art, { states, time = 0, ringed = [] })` draws the art scaled to the canvas:
    - `states` uses `RGBWK`, and `.` = undrawn (drawn as a dim socket).
    - The background is `#0A0E14` when `art.background` is null.
    - Bulb 0 is drawn as a five-point star.
    - `time` in ms drives twinkle, and `0` means static.
    - `ringed` is a list of answer indices to circle.

Page behaviour (spec "Viewer" table):
- **Cards:** one per tree.
  - **`run`:** one card, with the puzzle, the score summary, and no expected tree.
  - **`test`:** one card per example, with the expected tree beside the actual tree. Wrong bulbs are ringed on the actual tree, and the card has a PASS/FAIL badge.
  - **`show`:** one card, with the show length.
- **Animation:**
  - Frames play in index order as they arrive. A gap waits for its frame until `done`, then is skipped.
  - Interval: `IntervalMs`, which is 100 for all kinds.
  - A frame with `SkippedBefore > 0` is preceded by a 1.5 s caption, `999,999,900 passes later`, formatted with `toLocaleString('en-GB')`. `skippedAtEnd` gets the same caption before the final tree.
  - After the last frame, the final `Answer` tree twinkles.
- **Startup:** loads `/latest`, then listens to `/events`. A new publication replaces the content.

- [ ] **Step 1:** Add `TestEmbeddedAssets` in `internal/viewer`. `GET /static/tree.js` returns `200`, and `/` contains `<script type="module" src="/static/viewer.js">`. Run it. Expected: FAIL.
- [ ] **Step 2:** Implement the page. Run the test. Expected: PASS.
- [ ] **Step 3: Manual check.** Start the daemon with `go run ./cmd/hth viewer`. Post a publication and frames with `curl -H "X-HTH: 1"`. Check:
  - Frames animate in order.
  - The caption shows for skips.
  - Test cards show ringed wrong bulbs.
  - The 48-row tree reads clearly.
  - The page works in Chrome, Firefox, and Safari or Edge.
- [ ] **Step 4:** Commit: `feat: animated viewer page`.

### Task 11: Animations for `hth run` and `hth test`

**Files:**
- Create: `internal/animate/animate.go`, `internal/animate/animate_test.go`
- Modify: `cmd/hth/main.go`, `cmd/hth/run_test.go`

**Interfaces:**
- Consumes: `script.Sample`, `program.Run`, `viewer.Open`/`Session`, `art.WriteGIF`.
- Produces: `func Animate(ctx context.Context, argv []string, plan script.Plan, size int, final string, card string, sess *viewer.Session, gifPath string) (frames []string, err error)`. `sess` may be nil.
  - It runs every frame script through `program.Run` with `io.Discard` stderr, using 8 workers.
  - Each answer that passes `tree.Validate` is sent to `sess.Frame` as soon as it arrives. A failed or timed-out run is skipped.
  - When all runs end, it calls `sess.Done` and writes the GIF from the frames that succeeded, in plan order. A skipped frame's `SkippedBefore` carries to the next kept frame.
  - It returns the kept frame lines in plan order (without the final tree).
- `hth run`: after the score is printed, it opens the viewer and prints `Animating N frames…`. It calls `Animate` with `hth-run-<puzzle>.gif`, then prints `Saved hth-run-3.2.gif`. A viewer error prints `hth: viewer: <err>` and continues without one. A GIF error prints `hth: gif: <err>`. Neither changes the exit code.
- `hth test`: it opens one publication with a card per example. After the PASS/FAIL report, it animates each example in turn into `hth-test-<name>.gif`.
- Ctrl-C during animation stops the workers. For `hth run`, the submission has already happened.

- [ ] **Step 1: Write failing tests:**
  - `TestAnimateOrdersFramesAndWritesGIF`: the fake `count` on a 10-step script. The returned frames are 10 lines whose `R` sits at indices 1–10 in order, and the GIF has 11 images.
  - `TestAnimateSkipsSlowFrames`: the fake `slow-after 5` on a 10-step script, with `program.Timeout` lowered to 200 ms. The result is the frames whose scripts have at most 5 lines, and the call returns in under 3 s.
  - `TestAnimateRunsInParallel`: 16 frames of the fake `sleep-ms 300` finish in under 1.5 s.
  - In `cmd/hth`:
    - `TestRunSubmitsBeforeAnimating`: the fake API records the submission time. The fake `log` records each run's start. Only the first run (the real one) starts before the submission.
    - `TestRunSucceedsWhenViewerFails`: `HTH_VIEWER_PORT` points at an HTML server. Exit 0, stderr has `hth: viewer:`, and `hth-run-1.1.gif` exists in the temp working folder.
- [ ] **Step 2:** Run the tests. Expected: FAIL.
- [ ] **Step 3:** Implement with a buffered job channel and `sync.WaitGroup`.
- [ ] **Step 4:** Run the tests. Expected: PASS. Manual check: `go run ./cmd/hth test examples -- <fake count>` opens one tab. A second run updates the same tab, and the GIFs appear in the folder.
- [ ] **Step 5:** Commit: `feat: emulator-built animations for run and test`.

### Task 12: `hth show`

**Files:**
- Modify: `cmd/hth/main.go`
- Create: `cmd/hth/show_test.go`

**Interfaces:**
- Consumes: `api.Client.ShowPermission`/`UploadShow`, `script.Parse`/`Every`, `animate.Animate`.
- Steps for `hth show <file> -- <command…>`:
  1. Parse the file. `SIZE` must be 16, else `light shows use SIZE 16`. `Steps()` must be ≤ 500, else `show runs N steps; the limit is 500`.
  2. Call `ShowPermission`. A `403` prints the server's Santa message and exits 1. Nothing runs before this check passes.
  3. Run the full script through `program.Run` and validate the answer.
  4. Animate with `script.Every`, writing `hth-show-<base>.gif`.
  5. Upload the frames that `Animate` returned, plus the final tree as the last frame.
  6. Print `Show uploaded: N frames (about S seconds on the big screen)`, where `S = N / 10` rounded up (shows play at 10 fps).

  If any frame run failed, it still uploads, and prints `K steps were skipped because the program failed on them` first.

- [ ] **Step 1: Write failing tests** with a fake API:
  - `TestShowRefusedBeforeSanta`: the permission check gets `403`. Exit 1, the message has `Santa comes first`, and the program never ran (the fake `log` file stays empty).
  - `TestShowTooLong`: a 501-step script is refused before any network call.
  - `TestShowWrongSize`: `SIZE 4` is refused.
  - `TestShowUploadsEveryStep`: a 12-step script gives an upload of 13 frames (12 steps plus the final tree), all 136 chars, and the GIF is written.
- [ ] **Step 2:** Run the tests. Expected: FAIL.
- [ ] **Step 3:** Implement.
- [ ] **Step 4:** Run the tests. Expected: PASS.
- [ ] **Step 5:** Commit: `feat: hth show for light shows`.

### Task 13: Starter stubs

**Files:**
- Create: `stubs/python/emulator.py`, `stubs/javascript/emulator.js`, `stubs/java/Emulator.java`, `stubs/csharp/Emulator.csproj` and `stubs/csharp/Program.cs`, `stubs/go/main.go`, `stubs/README.md` (the `hth test examples -- …` command per language), `stubs/stubs_test.go`

Each stub reads all of stdin, takes `n` from the first non-blank line that is not a comment (`SIZE n`), and prints `n(n+1)/2` `K`s and a newline. Each has a comment saying where to start writing the emulator. No other logic.

- [ ] **Step 1: Write the failing test.** `TestStubs` is table-driven with these commands:
  - `python3 stubs/python/emulator.py`
  - `node stubs/javascript/emulator.js`
  - `java stubs/java/Emulator.java`
  - `dotnet run --project stubs/csharp`
  - `go run ./stubs/go`

  Each gets `SIZE 16` with a leading comment line, and the output must be exactly 136 `K` + `\n`. A test is skipped with `t.Skip` when its runtime is missing from `PATH`.
- [ ] **Step 2:** Run `go test ./stubs`. Expected: FAIL. Write the stubs. Run again. Expected: PASS for every installed runtime.
- [ ] **Step 3:** Commit: `feat: starter stubs in five languages`.

### Task 14: Release builds and participant README

**Files:**
- Create:
  - `.github/workflows/ci.yml`: on push, run `go vet` and `go test ./...` on ubuntu, macos, and windows. Set up Python, Node, Java, and .NET so the stub tests run.
  - `.github/workflows/release.yml`: on tag `v*`, build the five targets with `CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${tag}"`. Zip each as `hth-<os>-<arch>.zip` containing `hth[.exe]`, `examples/`, `stubs/`, and `README.md`, and attach the zips to the GitHub release.
- Modify: `README.md`

- [ ] **Step 1:** Write `README.md` for participants:
  - Install: download the zip and put `hth` on `PATH`. On macOS, run `xattr -d com.apple.quarantine hth` because the binary is unsigned.
  - Commands: `hth login`, `hth test examples -- <cmd>`, `hth run <puzzle> -- <cmd>`, and `hth show <file> -- <cmd>`.
  - Scoring and unlocks in one paragraph: 80 % unlocks the next puzzle, and the points rule (100 for perfect, `40 + 2 × (pct − 80)` from 80 % to 99 %, 0 below).
  - The program contract, from Global Constraints.
  - How animations work: up to 120 extra runs on cut-down scripts after submission, so a slow emulator animates slowly but never affects the score. The GIF names.
  - Light shows: unlocked by a perfect Secret Santa (7.1), 500-step limit, `SIZE 16`, latest show per team kept, not scored.
  - Troubleshooting: the port variable, the Windows wrapper note from Task 3, and what `403` and `410` mean.
  - A link to `docs/spec.md`.
- [ ] **Step 2:** Push a test tag to a fork, or run the release job with `workflow_dispatch`. Expected: five zips. On each OS available, unzip and run `hth --version` and `hth test examples -- python3 stubs/python/emulator.py`. Expected: the version prints and `0/8 passed` (the stub prints all `K`).
- [ ] **Step 3:** Commit: `ci: cross-platform release builds and participant readme`.
