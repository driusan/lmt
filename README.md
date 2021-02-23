# Literate Markdown For C++ Programmers

[lmt](https://github.com/driusan/lmt) is a tool for extracting text from the
code blocks in markdown files.  This file demonstrates all the lmt features in a
C++-centric way.

## Installing lmt

First, install the go language if you don't have it (homebrew: `brew install
go`).

Then build the tool.  The following assumes you have a `~/bin` directory in your
`PATH`:

```bash
git clone https://github.com/driusan/lmt
cd lmt
go build -o ~/bin
```

## Demo

To observe `lmt` at work, put this file in an empty directory, cd to that
directory, and `lmt Literate.md`.  Now look in the directory and you'll see
extracted files extracted from the code blocks alongside this markdown file.  In
literate programming lingo, this extraction is (somewhat counterintuitively)
called “tangling.”

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
   `#line` directives will be added so that when debugging the extracted code,
   your debugger will show you the line in the original source markdown file.
   (If you don't want this effect, just use an unrecognized language name like
   `cxx`).
2. `hello.cpp`: The code block will be written to the file `hello.cpp`.
3. `+=`: The code block will be appended to that file, rather than overwriting
   its content.  Since we haven't written anything to `hello.cpp` yet, the
   effect is the same, but since overwriting the code you've already extracted
   is kind of a nice case, to enable developing examples like those in `lmt`'s
   own source, you might want to use `+=` by default.
   

### Macro References

The `<<<`*string*`>>>` sequences in the body of the code block are called
“macro references.” An LMT “macro” is just a variable whose value can be extracted
from one or more code blocks, and will be substituted wherever its name appears
in triple angle brackets.

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

You need to specify both a file type and a destination (macro or file)if you
want the code block tangled:

No file type (`​``` bar.txt`—note the space):
``` bar.txt
This doesn't get tangled anywhere
```

No destination (`​```cpp`):
```cpp
auto x = "nor does this";
```

But any file type string and filename  (`​```arbitrary foo.txt`) will do

```arbitrary foo.txt
This gets tangled
into foo.txt.
```
