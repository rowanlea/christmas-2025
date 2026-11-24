import java.io.IOException;
import java.nio.charset.StandardCharsets;

public class Emulator {
    public static void main(String[] args) throws IOException {
        String script = new String(System.in.readAllBytes(), StandardCharsets.UTF_8);

        int n = 0;
        for (String raw : script.split("\n")) {
            String line = raw.trim();
            if (line.isEmpty() || line.startsWith("#")) continue;
            n = Integer.parseInt(line.split("\s+")[1]);
            break;
        }

        // Start here: replace this with your emulator. Read the script above, then
        // work out the colour of every bulb. Print debugging to stderr, never stdout.
        System.out.println("K".repeat(n * (n + 1) / 2));
    }
}
