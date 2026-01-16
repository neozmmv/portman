package rules

import (
	"os"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()

	b, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return string(b)
}

func TestOpenAddsRuleInOracleRulesFile(t *testing.T) {
	in := readFixture(t, "rules.v4")

	out, changed, err := Open(in, 3308, "tcp")
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true")
	}

	want := RuleLine(3308, TCP)
	if !strings.Contains(out, want) {
		t.Fatalf("expected output to contain:\n%s", want)
	}

	// rule before COMMIT
	idxRule := strings.Index(out, want)
	idxCommit := strings.Index(out, "\nCOMMIT\n")
	if idxRule == -1 || idxCommit == -1 || idxRule > idxCommit {
		t.Fatalf("expected rule to be before COMMIT")
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	in := readFixture(t, "rules.v4")

	out1, changed1, err := Open(in, 3308, "tcp")
	if err != nil {
		t.Fatalf("Open 1 error: %v", err)
	}
	if !changed1 {
		t.Fatalf("expected changed1=true")
	}

	out2, changed2, err := Open(out1, 3308, "tcp")
	if err != nil {
		t.Fatalf("Open 2 error: %v", err)
	}
	if changed2 {
		t.Fatalf("expected changed2=false")
	}
	if out2 != out1 {
		t.Fatalf("expected content unchanged on second open")
	}
}

func TestCloseRemovesRule(t *testing.T) {
	in := readFixture(t, "rules.v4")

	with, changed, err := Open(in, 3308, "tcp")
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}
	if !changed {
		t.Fatalf("expected changed=true from Open")
	}

	out, changedClose, err := Close(with, 3308, "tcp")
	if err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if !changedClose {
		t.Fatalf("expected changed=true from Close")
	}

	rule := RuleLine(3308, TCP)
	if strings.Contains(out, rule) {
		t.Fatalf("expected rule to be removed:\n%s", rule)
	}
}

func TestStatusWorks(t *testing.T) {
	in := readFixture(t, "rules.v4")

	st, err := Status(in, 8000, "tcp/udp")
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if !st[TCP] || !st[UDP] {
		t.Fatalf("expected 8000 tcp and udp to be open in oracle rules file")
	}

	st2, err := Status(in, 3308, "tcp")
	if err != nil {
		t.Fatalf("Status error: %v", err)
	}
	if st2[TCP] {
		t.Fatalf("expected 3308 tcp to be closed initially")
	}
}

func TestListReturnsPortsFromPortmanBlock(t *testing.T) {
	in := readFixture(t, "rules.v4")

	items, err := List(in)
	if err != nil {
		t.Fatalf("List error: %v", err)
	}

	want := []PortRule{
		{Port: 3306, Proto: TCP},
		{Port: 3307, Proto: TCP},
		{Port: 3307, Proto: UDP},
	}

	if len(items) != len(want) {
		t.Fatalf("expected %d items, got %d: %#v", len(want), len(items), items)
	}
	for i := range want {
		if items[i] != want[i] {
			t.Fatalf("at %d: expected %#v, got %#v", i, want[i], items[i])
		}
	}
}
