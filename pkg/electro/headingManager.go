package electro

import (
	"strconv"
	"strings"
)

const maxHeadingDepth = 6

// appendixLetterStart is the letter just before "A", so that incrementing it
// up front yields "A" for the first appendix.
const appendixLetterStart = 'A' - 1

type headingManagerT struct {
	atLevel           int
	headingNumber     []int
	appendixLetter    rune
	atLevelIsAppendix bool
}

func newHeadingManager(level int) *headingManagerT {
	return &headingManagerT{
		atLevel:        level,
		headingNumber:  make([]int, maxHeadingDepth+1),
		appendixLetter: appendixLetterStart,
	}
}

func (hm *headingManagerT) GetNextHeadingNumber(level int, isAppendix bool) string {
	// We only ever track a single appendix letter, associated with atLevel.
	// The isAppendix flag is therefore only honoured when the heading is at
	// atLevel. This means appendix headings always look like "A", "A.1",
	// "A.1.1" and never "1.A.1", which matches how our documents work.
	if isAppendix && level == hm.atLevel {
		// Bump to the next appendix letter. We never go past "Z"; we just keep
		// repeating "Z" so the behaviour stays defined beyond 26 appendices.
		hm.atLevelIsAppendix = true
		if hm.appendixLetter < 'Z' {
			hm.appendixLetter += 1
		}
	} else {
		// Bump this heading level. A normal heading at atLevel ends any
		// in-progress appendix numbering.
		if level == hm.atLevel {
			hm.atLevelIsAppendix = false
		}
		hm.headingNumber[level] += 1
	}
	// Reset lower levels
	for i := level + 1; i < maxHeadingDepth; i++ {
		hm.headingNumber[i] = 0
	}
	// Build heading number string. The atLevel component is the appendix letter
	// while appendix numbering is active; everything below it stays numeric.
	var parts []string
	for i := hm.atLevel; i <= level; i++ {
		if i == hm.atLevel && hm.atLevelIsAppendix {
			parts = append(parts, string(hm.appendixLetter))
		} else {
			parts = append(parts, strconv.Itoa(hm.headingNumber[i]))
		}
	}
	return strings.Join(parts, ".")
}
