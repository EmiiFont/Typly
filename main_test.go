package main

import "testing"

func TestParseSize(t *testing.T) {
	for _, test := range []struct {
		input  string
		width  int
		height int
		valid  bool
	}{
		{input: "320x180", width: 320, height: 180, valid: true},
		{input: "320X180", width: 320, height: 180, valid: true},
		{input: "320x180junk"},
		{input: "320x180 "},
		{input: "0x180"},
	} {
		width, height, err := parseSize(test.input)
		if test.valid {
			if err != nil || width != test.width || height != test.height {
				t.Errorf("parseSize(%q) = %dx%d, %v; want %dx%d", test.input, width, height, err, test.width, test.height)
			}
		} else if err == nil {
			t.Errorf("parseSize(%q) unexpectedly succeeded", test.input)
		}
	}
}
