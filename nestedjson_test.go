package nestedjson

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var jsonStrings = map[string]string{
	"s1": `{"a": 1, "b": "moo", "c": true, "d": 1.2}`,
	"complex": `{
	  "a": {
	    "b": {
	      "c": {
	        "h": [
	          [1, 2, 3], 
	          ["a", "b", "c"], 
	          [1.2, 4.5, 7.8], 
	          [
	            ["h", "i", "j"], 
	            ["k", "l", "m"]
	          ]
	        ], 
	        "e": "moo", 
	        "d": 1, 
	        "g": {
	          "y": [1.3, 1.5, 2.8], 
	          "x": [0, 1, 2], 
	          "z": [
	            {"a": "hello", "b": "world"}, 
	            {"a": 100.12, "b": 200.24}, 
	            {"a": 1, "c": "go rocks", "b": 2}
	          ]
	        }, 
	        "f": ["cow", "dog", "bird"]
	      }
	    }
	  }
	}`,
}

func getNestedJson(name string) *NestedJson {
	n, _ := Decode([]byte(jsonStrings[name]))
	return n
}

func TestSplitPath(t *testing.T) {
	var testPaths = []struct {
		path  string
		parts []interface{}
	}{
		{"a", []interface{}{"a"}},
		{"a.b", []interface{}{"a", "b"}},
		{"[0]", []interface{}{0}},
		{"[0][1][2]", []interface{}{0, 1, 2}},
		{"a.b.c[0][1].d[0]", []interface{}{"a", "b", "c", 0, 1, "d", 0}},
		{"[0][1].a", []interface{}{0, 1, "a"}},
		{"[0].a[1].b[2][3].c.a", []interface{}{0, "a", 1, "b", 2, 3, "c", "a"}},
	}

	for _, item := range testPaths {
		parts, err := splitPath(item.path)
		assert.Nil(t, err)
		assert.Equal(t, parts, item.parts)
	}
}

func TestSplitPathErrors(t *testing.T) {
	var errorPaths = []string{
		"",
		"a..b.",
		"..",
		"a[[2]",
		"[]",
		"a[0.",
		"a[0].[1]",
	}

	for _, item := range errorPaths {
		_, err := splitPath(item)
		assert.Error(t, err)
	}
}

func TestGetSimple(t *testing.T) {
	json := getNestedJson("s1")
	testPaths := []struct {
		path string
		val  interface{}
	}{
		{"a", 1},
		{"b", "moo"},
		{"c", true},
		{"d", 1.2},
	}

	for _, i := range testPaths {
		v, err := json.Interface(i.path)
		assert.Nil(t, err)
		assert.Equal(t, v, i.val)
	}
}

func TestGetComplex(t *testing.T) {
	json := getNestedJson("complex")
	testPaths := []struct {
		path string
		val  interface{}
	}{
		{"a.b.c.d", 1},
		{"a.b.c.e", "moo"},
		{"a.b.c.f", []interface{}{"cow", "dog", "bird"}},
		{"a.b.c.g.x[0]", 0},
		{"a.b.c.g.y[1]", 1.5},
		{"a.b.c.g.z[0].a", "hello"},
		{"a.b.c.g.z[1].b", 200.24},
		{"a.b.c.g.z[2].c", "go rocks"},
		{"a.b.c.h[0][0]", 1},
		{"a.b.c.h[0][1]", 2},
		{"a.b.c.h[0][2]", 3},
		{"a.b.c.h[3][0][0]", "h"},
		{"a.b.c.h[3][1][2]", "m"},
	}

	for _, i := range testPaths {
		v, err := json.Interface(i.path)
		assert.Nil(t, err)
		assert.Equal(t, v, i.val, i.path)
	}
}
