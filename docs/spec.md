# Hack the Halls: LSCC Christmas kata

Nov 19, 2026 · @Edward Farrow

## Summary

Teams build an emulator for a Christmas lights controller. A harness we supply pulls challenges from the online system, runs them through the team's code, renders the result as a PNG, and submits it for scoring against the system's own render.

The tree is a triangle of 136 bulbs, each red, green, blue, white, or off. The controller has eleven instructions, three of them classic light patterns, and teams get all of them from the start. What makes it hard is the way the challenges order and combine them, in scripts far too long to work through by hand.

Scores count correct bulbs, so an emulator that's nearly right still earns points. Any language works, and AI tools are allowed.

## The tree

The tree is a triangle of bulbs with 16 rows. Row r, counting from 0 at the star, holds r + 1 bulbs, so the tree has 136 in total. Position p counts from 0 at the left of each row, so 0 ≤ p ≤ r.

```
        (0,0)
     (1,0) (1,1)
  (2,0) (2,1) (2,2)
```

A bulb's neighbours are the two beside it in its row, two above it, and two below it. Bulbs on an edge have fewer.

| Direction | Position |
| --- | --- |
| Left | (r, p − 1) |
| Right | (r, p + 1) |
| Up-left | (r − 1, p − 1) |
| Up-right | (r − 1, p) |
| Down-left | (r + 1, p) |
| Down-right | (r + 1, p + 1) |

Each bulb is in one of five states. Every bulb starts Off.

| Char | State |
| --- | --- |
| R | Red |
| G | Green |
| B | Blue |
| W | White |
| K | Off |

## Input and output

The team's program reads a challenge script on stdin and prints one answer line on stdout. The harness handles everything else.

Script rules:

- The first line is `SIZE n`, the number of rows. Rounds 1 to 6 use 16, and the worked example uses 3.
- Each line holds one instruction. Tokens are uppercase and separated by single spaces. Blank lines and lines starting with `#` are ignored.
- `REPEAT` opens a block and `END` closes it. Blocks can nest.

The answer line holds one state character per bulb, read row by row from the star and left to right within each row. At 16 rows that is 136 characters, and in general it is n(n + 1)/2, with no separators or trailing whitespace.

| Contract |  |
| --- | --- |
| Time limit | 10 seconds per run |
| Errors | Anything on stderr is shown to the team. A non-zero exit submits nothing |

## Worked example

On a three-row tree, this script produces the answer `RRGRGW`.

```
SIZE 3
FILL G
SET 0 0 W
ROW 2 R
ROTATE
```

Before `ROTATE`, the star is white and the bottom row is red, which reads `WGGRRR`.

```
    W
   G G
  R R R
```

Rotating 120° clockwise carries the star to the bottom-right corner, and the bottom row becomes the left edge.

```
    R
   R G
  R G W
```

## Gameplay and scoring

Each puzzle is scored on accuracy, the share of bulbs a team's render gets right. Reaching 80% unlocks the next puzzle, and puzzles unlock one at a time, in order.

Points start at 80% and rise with accuracy. 80% is worth 40 points, and each percentage point above that adds 2, so 99% scores 78. A perfect tree scores 100. The extra 22 for the last bulb is deliberate, because an exact emulator is the point of the kata. Anything under 80% scores nothing, since a wrong rotation or mirror can reach 60% by accident.

A team's best submission on each puzzle counts.

A team's loop on each puzzle:

1. Run the harness against the puzzle with the team's program.
2. The harness fetches a freshly generated script, which must be submitted within five minutes.
3. The team's program produces the answer line, and the harness renders and submits it.
4. The system returns the score. The team debugs locally against the worked example, then fetches again.

Every fetch is a new script, so a resubmission tests the emulator rather than a remembered answer. Generated scripts run to thousands of lines, with nested and billion-pass loops. Nobody traces that by hand inside five minutes, so working code is the only way through.

## Challenge harness

The harness is a single static binary, `hth`, for macOS, Windows, and Linux. It runs the team's program as a subprocess, so any language that reads stdin and writes stdout works with no SDK.

```
hth login team-07
hth run 3.2 -- python emulator.py
hth run 3.2 -- java -jar emulator.jar
hth test examples/ -- ./emulator
```

`hth run` takes a puzzle number and the command that starts the team's program. It fetches a challenge and pipes the script in. It then checks the answer line's length against the size line and its characters, renders the PNG, submits it, and prints the score.

