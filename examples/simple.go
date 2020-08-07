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

	// Marshall the Colorized JSON
	colorjson.Marshal(os.Stdout, obj)
}
