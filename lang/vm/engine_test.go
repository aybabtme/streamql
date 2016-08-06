package vm

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aybabtme/streamql/lang/ast"
	"github.com/aybabtme/streamql/lang/parser"
	"github.com/aybabtme/streamql/lang/vm/msg"
)

func benchEngine(b *testing.B, query string, makeVM func(*ast.FilterStmt) Engine) {

	q, err := parser.NewParser(strings.NewReader(query)).Parse()
	if err != nil {
		b.Fatal(err)
	}
	f := q.Filters[0]

	vm := makeVM(f)
	msgs := loadMessages(b)

	start := time.Now()
	b.ResetTimer()
	vm.Filter(
		ArraySource(msgs),
		func(_ msg.Message) bool { return true },
	)
	b.StopTimer()

	kmsg := float64(len(msgs)) / 1e6

	b.Logf("%.2fM msg/s", kmsg/time.Since(start).Seconds())
}

func loadMessages(b *testing.B) []msg.Message {
	var msgs []msg.Message
	for i := 0; i < 1e5; i++ {
		msgs = append(msgs, msg.Naive(map[string]interface{}{
			"aString": "lol",
			"aNumber": 42,
			"anArray": []interface{}{
				"lol",
				42,
				map[string]interface{}{
					"aString": "lol",
					"aNumber": 42,
					"anArray": []interface{}{
						"lol",
						42,
						map[string]interface{}{},
						[]interface{}{},
					},
					"anObject": map[string]interface{}{},
				},
				[]interface{}{},
			},
			"anObject": map[string]interface{}{
				"aString": "lol",
				"aNumber": 42,
				"anArray": []interface{}{
					"lol",
					42,
					map[string]interface{}{},
					[]interface{}{},
				},
				"anObject": map[string]interface{}{},
			},
		}))
	}
	return msgs
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
