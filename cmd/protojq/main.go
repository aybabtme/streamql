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

	"github.com/aybabtme/streamql/lang/grammar"
	"github.com/aybabtme/streamql/lang/msg"
	"github.com/aybabtme/streamql/lang/msg/gomsg"
	"github.com/aybabtme/streamql/lang/msg/msgutil"
	"github.com/aybabtme/streamql/lang/vm"
	"github.com/aybabtme/streamql/lang/vm/astvm"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(0)
	log.SetPrefix("protojq: ")
	flag.Parse()
	query := strings.Join(flag.Args(), " ")

	tree, err := grammar.Parse(strings.NewReader(query))
	if err != nil {
		log.Fatalf("invalid query: %v", err)
	}

	in := os.Stdin
	out := os.Stdout

	builder := gomsg.Build()
	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)

	var wg sync.WaitGroup
	wg.Add(2)
	defer wg.Wait()

	inc := make(chan msg.Msg, runtime.NumCPU())
	go func() {
		defer wg.Done()
		for {
			var v interface{}
			switch err := dec.Decode(&v); err {
			case io.EOF:
				close(inc)
				return
			case nil:
				msgv, err := msgutil.FromGo(builder, v)
				if err != nil {
					panic(err)
				}
				inc <- msgv
			default:
				log.Fatalf("invalid input: %v", err)
			}
		}
	}()

	outc := make(chan msg.Msg, runtime.NumCPU())
	go func() {
		defer wg.Done()
		for m := range outc {
			v, err := msgutil.Reveal(m)
			if err != nil {
				panic(err)
			}
			if err := enc.Encode(v); err != nil {
				log.Fatalf("invalid output: %v", err)
			}
		}
	}()

	engine := astvm.Interpreter(tree, &vm.Options{})
	engine.Run(
		builder,
		func() (msg.Msg, bool, error) { msg, more := <-inc; return msg, more, nil },
		func(m msg.Msg) error { outc <- m; return nil },
	)
	close(outc)
}