`hth test` is the inner TDD loop. It runs the worked example locally without submitting, compares the answer with the expected one, and shows the expected and actual trees side by side in the viewer, with the wrong bulbs ringed.

We also supply starter stubs in common languages. Each one reads stdin and prints 136 Ks, which is enough to prove the pipe works before anyone writes real logic.

### Viewer

Every run shows its result as an animated tree in a viewer tab. On the first run, `hth` starts a local viewer page and opens it in the default browser. Each later run updates that tab rather than opening a new one. We use a browser tab rather than a native pop-up because it animates the same way on every operating system.

The bulbs sit on a drawn Christmas tree, deep green with layered branches and a trunk in a stand, on a plain dark background. The star bulb at (0, 0) is drawn star-shaped and takes that bulb's state. The decoration uses no palette colours, so it never reads as a bulb state.

The viewer canvas is precomputed and embedded in `hth`, so nothing is drawn from scratch at run time. Viewer trees are drawn on one of two grids: trees of up to 16 rows on the 16-row grid, and trees of 17 to 48 rows on the 48-row grid, where a smaller tree lights the top rows of its grid. For each grid, the build produces:

- a background image with the tree, trunk, and stand on a plain dark background, drawn at the canvas size
- a table of bulb centres and radii, in answer-line order

The viewer lays the background down once, then draws each frame by colouring bulbs from the table. Twinkle runs on top in the browser. Because the background is a fixed image, it can be hand-finished as pixel art rather than generated. The scoreboard page uses the same assets, so trees look identical on a laptop and on the big screen.

A script with any other size falls back to a plain dark background with bulbs drawn from the render spec, so nothing breaks if a team experiments.

| Run | Animation | Also shown |
| --- | --- | --- |
| `hth run` | The tree evolving through the script, then the final tree twinkling | Puzzle number and score |
| `hth test` | The same, for the worked example | Expected tree beside it, with wrong bulbs ringed |
| `hth show` | Every step of the show | Show length |

Frames come from the team's own emulator, so a broken emulator animates its own mistakes. The harness walks through the script's execution and picks up to 120 points spread across it. For each point it writes a shortened script that stops there, cutting any loop it's inside down to the passes so far, and runs the team's program on it. Each answer line becomes one frame.

A billion-pass loop shows its first passes, then a caption such as "999,999,880 passes later" before the final tree.

These runs start after the answer is submitted, so they never eat into the five-minute expiry. Up to eight run in parallel, and the viewer adds frames as they arrive. A run that times out is skipped.

Live puzzles never show the expected tree, because that would give the answer away.

Every run also saves its animation as a GIF in the team's working folder. The harness writes it from the same frames the viewer plays. The PNG that gets submitted is always the static final render, so the animation has no effect on scoring.

## Scoreboard page

The scoring system's webpage runs on the big screen all evening, with the scoreboard on the left and a live feed of submissions on the right.

Each submission arrives in the feed as a tile showing the team, the puzzle, the score out of the bulb count, and the tree. The page redraws the tree from the bulb states the scorer sampled, using the same decorated style as the harness viewer. The newest tile slides in at the top, and the feed keeps the latest eight.

Showing other teams' trees gives nothing away, because every fetch generates a different script. A correct tree only answers that team's own challenge.

Some submissions get a banner across the feed:

| Moment | Banner |
| --- | --- |
| First full-marks solve of a puzzle | The team and puzzle, with the tree enlarged for a few seconds |
| First team to finish a round | The team and round |
| Any Secret Santa solve | The team, with its tree enlarged |

The page needs live updates, through server-sent events or polling every few seconds.

The scoreboard also has a challenge ladder. It lists every level, with 1.1 at the bottom, 7.1 near the top and "Found Santa" above it. Each team's name appears once, beside the first level on which it has not yet reached 80 %, with a fill showing how far it has got there. The server works out the ladder, so the page only draws it.

Light shows get their own panel under the feed, playing each team's latest show in turn. A team's first show also gets a banner.

The host hides the standings on the big screen near the end. The final standings are revealed at prize-giving.

## Instructions

The controller has eleven instructions, all in the spec from the start: eight that paint and transform, and three classic light patterns. C stands for a state character.

