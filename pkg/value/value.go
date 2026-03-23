package value

import (
	"encoding/json"
	"fmt"
	"io"
)

type Value struct {
	prototype interface{}
	parent    *Value
	from      string
	castErr   error
	unsafe    *UnsafeValue
}

type UnsafeValue struct {
	*Value
}

func (u *UnsafeValue) Int() int {
	result, _ := u.Value.Int()
	return result
}

func (u *UnsafeValue) Bool() bool {
	result, _ := u.Value.Bool()
	return result
}

func (u *UnsafeValue) Float64() float64 {
	result, _ := u.Value.Float64()
	return result
}

func (u *UnsafeValue) String() string {
	result, _ := u.Value.String()
	return result
}

func (u *UnsafeValue) StringSlice() []string {
	result, _ := u.Value.StringSlice()
	return result
}

func (u *UnsafeValue) Slice() []interface{} {
	result, _ := u.Value.Slice()
	return result
}

func (u *UnsafeValue) ValueSlice() []*Value {
	result, _ := u.Value.ValueSlice()
	return result
}

func (u *UnsafeValue) Map() map[string]interface{} {
	result, _ := u.Value.Map()
	return result
}

func (u *UnsafeValue) StringMap() map[string]string {
	result, _ := u.Value.StringMap()
	return result
}

func (u *UnsafeValue) SuperGet(keyOrPoses ...interface{}) *Value {
	traveler := u.Value

	for _, keyOrPos := range keyOrPoses {
		if pos, ok := keyOrPos.(int); ok {
			traveler = traveler.valueAt(pos)
		} else if key, ok := keyOrPos.(string); ok {
			traveler = traveler.get(key)
		}
	}

	return traveler
}

func NewValue(prototype interface{}) *Value {
	value := &Value{
		prototype: prototype,
		parent:    nil,
		from:      "",
		castErr:   nil,
	}
	value.unsafe = &UnsafeValue{value}
	return value
}

func NewChildValue(prototype interface{}, parent *Value, from string) *Value {
	value := NewValue(prototype)
	value.parent = parent
	value.from = from
	return value
}

func NewValueFromReader(reader io.Reader) (*Value, error) {
	var prototype map[string]interface{}
	err := json.NewDecoder(reader).Decode(&prototype)
	if err != nil {
		return nil, err
	}
	return NewValue(prototype), nil
}

func (v *Value) Unsafe() *UnsafeValue {
	return v.unsafe
}

func (v *Value) Int() (int, error) {
	if result, ok := v.prototype.(int); ok {
		return result, nil
	}
	return 0, fmt.Errorf("value '%s' is not int type", v.prototype)
}

func (v *Value) Bool() (bool, error) {
	if result, ok := v.prototype.(bool); ok {
		return result, nil
	}
	return false, fmt.Errorf("value '%s' is not int type", v.prototype)
}

func (v *Value) Float64() (float64, error) {
	if result, ok := v.prototype.(float64); ok {
		return result, nil
	}
	return 0, fmt.Errorf("value '%s' is not float64 type", v.prototype)
}

func (v *Value) String() (string, error) {
	if result, ok := v.prototype.(string); ok {
		return result, nil
	}
	return "", fmt.Errorf("value '%s' is not string type", v.prototype)
}

func (v *Value) StringSlice() ([]string, error) {
	if result, ok := v.prototype.([]string); ok {
		return result, nil
	}
	return nil, fmt.Errorf("value '%s' is not []string type", v.prototype)
}

func (v *Value) Slice() ([]interface{}, error) {
	if result, ok := v.prototype.([]interface{}); ok {
		return result, nil
	}
	return nil, fmt.Errorf("value '%s' is not []interface{} type", v.prototype)
}

func (v *Value) ValueSlice() ([]*Value, error) {
	slice, err := v.Slice()
	if err != nil {
		return nil, err
	}
	result := make([]*Value, len(slice))
	for i, s := range slice {
		result[i] = NewChildValue(s, v, fmt.Sprintf("[%d]", i))
	}
	return result, nil
}

func (v *Value) ValueAt(poses ...int) *Value {
	traveler := v

	for _, pos := range poses {
		traveler = traveler.valueAt(pos)
	}

	return traveler
}

func (v *Value) valueAt(pos int) *Value {
	result := NewChildValue(nil, v, fmt.Sprintf("[%d]", pos))

	if first, err := v.ValueSlice(); err == nil {
		if pos < 0 || pos >= len(first) {
			result.castErr = fmt.Errorf("pos '%d' out of bounds 0 to %d", pos, len(first))
		} else {
			// Notice that the prototype cannot be Value struct.
			result.prototype = first[pos].prototype
		}
	} else if second, err := v.StringSlice(); err == nil {
		if pos < 0 || pos >= len(second) {
			result.castErr = fmt.Errorf("pos '%d' out of bounds 0 to %d", pos, len(second))
		} else {
			result.prototype = second[pos]
		}
	} else {
		result.castErr = fmt.Errorf("get value at '%d' error: parent is neither []interface{} nor []string type, but %s", pos, v.Type())
	}

	return result
}

