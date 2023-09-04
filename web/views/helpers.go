package views

import (
	"strconv"
	"time"
)

var HelperFuncMap = map[string]interface{}{
	"formatDate":    formatDate,
	"convertDate":   convertDate,
	"contrastColor": contrastColor}

func formatDate(date time.Time) string {
	return date.Format(time.Stamp)
}

func convertDate(date time.Time, timeZone string) time.Time {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return date
	}

	return date.In(loc)
}

// contrastColor takes a hex color code as input and returns either #000000 (black) or #FFFFFF (white) as output
// depending on which one contrasts the most with the input color
func contrastColor(hex string) string {
	// convert the hex color code to RGB values
	r, _ := strconv.ParseInt(hex[1:3], 16, 0)
	g, _ := strconv.ParseInt(hex[3:5], 16, 0)
	b, _ := strconv.ParseInt(hex[5:7], 16, 0)

	// calculate the luminance of the input color using a formula from https://www.w3.org/TR/WCAG20/#relativeluminancedef
	lum := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)

	// if the luminance is less than or equal to 128, return white as the contrast color
	// otherwise, return black as the contrast color
	if lum <= 128 {
		return "#FFFFFF"
	} else {
		return "#000000"
	}
}
