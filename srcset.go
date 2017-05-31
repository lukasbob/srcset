// Package srcset `srcset` provides a parser for the HTML5 `srcset` attribute, based on the
// [WHATWG reference algorithm](https://html.spec.whatwg.org/multipage/embedded-content.html#parse-a-srcset-attribute).
// TODO: This works, but I dislike the state manipulation.
// Use more go-like structures for reading and tokenization, like bufio.Scanner
package srcset

import (
	"regexp"
	"strconv"
)

// ImageSource is a structure that contains an image definition.
type ImageSource struct {
	URL     string
	Width   *int64
	Height  *int64
	Density *float64
}

// SourceSet is the result of parsing the value of a srcset attribute.
// A SourceSet consists of multiple ImageSource instances.
type SourceSet []ImageSource

const (
	comma       = ','
	leftParens  = '('
	rightParens = ')'
)

const (
	stateNone = iota
	stateInDescriptor
	stateInParens
	stateAfterDescriptor
)

var (
	regexLeadingSpaces         = regexp.MustCompile("^[ \t\n\r\u000c]+")
	regexLeadingCommasOrSpaces = regexp.MustCompile("^[, \t\n\r\u000c]+")
	regexLeadingNotSpaces      = regexp.MustCompile("^[^ \t\n\r\u000c]+")
	regexTrailingCommas        = regexp.MustCompile("[,]+$")
	regexNonNegativeInteger    = regexp.MustCompile(`^\d+$`)
	regexFloatingPoint         = regexp.MustCompile(`^-?(?:[0-9]+|[0-9]*\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)
)

// Parse takes the value of a srcset attribute and parses it.
func Parse(input string) SourceSet {
	var (
		url         string
		pos         = 0
		currState   = stateNone
		end         = len(input)
		candidates  = SourceSet{}
		descriptors = []string{}
	)

	collectChars := func(rx *regexp.Regexp) string {
		if match := rx.FindString(input[pos:]); match != "" {
			pos += len(match)
			return match
		}

		return ""
	}

	isSpace := func(c rune) bool {
		return (c == '\u0020' || // space
			c == '\u0009' || // horizontal tab
			c == '\u000A' || // new line
			c == '\u000C' || // form feed
			c == '\u000D') // carriage return
	}

	parseDescriptors := func() {
		var (
			isErr = false
			h     *int64
			w     *int64
			d     *float64
		)

		for _, desc := range descriptors {
			lastIdx := len(desc) - 1
			lastChar, numericVal := desc[lastIdx], desc[:lastIdx]
			intVal, intErr := strconv.ParseInt(numericVal, 10, 64)
			floatVal, floatErr := strconv.ParseFloat(numericVal, 64)

			if regexNonNegativeInteger.MatchString(numericVal) && lastChar == 'w' {
				if w != nil || d != nil {
					isErr = true
				}
				if intErr != nil || intVal == 0 {
					isErr = true
				} else {
					w = &intVal
				}
			} else if regexFloatingPoint.MatchString(numericVal) && lastChar == 'x' {
				if w != nil || d != nil || h != nil {
					isErr = true
				}
				if floatErr != nil || floatVal < 0 {
					isErr = true
				} else {
					d = &floatVal
				}
			} else if regexNonNegativeInteger.MatchString(numericVal) && lastChar == 'h' {
				if h != nil || d != nil {
					isErr = true
				}
				if intErr != nil || intVal == 0 {
					isErr = true
				} else {
					h = &intVal
				}
			} else {
				isErr = true
			}
		}

		if !isErr {
			candidates = append(candidates, ImageSource{
				URL:     url,
				Density: d,
				Width:   w,
				Height:  h,
			})
		}
	}

	tokenize := func() {
		collectChars(regexLeadingSpaces)
		currDescriptor := ""
		currState = stateInDescriptor

		for {
			if pos == len(input) {
				if currState != stateAfterDescriptor && currDescriptor != "" {
					descriptors = append(descriptors, currDescriptor)
				}

				parseDescriptors()
				return
			}

			c := rune(input[pos])

			switch currState {
			case stateInDescriptor:
				if isSpace(c) {
					if currDescriptor != "" {
						descriptors = append(descriptors, currDescriptor)
						currDescriptor = ""
						currState = stateAfterDescriptor
					}
				} else if c == comma {
					pos++
					if currDescriptor != "" {
						descriptors = append(descriptors, currDescriptor)
						parseDescriptors()
						return
					}
				} else if c == leftParens {
					currDescriptor += string(c)
					currState = stateInParens
				} else {
					currDescriptor += string(c)
				}

			case stateInParens:
				if c == rightParens {
					currDescriptor += string(c)
					currState = stateInDescriptor
				} else {
					currDescriptor += string(c)
				}

			case stateAfterDescriptor:
				if isSpace(c) {

				} else {
					currState = stateInDescriptor
					pos--
				}
			}

			pos++
		}
	}

	for {
		collectChars(regexLeadingCommasOrSpaces)
		if pos >= end {
			return candidates
		}

		url = collectChars(regexLeadingNotSpaces)
		descriptors = []string{}

		if url[len(url)-1] == ',' {
			url = regexTrailingCommas.ReplaceAllString(url, "")
			parseDescriptors()
		} else {
			tokenize()
		}
	}
}
