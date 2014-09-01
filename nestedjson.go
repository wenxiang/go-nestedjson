package nestedjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

var partRe, _ = regexp.Compile(`^([A-Za-z0-9_]*)((\[[0-9]+\])+)*$`)
var arrayIndexRe, _ = regexp.Compile(`\[([0-9]+)\]`)

type NestedJson struct {
	data map[string]interface{}
}

func splitPath(path string) ([]interface{}, error) {
	var rv []interface{}
	parts := strings.Split(path, ".")
	for i, part := range parts {
		if part == "" {
			return nil, errors.New("Invalid path: " + path)
		}
		partMatches := partRe.FindStringSubmatch(part)
		if len(partMatches) == 0 {
			return nil, errors.New("Invalid part: " + part)
		}
		// abc[0][1][2]
		objKey := partMatches[1]       //abc
		arrayIndexes := partMatches[2] // [0][1][2]

		if objKey == "" {
			if i > 0 {
				return nil, errors.New("Invalid path: " + path)
			}
		} else {
			rv = append(rv, objKey)
		}

		if arrayIndexes != "" {
			arrayIndexMatches := arrayIndexRe.FindAllStringSubmatch(arrayIndexes, -1)
			for _, indexMatch := range arrayIndexMatches {
				intIndex, _ := strconv.Atoi(indexMatch[1])
				rv = append(rv, intIndex)
			}
		}
	}
	return rv, nil
}

func getPart(obj interface{}, part interface{},
	createMissingObject bool) (interface{}, error) {

	switch p := part.(type) {
	case int:
		if arr, ok := obj.([]interface{}); ok {
			if p < len(arr) {
				return arr[p], nil
			} else {
				return nil, errors.New(fmt.Sprintf("Array index out of bounds: %d", p))
			}
		} else {
			return nil, errors.New(fmt.Sprintf("%s is not an array: %T", obj, obj))
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
				return nil, errors.New(fmt.Sprintf("Key does not exist: %s", p))
			}
		} else {
			return nil, errors.New(fmt.Sprintf("%s is not an object: %T", obj, obj))
		}
	}
	return nil, errors.New(fmt.Sprintf("Invalid Part: %T", part))
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

func (n *NestedJson) Encode() ([]byte, error) {
	return json.Marshal(&n.data)
}

func (n *NestedJson) Interface(path string) (interface{}, error) {

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
			return errors.New(fmt.Sprintf("Not an array: %s", curr))
		}

	case string:
		if m, ok := curr.(map[string]interface{}); ok {
			m[k] = val
		} else {
			return errors.New(fmt.Sprintf("Not an object: %s", curr))
		}
	}

	return nil

}

func (n *NestedJson) String(path string) (string, error) {
	o, err := n.Interface(path)
	if err != nil {
		return "", err
	}
	switch rv := o.(type) {
	case string:
		return rv, nil
	default:
		return "", errors.New(fmt.Sprintf("%s is not a string", path, o))
	}
}

func (n *NestedJson) Int(path string) (int, error) {
	o, err := n.Interface(path)
	if err != nil {
		return 0, err
	}
	switch rv := o.(type) {
	case int:
		return rv, nil
	case float64:
		return int(rv), nil
	default:
		return 0, errors.New(fmt.Sprintf("%s is not an integer", path, o))
	}
}

func (n *NestedJson) Float(path string) (float64, error) {
	o, err := n.Interface(path)
	if err != nil {
		return 0, err
	}
	switch rv := o.(type) {
	case int:
		return float64(rv), nil
	case float64:
		return rv, nil
	default:
		return 0, errors.New(fmt.Sprintf("%s is not a float", path, o))
	}
}

func (n *NestedJson) Bool(path string) (bool, error) {
	o, err := n.Interface(path)
	if err != nil {
		return false, err
	}
	switch rv := o.(type) {
	case bool:
		return rv, nil
	default:
		return false, errors.New(fmt.Sprintf("%s is not a bool", path, o))
	}
}

func (n *NestedJson) Array(path string) ([]interface{}, error) {
	o, err := n.Interface(path)
	if err != nil {
		return nil, err
	}
	switch rv := o.(type) {
	case []interface{}:
		return rv, nil
	default:
		return nil, errors.New(fmt.Sprintf("%s is not an Array", path, o))
	}
}

func (n *NestedJson) Map(path string) (map[string]interface{}, error) {
	o, err := n.Interface(path)
	if err != nil {
		return nil, err
	}
	switch rv := o.(type) {
	case map[string]interface{}:
		return rv, nil
	default:
		return nil, errors.New(fmt.Sprintf("%s is not a map", path, o))
	}
}
