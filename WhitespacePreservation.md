## Whitespace Preservation

Someone requested that code blocks preserve their indentation level when expanded.
It seems like a reasonable request, that should make it possible to do things
like use Python.

To do that, we'll have to group the whitespace token in the Replace Regex so
that we know the indentation level to preserve.

```go "Replace Regex"
replaceRe = regexp.MustCompile(`^([\s]*)<<<(.+)>>>[\s]*$`)
```

Then update the reference to it when looking up the replacement, since the
index will have changed. We'll also want to pass the whitespace token we
just found to Replace() to use as a prefix (in addition to the current prefix.)

(We probably should have used a named grouping, but for now we'll just change
the index to keep our changes simple.)

```go "Lookup replacement and add to ret"
bname := BlockName(matches[2])
if val, ok := blocks[bname]; ok {
	ret += val.Replace(prefix + matches[1])
} else {
	fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
	ret += CodeBlock(line)
}
```

We'll have to update our function signature too.

```go "Replace Declaration"
// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace(prefix string) (ret CodeBlock) {
	<<<Replace codeblock implementation>>>
}
```

and make sure we pass the empty string as the starting prefix when outputting
a new file.

```go "Output files"
for filename, codeblock := range files {
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

Then all that's left is, when replacing lines, we'll have to include the prefix
that we just passed, so that the lines get indented. But only if the line isn't
empty.

```go "Handle replace line"
matches := replaceRe.FindStringSubmatch(line)
if matches == nil {
	if line != "\n" {
		ret += CodeBlock(prefix)
	}
	ret += CodeBlock(line)
	continue
}
<<<Lookup replacement and add to ret>>>
```

And now our generated code looks prettier without needing an autoformatter
like `go fmt`.
