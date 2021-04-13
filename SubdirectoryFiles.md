# Writing Files to Subdirectories

There's a small bug in our implementation where, when trying to write to a
file in a subdirectory such as foo/bar.go, the file gets written to foo instead
of creating the directory.

This should be fairly easy to fix by updating our Output files block to use
a little help from the Go standard library `path/filepath` functions.

We'll get the directory of the string, and if it's not "." (filepath.Dir
returns ".", not "" for empty paths) call `os.MkdirAll` on it in order to create
the directory before creating the file.

```go "Output files"
for filename, codeblock := range files {
	if dir := filepath.Dir(string(filename)); dir != "." {
		if err := os.MkdirAll(dir, 0775); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}

	f, err := os.Create(string(filename))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		continue
	}
	fmt.Fprintf(f, "%s", codeblock.Replace(""))
	// We don't defer this so that it'll get closed before the loop finishes.
	f.Close() 

}
```

```go "main.go imports" +=
"path/filepath"
```


