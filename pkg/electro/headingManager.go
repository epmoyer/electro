package electro

import (
	"strconv"
	"strings"
)

const maxHeadingDepth = 6

type headingManagerT struct {
	atLevel       int
	headingNumber []int
	headingLetter []rune
	isLetter      []bool
}

func newHeadingManager(level int) *headingManagerT {
	hm := &headingManagerT{
		atLevel:       level,
		headingNumber: make([]int, maxHeadingDepth+1),
		headingLetter: make([]rune, maxHeadingDepth+1),
		isLetter:      make([]bool, maxHeadingDepth+1),
	}
	// Appendix letters start at "A" at every level.
	for i := range hm.headingLetter {
		hm.headingLetter[i] = 'A'
	}
	return hm
}

func (hm *headingManagerT) GetNextHeadingNumber(level int, isAppendix bool) string {
	if isAppendix {
		// Pull the current appendix letter at this level instead of a number.
		hm.isLetter[level] = true
	} else {
		// Bump this heading level
		hm.isLetter[level] = false
		hm.headingNumber[level] += 1
	}
	// Reset lower levels
	for i := level + 1; i < maxHeadingDepth; i++ {
		hm.headingNumber[i] = 0
		hm.headingLetter[i] = 'A'
		hm.isLetter[i] = false
	}
	// Build heading number string
	var parts []string
	for i := hm.atLevel; i <= level; i++ {
		if hm.isLetter[i] {
			parts = append(parts, string(hm.headingLetter[i]))
		} else {
			parts = append(parts, strconv.Itoa(hm.headingNumber[i]))
		}
	}
	if isAppendix {
		// Advance the appendix letter for the next heading at this level. We
		// never go past "Z"; we just keep repeating "Z" so the behaviour stays
		// defined beyond 26 appendices.
		if hm.headingLetter[level] < 'Z' {
			hm.headingLetter[level] += 1
		}
	}
	return strings.Join(parts, ".")
}
