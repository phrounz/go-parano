package util

import "strings"

//------------------------------------------------------------------------------

// WildcardMap is an object which stores several strings, each can have a wildcard.
type WildcardMap struct {
	noWildcard   map[string]interface{}
	withWildcard map[string]wcmData
}

type wcmData struct {
	index int
	value interface{}
}

// NewWildcardMap makes a WildcardMap.
// Each input values shall have zero or one wildcard character '*'.
func NewWildcardMap() WildcardMap {
	return WildcardMap{
		noWildcard:   map[string]interface{}{},
		withWildcard: map[string]wcmData{},
	}
}

// Count returns the number of elements in the WildcardMap.
func (ws *WildcardMap) Count() int {
	return len(ws.noWildcard) + len(ws.withWildcard)
}

func (ws *WildcardMap) Add(key string, value interface{}) {
	var index = strings.Index(key, "*")
	if index == -1 {
		ws.noWildcard[key] = value
	} else {
		ws.withWildcard[key] = wcmData{index: index, value: value}
	}
}

// Find checks if str matches an element in the WildcardMap.
func (ws *WildcardMap) Find(str string) (interface{}, bool) {
	if ws != nil {
		if value, ok := ws.noWildcard[str]; ok {
			return value, true
		}
		for key, wcmData := range ws.withWildcard {
			if wcmData.index != -1 {
				var beginning = key[:wcmData.index]
				var end = key[wcmData.index+1:]
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
