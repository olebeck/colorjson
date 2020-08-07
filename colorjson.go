package colorjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

const initialDepth = 0
const valueSep = ","
const null = "null"
const startMap = "{"
const endMap = "}"
const startArray = "["
const endArray = "]"

const emptyMap = startMap + endMap
const emptyArray = startArray + endArray

type Formatter struct {
	KeyColor        *color.Color
	StringColor     *color.Color
	BoolColor       *color.Color
	NumberColor     *color.Color
	NullColor       *color.Color
	StringMaxLength int
	Indent          int
	DisabledColor   bool
	RawStrings      bool
}

func NewFormatter() *Formatter {
	return &Formatter{
		KeyColor:        color.New(color.FgWhite),
		StringColor:     color.New(color.FgGreen),
		BoolColor:       color.New(color.FgYellow),
		NumberColor:     color.New(color.FgCyan),
		NullColor:       color.New(color.FgMagenta),
		StringMaxLength: 0,
		DisabledColor:   false,
		Indent:          0,
		RawStrings:      false,
	}
}

func (f *Formatter) sprintfColor(c *color.Color, format string, args ...interface{}) string {
	if f.DisabledColor || c == nil {
		return fmt.Sprintf(format, args...)
	}
	return c.SprintfFunc()(format, args...)
}

func (f *Formatter) sprintColor(c *color.Color, s string) string {
	if f.DisabledColor || c == nil {
		return fmt.Sprint(s)
	}
	return c.SprintFunc()(s)
}

func (f *Formatter) writeIndent(w *bufio.Writer, depth int) (int, error) {
	return w.WriteString(strings.Repeat(" ", f.Indent*depth))
}

func (f *Formatter) writeObjSep(w *bufio.Writer) (int, error) {
	if f.Indent != 0 {
		return w.WriteRune('\n')
	} else {
		return w.WriteRune(' ')
	}
}

func (f *Formatter) Marshal(w io.Writer, jsonObj interface{}) (int, error) {
	buf := bufio.NewWriter(w)
	n, err := f.marshalValue(jsonObj, buf, initialDepth)
	if err != nil {
		return n, err
	}

	err = buf.Flush()
	if err != nil {
		return n, err
	}

	return n, nil
}

func (f *Formatter) marshalMap(m map[string]interface{}, w *bufio.Writer, depth int) (int, error) {
	remaining := len(m)

	if remaining == 0 {
		return w.WriteString(emptyMap)
	}

	keys := make([]string, 0)
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var wr int
	n, err := w.WriteString(startMap)
	if err != nil {
		return wr, err
	}

	wr += n

	n, err = f.writeObjSep(w)
	if err != nil {
		return n, err
	}

	wr += n

	for _, key := range keys {
		n, err = f.writeIndent(w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = w.WriteString(f.KeyColor.Sprintf("\"%s\": ", key))
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = f.marshalValue(m[key], w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		remaining--
		if remaining != 0 {
			n, err = w.WriteString(valueSep)
			if err != nil {
				return wr, err
			}

			wr += n
		}

		n, err = f.writeObjSep(w)
		if err != nil {
			return wr, err
		}

		wr += n
	}

	n, err = f.writeIndent(w, depth)
	if err != nil {
		return wr, err
	}

	wr += n

	n, err = w.WriteString(endMap)
	if err != nil {
		return wr, err
	}

	wr += n

	return wr, nil
}

func (f *Formatter) marshalArray(a []interface{}, w *bufio.Writer, depth int) (int, error) {
	if len(a) == 0 {
		return w.WriteString(emptyArray)
	}

	var wr int

	n, err := w.WriteString(startArray)
	if err != nil {
		return n, err
	}

	wr += n

	n, err = f.writeObjSep(w)
	if err != nil {
		return wr, err
	}

	wr += n

	for i, v := range a {
		n, err = f.writeIndent(w, depth)
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = f.marshalValue(v, w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		if i < len(a)-1 {
			n, err = w.WriteString(valueSep)
			if err != nil {
				return wr, err
			}

			wr += n
		}

		n, err = f.writeObjSep(w)
		if err != nil {
			return wr, err
		}

		wr += n
	}
	n, err = f.writeIndent(w, depth)
	if err != nil {
		return wr, err
	}

	wr += n

	n, err = w.WriteString(endMap)
	if err != nil {
		return wr, err
	}

	wr += n

	return wr, nil
}

func (f *Formatter) marshalValue(val interface{}, w *bufio.Writer, depth int) (int, error) {
	switch v := val.(type) {
	case map[string]interface{}:
		return f.marshalMap(v, w, depth)
	case []interface{}:
		return f.marshalArray(v, w, depth)
	case string:
		return f.marshalString(v, w)
	case float64:
		return w.WriteString(f.sprintColor(f.NumberColor, strconv.FormatFloat(v, 'f', -1, 64)))
	case bool:
		return w.WriteString(f.sprintColor(f.BoolColor, strconv.FormatBool(v)))
	case nil:
		return w.WriteString(f.sprintColor(f.NullColor, null))
	case json.Number:
		return w.WriteString(f.sprintColor(f.NumberColor, v.String()))
	}

	return 0, nil
}

func (f *Formatter) marshalString(str string, w *bufio.Writer) (int, error) {
	if !f.RawStrings {
		strBytes, _ := json.Marshal(str)
		str = string(strBytes)
	}

	if f.StringMaxLength != 0 && len(str) >= f.StringMaxLength {
		str = fmt.Sprintf("%s...", str[0:f.StringMaxLength])
	}

	return w.WriteString(f.sprintColor(f.StringColor, str))
}

// Marshal JSON data with default options
func Marshal(w io.Writer, jsonObj interface{}) (int, error) {
	return NewFormatter().Marshal(w, jsonObj)
}
