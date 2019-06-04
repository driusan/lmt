# Expand Markup

lmt has chosen to be very selective about the markup it handles. By selecting
\<\<\<CODEBLOCKNAME>>> as a markup for insertion macros it has chosen well.
Sadly it is not perfect, but what is? A common pain point is using gofmt on
codeblocks and ending up with garbage. By expanding the macro markup with the
option to use //\<CODEBLOCKNAME>>> we are now able to gofmt our codeblocks
before passing through gofmt and ending up with perfect (well...) go code which
`go fmt` loves and respects.

A second pain point is the inability to embed ``` in codeblocks, at least in a
way which most or any other markdown interpreters accepts. lmt has chosen to
start a code block with three (or more) backticks (by first checking if the
first character in a line is a backtick, then if three of them are and last by
removing all backticks at the beginning of a line, but at least one), and
ending a codeblock with a line with exactly three backticks. The original
markdown specs does not have code fences at all, but expect the writer to
intend the code with at least four spaces, lmt does not handle this, and it is
not a common enough occurance to even consider for lmt. [GFM] does state "The
content of the code block consists of all subsequent lines, until a closing
code fence of the same type as the code block began with (backticks or tildes),
and with at least as many backticks or tildes as the opening code fence" and "A
fenced code block begins with a code fence, indented no more than three
spaces".  [CommonMark] does seem to be originator for the GFM description.
[Pandoc] make a case for "These begin with a row of three or more tildes (~)
and end with a row of tildes that must be at least as long as the starting row"
(but accepts backticks all the same in the next section). And [Markdown Extra]
fills in with "Fenced code blocks are like Markdown’s regular code blocks,
except that they’re not indented and instead rely on start and end fence lines
to delimit the code block. The code block starts with a line containing three
or more tilde ~ characters, and ends with the first line with the same number
of tilde ~" (but accepts backticks in the next section). Finally
[MultiMarkdown] is the outlier "These code blocks should begin with 3 to 5
backticks, an optional language specifier (if using a syntax highlighter), and
should end with the same number of backticks you started with".

[GFM]: https://github.github.com/gfm/#fenced-code-blocks
[CommonMark]: https://spec.commonmark.org/0.29/#fenced-code-blocks
[Pandoc]: https://pandoc.org/MANUAL.html#fenced-code-blocks
[Markdown Extra]: https://michelf.ca/projects/php-markdown/extra/#fenced-code-blocks
[MultiMarkdown]: https://fletcher.github.io/MultiMarkdown-4/syntax

It does seem to make sense to expand (or at least "correct") the handling of
codeblocks. It seems we can sum it up with:

- the character option of tildes and/or backticks
- to accept intendation
- and if so how much
- the number of characters
- should the block end with the same number of characters or at least the same?

In short this proposal says both characters, allow intendation, three or more
starting, and at least the same at the end. It does seem to break as few as
possible of the other markdown implementations and should not break lmt-users
current files, while still allowing for having code blocks in markdown
formatted documentation code blocks (Inception). it does not touch upon the
format of the "info string" following the starting code fence. It is my solemn
belief this should hurt nobody and help somebody.

As of now, these are our Regex implementations.

```go "Namedblock Regex"
namedBlockRe = regexp.MustCompile("^`{3,}\\s?(\\w*)\\s*\"(.+)\"\\s*([+][=])?$")
```

```go "Fileblock Regex"
fileBlockRe = regexp.MustCompile("^`{3,}\\s?(\\w+)\\s+([\\w\\.\\-\\/]+)\\s*([+][=])?$")
```

```go "Replace Regex"
replaceRe = regexp.MustCompile(`^([\s]*)<<<(.+)>>>[\s]*$`)
```

Lets replace them one by one with something which implements the wishes above.

Replace seems easiest. It changes `PREFIX<<<NAME>>>optionalspaces` to also
allowing `PREFIX//<NAME>>>optionalspaces`, where capital letters are named
capturing groups. We have decided NOT to capture the selection of <<< or //<
and not to allow for ending with //>.

