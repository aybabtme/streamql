package main

import (
	"flag"
	"log"
	"strings"

	"github.com/aybabtme/streamql/lang/parser"
)

func main() {
	flag.Parse()
	query := strings.Join(flag.Args(), " ")

	log.Printf("query=%q", query)
	tree, err := parser.NewParser(strings.NewReader(query)).Parse()
	if err != nil {
		log.Fatalf("invalid query: %v", err)
	}
	_ = tree
}
