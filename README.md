ColorJSON: The Fast Color JSON Marshaller for Go
================================================
![ColorJSON Output](https://i.imgur.com/pLtCXhb.png)
What is this?
-------------

This package is based heavily on TylerBrock/colorjson but uses direct-to-stream serialisation to a `io.Writer` interface via an internal buffer.


Installation
------------

```sh
go get -u github.com/relvacode/colorjson
```

Usage
-----

Setup

```go
import "github.com/relvacode/colorjson"

str := `{
  "str": "foo",
  "num": 100,
  "bool": false,
  "null": null,
  "array": ["foo", "bar", "baz"],
  "obj": { "a": 1, "b": 2 }
}`

// Create an intersting JSON object to marshal in a pretty format
var obj map[string]interface{}
json.Unmarshal([]byte(str), &obj)
```

Vanilla Usage

```go
// Use stdout, or any other standard io.Writer interface
_, err := colorjson.Marshal(os.Stdout, obj)
```

Customization (Custom Indent)
```go
f := colorjson.NewFormatter()
f.Indent = 2

_, err := f.Marshal(os.Stdout, obj)
```