| Instruction | Effect |
| --- | --- |
| `SET r p C` | Paint bulb (r, p) |
| `ROW r C` | Paint every bulb in row r |
| `FILL C` | Paint every bulb. `FILL K` clears the tree |
| `ROTATE` | Rotate the tree 120° clockwise. The star's state moves to the bottom-right corner, that corner's to the bottom-left, and the bottom-left's to the star |
| `MIRROR` | Reflect left to right, so the state at (r, p) moves to (r, r − p) |
| `GROW C` | Every Off bulb with at least one C neighbour becomes C. All bulbs update at once, from the tree as it stood before the instruction |
| `CYCLE` | Every lit bulb advances R → G → B → W → R. Off bulbs stay Off |
| `SEQUENCE C` | Switch on the first Off bulb along the wiring as C. Does nothing if no bulb is Off |
| `CHASE` | Every bulb's state moves one place along the wiring. The last bulb's state moves to the star |
| `FALL` | Every bulb's state drops straight down to (r + 2, p + 1). States that fall below the bottom row are lost, and rows 0 and 1 switch off |
| `REPEAT n ... END` | Run the block n times, where 0 ≤ n ≤ 1,000,000,000. Blocks nest |

### Wiring

The patterns follow the wiring, which snakes down the tree from the star. Row 1 runs left to right, row 2 runs right to left, and the direction alternates from there. The wiring is fixed to positions, so `ROTATE` and `MIRROR` move states but never rewire the tree.

The wiring order is not the answer-line order. On a four-row tree it runs:

```
(0,0) (1,0) (1,1) (2,2) (2,1) (2,0) (3,0) (3,1) (3,2) (3,3)
```

`FALL` moves two rows because that is straight down in the triangle. On a four-row tree with the star red, (1, 0) green, and (1, 1) blue, one `FALL` gives `KKKKRKKGBK`.

## Rounds

The evening has 23 puzzles in seven rounds, numbered by round and puzzle. Every puzzle uses the same eleven instructions, and each fetch generates a fresh script from the puzzle's template.

### Round 1: Lights On

Short scripts of up to 40 lines, which pairs can score before the emulator is finished.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 1.1 | Paint by Numbers | 30 random `SET` lines | Answer-line order |
| 1.2 | Fill Your Boots | `FILL`, then `ROW` and `SET` lines overwriting it | Row lengths and overwrite order |
| 1.3 | Turn of Events | 20 `SET` lines, then one `ROTATE` | Which corner the star goes to |
| 1.4 | Looking Glass | 20 `SET` lines, then one `MIRROR` | The mirror formula, especially on the middle bulb of rows with an odd number of bulbs |

### Round 2: Order, Order

Transform chains of up to 60 lines, where painting and order both matter.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 2.1 | Turn, Then Reflect | A pattern, then `ROTATE` and `MIRROR` | A working pair of transforms |
| 2.2 | Second Thoughts | The same template with `MIRROR` first | Assuming transforms commute. Teams that pass 2.1 and fail 2.2 learn why |
| 2.3 | Round in Circles | A pattern, then 30 random transforms | A single wrong step anywhere in the chain |
| 2.4 | Three Sides to Every Story | `ROW 15` in a different colour after each of three rotations, plus `SET` lines between | Painting in a rotated frame, and which edge owns each corner |

```
SIZE 16
ROW 15 R
ROTATE
ROW 15 G
SET 3 1 W
ROTATE
ROW 15 B
MIRROR
```

### Round 3: Spreading Cheer

`GROW` and `CYCLE` mixed with transforms.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 3.1 | Growing Pains | A few seeds, then `GROW` up to six times in one colour | In-place updates, which flood a row in a single pass |
| 3.2 | Turf War | Seeds in two colours, with `GROW R` and `GROW G` alternating | Contested bulbs, which go to whichever colour grows first |
| 3.3 | Cornered | Seeds on the corners, with `GROW` and `ROTATE` interleaved | Neighbour bounds at corners and along the bottom row |
| 3.4 | Change of Heart | Seeds, then `REPEAT` over `CYCLE` and `GROW R` | The colour that grows changes every pass, and Off bulbs never cycle |

```
SIZE 16
SET 0 0 R
SET 15 0 G
SET 15 15 B
REPEAT 9
  CYCLE
  GROW R
  ROTATE
END
```

### Round 4: Light Show

The three patterns, which all follow the wiring rather than the answer-line order.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 4.1 | Switched On | Seeds, then `SEQUENCE` up to 60 times, changing colour between runs | Following the answer-line order instead of the wiring |
| 4.2 | Chasing Tail | A pattern, then `CHASE` up to 200 times | The change of direction on each row, and the wrap from the last bulb back to the star |
| 4.3 | Falling Star | Painted rows, then `FALL` repeated, with `SET` lines between | Dropping one row instead of two, and states lost off the bottom |
| 4.4 | Light Show | All three patterns mixed with `ROTATE` and `MIRROR` | Treating the wiring as attached to states rather than positions |

