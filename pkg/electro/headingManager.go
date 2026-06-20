package electro

import (
	"strconv"
	"strings"
)

const maxHeadingDepth = 6

// appendixLetterStart is the letter just before "A", so that incrementing it
// up front yields "A" for the first appendix at a level.
const appendixLetterStart = 'A' - 1

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
	// Appendix letters start one before "A"; they are bumped to "A" on first use.
	for i := range hm.headingLetter {
		hm.headingLetter[i] = appendixLetterStart
	}
	return hm
}

func (hm *headingManagerT) GetNextHeadingNumber(level int, isAppendix bool) string {
	if isAppendix {
		// Bump to the next appendix letter at this level instead of a number. We
		// never go past "Z"; we just keep repeating "Z" so the behaviour stays
		// defined beyond 26 appendices. The letter is bumped up front so it stays
		// stable while lower-level headings reference it as a prefix (e.g. "A.1").
		hm.isLetter[level] = true
		if hm.headingLetter[level] < 'Z' {
			hm.headingLetter[level] += 1
		}
	} else {
		// Bump this heading level
		hm.isLetter[level] = false
		hm.headingNumber[level] += 1
	}
	// Reset lower levels
	for i := level + 1; i < maxHeadingDepth; i++ {
		hm.headingNumber[i] = 0
		hm.headingLetter[i] = appendixLetterStart
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
	return strings.Join(parts, ".")
}
