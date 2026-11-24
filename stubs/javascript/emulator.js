const script = require("fs").readFileSync(0, "utf8");

let n = 0;
for (const raw of script.split("\n")) {
  const line = raw.trim();
  if (line === "" || line.startsWith("#")) continue;
  n = parseInt(line.split(/\s+/)[1], 10);
  break;
}

// Start here: replace this with your emulator. Read the script above, then
// work out the colour of every bulb. Print debugging to stderr, never stdout.
console.log("K".repeat((n * (n + 1)) / 2));
