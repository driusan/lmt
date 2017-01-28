package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type File string
type CodeBlock string
type BlockName string

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock
var namedBlockRe *regexp.Regexp
var fileBlockRe *regexp.Regexp
var replaceRe *regexp.Regexp

// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader) error {
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
		if inBlock {
			if line == "```\n" {
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

				inBlock = false
				appending = false
				bname = ""
				fname = ""
				block = ""
				continue
			} else {
				block += CodeBlock(line)
			}
		} else {
			if line == "" {
				continue
			}

			switch line[0] {
			case '`':
				if len(line) >= 3 && line[0:3] == "```" {
					inBlock = true
				}
			case '#':
				fname, bname, appending = parseHeader(line)
				// We're outside of a block, so just blindly reset it.
				block = ""
			default:
				inBlock = false
				appending = false
				bname = ""
				fname = ""
				block = ""
			}
		}
	}
}
func parseHeader(line string) (File, BlockName, bool) {
	matches := namedBlockRe.FindStringSubmatch(line)
	if matches != nil {
		return "", BlockName(matches[2]), (matches[3] == "+=")
	}
	matches = fileBlockRe.FindStringSubmatch(line)
	if matches != nil {
		return File(matches[2]), "", (matches[3] == "+=")
	}
	return "", "", false
}

// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace(prefix string) (ret CodeBlock) {
	scanner := bufio.NewReader(strings.NewReader(string(c)))

	for {
		line, err := scanner.ReadString('\n')
		// ReadString will eventually return io.EOF and this will return.
		if err != nil {
			return
		}
		matches := replaceRe.FindStringSubmatch(line)
		if matches == nil {
			ret += CodeBlock(prefix) + CodeBlock(line)
			continue
		}
		bname := BlockName(matches[2])
		if val, ok := blocks[bname]; ok {
			ret += val.Replace(prefix + matches[1])
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
			ret += CodeBlock(line)
		}
	}
	return
}

func main() {
	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)
	namedBlockRe = regexp.MustCompile(`^([#]+)[\s]*"(.+)"[\s]*([+][=])?`)
	fileBlockRe = regexp.MustCompile(`^([#]+)[\s]*([\w\.\-\/]+)[\s]*([+][=])?`)
	replaceRe = regexp.MustCompile(`^([\s]*)<<<(.+)>>>[\s]*$`)

	// os.Args[0] is the command name, "lmt". We don't want to process it.
	for _, file := range os.Args[1:] {
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

	}
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
		fmt.Fprintf(f, "%s", codeblock.Replace(""))
		// We don't defer this so that it'll get closed before the loop finishes.
		f.Close()

	}
}
