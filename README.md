jgrep
=====

Simple JSON/YAML grep, usage : jgrep [-options] "path/path/path" [file]

__jgrep__ reads from file or stdin, unmarshal the input weather it's json or yaml, traversing the structure searching
the specified 'path', printing the values of what's been grep'ed in same format as input.

Options:
 - `-h` print help message
 - `-j` output in json format
 - `-y` output in yaml format
 - `-t` output in text format
 - `-s` separator character in text format, default is colon

Each 'path' element can be:
 - 'string' specifying a specific key in an object
 - 'number' specifying an index in an array
 - 'star' specifying all keys in object or all indexes in array
 - 'key=value' select hash'es containing the given key=value field
 - '.' stop processing and ignore everything after this point
 - comma separated list of the above, each will be evaluated and printed comma separated

INSTALL
-------

    brew tap frese/frese; brew install jgrep

TODO
----
* allow regular expressions in searches for keys and values
* lots