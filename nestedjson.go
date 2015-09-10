package nestedjson

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
)

type NestedJson struct {
	data map[string]interface{}
}

func splitPath(path string) ([]interface{}, error) {
	const (
		Blank = iota
		Key
		Dot
		StartArrayIndex
		ArrayIndex
		EndArrayIndex
	)

	var parts []interface{}
	startPos := 0
	pos := 0
	state := Blank

	for pos = 0; pos < len(path); pos++ {
		c := path[pos]
		switch {

		case c == '.':
			switch state {
			case Blank, StartArrayIndex, Dot, ArrayIndex:
				return nil, fmt.Errorf("Invalid path: %s, pos: %d", path, pos)
			case Key:
				parts = append(parts, path[startPos:pos])
				state = Dot
			case EndArrayIndex:
				state = Dot
			}

		case c == '[':
			switch state {
			case Blank, EndArrayIndex:
				state = StartArrayIndex
			case StartArrayIndex, Dot:
				return nil, fmt.Errorf("Invalid path: %s, pos: %d", path, pos)
			case Key:
				parts = append(parts, path[startPos:pos])
				state = StartArrayIndex
			}

		case c == ']':
			switch state {
			case Blank, Key, StartArrayIndex, EndArrayIndex, Dot:
				return nil, fmt.Errorf("Invalid path: %s, pos: %d", path, pos)
			case ArrayIndex:
				i, _ := strconv.Atoi(path[startPos:pos])
				parts = append(parts, i)
				state = EndArrayIndex
			}

		case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
			switch state {
			case Blank, Key:
				state = Key
			case Dot:
				state = Key
				startPos = pos
			case StartArrayIndex, EndArrayIndex, ArrayIndex:
				return nil, fmt.Errorf("Invalid path: %s, pos: %d", path, pos)
			}

		case '0' <= c && c <= '9':
			switch state {
			case Blank, Key:
				state = Key
			case Dot:
				state = Key
				startPos = pos
			case StartArrayIndex:
				state = ArrayIndex
				startPos = pos
			case ArrayIndex:
				state = ArrayIndex
			case EndArrayIndex:
				return nil, fmt.Errorf("Invalid path: %s, pos: %d", path, pos)
			}
		}
	}

	switch state {
	case EndArrayIndex:
	case Key:
		parts = append(parts, path[startPos:pos])
	default:
		return nil, fmt.Errorf("Invalid path: %s, pos: %d", path, pos)
	}

	return parts, nil
}

func getPart(obj interface{}, part interface{},
	createMissingObject bool) (interface{}, error) {

	switch p := part.(type) {
	case int:
		if arr, ok := obj.([]interface{}); ok {
			if p < len(arr) {
				return arr[p], nil
			} else {
				return nil, fmt.Errorf("Array index out of bounds: %d", p)
			}
		} else {
			return nil, fmt.Errorf("%s is not an array: %T", obj, obj)
		}

	case string:
		if m, ok := obj.(map[string]interface{}); ok {
			if rv, ok := m[p]; ok {
				return rv, nil
			} else {
				if createMissingObject {
					rv = make(map[string]interface{})
					m[p] = rv
					return rv, nil
				}
				return nil, fmt.Errorf("Key does not exist: %s", p)
			}
		} else {
			return nil, fmt.Errorf("%s is not an object: %T", obj, obj)
		}
	}
	return nil, fmt.Errorf("Invalid Part: %T", part)
}

func New(args ...map[string]interface{}) *NestedJson {
	var m *NestedJson
	switch len(args) {
	case 0:
		m = &NestedJson{make(map[string]interface{})}
	case 1:
		m = &NestedJson{args[0]}
	default:
		log.Panicf("NewNestedJson() received too many arguments: %d", len(args))
	}
	return m
}

func Decode(b []byte) (*NestedJson, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err == nil {
		return &NestedJson{m}, nil
	} else {
		return nil, err
	}
}

func DecodeStr(s string) (*NestedJson, error) {
	if n, err := Decode([]byte(s)); err == nil {
		return n, nil
	} else {
		return nil, err
	}
}

func (n *NestedJson) Encode() ([]byte, error) {
	return json.Marshal(&n.data)
}

func (n *NestedJson) EncodeStr() (string, error) {
	if b, err := n.Encode(); err == nil {
		return string(b), nil
	} else {
		return "", err
	}
}

func (n *NestedJson) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&n.data, "", "  ")
}

func (n *NestedJson) EncodePrettyStr() (string, error) {
	if b, err := n.EncodePretty(); err == nil {
		return string(b), nil
	} else {
		return "", err
	}
}

func (n *NestedJson) Get(path string) (interface{}, error) {

	parts, err := splitPath(path)
	if err != nil {
		return nil, err
	}

	var curr interface{} = n.data
	for _, part := range parts {
		curr, err = getPart(curr, part, false)
		if err != nil {
			return nil, err
		}
	}
	return curr, nil
}

func (n *NestedJson) Set(path string, val interface{}) error {
	parts, err := splitPath(path)
	if err != nil {
		return err
	}

	var curr interface{} = n.data
	for _, part := range parts[:len(parts)-1] {
		curr, err = getPart(curr, part, true)
		if err != nil {
			return err
		}
	}

	switch k := parts[len(parts)-1].(type) {
	case int:
		if arr, ok := curr.([]interface{}); ok {
			arr[k] = val
		} else {
			return fmt.Errorf("Not an array: %s", curr)
		}

	case string:
		if m, ok := curr.(map[string]interface{}); ok {
			m[k] = val
		} else {
			return fmt.Errorf("Not an object: %s", curr)
		}
	}

	return nil

}

func (n *NestedJson) Data() map[string]interface{} {
	return n.data
}

func (n *NestedJson) String(path string) (string, error) {
	o, err := n.Get(path)
	if err != nil {
		return "", err
	}
	switch rv := o.(type) {
	case string:
		return rv, nil
	default:
		return "", fmt.Errorf("%s is not a string", path, o)
	}
}

func (n *NestedJson) Int(path string) (int, error) {
	o, err := n.Get(path)
	if err != nil {
		return 0, err
	}
	switch rv := o.(type) {
	case int:
		return rv, nil
	case float64:
		return int(rv), nil
	default:
		return 0, fmt.Errorf("%s is not an integer", path, o)
	}
}

func (n *NestedJson) Float(path string) (float64, error) {
	o, err := n.Get(path)
	if err != nil {
		return 0, err
	}
	switch rv := o.(type) {
	case int:
		return float64(rv), nil
	case float64:
		return rv, nil
	default:
		return 0, fmt.Errorf("%s is not a float", path, o)
	}
}

func (n *NestedJson) Bool(path string) (bool, error) {
	o, err := n.Get(path)
	if err != nil {
		return false, err
	}
	switch rv := o.(type) {
	case bool:
		return rv, nil
	default:
		return false, fmt.Errorf("%s is not a bool", path, o)
	}
}

func (n *NestedJson) Array(path string) ([]interface{}, error) {
	o, err := n.Get(path)
	if err != nil {
		return nil, err
	}
	switch rv := o.(type) {
	case []interface{}:
		return rv, nil
	default:
		return nil, fmt.Errorf("%s is not an Array", path, o)
	}
}

func (n *NestedJson) Map(path string) (map[string]interface{}, error) {
	o, err := n.Get(path)
	if err != nil {
		return nil, err
	}
	switch rv := o.(type) {
	case map[string]interface{}:
		return rv, nil
	default:
		return nil, fmt.Errorf("%s is not a map", path, o)
	}
}
