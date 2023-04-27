# Module - "filepath"

```golang
filepath := import("filepath")
```

## Functions

### join
```golang
join(elem ...string) => string
```
Joins any number of path elements into a single path.

#### Example
```golang
fmt := import("fmt")
filepath := import("filepath")

// On Unix
fmt.println(filepath.join("a", "b", "c"))
fmt.println(filepath.join("a", "b/c"))
fmt.println(filepath.join("a/b", "c"))
fmt.println(filepath.join("a/b", "/c"))
fmt.println(filepath.join("a/b", "../../../xyz"))
```
```
Output:
a/b/c
a/b/c
a/b/c
a/b/c
../xyz
```

### file_exists
```golang
file_exists(path string) => bool
```
Returns whether a file exists at the specified path.

#### Example
```golang
fmt := import("fmt")
filepath := import("filepath")

fmt.println(filepath.file_exists("/etc/passwd"))
fmt.println(filepath.file_exists("/etc/not-a-file"))
fmt.println(filepath.file_exists("/etc"))
```
```
Output:
true
false
false
```

### dir_exists
```golang
dir_exists(path string) => bool
```
Returns whether a directory exists at the specified path.

#### Example
```golang
fmt := import("fmt")
filepath := import("filepath")

fmt.println(filepath.dir_exists("/etc/passwd"))
fmt.println(filepath.dir_exists("/etc/not-a-file"))
fmt.println(filepath.dir_exists("/etc"))
```
```
Output:
false
false
true
```

### base
```golang
base(path string) => string
```
Returns the last element of the path.

### dir
```golang
dir(path string) => string
```
Returns all but the last element of path, typically the path's directory.

### abs
```golang
abs(path string) => string/error
```
Returns an absolute representation of path.

### ext
```golang
ext(path string) => string
```
Returns the file name extension used by path.

### glob
```golang
glob(pattern string) []string/error
glob(pattern string, exclude_re string) []string/error
```
Returns the names of all files matching the shell pattern or nil if there is no matching file. Optionally can specify a regex string of the files to exclude.

### from_slash
```golang
from_slash(path string) string
```
Returns the result of replacing each slash ('/') character in path with a separator character.
