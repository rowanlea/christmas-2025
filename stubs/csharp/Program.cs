var script = Console.In.ReadToEnd();

var n = 0;
foreach (var raw in script.Split('\n'))
{
    var line = raw.Trim();
    if (line.Length == 0 || line.StartsWith('#')) continue;
    n = int.Parse(line.Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries)[1]);
    break;
}

// Start here: replace this with your emulator. Read the script above, then
// work out the colour of every bulb. Print debugging to stderr, never stdout.
Console.WriteLine(new string('K', n * (n + 1) / 2));
