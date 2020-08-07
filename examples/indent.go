package main

import (
	"encoding/json"
	"github.com/relvacode/colorjson"
	"os"
)

func main() {
	str := `{
      "str": "foo",
      "num": 100,
      "bool": false,
      "null": null,
      "array": ["foo", "bar", "baz"],
      "obj": { "a": 1, "b": 2 }
    }`

	var obj map[string]interface{}
	json.Unmarshal([]byte(str), &obj)

	// Make a custom formatter with indent set
	f := colorjson.NewFormatter()
	f.Indent = 4

	// Marshall the Colorized JSON to STDOUT
	f.Marshal(os.Stdout, obj)
}