```
SIZE 16
ROW 0 W
REPEAT 8
  FALL
  SEQUENCE G
END
ROTATE
REPEAT 40
  CHASE
END
```

### Round 5: Long Haul

Scripts too long to trace by hand. Each executes at most 200,000 instructions, so a straightforward emulator in any language finishes inside the time limit without cycle detection.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 5.1 | The Long List | 5,000 lines with no loops, using every instruction | Small errors that accumulate |
| 5.2 | Wheels Within Wheels | `REPEAT` nested four deep, with counts up to 20 | Matching each `END` to its `REPEAT` |
| 5.3 | Nothing Doing | `REPEAT 0`, `REPEAT 1`, empty blocks, and comments and blank lines inside blocks | Parser edge cases |

### Round 6: Silent Night

Loops of up to 1,000,000,000 passes, which only finish in time with cycle detection.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 6.1 | Round the Clock | A pattern, then a billion passes of `ROTATE` and `CYCLE` | Brute force. The tree repeats every 12 passes from the start |
| 6.2 | False Start | As 6.1 with `GROW` in the loop | Taking the pass count modulo 12 from the start. Growth takes dozens of passes to saturate before the repeat begins |
| 6.3 | Deep Sleep | A 1,000-pass outer loop that paints a bulb, then runs a billion-pass inner loop of `ROTATE`, `CHASE`, and `CYCLE` | Detecting cycles only in the outermost loop, and predicting periods instead of detecting them. The inner loop repeats every 372 passes |

```
SIZE 16
SET 2 1 R
SET 9 4 G
REPEAT 1000000000
  CYCLE
  GROW R
  ROTATE
END
```

### Round 7: Secret Santa

Secret Santa's correct end state is a picture: Santa in his sleigh with a sack, pulled by a red-nosed reindeer, on a green tree. The scene runs along the bottom 14 rows, where the tree is 35 to 48 bulbs wide, with snowfall above it. It opens as soon as a team reaches it.

| Puzzle | Name | Script | Catches |
| --- | --- | --- | --- |
| 7.1 | Secret Santa | A 48-row tree with 1,176 bulbs. The script paints a scrambled picture, then runs billion-pass loops of `ROTATE`, `MIRROR`, `CHASE`, and `CYCLE` that unscramble it | Any inexact instruction. One wrong step leaves a scramble instead of Santa |

The spec only promises 16 rows for Rounds 1 to 6, and the worked example uses 3 rows, so a team that read `SIZE` from the start loses nothing on the larger tree.

Every fetch varies the picture. The sleigh faces left or right, sits a few bulbs to either side, and carries a sack in a different colour. So a team that spots Santa on the feed still has to run its own script to find out which version it got.

A perfect solve puts Santa on the big screen. A near miss puts up a recognisable but broken one, and the room can see how close the team got.

### After Secret Santa: light shows

Once a team has a Secret Santa submission with every bulb correct, the system unlocks light shows for that team. Until then, `hth show` is refused with a message saying Santa comes first.

A light show is a script the team writes itself, using the same eleven instructions on a 48-row tree (the same grid as Secret Santa), and runs through its own emulator:

```
hth show my-show.txt -- python emulator.py
```

Frames are made the same way as for puzzles, except that every step is kept rather than a sample. Shows that run past 500 steps are refused.

The harness previews the show in the viewer, saves it as a GIF, and uploads the frames for the big screen. Nothing is scored.

At 10 frames a second, a 500-instruction show plays for about 50 seconds. A team can resubmit as often as it likes, and only its latest show is kept.

## Run sheet

The evening runs 18:30 to 21:30 at the Codurance office, with participants working in pairs. The harness and starter stubs go out with registration, so laptops are set up before arrival.

| Time |  |
| --- | --- |
| 18:30 | Food and arrivals |
| 18:50 | Ten-minute intro: the tree, the harness, and the worked example |
| 19:00 | Pairs form, and puzzle 1.1 opens |
| 20:45 | Host hides the standings on the big screen |
| 20:50 | Reveal of the intended approach, then two or three pairs show their designs and a few light shows play |
| 21:15 | Prizes |
| 21:30 | Pub |
