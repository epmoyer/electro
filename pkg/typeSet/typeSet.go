package typeSet

import (
	"fmt"
	"slices"
	"strings"
)

// Set is a custom set type
type Set map[string]struct{}

func NewSet() Set {
	return make(Set)
}

// AsSlice the names as a string slice
func (s Set) AsSlice() []string {
	result := make([]string, 0, len(s))
	for key := range s {
		result = append(result, key)
	}
	return result
}

// Add adds an element to the set
func (s Set) Add(value string) {
	s[value] = struct{}{}
}

// Remove removes an element from the set
func (s Set) Remove(value string) {
	delete(s, value)
}

// Contains checks if the set contains the given element
func (s Set) Contains(value string) bool {
	_, exists := s[value]
	return exists
}

// Check is the set is equal to another set
func (s Set) Equal(other Set) bool {
	if len(s) != len(other) {
		return false
	}
	for key := range s {
		if !other.Contains(key) {
			return false
		}
	}
	return true
}

func (s Set) Difference(other Set) Set {
	result := make(Set)
	for key := range s {
		if !other.Contains(key) {
			result.Add(key)
		}
	}
	return result
}

// String returns a string representation of the set, as a slice, for printing etc.
func (s Set) String() string {
	// return fmt.Sprintf("%v", s.AsSlice())
	elements := make([]string, 0, len(s))
	for k := range s {
		elements = append(elements, fmt.Sprintf("%q", k))
	}
	// Sort so that the string representation is deterministic.
	slices.Sort(elements)
	return "[" + strings.Join(elements, ", ") + "]"
}

func MakeSetFromSlice(slice []string) Set {
	result := make(Set)
	for _, value := range slice {
		result.Add(value)
	}
	return result
}
