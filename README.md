# streamql

`streamql` is a package implementing:
- a **small query language** used to slice and dice streams of messages, and
- a **virtual machine** that can interpret the language and execute it against a stream of _message_.

## Description

_Messages_ in `streamql` can have the following types:
- **string**: is an ordered list of letters
- **number**: is a real value
- **boolean**: is a true or false value
- **array**: is an ordered list of _messages_.
- **object**: is an unordered map of **string** to _message_.

The query language can then be used to select parts of individual message and filter messages
based on their content.

Basically it works quite a bit like [`./jq`](https://stedolan.github.io/jq/) but implements only part of a similar language and an engine to process arbitrary structured messages (not just JSON). The idea is that you can make `jq`-like tools for arbitrary structured message formats that support the types in the shape of a _message_.

## Status

The API is very much a work in progress. The `msg.Sink`, `msg.Source` and `msg.Builder` are a bit cumbersome and I need to think more about how to make this cleaner.

## Goals

* performance: execution of queries on streams must be rapid... _right now, the AST and the VMs are pretty innefficient :)_
* memory: the VM must use a bounded amount of memory to process any stream of message.
* simple language: features in the language must remain minimalistic and be orthogonal to one another.
* generic semantic: the language should be appliable on any dataformat that respects the __message__ semantics.

## Example queries

For now the language is better described by it's [tests](https://github.com/aybabtme/streamql/blob/master/lang/vm/vmtest/basic.go#L25-L961), but here's a couple of valid queries:

```streamql
.hello
.hello["an awkward key"] + "world"
.[]
.[42]
.[:42]
.[42:]
.hello.world | select(. > 4.0)
select(.keep) | .name
.lol[0:1] | select(.is_red && string(.size) == "large") | select(.)
```

## Stability

Consider the package to be version 0. It will change, you should vendor the package if you want to have reproducible builds. When the grammar and VMs support the basic features of a v1, a git tag will be set.

## Contributing

Wait before firing up your editor! Please see [`CONTRIBUTING`](CONTRIBUTING.md). Not much process, but just to keep my sanity as a very part time open-source person.

## License

MIT.
