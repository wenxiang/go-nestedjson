package nestedjson

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var jsonStrings = map[string]string{
	"s1": `{
		"a": 1, 
		"b": "moo", 
		"c": true, 
		"d": 1.2
	}`,

	"s2": `{
		"a": {
			"b": "moo",
			"c": 1,
			"d": false
		},
		"b": 0,
		"c": [1,2,3],
		"d": [[0, 1], {"a": 1}, [{"b": 2}, {"c": 3}]]
	}`,

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

func getTestJson(t *testing.T, name string) *NestedJson {
	n, err := DecodeStr(jsonStrings[name])
	assert.NoError(t, err, "JSON Decode failed: %s", name)
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
		"a[2",
		"[]",
		"a[0.",
		"a[0].[1]",
	}

	for _, item := range errorPaths {
		_, err := splitPath(item)
		assert.Error(t, err, "Error was expected for path: %s", item)
	}
}

func TestGetSimple(t *testing.T) {
	json := getTestJson(t, "s1")
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
		v, err := json.Get(i.path)
		assert.Nil(t, err)
		assert.EqualValues(t, v, i.val)
	}
}

func TestGetComplex(t *testing.T) {
	json := getTestJson(t, "complex")
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
		v, err := json.Get(i.path)
		assert.Nil(t, err)
		assert.EqualValues(t, v, i.val, i.path)
	}
}

func TestGetErrors(t *testing.T) {
	json := getTestJson(t, "s2")
	testPaths := []struct {
		path      string
		errString string
	}{
		{"a.b.e", "moo is not an object"},
		{"a.f.m.a", "Key does not exist"},
		{"a[0]", "not an array"},
		{"c[10]", "out of bounds"},
		{"d[0][5]", "out of bounds"},
		{"d[1].b", "does not exist"},
		{"d[2][0].b.e", "not an object"},
		{"d[2][0].c", "does not exist"},
	}

	for _, i := range testPaths {
		_, err := json.Get(i.path)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), i.errString, i.path)
		}
	}
}

func TestSetNew(t *testing.T) {
	json := New()
	tests := []struct {
		path string
		val  interface{}
	}{
		{"a.b.c", 1},
		{"a.b.d", "moo"},
		{"b", []interface{}{1, 2, 3}},
		{"b[0]", 4},
		{"c", map[string]interface{}{
			"A": 1, "B": 1.2, "C": true,
		}},
		{"c.A", false},
		{"c.A", "X"},
		{"c.B", 4.5},
		{"b[0]", []interface{}{1.2, 1.3, 1.4}},
		{"b[0][0]", []interface{}{"a", "b", "c"}},
		{"b[0][0][1]", "FUU"},
	}

	for _, i := range tests {
		json.Set(i.path, i.val)
		v, err := json.Get(i.path)
		assert.NoError(t, err)
		assert.Equal(t, i.val, v, "%s != %s", i.path, i.val)
	}

	jsonString, err := json.EncodeStr()
	assert.NoError(t, err)
	assert.Equal(t, jsonString,
		`{"a":{"b":{"c":1,"d":"moo"}},"b":[[["a","FUU","c"],`+
			`1.3,1.4],2,3],"c":{"A":"X","B":4.5,"C":true}}`)
}

func TestSetExisting(t *testing.T) {
	json := getTestJson(t, "s2")
	tests := []struct {
		path string
		val  interface{}
	}{
		{"a.b", map[string]interface{}{
			"x": 0.5, "y": 10,
		}},
		{"c[0]", "xxx"},
		{"b", []interface{}{1, 2, 3, 4, 5}},
		{"d[1].a", "zzz"},
	}

	for _, i := range tests {
		json.Set(i.path, i.val)
		v, err := json.Get(i.path)
		assert.NoError(t, err)
		assert.Equal(t, i.val, v, "%s != %s", i.path, i.val)
	}

	jsonString, err := json.EncodeStr()
	assert.NoError(t, err)
	assert.Equal(t, jsonString,
		`{"a":{"b":{"x":0.5,"y":10},"c":1,"d":false},`+
			`"b":[1,2,3,4,5],"c":["xxx",2,3],"d":[[0,1],`+
			`{"a":"zzz"},[{"b":2},{"c":3}]]}`)
}
