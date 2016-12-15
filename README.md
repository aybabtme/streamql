# streamql

`streamql` is a package implementing:
- a **small query language** used to slice and dice streams of messages, and
- a **virtual machine** that can interpret the language and execute it against a stream of _message_.

## Description

_Messages_ in `streamql` can have the following types:
- **string**: is an ordered list of letters (currently the grammar only supports letters in `[a-Z]` 😅 )
- **number**: is a real value (currently the grammar has a bug and only supports the integer part of numbers 😅 )
- **array**: is an ordered list of _messages_.
- **object**: is an unordered map of **string** to _message_.

The query language can then be used to select parts of individual message and (to be done 😅) filter messages
based on their content.

Basically it works quite a bit like [`./jq`](https://stedolan.github.io/jq/) but implements only part of a similar language and an engine to process arbitrary structured messages (not just JSON). The idea is that you can make `jq`-like tools for arbitrary structured message formats that support the types in the shape of a _message_.

## Usage

```go
import (
	"github.com/aybabtme/streamql/lang/parser"
	"github.com/aybabtme/streamql/lang/vm"
	"github.com/aybabtme/streamql/lang/vm/msg"
)

// with streams of input and output messages
inc := make(chan msg.Message)
outc := make(chan msg.Message)

// parse a query
ast, err := parser.NewParser(query).Parse()
if err != nil {
    log.Fatalf("invalid query: %v", err)
}

// execute the filters (a query can contain many filters)
filter := tree.Filters[0]
engine := vm.ASTInterpreter(filter)
engine.Filter(
    func() (msg.Message, bool) { msg, more := <-inc; return msg, more },
    func(m msg.Message) bool { outc <- m; return true },
)
```

## Goals

* performance: the language must be fast to parse, the grammar be unambiguous, and execution of queries on streams must be rapid.
* memory: the VM must use a bounded amount of memory to process any stream of message.
* simple language: features in the language must remain minimalistic and be orthogonal to one another.
* generic semantic: the language should be appliable on any dataformat that respects the __message__ semantics.

## Example queries

```streamql
.string
.string.string
.[]
.string[]
.[42]
.[].string
.string[42]
.[42].string
.[42:42]
.[][]
.string[42:42]
.[][42]
.[42:42].string
.[42][]
.[42][42]
.[42:42][]
.[][42:42]
.[42:42][42]
.[42][42:42]
.[42:42][42:42]
```

## Stability

Consider the package to be version 0. It will change, you should vendor the package if you want to have reproducible builds. When the grammar and VMs support the basic features of a v1, a git tag will be set.

## Contributing

Wait before firing up your editor! Please see [`CONTRIBUTING`](CONTRIBUTING.md). Not much process, but just to keep my sanity as a very part time open-source person.

## License

MIT.
