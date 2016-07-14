package grammar

import "testing"

func TestGrammar(t *testing.T) {
	tests := []struct {
		sentence string
		tokens   []token
	}{
		{sentence: ".", tokens: []token{thisSymbol}},
	}
	for _, tt := range tests {
		t.Logf("parsing: %s", tt.sentence)
		if err := parser(tt.tokens); err != nil {
			t.Fatalf("can't parse input: %v", err)
		}
	}
}
