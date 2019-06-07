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
//line LineNumbers.md:99
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
//line LineNumbers.md:145
		if inBlock {
			if line.text == "```\n" {
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
//line LineNumbers.md:148
				continue
			}
//line LineNumbers.md:48
			block = append(block, line)
//line LineNumbers.md:151
			continue
		}
//line LineNumbers.md:170
		if len(line.text) >= 3 && (line.text[0:3] == "```") {
			inBlock = true
			// We were outside of a block, so just blindly reset it.
			block = make(CodeBlock, 0)
//line LineNumbers.md:181
			fname, bname, appending, line.lang = parseHeader(line.text)
//line LineNumbers.md:175
		}
//line LineNumbers.md:111
	}
//line LineNumbers.md:121
}
//line LineNumbers.md:190
func parseHeader(line string) (File, BlockName, bool, language) {
	line = strings.TrimSpace(line)
//line LineNumbers.md:197
	var matches []string
	if matches = namedBlockRe.FindStringSubmatch(line); matches != nil {
		return "", BlockName(matches[2]), (matches[3] == "+="), language(matches[1])
	}
	if matches = fileBlockRe.FindStringSubmatch(line); matches != nil {
		return File(matches[2]), "", (matches[3] == "+="), language(matches[1])
	}
	return "", "", false, ""
//line LineNumbers.md:193
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
//line LineNumbers.md:306
			case "bash", "shell", "sh", "perl":
				formatstring = "#line %v \"%v\"\n"
			case "go", "golang":
				formatstring = "//line %[2]v:%[1]v\n"
			case "C", "c":
				formatstring = "#line %v \"%v\"\n"
//line LineNumbers.md:290
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
//line README.md:74

func main() {
//line README.md:157
	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)
//line LineNumbers.md:208
	namedBlockRe = regexp.MustCompile("^`{3,}\\s?(\\w*)\\s*\"(.+)\"\\s*([+][=])?$")
//line LineNumbers.md:212
	fileBlockRe = regexp.MustCompile("^`{3,}\\s?(\\w+)\\s+([\\w\\.\\-\\/]+)\\s*([+][=])?$")
//line WhitespacePreservation.md:11
	replaceRe = regexp.MustCompile(`^([\s]*)<<<(.+)>>>[\s]*$`)
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
//line LineNumbers.md:318
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
