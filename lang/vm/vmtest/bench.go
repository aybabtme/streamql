package vmtest

import (
	"testing"

	"strings"

	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/grammar"
	"github.com/aybabtme/streamql/lang/msg"
	"github.com/aybabtme/streamql/lang/vm"
)

func Bench(b *testing.B, mkBuild func() msg.Builder, mkVM func(*ast.AST, *vm.Options) vm.VM) {

	bd := mkBuild()

	tests := []struct {
		name    string
		query   string
		samples []msg.Msg
	}{
		{"passthru_bool", `.`, list(
			mustBool(bd, true),
		)},

		{"passthru_obj", `.`, list(
			mustObject(bd, map[string]msg.Msg{}),
		)},

		{"passthru_pipe_bool", `. | .`, list(
			mustBool(bd, true),
		)},

		{"passthru_pipe_obj", `. | .`, list(
			mustObject(bd, map[string]msg.Msg{}),
		)},

		{"string regexp", `regexp(., "a")`, list(
			mustString(bd, "aaaa"),
		)},

		{"string contains", `contains(., "a")`, list(
			mustString(bd, "aaaa"),
		)},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {

			query := tt.query
			n := b.N

			var input []msg.Msg
			for i := 0; i < n; i++ {
				input = append(input, tt.samples[i%len(tt.samples)])
			}

			tree, err := grammar.Parse(strings.NewReader(query))
			if err != nil {
				b.Fatal(err)
			}

			vm := mkVM(tree, &vm.Options{})

			b.ReportAllocs()
			b.ResetTimer()

			err = vm.Run(mkBuild(), ArraySource(input), func(m msg.Msg) error { return nil })
			if err != nil {
				b.Fatal(err)
			}

		})
	}

}
