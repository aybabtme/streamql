# protojq

A demo CLI tool that acts a bit like `jq`. Some queries will behave the same, 
but not all capabilities of `jq` are implemented (no user defined function, no allocation of
variables, no construction of _object_ or _arrays_), and some keywords differ (boolean algebra).

Still, you can do:

```bash
$ echo '{"hello":"world", "keep":true}' | protojq 'select(has("hello")) | .hello'
"world"
$ echo '{"hello":"world", "keep":true}' | jq 'select(has("hello")) | .hello'
"world"
```
