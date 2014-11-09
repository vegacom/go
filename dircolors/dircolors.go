/*
Package dircolors adds support for dircolors(5) codes and color init strings.

Examples:

* Key("/tmp") returns "di" assuming /tmp is a subdir.

* ForTTY("/tmp") returns a string that contains "tmp" and a terminal color sequence.

DEPENDENCIES:
* glog

PLATFORMS:
* Linux

MORE:
  http://linux.die.net/man/5/dir_colors

  % dircolors --print-database
*/
package dircolors

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/golang/glog"
)

const (
	Reset         = "rs" // Normal
	Dir           = "di"
	Link          = "ln"
	MultiHardlink = "mh"
	NamedPipe     = "pi"
	Socket        = "so"
	BlockDevice   = "bd"
	CharDevice    = "cd"
	Orphan        = "or"
	Setuid        = "su"
	Setgid        = "sg"
	Executable    = "ex"
)

// NotNormal are dircolors codes for non-regular files.
var NotNormal = map[string]bool{
	Orphan:      true,
	Dir:         true,
	Link:        true,
	NamedPipe:   true,
	CharDevice:  true,
	BlockDevice: true,
	Socket:      true,
}

// Key returns the dircolor code for fpath based on its type. Eg. "di" for dir.
func Key(fpath string) (string, error) {
	info, err := os.Lstat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return Orphan, nil
		}
		return "", err
	}
	if info.IsDir() {
		return Dir, nil
	}
	mode := info.Mode()
	if mode&os.ModeSymlink != 0 {
		if _, err := os.Readlink(fpath); err != nil {
			return Orphan, nil
		}
		return Link, nil
	}
	if mode&os.ModeNamedPipe != 0 {
		return NamedPipe, nil
	}
	if mode&os.ModeSocket != 0 {
		return Socket, nil
	}
	if mode&os.ModeDevice != 0 {
		if mode&os.ModeCharDevice != 0 {
			return CharDevice, nil
		}
		return BlockDevice, nil
	}
	if mode&os.ModeSetuid != 0 {
		return Setuid, nil
	}
	if mode&os.ModeSetgid != 0 {
		return Setgid, nil
	}
	if mode&0111 != 0 {
		return Executable, nil
	}
	sys, ok := info.Sys().(*syscall.Stat_t) // assume it's Linux
	if !ok {
		return "", errors.New("unsupported system")
	}
	if sys.Nlink > 1 {
		return MultiHardlink, nil
	}
	ext := filepath.Ext(fpath)
	if ext != "" {
		return "*" + ext, nil
	}

	// TODO: investigate what these codes are for:
	// do door
	// tw
	// ca capability
	// ow other writable
	// st sticky

	return Reset, nil
}

var tty map[string]string

// ForTTY wraps fpath with color sequences based on its type. Eg. dirs are colored blue.
func ForTTY(fpath string) string {
	if tty == nil {
		tty = parseDircolors()
	}

	key, err := Key(fpath)
	if err != nil || tty[key] == "" {
		return fpath
	}
	return "\033[" + tty[key] + "m" + fpath + "\033[0m"
}

// parseDircolors returns a map of dircolor code to color. It first retrieves the dircolors string
// from the LS_COLORS env var or from the the output of the dircolors command.
func parseDircolors() map[string]string {
	s := os.Getenv("LS_COLORS")
	if s == "" {
		var err error
		s, err = execDircolors()
		if err != nil {
			glog.Warningf("no dircolor support: %s", err)
		}
	}

	m := make(map[string]string)
	for _, pair := range strings.Split(s, ":") {
		items := strings.SplitN(pair, "=", 2)
		if len(items) == 2 {
			m[items[0]] = items[1]
		}
	}
	return m
}

func execDircolors() (string, error) {
	b, err := exec.Command("dircolors", "-b").Output()
	if err != nil {
		return "", err
	}
	data := string(b)
	start, end := strings.Index(data, "'"), strings.LastIndex(data, "'")
	if start == -1 || end == -1 {
		return "", fmt.Errorf("`dircolors -b` returned bad format (%q)", data)
	}
	return data[start+1 : end], nil
}
