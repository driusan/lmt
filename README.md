# Literate Markdown Tangle

[lmt](https://github.com/driusan/lmt) is a tool for extracting text from the
code blocks in markdown files.  It allows programmers to write in a [literate programming](https://en.wikipedia.org/wiki/Literate_programming) style using
markdown as the source language.

## Installing lmt

lmt is a self-contained Go program written using the LP paradigm. The source is committed alongside the markdown source of this repository for bootstrapping
purposes.

You require the [Go language](https://golang.org/) if you don't already have it.

To build the tool:

```bash
git clone https://github.com/driusan/lmt
cd lmt
go build
```

This will build the binary named `lmt` for your platform in the current
directory. You can use the `-o $path` argument to `go build` to build
the binary in a different location. (i.e. `go build -o ~/bin/` to put the
binary in `~/bin/`.)

#### A note for Nix(OS) users
This repo also comes with a `shell.nix` file. While an existing version is included, like lmt itself, this is mainly for bootstrapping purposes. To compile it, use `lmt Nix-Shell.md`

## Demo

To observe `lmt` at work, put this file in an empty directory, cd to that
directory, and run `lmt README.md`.  Now look in the directory and you'll see
files extracted from the code blocks alongside this markdown file.  In
literate programming lingo, this extraction is (somewhat counterintuitively)
called "tangling." Generating documentation from the source is called
"weaving", and `lmt` leaves that to existing markdown renderers (such as
the GitHub frontend.)

`lmt` is language agnostic. The below demonstration of features is written
in (very trivial) C++ to demonstrate using other languages.

### Tangling into a file.

The markup for the code block below starts with `​```cpp hello.cpp +=`:
    
```cpp hello.cpp +=
<<<copyright>>>
<<<includes>>>

int main() {
    <<<body of main>>>
}
```

The header says 3 things:

1. `cpp`: the code block is written in C++. In the rendered markdown output, that
   affects syntax highlighting, to lmt it means that language-appropriate
   pragma directives will be added so that when debugging the extracted code,
   your debugger will show you the line in the original markdown source file.
   (If you don't want this effect, you can just use an unrecognized language
   name like `cxx`).
2. `hello.cpp`: The code block will be written to the file `hello.cpp`.
3. `+=`: The code block will be appended to the most recent code block
   defining that file, rather than overwriting its content.  Since we haven't
   written anything to `hello.cpp` yet, the effect is the same, but this
   demonstrates the ability to use it. 
   

### Macro References

The `<<<`*string*`>>>` sequences in the body of the code block are called
"macro references." An LMT "macro" is just a variable whose value can be
extracted from one or more code blocks, and will be substituted wherever
its name appears in triple angle brackets on a line. There are no arguments
to `lmt` macros.

If we were to run lmt on this file at this point, we would get the warnings:

```
Warning: Block named copyright referenced but not defined.
Warning: Block named includes referenced but not defined.
Warning: Block named body of main referenced but not defined.
```

This allows us to stub in a macro reference whenever we want in our code,
and only later define them in whatever order best fits our prose. When there
are no more warnings, the `hello.cpp` file should build (assuming we didn't
include any syntax or other compiler errors.)

### Macro Content

The markup for the code block below starts with `​```cpp "body of main"`

```cpp "body of main"
std::cout << "Hello, werld!" << std::endl;
```

The double quotes around `body of main` mean that the code block will be
extracted into a macro of that name.  You can see where its value will be
injected into hello.cpp via `<<<body of main>>>`,
[above](#tangling-into-a-file).  Since there's no `+=` at the end of the block's
first line of markup, this code block overwrites any existing value the macro
might already have (but since it has no existing value, it's a wash).

`lmt` uses quotation marks to differentiate between macros and file
destinations. If a name is encased in quotes, it's a macro, if not, it's
a file.

We can later re-define a macro to overwrite it (`​```cpp "body of main"`,
again)

```cpp "body of main"
std::cout << "Hello, world!" << std::endl;
```

`lmt` parses each file passed on the command line in order. The last
definition of a macro will be used for all references to that macro in
other code blocks (including blocks which preceeded it in the source.)

### Appending To A Macro

We can use `#include`s to demonstrate `+=` on macros.  There are two includes in
this program. The markup for the following block starts with `​```cpp
"includes"`, which causes the (empty) value of the `includes` macro to be
overwritten.

```cpp "includes"
#include <iostream>
```

The markup for the next code block, however, starts with `​```cpp "includes" +=`,
which causes the block to be appended to the `includes` macro.

```cpp "includes" +=
#include <numeric>
```

Its value is now:

```cpp
#include <iostream>
#include <numeric>
```

(the code block above is not being tangled).

### Hidden content.

The raw markdown in this file contains a comment containing a code block with a
copyright notice.  It looks a bit like this one:

    <!-- 
    ```cpp "copyright"
    // Copyright 42 BCE not the actual copyright
    ```
    -->

If you're reading the rendered markdown in your browser, you can't see the
*actual* comment, but it still gets tangled into the `copyright` macro, which is
substituted into hello.cpp by the `<<<copyright>>>` macro reference.  This
technique lets you tangle content that you don't want showing up in the
documentation.

<!-- 
```cpp "copyright"
// Copyright 2020 Me, myself, and I
```
-->

### What Tangles and What Doesn't.

We can tangle into a random data file (`​```csv data.csv`)


```csv data.csv
foo, bar, baz,
qix, qux, quux,
```

You need to specify both a language and a destination (macro or file) if
you want the code block tangled:

No language (`​``` bar.txt`—note the space):

``` bar.txt
This doesn't get tangled anywhere
```

No destination, but includes syntax highlighting (`​```cpp`)

```cpp
auto x = "nor does this";
```

But any language string and filename  (`​```arbitrary foo.txt`) will do

```arbitrary foo.txt
This gets tangled
into foo.txt.
```

Running `lmt` on this file at this point should generate the files `data.csv`,
`foo.txt`, and `hello.cpp` with the expected contents and produce no warnings.

## Building lmt from source

While the tangled source of `lmt` is included for bootstrapping purposes,
the markdown is considered the canonical version. The Go source can
be re-extracted with:

```shell
lmt Implementation.md WhitespacePreservation.md SubdirectoryFiles.md LineNumbers.md IndentedBlocks.md
```

If you'd like to read the source, the order of the files and patches were
written is the same as passed on the command line.

 1. [Basic Implementation](Implementation.md)
 2. [Whitespace Preservation](WhitespacePreservation.md)
 3. [Subdirectory Files](SubdirectoryFiles.md)
 4. [Line Numbers](LineNumbers.md)
 5. [Indented Blocks](IndentedBlocks.md)

Small bug fixes can be contributed by modifying the prose and code in
the existing files. Larger features can be included as a patch in a
new file.

## Credits

`lmt` is primarily authored by Dave MacFarlane ([@driusan](https://github.com/driusan/)). Bryan Allred ([@bmallred](https://github.com/bmallred/)) improved the
parsing code to include the metadata in the code block header rather than a
rendered markdown header. [@mek-apelsin](https://github.com/mek-apelsin/)
wrote the patch to include pragmas for line numbers, and Dave Abrahams
([@dabrahams](https://github.com/dabrahams/)) wrote the demo of features in
this README, making it more user-focused (it previously dove straight into
implementation.)
