package colorjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/xo/terminfo"
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
	Buffer          *bufio.Writer
	BackColor       color.PrinterFace
	KeyColor        color.PrinterFace
	StringColor     color.PrinterFace
	BoolColor       color.PrinterFace
	NumberColor     color.PrinterFace
	NullColor       color.PrinterFace
	StringMaxLength int
	Indent          int
	DisabledColor   bool
	RawStrings      bool
}

func init() {
	color.ForceSetColorLevel(terminfo.ColorLevelMillions)
}

func NewFormatter(w io.Writer) *Formatter {
	f := &Formatter{
		Buffer:          bufio.NewWriter(w),
		BackColor:       color.FgWhite,
		KeyColor:        color.C256(250),
		StringColor:     color.FgGreen,
		BoolColor:       color.FgYellow,
		NumberColor:     color.FgCyan,
		NullColor:       color.FgMagenta,
		StringMaxLength: 0,
		DisabledColor:   false,
		Indent:          0,
		RawStrings:      false,
	}
	return f
}

func (f *Formatter) sprintfColor(c color.PrinterFace, format string, args ...interface{}) string {
	if f.DisabledColor || c == nil {
		return fmt.Sprintf(format, args...)
	}
	return c.Sprintf(format, args...)
}

func (f *Formatter) sprintColor(c color.PrinterFace, s string) string {
	if f.DisabledColor || c == nil {
		return fmt.Sprint(s)
	}
	return c.Sprint(s)
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

func (f *Formatter) Encode(jsonObj interface{}) error {
	if s, ok := jsonObj.(string); ok {
		f.Buffer.WriteString(s)
		return f.Buffer.Flush()
	}
	_, err := f.marshalValue(reflect.ValueOf(jsonObj), f.Buffer, initialDepth)
	if err != nil {
		return err
	}

	err = f.Buffer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (f *Formatter) marshalStruct(m reflect.Value, w *bufio.Writer, depth int) (int, error) {
	remaining := m.NumField()
	t := m.Type()

	if remaining == 0 {
		return w.WriteString(f.sprintColor(f.BackColor, emptyMap))
	}

	var wr int
	n, err := w.WriteString(f.sprintColor(f.BackColor, startMap))
	if err != nil {
		return wr, err
	}

	wr += n

	n, err = f.writeObjSep(w)
	if err != nil {
		return n, err
	}

	wr += n

	for i := 0; i < m.NumField(); i++ {
		n, err = f.writeIndent(w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		keyName := t.Field(i).Name

		n, err = w.WriteString(f.KeyColor.Sprintf("\"%s\": ", keyName))
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = f.marshalValue(m.Field(i), w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		remaining--
		if remaining != 0 {
			n, err = w.WriteString(f.sprintColor(f.BackColor, valueSep))
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

	n, err = w.WriteString(f.sprintColor(f.BackColor, endMap))
	if err != nil {
		return wr, err
	}

	wr += n

	return wr, nil
}

func (f *Formatter) marshalMap(m reflect.Value, w *bufio.Writer, depth int) (int, error) {
	remaining := m.Len()

	if remaining == 0 {
		return w.WriteString(f.sprintColor(f.BackColor, emptyMap))
	}

	var wr int
	n, err := w.WriteString(f.sprintColor(f.BackColor, startMap))
	if err != nil {
		return wr, err
	}

	wr += n

	n, err = f.writeObjSep(w)
	if err != nil {
		return n, err
	}

	wr += n

	for _, key := range m.MapKeys() {
		n, err = f.writeIndent(w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = w.WriteString(f.KeyColor.Sprintf("\"%s\": ", key.String()))
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = f.marshalValue(m.MapIndex(key), w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		remaining--
		if remaining != 0 {
			n, err = w.WriteString(f.sprintColor(f.BackColor, valueSep))
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

	n, err = w.WriteString(f.sprintColor(f.BackColor, endMap))
	if err != nil {
		return wr, err
	}

	wr += n

	return wr, nil
}

func (f *Formatter) marshalArray(a reflect.Value, w *bufio.Writer, depth int) (int, error) {
	if a.Len() == 0 {
		return w.WriteString(f.sprintColor(f.BackColor, emptyArray))
	}

	var wr int

	n, err := w.WriteString(f.sprintColor(f.BackColor, startArray))
	if err != nil {
		return n, err
	}

	wr += n

	n, err = f.writeObjSep(w)
	if err != nil {
		return wr, err
	}

	wr += n

	for i := 0; i < a.Len(); i++ {
		n, err = f.writeIndent(w, depth)
		if err != nil {
			return wr, err
		}

		wr += n

		n, err = f.marshalValue(a.Index(i), w, depth+1)
		if err != nil {
			return wr, err
		}

		wr += n

		if i < a.Len()-1 {
			n, err = w.WriteString(f.sprintColor(f.BackColor, valueSep))
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

	n, err = w.WriteString(f.sprintColor(f.BackColor, endArray))
	if err != nil {
		return wr, err
	}

	wr += n

	return wr, nil
}

func (f *Formatter) marshalValue(val reflect.Value, w *bufio.Writer, depth int) (int, error) {
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	switch val.Type().Kind() {
	case reflect.Map:
		return f.marshalMap(val, w, depth)
	case reflect.Slice:
		return f.marshalArray(val, w, depth)
	case reflect.String:
		return f.marshalString(val.String(), w)
	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var s string
		if val.CanFloat() {
			s = strconv.FormatFloat(val.Float(), 'f', -1, 64)
		} else if val.CanInt() {
			s = strconv.FormatInt(val.Int(), 10)
		}
		return w.WriteString(f.sprintColor(f.NumberColor, s))
	case reflect.Bool:
		return w.WriteString(f.sprintColor(f.BoolColor, strconv.FormatBool(val.Bool())))
	case reflect.Invalid:
		return w.WriteString(f.sprintColor(f.NullColor, null)) // nil todo
	case reflect.Struct:
		return f.marshalStruct(val, w, depth)
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
func Marshal(w io.Writer, jsonObj interface{}) error {
	return NewFormatter(w).Encode(jsonObj)
}
