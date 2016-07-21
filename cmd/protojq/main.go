package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aybabtme/streamql/lang/parser"
	"github.com/aybabtme/streamql/lang/vm"
	"github.com/aybabtme/streamql/lang/vm/msg"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("protojq: ")
	flag.Parse()
	query := strings.Join(flag.Args(), " ")

	tree, err := parser.NewParser(strings.NewReader(query)).Parse()
	if err != nil {
		log.Fatalf("invalid query: %v", err)
	}
	if len(tree.Filters) > 1 {
		log.Fatalf("can only accept 1 filter")
	}

	in := os.Stdin
	out := os.Stdout
	if len(tree.Filters) < 0 {
		_, err := io.Copy(in, out)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)

	engine := vm.ASTInterpreter(tree.Filters[0])
	engine.Filter(
		func() (msg.Message, bool) {
			var v interface{}
			switch err := dec.Decode(&v); err {
			case io.EOF:
				return nil, false
			case nil:
				return msg.Naive(v), true
			default:
				log.Fatalf("invalid input: %v", err)
				return nil, false
			}
		},
		func(m msg.Message) bool {
			if err := enc.Encode(m.(*msg.NaiveMessage).Orig()); err != nil {
				log.Fatalf("invalid output: %v", err)
			}
			return true
		},
	)
}
