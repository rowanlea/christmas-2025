package script

import (
	"math/big"
	"os"
	"strings"
	"testing"
)

func load(t *testing.T, name string) Script {
	t.Helper()
	b, err := os.ReadFile("../../examples/" + name)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Parse(string(b))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func step(l string) Node { return Node{Line: l} }

func loop(c int64, body ...Node) Node { return Node{Loop: true, Count: c, Body: body} }

func equalNodes(a, b []Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Line != b[i].Line || a[i].Loop != b[i].Loop || a[i].Count != b[i].Count || !equalNodes(a[i].Body, b[i].Body) {
			return false
		}
	}
	return true
}

func TestParseNested(t *testing.T) {
	s := load(t, "5-long-haul.txt")
	want := []Node{
		step("SET 0 0 R"), step("ROW 3 B"),
		loop(0, step("FILL W")),
		loop(1,
			loop(2),
			loop(3, step("CYCLE"), loop(2, step("CHASE")))),
		step("SET 3 1 G"),
	}
	if s.Size != 4 || !equalNodes(s.Body, want) {
		t.Fatalf("got %+v", s)
	}
}

func TestParseTrimsIndentAndCR(t *testing.T) {
	s, err := Parse("SIZE 3\r\n\tREPEAT 2 \r\n  CYCLE\r\nEND\r\n")
	if err != nil || !equalNodes(s.Body, []Node{loop(2, step("CYCLE"))}) {
		t.Fatalf("got %+v %v", s, err)
	}
}

func TestParseErrors(t *testing.T) {
	cases := map[string]string{
		"# c\n\nSET 0 0 R\n":                   "line 3: missing SIZE",
		"SIZE 3\nEND\n":                        "line 2: unmatched END",
		"SIZE 3\nSET 0 0 R\nREPEAT 2\nCYCLE\n": "line 3: unclosed REPEAT",
		"SIZE 3\nREPEAT x\nEND\n":              "line 2: bad REPEAT count",
		"SIZE 3\nREPEAT -1\nEND\n":             "line 2: bad REPEAT count",
		"SIZE 3\nREPEAT 1000000001\nEND\n":     "line 2: bad REPEAT count",
		"SIZE 3\nREPEAT\nEND\n":                "line 2: bad REPEAT count",
	}
	for in, want := range cases {
		if _, err := Parse(in); err == nil || err.Error() != want {
			t.Errorf("Parse(%q) error %v, want %q", in, err, want)
		}
	}
	if _, err := Parse("SIZE 3\nREPEAT 1000000000\nEND\n"); err != nil {
		t.Errorf("max count rejected: %v", err)
	}
}

func TestStepsCountsArithmetically(t *testing.T) {
	s := load(t, "6-silent-night.txt")
	if got := s.Steps().String(); got != "3000000002" {
		t.Fatalf("got %s", got)
	}
	n, err := Parse("SIZE 3\nREPEAT 1000\nREPEAT 1000000000\nA\nB\nC\nEND\nEND\n")
	if err != nil {
		t.Fatal(err)
	}
	if n.Steps().Cmp(big.NewInt(3_000_000_000_000)) != 0 {
		t.Fatalf("got %s", n.Steps())
	}
	deep := "SIZE 3\n" + strings.Repeat("REPEAT 1000000000\n", 5) + "A\n" + strings.Repeat("END\n", 5)
	d, _ := Parse(deep)
	if d.Steps().String() != "1"+strings.Repeat("0", 45) {
		t.Fatalf("got %s", d.Steps())
	}
}
