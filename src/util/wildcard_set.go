package util

import "strings"

//------------------------------------------------------------------------------

// WildcardSet is an object which stores several strings, each can have a wildcard.
type WildcardSet map[string]int

// NewWildcardSet makes a WildcardSet for using with InWildcardSet().
// Each input values shall have zero or one wildcard character '*'.
func NewWildcardSet(input []string) WildcardSet {
	var ws = make(map[string]int)
	for _, element := range input {
		ws[element] = strings.Index(element, "*")
	}
	return ws
}

// InWildcardSet checks if str matches an element in the WildcardSet.
func InWildcardSet(str string, ws WildcardSet) bool {
	if ws != nil {
		if wildcardIndex, ok := ws[str]; ok {
			if wildcardIndex == -1 { // If the value in the map is -1, it means that it has no wildcard.
				return true
			}
		}
		for key, wildcardIndex := range ws {
			// If the value in the map is !=-1, it means that the key has one character '*'
			// which is a wilcard when finding str.
			if wildcardIndex != -1 {
				var beginning = key[:wildcardIndex]
				var end = key[wildcardIndex+1:]
				if len(str) >= len(beginning)+len(end) &&
					(len(beginning) == 0 || beginning == str[:len(beginning)]) &&
					(len(end) == 0 || end == str[len(str)-len(end):]) {
					//fmt.Printf("matches: %s %s\n", str, key)
					return true
				}
			}
		}
	}
	return false
}

//------------------------------------------------------------------------------
