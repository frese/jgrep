jgrep (go-lang version)
=======================

Simple JSON grep, usage : jgrep "path/path/path" [filename]

__jgrep__ reads from file or stdin, traversing the json structure searching
the specified 'path', printing the value of the last key.

Each 'path' element can be:
* a string specifying a key in an object
* a number specifying an index in an array
* a 'star' specifying all keys in object or indexes in array
* a comma separated list of keys. The combined output is comma separated.
* an 'equal' key=value to select a specific key

TODO
----
* allow regular expressions in searches for keys and values
