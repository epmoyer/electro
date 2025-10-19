package electro

import (
	"strconv"
	"strings"
)

const maxHeadingDepth = 6

type headingManagerT struct {
	atLevel       int
	headingNumber []int
}

func newHeadingManager(level int) *headingManagerT {
	hm := &headingManagerT{
		atLevel:       level,
		headingNumber: make([]int, maxHeadingDepth),
	}
	return hm
}

func (hm *headingManagerT) GetNextHeadingNumber(level int) string {
	// Bump this heading level
	hm.headingNumber[level] += 1
	// Reset lower levels
	for i := level + 1; i < maxHeadingDepth; i++ {
		hm.headingNumber[i] = 0
	}
	// Build heading number string
	var parts []string
	for i := hm.atLevel; i <= level; i++ {
		parts = append(parts, strconv.Itoa(hm.headingNumber[i]))
	}
	return strings.Join(parts, ".")
}
