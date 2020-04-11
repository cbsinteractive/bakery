package parsers

import (
	"regexp"
)

// splitAfter splits a string after the matches of the specified regexp
func splitAfter(s string, re *regexp.Regexp) []string {
	var splitResults []string
	var position int
	indices := re.FindAllStringIndex(s, -1)
	if indices == nil {
		return append(splitResults, s)
	}
	for _, idx := range indices {
		section := s[position:idx[1]]
		splitResults = append(splitResults, section)
		position = idx[1]
	}
	return append(splitResults, s[position:])
}

//validatePositiveRange will check if x and y are a valid
//range greater than zero, and less than a provided max value
func validatePositiveRange(x, y, max int) bool {
	if x > y ||
		0 > x ||
		y > max ||
		x == y {
		return false
	}

	return true
}
