# gojq
Fast JSON (un)marshal &amp; query in Go

Shows ~ 10x speed improvement over standard `json.Unmarshal` with `map[string]interface{}` in parsing some types of JSON payload.

# disclaimer
- Designed for a very specific use case; make sure it fits yours
- Still a work in progress
- Doesn't handle any kind of whitespaces

# usage
Basic usage is simply:

```
buffer, err := ioutil.ReadAll(r.Body)
if err != nil {
    return
}

value := jq.Value{}
if err := value.Unmarshal(buffer); err != nil {
    return
}
```

But, consider making a pool of `jq.Value` for maximum performance:

```
values := sync.Pool{
    New: func() interface{} { return &Value{} }
}

...

value := values.Get().(*Value)

n, err := value.UnmarshalFrom(r.Body)
if err != nil {
    return
}

...

values.Put(value)
```

Note that this is the technique used in benchmarks.

# query
Once unmarshaled, use a `jq.Query` to navigate and extract information from the JSON. For example, here is how to get a string field:

```
q := value.NewQuery()

domain, err := q.String("bid", "request", "site", "domain")
if err != nil {
    return
}
```

Queries are nicer albeit slower (~ x2) than the equivalent required `map[string]interface{}` gymnastic. Because part of the parsing is done lazily, some conversions only happen inside queries. This is by design and optimize the case where you need to query only a few fields from the JSON payload.

# benchmarks
Using 10,000 samples found in file samples.json.

```
$ go test -v -bench . -run none -benchmem ./...
PASS
BenchmarkExtractGoAndFindJSON-32    10000000           133 ns/op           0 B/op          0 allocs/op
BenchmarkExtractAndFindJSONFast-32   5000000           264 ns/op          32 B/op          1 allocs/op
BenchmarkParseGoJSON-32                30000         58192 ns/op        7871 B/op        186 allocs/op
BenchmarkParseJSON-32                 300000          3941 ns/op          78 B/op          0 allocs/op
```