func (v *Value) Head() *Value {
	result := NewChildValue(nil, v, "[0]")

	if first, err := v.ValueSlice(); err == nil {
		if len(first) <= 0 {
			result.castErr = fmt.Errorf("out of bounds")
		} else {
			// Notice that the prototype cannot be Value struct.
			result.prototype = first[0].prototype
		}
	} else if second, err := v.StringSlice(); err == nil {
		if len(second) <= 0 {
			result.castErr = fmt.Errorf("out of bounds")
		} else {
			result.prototype = second[0]
		}
	} else {
		result.castErr = fmt.Errorf("get head error: parent is neither []interface{} nor []string type, but %s", v.Type())
	}

	return result
}

func (v *Value) Tail() *Value {
	result := NewChildValue(nil, v, "[-1]")

	if first, err := v.ValueSlice(); err == nil {
		if len(first) <= 0 {
			result.castErr = fmt.Errorf("out of bounds")
		} else {
			// Notice that the prototype cannot be Value struct.
			result.prototype = first[len(first)-1].prototype
		}
	} else if second, err := v.StringSlice(); err == nil {
		if len(second) <= 0 {
			result.castErr = fmt.Errorf("out of bounds")
		} else {
			result.prototype = second[len(second)-1]
		}
	} else {
		result.castErr = fmt.Errorf("get tail error: parent is neither []interface{} nor []string type, but %s", v.Type())
	}

	return result
}

func (v *Value) Map() (map[string]interface{}, error) {
	if result, ok := v.prototype.(map[string]interface{}); ok {
		return result, nil
	}
	return nil, fmt.Errorf("value '%s' is not map[string]interface{} type", v.prototype)
}

func (v *Value) StringMap() (map[string]string, error) {
	if result, ok := v.prototype.(map[string]string); ok {
		return result, nil
	}
	return nil, fmt.Errorf("value '%s' is not map[string]string type", v.prototype)
}

func (v *Value) Raw() interface{} {
	return v.prototype
}

func (v *Value) Error() error {
	return v.castErr
}

func (v *Value) Success() bool {
	return v.castErr == nil
}

func (v *Value) Type() string {
	switch t := v.prototype.(type) {
	case int:
		return "int"
	case float64:
		return "float64"
	case string:
		return "string"
	case []string:
		return "[]string"
	case []interface{}:
		return "[]interface{}"
	case map[string]string:
		return "map[string]string"
	case map[string]interface{}:
		return "map[string]interface{}"
	default:
		return fmt.Sprintf("unsupported type %v", t)
	}
}

func (v *Value) Trace() string {
	trace := v.from
	parent := v.parent

	for {
		if parent == nil {
			break
		}
		trace = parent.from + "." + trace
		parent = parent.parent
	}

	return "source" + trace
}

func (v *Value) Debug() string {
	trace := ""
	parent := v
	rootCause := "success"
	seek := 0

	for {
		if parent == nil {
			break
		}
		if trace == "" {
			trace = parent.from
		} else {
			trace = parent.from + "." + trace
		}
		if parent.castErr != nil && parent.parent == nil {
			rootCause = parent.castErr.Error()
			seek = len(trace) - 1
		}
		if parent.castErr != nil && parent.parent != nil && parent.parent.castErr == nil {
			rootCause = parent.castErr.Error()
			seek = len(trace) - 1
		}
		parent = parent.parent
	}

	trace = "source" + trace

	if seek == 0 {
		return trace + ": " + rootCause
	}

	shift := func(input *string, blank int) {
		for i := 0; i < blank; i++ {
			*input += " "
		}
	}

	candidate := trace + "\n"
	shift(&candidate, len(trace)-seek)
	candidate += "x\n"
	shift(&candidate, len(trace)-seek)
	candidate += "|\n"
	shift(&candidate, len(trace)-seek)
	candidate += rootCause

	return candidate
}

func (v *Value) Get(keys ...string) *Value {
	traveler := v

	for _, key := range keys {
		traveler = traveler.get(key)
	}

	return traveler
}

func (v *Value) get(key string) *Value {
	result := &Value{
		prototype: nil,
		parent:    v,
		from:      fmt.Sprintf("get('%s')", key),
		castErr:   nil,
	}
	result.unsafe = &UnsafeValue{result}

	if first, err := v.Map(); err == nil {
		if target, ok := first[key]; ok {
			result.prototype = target
		} else {
			result.castErr = fmt.Errorf("key '%s' not found", key)
		}
	} else if second, err := v.StringMap(); err == nil {
		if target, ok := second[key]; ok {
			result.prototype = target
		} else {
			result.castErr = fmt.Errorf("key '%s' not found", key)
		}
	} else {
		result.castErr = fmt.Errorf("get key '%s' error: parent is neither map[string]interface{} nor map[string]string type, but %s", key, v.Type())
	}

	return result
}
