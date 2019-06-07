# Linenumber directives

Most languages have linenumber directives. They are used for generated code so
debuggers and error messages are able to tell which line from the source file
is responsible instead of the line from the intermediate representation. Go
uses '//line [filename]:[linenumber]', where [XXX] denotes an optional
parameter. Many languages from the C-family instead use the format `#line
lineno ["filename"]`.

This is quite useful, since `go build`, `go run` etc will provide references to
the markdown file with the bug (well, at least in theory). IF we want to debug
the intermediate file, a reference above the specific code is always readily
avialible a quick right click (Acme) or `gF` away (vim).

If we want to implement this for lmt, we need to change quite a few things, the
most obvious is the internal representation of a line, which needs quite a lot
more metadata such as the filename, the line number and the language of the
codeblock.

We need to change CodeBlock to a slice of CodeLines. For every line we don't
only record the text but the name of the source file, the linenumber and the
language (for different line directives).

```go "global block variables"
type File string
type CodeBlock []CodeLine
type BlockName string
type language string
<<<Codeline type definition>>>

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock
```

```go "Codeline type definition"
type CodeLine struct {
	text   string
	file   File
	lang   language
	number int
}
```

Handling a block line is now not as simple as just joining two strings. We
start appending CodeLines into CodeBlocks.

```go "Handle block line"
block = append(block, line)
```

This cascades. The CodeBlocks in the maps files and blocks can't be joined, but
have to be appended. We take a few moments to clean up the emptying of metadata
per line, i.e. if a block is ending we can unset inBlock.

```go "Handle block ending"
inBlock = false
// Update the files map if it's a file.
if fname != "" {
	if appending {
		files[fname] = append(files[fname], block...)
	} else {
		files[fname] = block
	}
}

// Update the named block map if it's a named block.
if bname != "" {
	if appending {
		blocks[bname] = append(blocks[bname], block...)
	} else {
		blocks[bname] = block
	}
}
```

To get this to work we need to start using CodeLine as line in our
implementation of the file processing. We also should to save the filename
directly when opening the file. To simplify we break out our variables from the
implementation of file processing.

```go "process file implementation variables"
scanner := bufio.NewReader(r)
var err error

var line CodeLine
line.file = File(inputfilename)

var inBlock, appending bool
var bname BlockName
var fname File
var block CodeBlock
```

When processing our files we need to increment the file line counter for every
line read.

```go "process file implementation"
<<<process file implementation variables>>>
for {
	line.number++
	line.text, err = scanner.ReadString('\n')
	switch err {
	case io.EOF:
		return nil
	case nil:
		// Nothing special
	default:
		return err
	}
	<<<Handle file line>>>
}
```

Sadly, there is no indicator from the Reader about which file we are reading so
we send it along with ProcessFile.

```go "ProcessFile Declaration"
// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader, inputfilename string) error {
	<<<process file implementation>>>
}
```

Of course we need to call ProcessFile with the filename too.

```go "Open and process file"
f, err := os.Open(file)
if err != nil {
	fmt.Fprintln(os.Stderr, "error: ", err)
	continue
}

if err := ProcessFile(f, file); err != nil {
	fmt.Fprintln(os.Stderr, "error: ", err)
}
// Don't defer since we're in a loop, we don't want to wait until the function
// exits.
f.Close()
```

When testing if the line starts with '```' we need to look at line.text,
instead of directly at line as before.

