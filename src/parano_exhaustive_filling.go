package main

import (
	"strings"

	"./fileparser"
	"./util"
)

//------------------------------------------------------------------------------

const constExaustiveFilling = "//!PARANO__EXHAUSTIVE_FILLING"

//------------------------------------------------------------------------------

type featureExhaustiveFilling struct {
	exhaustiveFillingStructs map[string]map[string]bool // true by field by struct name
}

//------------------------------------------------------------------------------

func paranoExhaustiveFillingInit() *featureExhaustiveFilling {
	return &featureExhaustiveFilling{
		exhaustiveFillingStructs: make(map[string]map[string]bool),
	}
}

//------------------------------------------------------------------------------

func paranoExhaustiveFillingVisit(n *fileparser.Node, featureExhaustiveFilling *featureExhaustiveFilling) {

	if n.IsCommentGroupWithComment(constExaustiveFilling) && n.Father != nil {
		var nextNode = n.NextNode()
		if nextNode != nil && nextNode.TypeStr == "TypeSpec" {
			if debugInfo {
				util.DebugPrintf("....... ExhaustiveFilling: >=%s %s<=", nextNode.Name, nextNode.TypeStr)
			}
			var keys = make(map[string]bool)
			for _, child1 := range nextNode.Children {
				if child1.TypeStr == "StructType" {
					for _, child2 := range child1.Children {
						if child2.TypeStr == "FieldList" {
							for _, field := range child2.Children {
								if field.Children[0].TypeStr == "Ident" {
									keys[field.Children[0].Name] = true
								}
							}
						}
					}
				}
			}
			featureExhaustiveFilling.exhaustiveFillingStructs[nextNode.Name] = keys
		}
	}
}

//------------------------------------------------------------------------------

func paranoExhaustiveFillingCheck(n *fileparser.Node, packageName string, featureExhaustiveFilling *featureExhaustiveFilling, filename1 string, filename2 string) (failedAtLeastOnce bool) {
	if fieldsStruct, ok := featureExhaustiveFilling.exhaustiveFillingStructs[n.Name]; ok {
		failedAtLeastOnce = commonCheckExhaustiveFilling(n, fieldsStruct, filename1, filename2)
	}
	return
}

//------------------------------------------------------------------------------

func paranoExhaustiveFillingCheckGlobal(mInfosByPackageName map[string]*packageInfos) (failedAtLeastOnce bool) {

	var mGlobalExhaustiveFillingStructs = make(map[string]map[string]bool)

	for _, packageInfos := range mInfosByPackageName {
		for _, fileInfos := range packageInfos.infosByFile {
			for structName, mFields := range fileInfos.featureExhaustiveFilling.exhaustiveFillingStructs {
				mGlobalExhaustiveFillingStructs[fileInfos.packageName+"."+structName] = mFields
			}
		}
	}

	for _, packageInfos := range mInfosByPackageName {
		for filename1, fileInfos := range packageInfos.infosByFile {
			fileInfos.rootNode.Visit(func(n *fileparser.Node) {
				if fieldsStruct, ok := mGlobalExhaustiveFillingStructs[n.Bytes]; ok {
					failedAtLeastOnce = commonCheckExhaustiveFilling(n, fieldsStruct, filename1, "???")
				}
			})
		}
	}
	return
}

//------------------------------------------------------------------------------

//!PARANO__PRIVATE_TO_FILE
func commonCheckExhaustiveFilling(n *fileparser.Node, fieldsStruct map[string]bool, filename1 string, filename2 string) (failedAtLeastOnce bool) {
	if n.Father.TypeStr == "CompositeLit" {
		var fields = make(map[string]bool)
		for _, keyValue := range n.Father.Children {
			if keyValue.TypeStr == "KeyValueExpr" {
				if keyValue.Children[0].TypeStr == "Ident" {
					fields[keyValue.Children[0].Name] = true
				}
			}
		}
		var missingFields = make([]string, 0)
		for field := range fieldsStruct {
			if _, ok := fields[field]; !ok {
				missingFields = append(missingFields, field)
			}
		}

		if len(missingFields) > 0 {
			util.NotPass("missing fields(s) %s in declaration \"%s{}\" in %s, type declared with %s in %s",
				strings.Join(missingFields, ", "), n.Bytes, filename1, constExaustiveFilling, filename2)
			failedAtLeastOnce = true
		}
	}
	return
}

//------------------------------------------------------------------------------
