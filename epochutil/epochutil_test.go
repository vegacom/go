package main

import (
	"fmt"
	"testing"
	"time"
)

func TestZones(t *testing.T) {
	if err := initTimeZones(); err != nil {
		t.Fatal(err)
	}
	if got := len(zones); got == 0 {
		t.Errorf("got 0 zones")
	}
	if got := filterZones("pacific"); len(got) == 0 {
		t.Errorf("got 0 zones in pacific")
	}
}

func TestParse(t *testing.T) {
	for _, tt := range []struct {
		descr string
		input string
		want  time.Time
	}{
		{
			descr: "nano",
			input: "1388536210123456789",
			want:  time.Unix(0, 1388536210123456789),
		},
		{
			descr: "milli",
			input: "1388536210123",
			want:  time.Unix(0, 1388536210123000000),
		},
		{
			descr: "string date",
			input: "2014-01-01 00:30:10.123456789 +0000 UTC",
			want:  time.Unix(0, 1388536210123456789),
		},
	} {
		got, err := parse(tt.input)
		if err != nil {
			t.Fatalf("%s: %v", tt.descr, err)
		}
		if !got.Equal(tt.want) {
			t.Errorf("%s: got %d want %d", tt.descr, got.UnixNano(), tt.want.UnixNano())
		}
	}
}

func ExamplePtimePrint() {
	mex, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		fmt.Println(err)
	}

	pt := &ptime{
		Input:    "1388536210123456789",
		Location: mex,
		Time:     time.Unix(0, 1388536210123456789),
	}
	pt.Print()
	// Output:
	//Tuesday 2013-12-31 18:30:10.123456789 -0600 CST
	//nano  = 1388536210123456789
	//micro : 1388536210123456
	//milli : 1388536210123
	//sec   : 1388536210
}

func ExamplePtimePrintDelta() {
	pt := &ptime{

		Delta: time.Hour * 9,
		Time:  time.Unix(0, 1388536210123456789),
	}
	pt.Print()
	// Output:
	//Wednesday 2014-01-01 18:30:10.123456789 +0900 JST
	//nano  : 1388568610123456789
	//micro : 1388568610123456
	//milli : 1388568610123
	//sec   : 1388568610
}
