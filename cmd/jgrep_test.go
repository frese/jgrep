package main

import (
	"encoding/json"
	"testing"
)


func TestJgrep(t *testing.T) {

	var source interface{}

	data := []byte(`{"key": "value"}`)
	path := []string{ "key" }

	err := json.Unmarshal(data, &source)
	if err != nil {
		t.Fatal("Cannot Unmarshal json data")
	}
	res := jgrep(source, path)
	out, err := json.Marshal(res)
    if err != nil {
        t.Fatal("Cannot Marshal result")
    }
	if string(out) != "\"value\"" {
		t.Errorf("Got %s but wanted 'value'", string(out))
	}

}