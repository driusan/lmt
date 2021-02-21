# Parsing Indented Blocks

GitHub flavoured markdown supports blocks that are indented,
for instance when a code block is in a list. Our current
implementation does a simple "does the line starts with ```?"
check in order to determine whether a line is the start of
a code block.

We should support code blocks such as:

1. Hello
   ```
   This is code
   ```
2. Point 2

Or

- Hi
  ```
  Still code
  ```
- Unnumbered point


Recall that our existing Check block start line was:

```go
if len(line.text) >= 3 && (line.text[0:3] == "```") {
    inBlock = true
    // We were outside of a block, so just blindly reset it.
    block = make(CodeBlock, 0)
    <<<Check block header>>>
}
```

We could either trim the whitespace on the line before doing the
check, or replace it with a regular expression that handles whitespace.
Trimming would be easier, but note that in the rendered markdown the
whitespace at the start of the line inside of the block is trimmed.
This suggests that we need to keep track of what the whitespace that
preceeded the "```" was on a per-block basis, in order to trim it
while rendering and should use a regex instead.

We only need to keep track of the whitespace prefix for the block
that we're currently in, since code blocks can't be embedded. After
extracting with a regex, we can keep the prefix in a simple variable
and use `strings.TrimPrefix` to trim the whitespace of each individual
line while parsing, rather than modifying our CodeBlock structure and
doing it at output time.

Our block start would check would then look something like:

```go "Check block start"
if matches := blockStartRe.FindStringSubmatch(line.text); matches != nil {
    inBlock = true
    blockPrefix = matches[1]
    line.text = strings.TrimPrefix(line.text, blockPrefix)
    // We were outside of a block, so just blindly reset it.
    block = make(CodeBlock, 0)
    <<<Check block header>>>
}
```

We'll have to define our blockStartRe global:

```go "global variables" +=
var blockStartRe *regexp.Regexp
```

and initialize it along with our other regular expressions. Add 
it to the list of Regexs that are initialized:

```go "Initialize" +=
<<<Block Start Regex>>>
```

And initialize it. Any whitespace at the start of the
line, followed by "```" should match.

```go "Block Start Regex"
blockStartRe = regexp.MustCompile("^([\\s]*)```")
```

We'll have to declare our blockPrefix at the start of
our "process file implementation variables" block, so that it's
in scope both while parsing a block (`inBlock` is true)
and looking for a block start (the code we just wrote).

```go "process file implementation variables" +=
var blockPrefix string
```

Now we've handled starting a block and determining the
whitespace prefix, we need to handle the `inBlock`
case. If we strip the blockPrefix from every line as we
parse it, the rest of our code shouldn't have to change.

Recall that our previous implementation was:

```go
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

We'll strip the prefix from line.text at the start of inBlock,
and the rest of our code can go on pretending it was a non-indented
block.

```go "Handle file line"
if inBlock {
    line.text = strings.TrimPrefix(line.text, blockPrefix)
    if line.text == "```\n" {
        <<<Handle block ending>>>
        continue
    }
    <<<Handle block line>>>
    continue
}
<<<Handle nonblock line>>>
```

With these changes, `lmt` should now be able to handle indented code
blocks.
