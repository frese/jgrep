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
)

func main() {
	var source interface{}
	var paths []string

	helpOpt := flag.Bool("h", false, "Print usage message")
	flag.Parse()
	
	// get the 'path' argument that we want to grep for
	pathArg := flag.Args()

	if len(pathArg) == 0 || *helpOpt {
		fmt.Println("Usage: jgrep 'path/path/...' [file]")
        fmt.Println("Where path is:")
        fmt.Println("- 'string' specifying a particular key in an object")
        fmt.Println("- 'number' specifying an index in an array")
        fmt.Println("- '*' all objects in given hash or array")
        fmt.Println("- 'key=value' selects hash'es containing the given key=value field")
        fmt.Println("- comma separated list of the above, each will be evaluated and printed comma separated")
        fmt.Println("If no file is specified, jgrep reads from stdin.")
        fmt.Println("The value of the last object(s) will be printed to stdout")
		os.Exit(0)
	}

	paths = strings.Split(pathArg[0], "/")

	// check if pathArg[1] is defined, this is filename to read, if not defined read from stdin
	// TODO

	// read json from stdin to buf
	var buf bytes.Buffer
    reader := bufio.NewReader(os.Stdin)

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

	// Unmarshal the json from buffer
	err := json.Unmarshal(buf.Bytes(), &source)
	if err != nil {
		log.Fatalf("Failed to parse json: %s", err)
	}

	// run through the json, hunting after wanted "path"s
	res := jgrep(source, paths)
	js, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		log.Fatalf("Failed to marchal json: %s", err)
	}
	fmt.Println(trimQuotes(string(js)))
}

func jgrep(src interface{}, paths []string) interface{} {
	
	var res []interface{}

	// fmt.Printf("DEBUG: interface: %v - paths: %v\n", src, paths)
	// check that we have any path's left, 
	// if not, print whats left of src in json format and return
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
	case strings.Compare(p1, "*") == 0:
		switch t:= src.(type) {
		case []interface{}:
			for _, element := range t {
				res = append(res, jgrep(element, paths[1:]))
			}
		case map[string]interface{}:
			for _, element := range t {
				res = append(res, jgrep(element, paths[1:]))
			}
		default:
			log.Fatalf("Expected an array, got %v", src)
		}
		return res

	//------------------------------------------
	case strings.Contains(p1, "="):

		parts := strings.Split(p1, "=")
		if len(parts) != 2 {
			log.Fatalln("Expected simple 'key=value' as expression")
		}
		k := parts[0]
		v := parts[1]

		switch t:= src.(type) {
		case []interface{}:
			for _, element := range t {
				switch m := element.(type) {
				case map[string]interface{}:
					if v == fmt.Sprintf("%v", m[k]) {
						res = append(res, jgrep(element, paths[1:]))
					}
				}
			}
			return res
		default:
			log.Fatalf("Expected an array of maps, got %v", src)
		}

	//------------------------------------------
	case IsNumeric(p1):
		switch t:= src.(type) {
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
		switch t:= src.(type) {
		case map[string]interface{}:
			return jgrep(t[p1], paths[1:])
		default:
			log.Fatalf("Expected map of strings, got %v", src)
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