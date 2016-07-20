package vm

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/aybabtme/streamql/lang/parser"
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
		vm := ASTInterpreter(query)
		var toFilter []Message
		for _, input := range tt.input {
			msg, err := ReadJSONMessage([]byte(input))
			if err != nil {
				t.Fatal(err)
			}
			toFilter = append(toFilter, msg)
		}
		out := vm.Filter(toFilter)
		var got [][]string
		for _, outelems := range out {
			var gotelems []string
			for _, elem := range outelems {
				str := string(WriteJSONMessage(elem.(*NaiveMessage)))
				gotelems = append(gotelems, str)
			}
			got = append(got, gotelems)
		}
		if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("want=%#v", tt.want)
			t.Fatalf(" got=%#v", got)
		}
	}
}

func ReadJSONMessage(data []byte) (Message, error) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &NaiveMessage{v: v}, nil
}

func WriteJSONMessage(msg *NaiveMessage) []byte {
	data, _ := json.Marshal(msg.v)
	return data
}

type NaiveMessage struct {
	v interface{}
}

func (msg *NaiveMessage) Member(k string) (Message, bool) {
	m, ok := msg.v.(map[string]interface{})
	if !ok {
		return nil, false
	}
	child, ok := m[k]
	if !ok {
		return nil, false
	}
	return &NaiveMessage{v: child}, true
}

func (msg *NaiveMessage) Each() ([]Message, bool) {
	elems, ok := msg.v.([]interface{})
	if !ok {
		return nil, false
	}
	var out []Message
	for _, el := range elems {
		out = append(out, &NaiveMessage{v: el})
	}
	return out, true
}

func (msg *NaiveMessage) Range(from, to int) ([]Message, bool) {
	elems, ok := msg.v.([]interface{})
	if !ok {
		return nil, false
	}
	if from >= len(elems) || to > len(elems) {
		return nil, false
	}
	var out []Message
	for _, el := range elems[from:to] {
		out = append(out, &NaiveMessage{v: el})
	}
	return out, true
}

func (msg *NaiveMessage) Index(i int) (Message, bool) {
	elems, ok := msg.v.([]interface{})
	if !ok {
		return nil, false
	}
	if i >= len(elems) {
		return nil, false
	}
	return &NaiveMessage{v: elems[i]}, true
}
