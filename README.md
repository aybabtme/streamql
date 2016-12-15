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

## Stability

Consider the package to be version 0. It will change, you should vendor the package if you want to have reproducible builds. When the grammar and VMs support the basic features of a v1, a git tag will be set.

## Contributing

Wait before firing up your editor! Please see [`CONTRIBUTING`](CONTRIBUTING.md). Not much process, but just to keep my sanity as a very part time open-source person.

## License

MIT.

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
