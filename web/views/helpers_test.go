package views

import "testing"

// TestContrastColor tests the contrastColor function with some sample inputs and outputs
func TestContrastColor(t *testing.T) {
	// create a table of test cases with input and expected output
	testCases := []struct {
		input  string
		output string
	}{
		{"#000000", "#FFFFFF"}, // black -> white
		{"#FFFFFF", "#000000"}, // white -> black
		{"#FF0000", "#FFFFFF"}, // red -> white
		{"#00FF00", "#000000"}, // green -> black
		{"#0000FF", "#FFFFFF"}, // blue -> white
		{"#FFFF00", "#000000"}, // yellow -> black
		{"#FF00FF", "#FFFFFF"}, // magenta -> white
		{"#00FFFF", "#000000"}, // cyan -> black
		{"#808080", "#FFFFFF"}, // gray -> white
	}

	// loop through the test cases and compare the actual output with the expected output
	for _, tc := range testCases {
		actual := contrastColor(tc.input)
		if actual != tc.output {
			t.Errorf("contrastColor(%s): got %s; want %s", tc.input, actual, tc.output)
		}
	}
}