```go "Replace Regex"
replaceRe = regexp.MustCompile(`^(?P<prefix>\s*)(?:<<|//)<(?P<name>.+)>>>\s*$`)
```

We have to introduce some way of recording how the codefence is introduced, and
how many characters is in it, to be able to compare it with a "probable" ending
code fence, ergo a fence has one type of character and a lenght.

```go "global block variables" +=
type codefence struct {
	char  string // This should probably be a rune for purity
	count int
}
```

Named blocks seems to be a wee bet harder than replace, but it is only at first
sight. First we have a "fence" of three or more tildes OR backticks, followed
by an optional space, the (optional) "language", optional spaces, the "name" of
the block in quotes and it ends with an "append" which is exactly `+=` if it
exists.

```go "Namedblock Regex"
namedBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w*)\\s*\"(?P<name>.+)\"\\s*(?P<append>[+][=])?$")
```

File blocks are not that different. The "language" is no longer optional, and
must be followed by at least a space, and the name (or "file" in our
implementation) does not allow for a lot of different characters, but it is
probably good enough.

```go "Fileblock Regex"
fileBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w+)\\s+(?P<file>[\\w\\.\\-\\/]+)\\s*(?P<append>[+][=])?$")
```

The worst thing is that both of the last regexps changes the matching order, by
introducing a new capturing group. All of our implementations need to change to
accommodate for this, and if we EVER change this again, they need to be
rechecked. To end this horror we've introduced names on our capturing groups,
but they aren't free to extract, we have to write some code for this.

First we change the parsing of the header to use a a map of string to string
and introduce a new value "fence", which we return as a "codefence", introduced
earlier. We start to use a new function namedMatchesfromRe, which should return
a map from the names of captured groups to the contents therein provided a
regular expression and a string to match.

```go "ParseHeader Declaration"
func parseHeader(line string) (File, BlockName, bool, language, codefence) {
	line = strings.TrimSpace(line) // remove indentation and trailing spaces

	// lets iterate over the regexps we have.
	for _, re := range []*regexp.Regexp{namedBlockRe, fileBlockRe} {
		if m := namedMatchesfromRe(re, line); m != nil {
			var fence codefence
			fence.char = m["fence"][0:1]
			fence.count = len(m["fence"])
			return File(m["file"]), BlockName(m["name"]), (m["append"] == "+="), language(m["language"]), fence
		}
	}

	// An empty return value for unnamed or broken fences to codeblocks.
	return "", "", false, "", codefence{}
}
```

Our function to extract named groups from a string with a regular expression
does need to return nil if the regexp does not match the provided string!
Luckily this is easy, we could return early whenever this happens. The rest
follows, we extract the names from the regex, and runs through the matches,
using the names as indexes for our return value and lastly remove the result
from the unnamed groups.

```go "Extract named matches from regexps"

// namedMatchesfromRe takes an regexp and a string to match and returns a map
// of named groups to the matches. If not matches are found it returns nil.
func namedMatchesfromRe(re *regexp.Regexp, toMatch string) (ret map[string]string) {
	substrings := re.FindStringSubmatch(toMatch)
	if substrings == nil {
		return nil
	}

	ret = make(map[string]string)
	names := re.SubexpNames()

	for i, s := range substrings {
		ret[names[i]] = s
	}
	// The names[0] and names[x] from unnamed string are an empty string.
	// Instead of checking every names[x] we simply overwrite the last ret[""]
	// and discard it at the end.
	delete(ret, "")
	return
}
```

Add this to the functions for lmt.

```go "other functions" +=
<<<Extract named matches from regexps>>>
```

Parseheader returns a new value which we put in the variable fence.

```go "Check block header"
fname, bname, appending, line.lang, fence = parseHeader(line.text)
```

The new variable needs to be defined before we process the file.

```go "process file implementation variables" +=
var fence codefence
```

And finally we're ready to start looking for codeblock ends which contain
something else than just three backticks. The length of the (trimmed) line must
be of equal length or longer than number of characters in the fence.  The line
should also only contain fence characters, which we check by replacing all
fence characters and check if all that is left is the empty string. To
summarize: if we're not in a codeblock we should handle the line as a nonblock
line, otherwise we should look for an end to the codeblock OR handle block
lines.

Note: strings.ReplaceAll was not introduced until go 1.12, which isn't
generally available in repositories.

```go "Handle file line"
if !inBlock {
	<<<Handle nonblock line>>>
	continue
}
if l := strings.TrimSpace(line.text); len(l) >= fence.count && strings.Replace(l, fence.char, "", -1) == "" {
	<<<Handle block ending>>>
	continue
}
<<<Handle block line>>>
```

Lastly, we check block starts by looking at the three first characters, and
accept not only backticks but tildes as well.

```go "Check block start"
if len(line.text) >= 3 && (line.text[0:3] == "```" || line.text[0:3] == "~~~") {
	inBlock = true
	// We were outside of a block, so just blindly reset it.
	block = make(CodeBlock, 0)
	<<<Check block header>>>
}
```
