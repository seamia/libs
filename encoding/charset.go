package encoding

import (
	"strings"
)

const (
	left	= "=?"	// "=?windows-1251?Q?"
	middle	= "?Q?"
	middle2	= "?q?"
	right 	= "?="
)

func Decode(src string) string {
	result := ""
	for {
		start := strings.Index(src, left)
		if start < 0 {
			result += src
			break
		}
		result += src[:start]
		src = src[start+len(left):]

		center := strings.Index(src, middle)
		end := strings.Index(src, right)

		if end < 0 {
			result += src
			break
		}
		if center < 0 || center > end {
			center = strings.Index(src, middle2)
		}
		if center < 0 || center > end {
			result += src
			break
		}
		end = center+len(middle) + strings.Index(src[center+len(middle):], right)

		encoding := src[:center]
		payload := src[center+len(middle):end]

		result += decodeIt(payload, encoding)

		src = src[end+len(right):]
		src = strings.TrimPrefix(src, " ")
	}

	return result
}

var (
	mapping = map[string]string {
		"=3F": "?",

		"=91": "‘",
		"=92": "’",
		"=96": "-",
		"=97": "|",

		"=A0": " ",	// non-brekable space?
	}
)

func decodeIt(what string, encd string) string {
	result := strings.Replace(what, "_", " ", -1)

	for from,to := range mapping {
		result = strings.Replace(result, from, to, -1)
	}

	if strings.Index(result, "=9") >= 0 {
		// fmt.Println("found unhandled escape:", result)	// todo: remove?
	}

	return result
}
