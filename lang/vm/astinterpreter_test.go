package vm

import (
	"encoding/json"
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

// Array source

type arraySource struct {
	idx  int
	msgs []msg.Message
}

func ArraySource(in []msg.Message) Source {
	return (&arraySource{msgs: in}).Next
}

func (src *arraySource) Next() (msg.Message, bool) {
	if src.idx >= len(src.msgs) {
		return nil, false
	}
	msg := src.msgs[src.idx]
	src.idx++
	return msg, true
}

// JSON

func ReadJSONMessage(data []byte) (msg.Message, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return msg.Naive(v), nil
}

func WriteJSONMessage(msg *msg.NaiveMessage) []byte {
	data, _ := json.Marshal(msg.Orig())
	return data
}
