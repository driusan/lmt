//line Implementation.md:60
package main

import (
//line Implementation.md:157
	"fmt"
	"os"
	"io"
//line Implementation.md:223
	"bufio"
//line Implementation.md:406
	"regexp"
//line Implementation.md:542
	"strings"
//line SubdirectoryFiles.md:35
	"path/filepath"
//line Implementation.md:64
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
//line Implementation.md:424
var namedBlockRe *regexp.Regexp
//line Implementation.md:458
var fileBlockRe *regexp.Regexp
//line Implementation.md:549
var replaceRe *regexp.Regexp
//line IndentedBlocks.md:68
var blockStartRe *regexp.Regexp
//line Implementation.md:67

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
//line IndentedBlocks.md:91
	var blockPrefix string
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
//line IndentedBlocks.md:118
		if inBlock {
		    line.text = strings.TrimPrefix(line.text, blockPrefix)
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
//line IndentedBlocks.md:122
		        continue
		    }
//line LineNumbers.md:48
		    block = append(block, line)
//line IndentedBlocks.md:125
		    continue
		}
//line IndentedBlocks.md:55
		if matches := blockStartRe.FindStringSubmatch(line.text); matches != nil {
		    inBlock = true
		    blockPrefix = matches[1]
		    line.text = strings.TrimPrefix(line.text, blockPrefix)
		    // We were outside of a block, so just blindly reset it.
		    block = make(CodeBlock, 0)
//line LineNumbers.md:181
		    fname, bname, appending, line.lang = parseHeader(line.text)
//line IndentedBlocks.md:62
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
			case "C", "c", "cpp":
				formatstring = "#line %v \"%v\"\n"
            default:
				ret += l.text
				continue
			}
			ret += fmt.Sprintf(formatstring, l.number, l.file)
		}
		ret += l.text
		linenumber = l.number
		file = l.file
	}
	return
}
//line Implementation.md:69

func main() {
//line Implementation.md:166
	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)
//line LineNumbers.md:208
	namedBlockRe = regexp.MustCompile("^`{3,}\\s?([\\w\\+]*)\\s*\"(.+)\"\\s*([+][=])?$")
//line LineNumbers.md:212
	fileBlockRe = regexp.MustCompile("^`{3,}\\s?([\\w\\+]+)\\s+([\\w\\.\\-\\/]+)\\s*([+][=])?$")
//line WhitespacePreservation.md:11
	replaceRe = regexp.MustCompile(`^([\s]*)<<<(.+)>>>[\s]*$`)
//line IndentedBlocks.md:82
	blockStartRe = regexp.MustCompile("^([\\s]*)```")
//line Implementation.md:144

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
//line Implementation.md:148

	}
//line LineNumbers.md:311
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
//line Implementation.md:72
}
