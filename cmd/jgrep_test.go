package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

var testSets = []struct {
	data   []byte
	path   string
	result []byte
}{
	{
		data:   []byte(`{"key": "value"}`),
		path:   "key",
		result: []byte(`"value"`),
	},
	{
		data:   []byte(`{"key": "value"}`),
		path:   "notkey",
		result: []byte(`null`),
	},
	{
		data:   []byte(`{"key": "value", "key2": "value2"}`),
		path:   "key",
		result: []byte(`"value"`),
	},
	{
		data:   []byte(`{"hash": { "key": "value"} }`),
		path:   "hash/key",
		result: []byte(`"value"`),
	},
	{
		data:   []byte(`{"hash": [ { "key": "value1"}, { "key": "value2" } ]}`),
		path:   "hash/*/key",
		result: []byte(`["value1","value2"]`),
	},
	{
		data:   []byte(`{"hash": [ { "key": "value1"}, { "key": "value2" }, { "notkey": "value3" } ]}`),
		path:   "hash/*/key",
		result: []byte(`["value1","value2"]`),
	},
	{
		data: []byte(`{"hash": { "subhash1": { "key": "value1"}, 
                                 "subhash2": { "key": "value2" }, 
                                 "subhash3": { "notkey": "value3" } }}`),
		path:   "hash/*/key",
		result: []byte(`["value1","value2"]`),
	},
	{
		data:   []byte(`{"items": [ {"cat": "color", "color": "red"}, {"cat": "food", "fruit": "apple"}, {"cat": "color", "color": "blue"} ]}`),
		path:   "items/cat=color",
		result: []byte(`[{"cat":"color","color":"red"},{"cat":"color","color":"blue"}]`),
	},
	{
		data:   []byte(`{"items": [ {"cat": "color", "color": "red"}, {"cat": "food", "fruit": "apple"}, {"cat": "color", "color": "blue"} ]}`),
		path:   "items/cat=color/color",
		result: []byte(`["red","blue"]`),
	},
	{
		data:   []byte(`[ {"cat": "color", "color": "red"}, {"cat": "food", "color": "green"}, {"cat": "color", "color": "blue"} ]`),
		path:   "*/cat,color",
		result: []byte(`[["color","red"],["food","green"],["color","blue"]]`),
	},
}

func TestJgrep(t *testing.T) {

	var source interface{}

	for _, set := range testSets {
		err := json.Unmarshal(set.data, &source)
		if err != nil {
			t.Fatal("Cannot Unmarshal json data")
		}
		res := jgrep(source, strings.Split(set.path, "/"))
		out, err := json.Marshal(res)
		if err != nil {
			t.Fatal("Cannot Marshal result")
		}
		match := bytes.Compare(out, set.result)
		if match != 0 {
			t.Errorf("Got <%s> but wanted <%s>", string(out), string(set.result))
		}
	}
}
