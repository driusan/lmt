# lmt - literate markdown tangle

This README describes a tangle program for a literate programming style
where the code is weaved into markdown code blocks. There is no corresponding
weave, because the markdown itself can already be read as the documentation,
either through a text editor or through an online system that already renders
markdown such as GitHub.

## Why?

[Literate programming](https://en.wikipedia.org/wiki/Literate_programming) is a
style of programming where, instead of directly writing source code, the programmer
writes their reasoning in human prose, and intersperses fragments of code which
can be extracted into the compilable source code with one tool (called "tangle"),
and conversely can be converted to a human readable document explaining the code
with another (called "weave").

[Markdown](http://daringfireball.net/projects/markdown/syntax) is a plaintextish
format popular with programmers. It's simple,  easy and already has support
for embedding code blocks using triple backticks (```), mostly for the purposes
of syntax highlighting in documentation.

The existing literate programming for markdown tools seem too heavyweight for me,
and too much like learning a new domain specific language which defeats the
purpose of using markdown.

I started tangling [the shell](https://github.com/driusan/dsh) that I was writing
to experiment with literate programming using copy and paste. It works, but is
cumbersome. This is a tool to automate that process.

It's written in Go, because the Go tooling (notably `go fmt`) lends itself well
to writing in this paradigm.

## Syntax

To be useful for literate programming code blocks need a few features that don't
exist in standard markdown:

1. The ability to embed macros, which will get expanded upon tangle.
2. The ability to denote code blocks as the macro to be expanded when referenced.
3. The ability to either append to or replace code blocks/macros, so that we can
   expand on our train of thought incrementally.
4. The ability to redirect a code block into a file (while expanding macros.)

Since markdown codeblocks will already let you specify the language of the block
for syntax highlighting purposes by naming the language after the three backticks,
my first thought was to put the file/codeblock name on the same line, after the
language name.

For a convention, we'll say that a string with quotations denotes the name of a
code block, and a string without quotations denotes a filename to put the code
block into. If a code block header ends in `+=` it'll mean "append to the named
code block", otherwise it'll mean "create or replace the existing code block."
We'll use a line inside of a code block containing nothing but a title inside
`<<<` and `>>>` (with optional whitespace) as a macro to expand, because it's a
convention that's unlikely to be used otherwise inside of source code in any
language.

### Implementation/Example.

The above paragraph fully defines our spec. So, an example of a file code block
might look like this:

```go main.go
package main

import (
	<<<main.go imports>>>
)

<<<global variables>>>

<<<other functions>>>

func main() {
	<<<main implementation>>>
}
```

For our implementation, we'll need to parse the markdown file (which file? We'll
use the arguments from the command line) one line at a time, starting from the
top to ensure we replace code blocks in the right order. If there are multiple
files, we'll process them in the order they were passed on the command line.

For now, we don't need to process any command line arguments, we'll just assume
everything passed is a file.

So an example of a named code block is like this:

```go "main implementation"
files := os.Args
for _, file := range files {
	<<<process file>>>
}
```

How do we process a file? We'll need to keep 2 maps: one for named macros, and
one for file output content. We won't do any expansion until all the files have
been processed, because a block might refer to another block that either hasn't
been defined yet, or later has its definition changed. Let's define our maps,
define a stub of a `process file` function, and redefine our `main implementation`
to take that into account.

Our maps, with some types defined for good measure:

```go "global variables"
type File string
type CodeBlock string
type BlockName string

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock
```

Our ProcessFile function:

```go "other functions"
// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader) error {
	<<<process file implementation>>>
}
```

And our new main:

```go "main implementation"
<<<Initialize>>>

// os.Args[0] is the command name, "lmt". We don't want to process it.
for _, file := range os.Args[1:] {
	<<<Open and process file>>>

}
<<<Output files>>>
```

We used a few packages, so let's import them before declaring the blocks we
just used.

```go "main.go imports"
"fmt"
"os"
"io"
```

Initializing the maps is pretty straight forward:

```go "Initialize"
// Initialize the maps
blocks = make(map[BlockName]CodeBlock)
files = make(map[File]CodeBlock)
```

As is opening the files, since we already declared the ProcessFile function and
we just need to open the file to turn it into an `io.Reader`:

```go "Open and process file"
f, err := os.Open(file)
if err != nil {
	fmt.Fprintln(os.Stderr, "error: ", err)
	continue
}

if err := ProcessFile(f); err != nil {
	fmt.Fprintln(os.Stderr, "error: ", err)
}
// Don't defer since we're in a loop, we don't want to wait until the function
// exits.
f.Close()
```

### Processing Files

Now that we've got the obvious overhead out of the way, we need to begin
implementing the code which parses a file.

We'll start by scanning each line. The Go `bufio` package has a Reader which
has a `ReadString` method that will stop at a delimiter (in our case, '\n')

We can do use this bufio Reader to iterate through lines like so:

```go "process file implementation"
scanner := bufio.NewReader(r)
var err error
var line string
for {
	line, err = scanner.ReadString('\n')
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

We'll need to import the `bufio` package which we just used too:

```go "main.go imports" +=
"bufio"
```

How do we handle a line? We'll need to keep track of a little state:

1. Are we in a code block?
2. If so, what name or file is it for?
3. Are we ending a code block? If so, update the map (either replace or append.)

So let's add a little state to our implementation:

```go "process file implementation"
scanner := bufio.NewReader(r)
var err error
var line string

var inBlock, appending bool
var bname BlockName
var fname File
var block CodeBlock

for {
	line, err = scanner.ReadString('\n')
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

We'll replace all of the variables with their zero value when we're not in a
block.

The flow of handling a line will be something like:

```go "Handle file line"
if inBlock {
	if line == "```\n" {
		<<<Handle block ending>>>
		continue
	} else {
		<<<Handle block line>>>
	}
} else {
	<<<Handle nonblock line>>>
}
```

Handling a code block line is easy, we just add it to the `block` if it's not
a block ending, and update the map/reset all the variables if it is.

```go "Handle block line"
block += CodeBlock(line)
```

```go "Handle block ending"
// Update the files map if it's a file.
if fname != "" {
	if appending {
		files[fname] += block
	} else {
		files[fname] = block
	}
}

// Update the named block map if it's a named block.
if bname != "" {
	if appending {
		blocks[bname] += block
	} else {
		blocks[bname] = block
	}
}

<<<Reset block flags>>>
```

```go "Reset block flags"
inBlock = false
appending = false
bname = ""
fname = ""
block = ""
```

#### Processing Non-Block lines

Processing non-block lines is easy, and we don't have to do anything since we
are only concerned with code blocks.
we don't need to care and can just reset the flags.
Otherwise, for triple backticks, we can just check the first three characters
of the line (we don't care if there's a language specified or not).

```go "Handle nonblock line"
if line == "" {
	continue
}

switch line[0] {
case '`':
	<<<Check block start>>>
default:
	<<<Reset block flags>>>
}
```

When a code block is reached we will need to reset the flags and parse the line
for the following information:

 - a filename
 - a block name/label
 - an append flag

```go "Check block start"
if len(line) >= 3 && line[0:3] == "```" {
	inBlock = true
	<<<Check block header>>>
}
```

#### Parsing Headers With a Regex

Parsing headers is a little more difficult, but shouldn't be too hard with
a regular expression. There's four potential components:

 1. 3 or more '`' characters. We don't care how many there are.
 2. 0 or more non-whitespace characters, which will may be the language type.
 3. 0 or more alphanumeric characters, which can be a file name.
 4. 0 or 1 string enclosed in quotation marks.
 5. It may or may not end in `+=`.

So the regex will look something like ```/^(`+)([a-zA-Z0-9\.]*)("[.*]"){0,1}(+=){0,1}$/```
(there are more characters that might be in a file name, but to keep the regex simple
we'll just assume letters, numbers, and dots.)

That regex is already starting to look hairy, so instead let's split it up into
two: one for checking if it's a named block, and if that fails one for checking
if it's a file name. It means we can't have a block which is *both* a named
block and *also* goes into a filename, but that's probably not a very useful
case and can always be done with two blocks (one named, and a file which only
contains a macro expanding to the named block.)

In fact, we'll put the whole thing into a function to make it easier to debug
and write tests if we want to.

```go "Check block header"
fname, bname, appending = parseHeader(line)
// We're outside of a block, so just blindly reset it.
block = ""
```

Then we need to define our parseHeader function:

```go "other functions" +=
func parseHeader(line string) (File, BlockName, bool) {
	line = strings.TrimSpace(line)
	<<<parseHeader implementation>>>
}
```

Our implementation is going to use a regex for a namedBlock, and compare the
line against it, so let's start by importing the regex package.

```go "main.go imports" +=
"regexp"
```

```go "parseHeader implementation"
namedBlockRe := regexp.MustCompile("^([`]+\\s?)[\\w]*[\\s]*\"(.+)\"[\\s]*([+][=])?$")
matches := namedBlockRe.FindStringSubmatch(line)
if matches != nil {
	return "", BlockName(matches[2]), (matches[3] == "+=")
}
<<<Check filename header>>>
return "", "", false
```

There's no reason to constantly be re-compiling the namedBlockRe, we can just
make it global and compile it once on initialization.

```go "global variables" +=
var namedBlockRe *regexp.Regexp
```

```go "Initialize" +=
namedBlockRe = regexp.MustCompile("^([`]+\\s?)[\\w]+[\\s]+\"(.+)\"[\\s]*([+][=])?$")
```

Then our parse implementation without the MustCompile is:

```go "parseHeader implementation"
matches := namedBlockRe.FindStringSubmatch(line)
if matches != nil {
	return "", BlockName(matches[2]), (matches[3] == "+=")
}
<<<Check filename header>>>
return "", "", false
```

Checking a filename header is fairly simple: just make sure there's alphanumeric
characters or dots and no spaces. If it's neither, we can just return the zero
value, since the header must immediately preceed the code block according to our
specification.

This time, we'll just go straight to declaring the regex as a global.

```go "global variables" +=
var fileBlockRe *regexp.Regexp
```

```go "Initialize" +=
fileBlockRe = regexp.MustCompile("^([`]+\\s?)[\\w]+[\\s]+([\\w\\.\\-\\/]+)[\\s]*([+][=])?$")
```

```go "Check filename header"
matches = fileBlockRe.FindStringSubmatch(line)
if matches != nil {
	return File(matches[2]), "", (matches[3] == "+=")
}
```

### Outputting The Files

Now, we've finally finished processing the file, all that remains is going through
the output files that were declared, expanding the macros, and writing them to
disk. Since our files is a `map[File]CodeBlock`, we can define methods on
`CodeBlock` as needed for things like expanding the macros.

Let's start by just ranging through our files map, and assuming there's a method
on code block which does the replacing.

```go "Output files"
for filename, codeblock := range files {
	f, err := os.Create(string(filename))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		continue
	}
	fmt.Fprintf(f, "%s", codeblock.Replace())
	// We don't defer this so that it'll get closed before the loop finishes.
	f.Close()

}
```

Now, we'll have to declare the Replace() method that we just used. The Replace()
will take a codeblock, go through it line by line, check if the current line is
a macro, and if so replace the content (recursively). We can use another regex
to determine if it's a macro line, and we can use a scanner similar to our
markdown line scanner to our previous one,

```go "other functions" +=
<<<Replace Declaration>>>
```

```go "Replace Declaration"
// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace() (ret CodeBlock) {
	<<<Replace codeblock implementation>>>
}
```

```go "Replace codeblock implementation"
scanner := bufio.NewReader(strings.NewReader(string(c)))

