//line README.md:65
package main

import (
//line README.md:149
	"fmt"
	"os"
	"io"
//line README.md:212
	"bufio"
//line README.md:385
	"regexp"
//line README.md:510
	"strings"
//line SubdirectoryFiles.md:35
	"path/filepath"
//line README.md:69
)

//line LineNumbers.md:25
type File string
type CodeBlock []CodeLine
type BlockName string
type language string
//line LineNumbers.md:36
type CodeLine struct {
	text   string
	file   File
	lang   language
	number int
}
//line LineNumbers.md:30

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock
//line MarkupExpansion.md:90
type codefence struct {
	char  string // This should probably be a rune for purity
	count int
}
//line README.md:402
var namedBlockRe *regexp.Regexp
//line README.md:432
var fileBlockRe *regexp.Regexp
//line README.md:516
var replaceRe *regexp.Regexp
//line README.md:72

//line LineNumbers.md:118
// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader, inputfilename string) error {
//line LineNumbers.md:82
	scanner := bufio.NewReader(r)
	var err error

	var line CodeLine
	line.file = File(inputfilename)

	var inBlock, appending bool
	var bname BlockName
	var fname File
	var block CodeBlock
//line MarkupExpansion.md:192
	var fence codefence
//line LineNumbers.md:99
	for {
		line.text, err = scanner.ReadString('\n')
		line.number++
		switch err {
		case io.EOF:
			return nil
		case nil:
			// Nothing special
		default:
			return err
		}
//line MarkupExpansion.md:208
		if !inBlock {
//line MarkupExpansion.md:223
			if len(line.text) >= 3 && (line.text[0:3] == "```" || line.text[0:3] == "~~~") {
				inBlock = true
				// We were outside of a block, so just blindly reset it.
				block = make(CodeBlock, 0)
//line MarkupExpansion.md:186
				fname, bname, appending, line.lang, fence = parseHeader(line.text)
//line MarkupExpansion.md:228
			}
//line MarkupExpansion.md:210
			continue
		}
		if l := strings.TrimSpace(line.text); len(l) >= fence.count && strings.Replace(l, fence.char, "", -1) == "" {
//line LineNumbers.md:56
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
//line MarkupExpansion.md:214
			continue
		}
//line LineNumbers.md:48
		block = append(block, line)
//line LineNumbers.md:111
	}
//line LineNumbers.md:121
}
//line MarkupExpansion.md:128
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
//line WhitespacePreservation.md:34
// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace(prefix string) (ret CodeBlock) {
//line LineNumbers.md:251
	var line string
	for _, v := range c {
		line = v.text
//line LineNumbers.md:234
		matches := replaceRe.FindStringSubmatch(line)
		if matches == nil {
			if v.text != "\n" {
				v.text = prefix + v.text
			}
			ret = append(ret, v)
			continue
		}
//line LineNumbers.md:220
		bname := BlockName(matches[2])
		if val, ok := blocks[bname]; ok {
			ret = append(ret, val.Replace(prefix+matches[1])...)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
			ret = append(ret, v)
		}
//line LineNumbers.md:255
	}
	return
//line WhitespacePreservation.md:38
}
//line LineNumbers.md:277

// Finalize extract the textual lines from CodeBlocks and (if needed) prepend a
// notice about "unexpected" filename or line changes, which is extracted from
// the contained CodeLines. The result is a string with newlines ready to be
// pasted into a file.
func (c CodeBlock) Finalize() (ret string) {
	var file File
	var formatstring string
	var linenumber int
	for _, l := range c {
		if linenumber+1 != l.number || file != l.file {
			switch l.lang {
			case "go", "golang":
				formatstring = "//line %[2]v:%[1]v\n"
			default:
				formatstring = "#line %v \"%v\"\n"
			}
			ret += fmt.Sprintf(formatstring, l.number, l.file)
		}
		ret += l.text
		linenumber = l.number
		file = l.file
	}
	return
}
//line MarkupExpansion.md:154

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
//line README.md:74

func main() {
//line README.md:157
	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)
//line MarkupExpansion.md:103
	namedBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w*)\\s*\"(?P<name>.+)\"\\s*(?P<append>[+][=])?$")
//line MarkupExpansion.md:112
	fileBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w+)\\s+(?P<file>[\\w\\.\\-\\/]+)\\s*(?P<append>[+][=])?$")
//line MarkupExpansion.md:82
	replaceRe = regexp.MustCompile(`^(?P<prefix>\s*)(?:<<|//)<(?P<name>.+)>>>\s*$`)
//line README.md:136

	// os.Args[0] is the command name, "lmt". We don't want to process it.
	for _, file := range os.Args[1:] {
//line LineNumbers.md:127
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
//line README.md:140

	}
//line LineNumbers.md:308
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
//line README.md:77
}
