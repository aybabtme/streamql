package vm

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/parser"
	"github.com/aybabtme/streamql/lang/vm/msg"
)

func TestASTInterpreter(t *testing.T) {

	tests := []struct {
		query string
		input []string
		want  [][]string
	}{
		{
			query: ".",
			input: []string{
				"{}",
			},
			want: [][]string{
				{"{}"},
			},
		},
		{
			query: ".,.",
			input: []string{
				"{}",
			},
			want: [][]string{
				{"{}"},
				{"{}"},
			},
		},
		{
			query: ".|.",
			input: []string{
				"{}",
			},
			want: [][]string{
				{"{}"},
			},
		},
		{
			query: ".hello",
			input: []string{
				"{}",
			},
			want: [][]string{nil},
		},
		{
			query: ".hello,.world",
			input: []string{
				"{}",
			},
			want: [][]string{nil, nil},
		},
		{
			query: ".hello,.world",
			input: []string{
				`{"hello":"world"}`,
				`{"hello":"lol"}`,
			},
			want: [][]string{
				{`"world"`, `"lol"`},
				nil,
			},
		},
		{
			query: ".hello,.world",
			input: []string{
				`{"hello":"world"}`,
				`{"world":"lol"}`,
			},
			want: [][]string{
				{`"world"`},
				{`"lol"`},
			},
		},
	}

	for _, tt := range tests {
		query, err := parser.NewParser(strings.NewReader(tt.query)).Parse()
		if err != nil {
			t.Fatal(err)
		}
		var got [][]string

		for _, f := range query.Filters {
			vm := ASTInterpreter(f)
			var toFilter []msg.Message
			for _, input := range tt.input {
				msg, err := ReadJSONMessage([]byte(input))
				if err != nil {
					t.Fatal(err)
				}
				toFilter = append(toFilter, msg)
			}

			var fout []string
			vm.Filter(
				ArraySource(toFilter),
				func(m msg.Message) bool {
					str := string(WriteJSONMessage(m.(*msg.NaiveMessage)))
					fout = append(fout, str)
					return true
				},
			)
			got = append(got, fout)
		}
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("want=%#v", tt.want)
			t.Fatalf(" got=%#v", got)
		}
	}
}

func BenchmarkASTEngine_Everything(b *testing.B) {
	benchEngine(b, ".", ASTInterpreter)
}

func BenchmarkASTEngine_NoResult(b *testing.B) {
	benchEngine(b, `.unexisting`, ASTInterpreter)
}

func BenchmarkASTEngine_SimpleString(b *testing.B) {
	benchEngine(b, ".aString", ASTInterpreter)
}

func BenchmarkASTEngine_SimpleNumber(b *testing.B) {
	benchEngine(b, ".aNumber", ASTInterpreter)
}

func BenchmarkASTEngine_SimpleArray(b *testing.B) {
	benchEngine(b, ".anArray", ASTInterpreter)
}

func BenchmarkASTEngine_SimpleObject(b *testing.B) {
	benchEngine(b, ".anObject", ASTInterpreter)
}

func BenchmarkASTEngine_ExplodeArray(b *testing.B) {
	benchEngine(b, ".anArray[]", ASTInterpreter)
}

func BenchmarkASTEngine_IndexArray(b *testing.B) {
	benchEngine(b, ".anArray[0]", ASTInterpreter)
}

func BenchmarkASTEngine_RangeArray(b *testing.B) {
	benchEngine(b, ".anArray[0:2]", ASTInterpreter)
}

func BenchmarkASTEngine_ExplodeArraySelect(b *testing.B) {
	benchEngine(b, ".anArray[].aString", ASTInterpreter)
}

func BenchmarkASTEngine_IndexArraySelect(b *testing.B) {
	benchEngine(b, ".anArray[0].aString", ASTInterpreter)
}

func BenchmarkASTEngine_RangeArraySelect(b *testing.B) {
	benchEngine(b, ".anArray[0:2].aString", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectString(b *testing.B) {
	benchEngine(b, ".anObject.aString", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectNumber(b *testing.B) {
	benchEngine(b, ".anObject.aNumber", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectArray(b *testing.B) {
	benchEngine(b, ".anObject.anArray", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectObject(b *testing.B) {
	benchEngine(b, ".anObject.anObject", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectExplodeArray(b *testing.B) {
	benchEngine(b, ".anObject.anArray[]", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectIndexArray(b *testing.B) {
	benchEngine(b, ".anObject.anArray[0]", ASTInterpreter)
}

func BenchmarkASTEngine_ObjectSelectRangeArray(b *testing.B) {
	benchEngine(b, ".anObject.anArray[0:2]", ASTInterpreter)
}