for {
	line, err := scanner.ReadString('\n')
	// ReadString will eventually return io.EOF and this will return.
	if err != nil {
		return
	}
	<<<Handle replace line>>>
}
return
```

We'll have to import the strings package we just used to convert our CodeBlock
into an io.Reader:

```go "main.go imports" +=
"strings"
```

Now, our replacement regex should be fairly simple:

```go "global variables" +=
var replaceRe *regexp.Regexp
```

```go "Initialize" +=
<<<Replace Regex>>>
```

```go "Replace Regex"
replaceRe = regexp.MustCompile(`^[\s]*<<<(.+)>>>[\s]*$`)
```

Okay, so let's do the actual line handling. If it doesn't match, add it to `ret`
and go on to the next line. If it matches, look up the part that matched in
blocks and include the replaced CodeBlock from there. (If it doesn't exist,
we'll add the line unexpanded and print a warning.)

```go "Handle replace line"
matches := replaceRe.FindStringSubmatch(line)
if matches == nil {
	ret += CodeBlock(line)
	continue
}
<<<Lookup replacement and add to ret>>>
```

Looking up a replacement is fairly straight forward, since we have a map by the
time this is called.

```go "Lookup replacement and add to ret"
bname := BlockName(matches[1])
if val, ok := blocks[bname]; ok {
	ret += val.Replace()
} else {
	fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
	ret += CodeBlock(line)
}
```

## Fin

And now, our tool is finally done! We've finally implemented our `lmt` tool tangle
tool, and can use it to write other literate markdown style programs with the
same syntax.

The output of running it on itself (included [patches](#patches) and then running `go fmt`)
is in this repo to make it a go-gettable executable for bootstrapping purposes.

To use it after installing it just run, for example

```shell
lmt README.md WhitespacePreservation.md SubdirectoryFiles.md
```

## Patches

 1. [Whitespace Preservation](WhitespacePreservation.md)
 2. [Subdirectory Files](SubdirectoryFiles.md)
