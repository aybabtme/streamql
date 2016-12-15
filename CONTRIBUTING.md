# PRs

## Branch

PRs must be made against the `next` branch, **not against master**.

## Before starting

Before starting work on a PR, consider what kind of PR you think you'll send:

- **small bug fix, performance tweaks**: go ahead!
- **refactors, larger changes**: please file an issue describing your intent.
- **new features**: please file an issue describing the feature.

## Be concise

Please try to keep your PR on only one topic. That is, don't mix refactoring code with bugfixes, or new features with bug fixes.

## Important considerations

In accordance with the goals of the language, here are some important questions to take into consideration.

### for new features

The package is meant to work well on streams of messages, so the overhead to process the stream should be small. That means changes to the query language should ensure that the virtual machine holds `O(1)` objects in memory to process any message.

For example: exploding an array with `.[]` may seem to require `O(n)` memory, but because the explosion of the array is sent in the `Sink` type for each element, processing a `.[]` query does not require more than `O(1)` objects. On the other hand, if we were to allow a `unique` keyword that enforces unicity of some message property, the virtual machine would have to hold an unbounded index of all the properties it saw before. That index would be `O(n)` (unless using a bloomfilter or a leaky cache), and thus would not be accepted in the language.

The idea is that users of the package need to be able to trust that the virtual machines will not use an unbounded amount of memory.

### for changes to the grammar

The specification of the grammar needs to remain LL(1). The specification is provided in `streamql.ebnf` in a grammar language quasi compatible with the grammar verifier tool at http://smlweb.cpsc.ucalgary.ca/start.html.
