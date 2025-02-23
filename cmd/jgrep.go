package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func main() {
	var source interface{}
	var paths []string
	var outputFormat string

	outputFormat = ""
	helpOpt := flag.Bool("h", false, "Print usage message")
	jsonOut := flag.Bool("j", false, "Output in json format")
	yamlOut := flag.Bool("y", false, "Output in yaml format")
	textOut := flag.Bool("t", false, "Output in text format")
	separator := flag.String("s", ":", "Separator character for text format")
	flag.Parse()

	// get the 'path' argument that we want to grep for
	pathArg := flag.Args()

	if len(pathArg) == 0 || *helpOpt {
		fmt.Println("Usage: jgrep [-options] 'path/path/...' [file]")
		fmt.Println("Options:")
		fmt.Println("  -j json output")
		fmt.Println("  -y yaml output")
		fmt.Println("  -t text output")
		fmt.Println("  -s separator character for text output, default is ':'")
		fmt.Println("  -h print help")
		fmt.Println("Where path is:")
		fmt.Println("- 'string' specifying a particular key in an object")
		fmt.Println("- 'number' specifying an index in an array")
		fmt.Println("- '*' all keys or values in given hash or all entries in an array")
		fmt.Println("- '.' stops and ignore everything after this point")
		fmt.Println("- 'key=value' selects hash'es containing the given key=value field")
		fmt.Println("- comma separated list of the above, each will be evaluated and printed comma separated")
		fmt.Println("If no file is specified, jgrep reads from stdin.")
		fmt.Println("The value of the last object(s) will be printed to stdout")
		os.Exit(0)
	}

	paths = strings.Split(pathArg[0], "/")
	var buf bytes.Buffer
	var reader bufio.Reader

	// check if pathArg[1] is defined, this is filename to read, if not defined read from stdin
	if len(pathArg) > 1 {
		f, err := os.Open(pathArg[1])
		if err != nil {
			log.Fatal("Cannot open file.")
		}
		reader = *bufio.NewReader(f)
		defer f.Close()
	} else {
		// read json/yaml from stdin to buf
		reader = *bufio.NewReader(os.Stdin)
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				buf.WriteString(line)
				break // end of the input
			} else {
				fmt.Println(err.Error())
				os.Exit(1) // something bad happened
			}
		}
		buf.WriteString(line)
	}

	// Unmarshal the json/yaml from buffer
	err := json.Unmarshal(buf.Bytes(), &source)
	if err != nil {
		err := Unmarshal(buf.Bytes(), &source)
		if err != nil {
			log.Fatalf("Failed to parse input as either json or yaml: %s", err)
		}
		outputFormat = "yaml" // set default output format
	} else {
		outputFormat = "json" // set default output format
	}

	// overwrite output format from flags
	if *jsonOut {
		outputFormat = "json"
	}
	if *yamlOut {
		outputFormat = "yaml"
	}
	if *textOut {
		outputFormat = "text"
	}

	// run through the json, hunting after wanted "path"s
	res := jgrep(source, paths)

	// Write the results
	switch outputFormat {
	case "text":
		textOutput(res, "", *separator)
	case "json":
		out, err := json.MarshalIndent(res, "", " ")
		if err != nil {
			log.Fatalf("Failed to marchal json: %s", err)
		}
		fmt.Println(trimQuotes(string(out)))
	case "yaml":
		out, err := yaml.Marshal(res)
		if err != nil {
			log.Fatalf("Failed to marchal yaml: %s", err)
		}
		fmt.Println(trimQuotes(string(out)))
	}
}

func jgrep(src interface{}, paths []string) interface{} {
	var res []interface{}

	// fmt.Printf("DEBUG: interface: %v - paths: %v\n", src, paths)
	// check that we have any path's left, if not, return whats left of src
	if len(paths) == 0 {
		return src
	}

	// lets work on the first path's
	p1 := paths[0]
	switch {

	//------------------------------------------
	case strings.Contains(p1, ","):
		ps := strings.Split(p1, ",")
		for _, part := range ps {
			res = append(res, jgrep(src, append([]string{part}, paths[1:]...)))
		}
		return res

	//------------------------------------------
	case strings.Compare(p1, ".") == 0:
		if len(paths) != 1 {
			log.Fatalln("The path expression '.' needs to be the last.")
		}
		return nil

	//------------------------------------------
	case strings.Compare(p1, "*") == 0:
		switch t := src.(type) {
		case []interface{}:
			var result []interface{}
			for _, element := range t {
				res := jgrep(element, paths[1:])
				if res != nil {
					result = append(result, res)
				}
			}
			return result
		case map[string]interface{}:
			var result []interface{}
			for _, element := range t {
				res := jgrep(element, paths[1:])
				if res != nil {
					result = append(result, res)
				}
			}
			return result
		default:
			return src
		}

	//------------------------------------------
	case strings.Contains(p1, "="):

		parts := strings.Split(p1, "=")
		if len(parts) != 2 {
			log.Fatalln("Expected simple 'key=value' as expression")
		}
		k := parts[0]
		v := parts[1]

		switch t := src.(type) {
		case []interface{}:
			for _, element := range t {
				switch m := element.(type) {
				case map[string]interface{}:
					if v == fmt.Sprintf("%v", m[k]) {
						res = append(res, jgrep(m, paths[1:]))
					}
				}
			}
			return res
		default:
			log.Fatalf("Expected an array of maps, got %v", src)
		}

	//------------------------------------------
	case IsNumeric(p1):
		switch t := src.(type) {
		case []interface{}:
			int2, err := strconv.Atoi(p1)
			if err != nil {
				log.Fatalln("Expected integer")
			}
			return jgrep(t[int2], paths[1:])
		default:
			log.Fatalf("Expected an array, got %v", src)
		}

	//------------------------------------------
	default:
		switch t := src.(type) {
		case map[string]interface{}:
			return jgrep(t[p1], paths[1:])
		default:
			log.Fatalf("Expected map of strings, got %+v", src)
		}
	}
	return nil
}

// IsNumeric returns if a string can be interpreted as numeric
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func textOutput(src interface{}, prefix string, separator string) {
	switch t := src.(type) {
	case []interface{}:
		for _, element := range t {
			textOutput(element, prefix, separator)
		}
	case map[string]interface{}:
		for k, v := range t {
			var newprefix string
			if prefix == "" {
				newprefix = k
			} else {
				newprefix = prefix + separator + k
			}
			textOutput(v, newprefix, separator)
		}
	case nil:
		fmt.Printf("%v\n", prefix)
	default:
		if prefix == "" {
			fmt.Printf("%v\n", t)
		} else {
			fmt.Printf("%v%s%v\n", prefix, separator, t)
		}
	}
}

// Unmarshal YAML to map[string]interface{} instead of map[interface{}]interface{}.
func Unmarshal(in []byte, out interface{}) error {
	var res interface{}

	if err := yaml.Unmarshal(in, &res); err != nil {
		return err
	}
	*out.(*interface{}) = cleanupMapValue(res)

	return nil
}

func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
