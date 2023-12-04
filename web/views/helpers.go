package views

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andrei0427/go-changediff/internal/app/models"
	"github.com/dustin/go-humanize"
)

var HelperFuncMap = map[string]interface{}{
	"formatDate":              formatDate,
	"formatDateShort":         formatDateShort,
	"formatHTMLInputDateTime": formatHTMLInputDateTime,
	"parseDateTime":           parseDateTime,
	"convertDate":             convertDate,
	"contrastColor":           contrastColor,
	"CDNUrl":                  CDNUrl,
	"isLast":                  isLast,
	"formatDuration":          formatDuration,
}

func isLast(elm interface{}, slice ...any) bool {
	last := slice[len(slice)-1]
	if last == nil {
		return false
	}

	if arr, ok := last.([]models.RoadmapPostActivityModel); ok {
		return len(arr) > 0 && arr[len(arr)-1] == elm
	}

	arr := last.([]interface{})
	return len(arr) > 0 && arr[len(arr)-1] == elm
}

func formatDuration(date time.Time) string {
	return humanize.Time(date)
}

func formatDate(date time.Time) string {
	return date.Format(time.RFC822)
}

func formatDateShort(date time.Time) string {
	return date.Format(time.DateOnly)
}

func formatHTMLInputDateTime(date time.Time) string {
	return date.Format("2006-01-02T15:04")
}

func parseDateTime(dateTime string) time.Time {
	parsed, err := time.Parse(time.DateTime, dateTime)
	if err != nil {
		return time.Now()

	}

	return parsed
}

func convertDate(date time.Time, timeZone string) time.Time {
	loc, err := time.LoadLocation(timeZone)
	if err != nil {
		return date
	}

	return date.In(loc)
}

func CDNUrl(filePath string) string {
	if strings.Contains(filePath, "http") {
		return filePath
	}
	return os.Getenv("CLOUDFLARE_R2_PUBLIC_URL") + "/" + filePath
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
