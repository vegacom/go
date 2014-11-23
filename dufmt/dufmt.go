/*
dufmt prints the disk usage ordered by size with human readable output.
*/
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unsafe" // for isTerminal

	humanize "github.com/dustin/go-humanize"
	"github.com/golang/glog"
	"github.com/vegacom/go/dircolors"
)

const (
	usageMessage = `NAME:
    dufmt - Prints the disk usage ordered by size with human readable output.

USAGE:
    dufmt [flags] [FILE1 FILE2 ...]

VERSION:
    0.1

EXAMPLES:
  Print the usage for all files and dirs in the current working dir:
    dufmt

FLAGS:
`
)

// TODO: print percentage bars.
// TODO: use cmd.StdoutPipe and scanner.

// flags
var (
	color = flag.String("color", "auto",
		"Surround filepaths with escape sequences to display them in color on the terminal. Valid values are never, always or auto.")
)

type duInfo struct {
	name          string
	size          int
	childrenCount int
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageMessage)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Determine what paths to pass to du depending on len(args).
	paths, err := paths(flag.Args())
	if err != nil {
		glog.Exitf("Bad args: %s", err)
	}
	// Call du and parse the output into a slice of *duInfo.
	data, err := execDu(paths)
	if err != nil {
		glog.Exitf("Cannot execute du: %s", err)
	}
	s, err := newDuSlice(data)
	if err != nil {
		glog.Exit(err)
	}

	sort.Sort(s)

	s = append(s, &duInfo{
		name: "total",
		size: total(s),
	})

	s.WriteAll(os.Stdout)
}

var baseDuArgs = []string{"--summarize", "--bytes", "--"}

func execDu(paths []string) ([]byte, error) {
	duArgs := append(baseDuArgs, paths...)
	return exec.Command("du", duArgs...).Output()
}

type duSlice []*duInfo

func (s duSlice) Len() int           { return len(s) }
func (s duSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s duSlice) Less(i, j int) bool { return s[i].size < s[j].size }
func (s duSlice) WriteAll(writer io.Writer) {
	colorize := *color == "always" || (*color == "auto" && isTerminal(syscall.Stdout))
	for _, info := range s {
		humanized := humanize.IBytes(uint64(info.size))
		name := info.name
		if colorize {
			name = dircolors.ForTTY(name)
		}
		io.WriteString(writer, fmt.Sprintf("%12s %s %s\n", humanized, formatCount(info.childrenCount), name))
	}
}

func formatCount(n int) string {
	if n <= 1 {
		return "         "
	}
	return fmt.Sprintf("%9d", n)
}

func count(fpath string) (n int) {
	walk := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		n++
		return nil
	}
	filepath.Walk(fpath, walk)
	return
}

// newDuSlice parses data from du output. The expected format is: "\d+\t.*"
func newDuSlice(data []byte) (duSlice, error) {
	var s duSlice
	for _, b := range bytes.Split(data, []byte("\n")) {
		line := string(b)
		if line == "" {
			continue
		}
		items := strings.SplitN(line, "\t", 2)
		size, err := strconv.Atoi(items[0])
		if err != nil {
			return nil, fmt.Errorf("unexpected format: %q: %s", line, err)
		}
		s = append(s, &duInfo{
			name:          items[1],
			size:          size,
			childrenCount: count(items[1]),
		})
	}
	return s, nil
}

func paths(args []string) ([]string, error) {
	if len(args) == 0 {
		return filepath.Glob("*")
	}

	if len(args) == 1 {
		first := args[0]
		fi, err := os.Stat(first)
		if err != nil {
			return nil, err
		}
		if fi.IsDir() {
			return filepath.Glob(filepath.Join(first, "*"))
		}
	}
	return args, nil
}

func total(s duSlice) (total int) {
	for _, x := range s {
		total += x.size
	}
	return
}

// code copied from golang.org/x/crypto/ssh/terminal:

const ioctlReadTermios = 0x5401 // syscall.TCGETS

// isTerminal returns true if the given file descriptor is a terminal.
func isTerminal(fd int) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}
