/*
Binary epochutil formats and parses dates. This is a command line interface tool for
        the Go time package.
*/
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

const (
	usageMessage = `NAME:
    epochutil - Binary epochutil formats and parses dates. This is a command line interface tool for
        the Go time package.

USAGE:
    epochutil [flags] command

VERSION:
    0.3

COMMANDS:
    zones                  Lists the available timezones.
    now                    Prints the current time.
	  STRING [STRING ...]    Parses the input and prints the time.
    help, h                Shows a list of available commands.

EXAMPLES:
  Print the time now in the default timezone:
    epochutil now

  Print the time now in Mexico City:
    epochutil --zone mexico_city now

  Print the time four hours ago:
    epochutil --delta -4h now

  Parse the epoch time in nanoseconds, print the time:
    epochutil 1388586612345678901

  Parse the epoch time in seconds, and print the time:
    epochutil 1388586612

  Parse the date and print the time:
    epochutil "2014-01-01"

  Parse the date and print the time one day after in Mexico City:
    epochutil --delta 24h --zone mexico_city "2014-01-01 09:30:10.267 +0900 JST"

MORE:
    http://en.wikipedia.org/wiki/Unix_time

FLAGS:
`
)

// flags
var (
	zone  = flag.String("zone", "", "Sets the timezone. Use `epochutil zones` to list all available timezones")
	delta = flag.Duration("delta", 0, "Adds (or substracts) a duration, eg. -8h")
)

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageMessage)
		flag.PrintDefaults()
	}
	flag.Parse()

	if err := proc(flag.Args()); err != nil {
		glog.Exit(err)
	}
}

func usage() {
	flag.Usage()
	os.Exit(2)
}

func proc(args []string) error {
	err := initTimeZones()
	if err != nil {
		return err
	}
	var location *time.Location
	if *zone != "" {
		candidates := filterZones(*zone)
		if len(candidates) == 0 {
			return fmt.Errorf("%s does not match any zone, use `epochutil zones`", *zone)
		}
		if len(candidates) > 1 {
			// TODO(vegacom): sometimes multi matches refer to the same tz (eg. utc).
			return fmt.Errorf("%s matches these zones:\n%s", *zone, strings.Join(candidates, "\n"))
		}
		location, err = time.LoadLocation(candidates[0])
		if err != nil {
			return err
		}
	}

	first := ""
	if len(args) > 0 {
		first = args[0]
	}

	switch first {
	case "", "h", "help":
		usage()

	case "zones":
		fmt.Println(strings.Join(zones, "\n"))

	case "now":
		pt := &ptime{
			Delta:    *delta,
			Location: location,
			Time:     time.Now(),
		}
		pt.Print()

	default:
		for _, input := range args {
			t, err := parse(input)
			if err != nil {
				glog.Error(err)
				continue
			}
			pt := &ptime{
				Input:    input,
				Delta:    *delta,
				Location: location,
				Time:     t,
			}
			pt.Print()
		}
	}
	return nil
}

type ptime struct {
	// The input string as provided by the user for each arg.
	Input string
	// The time offset to use, could be negative.
	Delta time.Duration
	// The parsed time zone.
	Location *time.Location
	// The time before applying delta or location.
	Time time.Time
}

var timeUnits = []string{
	"nano",
	"micro",
	"milli",
	"sec",
}

// Print prints the time after applying the delta and/or location.
func (pt *ptime) Print() {
	t := pt.Time.Add(pt.Delta)
	if pt.Location != nil {
		t = t.In(pt.Location)
	}

	fmt.Println(t.Weekday(), t)
	v := t.UnixNano()
	for _, unit := range timeUnits {
		s := strconv.FormatInt(v, 10)
		sep := ":"
		if s == pt.Input {
			sep = "="
		}
		fmt.Printf("%-5s %s %9s\n", unit, sep, s)
		v /= 1e3
	}
}

var numbersRE = regexp.MustCompile("^[0-9]+$")

func parse(s string) (time.Time, error) {
	if numbersRE.MatchString(s) {
		return parseEpoch(s)
	}
	return parseDateTime(s)
}

// parseEpoch parses the epoch (numeric) and returns the time value it represents; epoch could be in
// nano-seconds, micro-seconds, milli-seconds or seconds.
//
// ⚠ Dates in 1970 are not properly supported.
func parseEpoch(epoch string) (time.Time, error) {
	x, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	var t time.Time
	for j := 0; j < 4; j++ {
		t = time.Unix(0, x)
		if t.Year() != 1970 {
			return t, nil
		}
		x *= 1000
	}

	return t, nil
}

var dateLayouts = []string{
	"2006-01-02 15:04:05 -0700 MST",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// parseDateTime parses a formatted string and returns the time value it represents.
func parseDateTime(s string) (time.Time, error) {
	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date-time %q", s)
}

var zones []string

func initTimeZones() error {
	zipPath := runtime.GOROOT() + "/lib/time/zoneinfo.zip"
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	zones = make([]string, 0, len(r.File))
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		zones = append(zones, f.Name)
	}
	return nil
}

func filterZones(x string) []string {
	var out []string
	x = strings.ToLower(x)
	for _, z := range zones {
		if strings.Contains(strings.ToLower(z), x) {
			out = append(out, z)
		}
	}
	return out
}
