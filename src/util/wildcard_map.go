package util

import "strings"

//------------------------------------------------------------------------------

type wcmData struct {
	wildcardIndex int
	value         interface{}
}

// WildcardMap is an object which stores several strings, each can have a wildcard.
type WildcardMap map[string]wcmData

// NewWildcardMap makes a WildcardMap.
// Each input values shall have zero or one wildcard character '*'.
func NewWildcardMap() WildcardMap {
	return make(map[string]wcmData)
}

func (ws WildcardMap) Add(key string, value interface{}) {
	ws[key] = wcmData{
		wildcardIndex: strings.Index(key, "*"),
		value:         value,
	}
}

// Find checks if str matches an element in the WildcardMap.
func (ws WildcardMap) Find(str string) (interface{}, bool) {
	if ws != nil {
		if wcmData, ok := ws[str]; ok {
			if wcmData.wildcardIndex == -1 { // If the value in the map is -1, it means that it has no wildcard.
				return wcmData.value, true
			}
		}
		for key, wcmData := range ws {
			// If the value in the map is !=-1, it means that the key has one character '*'
			// which is a wilcard when finding str.
			if wcmData.wildcardIndex != -1 {
				var beginning = key[:wcmData.wildcardIndex]
				var end = key[wcmData.wildcardIndex+1:]
				if len(str) >= len(beginning)+len(end) &&
					(len(beginning) == 0 || beginning == str[:len(beginning)]) &&
					(len(end) == 0 || end == str[len(str)-len(end):]) {
					//fmt.Printf("matches: %s %s\n", str, key)
					return wcmData.value, true
				}
			}
		}
	}
	return nil, false
}

//------------------------------------------------------------------------------