```go "Handle file line"
if inBlock {
	if line.text == "```\n" {
		<<<Handle block ending>>>
		continue
	}
	<<<Handle block line>>>
	continue
}
<<<Handle nonblock line>>>
```

Handle nonblock line is actually longer than needed.  We only need to check if
it is a block start now, since markdown headers aren't handled as code block
descriptors. Lets clean it up.

```go "Handle nonblock line"
<<<Check block start>>>
```

We check if a block starts by reading the first few characters (if the line is
long enough to hold them). If we are in the start of a new block, we reset
block (the variable we save every new block in), set inBlock to true and then
check the headers.

```go "Check block start"
if len(line.text) >= 3 && (line.text[0:3] == "```") {
	inBlock = true
	// We were outside of a block, so just blindly reset it.
	block = make(CodeBlock, 0)
	<<<Check block header>>>
}
```

We parse the header for a new variable: the language of the codeblock.

```go "Check block header"
fname, bname, appending, line.lang = parseHeader(line.text)
```

Parseheader must actually extract the new value, this quickly becomes rather
noisy, since both declaration and implementation changes. Worst part is that
the regexp changes too. We still probably should try looking at naming the
variables in the regexp.

```go "ParseHeader Declaration"
func parseHeader(line string) (File, BlockName, bool, language) {
	line = strings.TrimSpace(line)
	<<<parseHeader implementation>>>
}
```

```go "parseHeader implementation"
var matches []string
if matches = namedBlockRe.FindStringSubmatch(line); matches != nil {
	return "", BlockName(matches[2]), (matches[3] == "+="), language(matches[1])
}
if matches = fileBlockRe.FindStringSubmatch(line); matches != nil {
	return File(matches[2]), "", (matches[3] == "+="), language(matches[1])
}
return "", "", false, ""
```

```go "Namedblock Regex"
namedBlockRe = regexp.MustCompile("^`{3,}\\s?(\\w*)\\s*\"(.+)\"\\s*([+][=])?$")
```

```go "Fileblock Regex"
fileBlockRe = regexp.MustCompile("^`{3,}\\s?(\\w+)\\s+([\\w\\.\\-\\/]+)\\s*([+][=])?$")
```

The (actual) replacement in Replace also joins strings, we need to append to
ret instead. The returned codeblock from the map blocks are replaced with the
new prefix and "expanded" (...) before appended.

```go "Lookup replacement and add to ret"
bname := BlockName(matches[2])
if val, ok := blocks[bname]; ok {
	ret = append(ret, val.Replace(prefix+matches[1])...)
} else {
	fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
	ret = append(ret, v)
}
```

Even lines which aren't replaced need to be appended instead of joining lines.
While we are at it let's cleanup empty lines when passing through (they should
not have prefix).

```go "Handle replace line"
matches := replaceRe.FindStringSubmatch(line)
if matches == nil {
	if v.text != "\n" {
		v.text = prefix + v.text
	}
	ret = append(ret, v)
	continue
}
<<<Lookup replacement and add to ret>>>
```

Finally something we can replace with a simpler implementation. Instead of
reading line by line from c (a CodeBlock, earlier a long string with newlines),
we're now able to use range. The string `line` is still analyzed, "de-macro-ed"
and prefixed.

```go "Replace codeblock implementation"
var line string
for _, v := range c {
	line = v.text
	<<<Handle replace line>>>
}
return
```

We need to use all the information we gathered, lets add a "Finalize"-function
which handles CodeLines and outputs the final strings with the line-directives
we been working for.

```go "other functions" +=
<<<Finalize Declaration>>>
```

Let us implement a method which acts upon a codeblock and returns a string
containing the ready to be used textual representation of the code (or other
files). We check if the filename is the same as the previous line, and if the
linenumber has been increased by exactly one. If any of those have changed it
is because the source for this line are not the same as the previous and we
interject a "line directive". We know of two variants of these: Go and the C
variant. Lastly we save state, so we have something to compare to on the next
line.

```go "Finalize Declaration"

// Finalize reads the textual lines from CodeBlocks and (if needed) prepend a
// notice about "unexpected" filename or line changes, which is extracted from
// the contained CodeLines. The result is a string with newlines ready to be
// pasted into a file.
func (block CodeBlock) Finalize() (ret string) {
	var prev CodeLine
	var formatstring string

	for _, current := range block {
		if prev.number+1 != current.number || prev.file != current.file {
			switch current.lang {
			<<<Format strings for languages>>>
			}
			if formatstring != "" {
				ret += fmt.Sprintf(formatstring, current.number, current.file)
			}
		}
		ret += current.text
		prev = current
	}
	return
}
```

By putting the format strings for different variables in its own codeblock it
should be simpler to add more later on.

```go "Format strings for languages"
case "bash", "shell", "sh", "perl":
	formatstring = "#line %v \"%v\"\n"
case "go", "golang":
	formatstring = "//line %[2]v:%[1]v\n"
case "C", "c":
	formatstring = "#line %v \"%v\"\n"
```

And finally, lets use our Finalize on the Replace-d codeblock, right before we
print it out in the file.

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
	fmt.Fprintf(f, "%s", codeblock.Replace("").Finalize())
	// We don't defer this so that it'll get closed before the loop finishes.
	f.Close()
}
```
