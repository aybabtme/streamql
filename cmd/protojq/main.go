package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/aybabtme/streamql/lang/parser"
	"github.com/aybabtme/streamql/lang/vm"
	"github.com/aybabtme/streamql/lang/vm/msg"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
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

	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	inc := make(chan msg.Message, runtime.NumCPU())
	go func() {
		defer wg.Done()
		for {
			var v interface{}
			switch err := dec.Decode(&v); err {
			case io.EOF:
				close(inc)
				return
			case nil:
				inc <- msg.Naive(v)
			default:
				log.Fatalf("invalid input: %v", err)
			}
		}
	}()

	outc := make(chan msg.Message, runtime.NumCPU())
	go func() {
		defer wg.Done()
		for m := range outc {
			if err := enc.Encode(m.(*msg.NaiveMessage).Orig()); err != nil {
				log.Fatalf("invalid output: %v", err)
			}
		}
	}()

	engine := vm.ASTInterpreter(tree.Filters[0])
	engine.Filter(
		func() (msg.Message, bool) { msg, more := <-inc; return msg, more },
		func(m msg.Message) bool { outc <- m; return true },
	)
	close(outc)
}
