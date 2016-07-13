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
		input := make(chan token, len(tt.tokens))
		for _, t := range tt.tokens {
			input <- t
		}
		close(input)
		if err := parser(input); err != nil {
			t.Fatalf("can't parse input: %v", err)
		}
	}
}
