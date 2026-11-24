import sys

script = sys.stdin.read()

n = 0
for line in script.splitlines():
    line = line.strip()
    if line == "" or line.startswith("#"):
        continue
    n = int(line.split()[1])
    break

# Start here: replace this with your emulator. Read the script above, then
# work out the colour of every bulb. Print debugging to stderr, never stdout.
print("K" * (n * (n + 1) // 2))
