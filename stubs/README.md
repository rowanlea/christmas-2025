# Starter stubs

Each stub reads the script from stdin and prints a tree of all-off bulbs (`K`). Start writing your emulator where the comment says so. Run it against the examples from the repo root:

- Python (`stubs/python/emulator.py`): `hth test examples -- python3 stubs/python/emulator.py`
- JavaScript (`stubs/javascript/emulator.js`): `hth test examples -- node stubs/javascript/emulator.js`
- Java (`stubs/java/Emulator.java`): `hth test examples -- java stubs/java/Emulator.java`
- C# (`stubs/csharp/Program.cs`): build first with `dotnet build stubs/csharp`, then `hth test examples -- dotnet run --no-build --project stubs/csharp`
- Go (`stubs/go/main.go`): `hth test examples -- go run stubs/go/main.go`

## C# troubleshooting

`dotnet run` builds your code first and prints compiler warnings on stdout, which `hth` reads as extra answer lines. That is why the command above uses `--no-build`. Run `dotnet build stubs/csharp` again after every edit. If `hth` says "expected 1 answer line, got 2" after an edit, build output went to stdout: rebuild, then run again.
