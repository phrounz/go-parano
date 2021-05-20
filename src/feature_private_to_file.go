package src

import (
	"regexp"

	"github.com/phrounz/go-parano/src/fileparser"
	"github.com/phrounz/go-parano/src/util"
)

//------------------------------------------------------------------------------

const constPrivateToFileComment = "//!PARANO__PRIVATE_TO_FILE"
const constLocalPrivateStuffLineRegexp1 = "\n//\\s+LOCAL PRIVATE STUFF\\s*\n"
const constLocalPrivateStuffLineRegexp2 = "\n//\\s+PRIVATE LOCAL STUFF\\s*\n"
const constLocalPrivateStuffLineRegexp3 = "\n//\\s+LOCAL PRIVATE STUFF \\(= the content below is not expected to be used outside this file\\)\\s*\n"

//------------------------------------------------------------------------------

type featurePrivateToFile struct {
	locationLocalPrivateStuff int
	privateToFileDecl         map[string]bool
}

//------------------------------------------------------------------------------

func ParanoPrivateToFileInit(fileBytes []byte) *featurePrivateToFile {

	var locationLocalPrivateStuff = -1
	var loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp1).FindIndex(fileBytes)
	if len(loc) > 0 {
		locationLocalPrivateStuff = loc[1]
	} else {
		loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp2).FindIndex(fileBytes)
		if len(loc) > 0 {
			locationLocalPrivateStuff = loc[1]
		} else {
			loc = regexp.MustCompile(constLocalPrivateStuffLineRegexp3).FindIndex(fileBytes)
			if len(loc) > 0 {
				locationLocalPrivateStuff = loc[1]
			}
		}
	}
	if util.IsDebug() {
		util.DebugPrintf("  locationLocalPrivateStuff: %d", locationLocalPrivateStuff)
	}
	return &featurePrivateToFile{
		locationLocalPrivateStuff: locationLocalPrivateStuff,
		privateToFileDecl:         make(map[string]bool),
	}
}

//------------------------------------------------------------------------------

func ParanoPrivateToFileVisit(n *fileparser.Node, feat *featurePrivateToFile) {

	if feat.locationLocalPrivateStuff != -1 && n.BytesIndexBegin > feat.locationLocalPrivateStuff && n.DepthLevel <= 2 && len(n.Children) > 0 {
		checkPrivateToFile(n.Children[0], feat)
	}

	if n.IsCommentGroupWithComment(constPrivateToFileComment) && n.Father != nil {
		checkPrivateToFile(n, feat)
	}
}

//------------------------------------------------------------------------------

func checkPrivateToFile(n *fileparser.Node, feat *featurePrivateToFile) {

	if n.Father.TypeStr == "GenDecl" {
		for _, n2 := range n.Father.Children {
			if n2.TypeStr == "ValueSpec" {
				if len(n2.Children) >= 2 {
					var name = n2.Children[0].Bytes
					if util.IsDebug() {
						util.DebugPrintf("....... PrivateToFile: ValueSpec: >= %s <=", name)
					}
					feat.privateToFileDecl[name] = true
					break
				}
			}
		}
	} else if n.Father.TypeStr == "FuncDecl" {
		if util.IsDebug() {
			util.DebugPrintf("....... PrivateToFile: FuncDecl: >= %s %s <=", n.Father.Name, n.Father.TypeStr)
		}
		feat.privateToFileDecl[n.Father.Name] = true
	} else {
		var nextNode = n.NextNode()
		if nextNode != nil && nextNode.TypeStr == "TypeSpec" {
			if util.IsDebug() {
				util.DebugPrintf("....... PrivateToFile: TypeSpec: >= %s %s <=", nextNode.Name, nextNode.TypeStr)
			}
			feat.privateToFileDecl[nextNode.Name] = true
		}
	}

}

//------------------------------------------------------------------------------

func ParanoPrivateToFileCheck(n *fileparser.Node, featurePrivateToFile *featurePrivateToFile, filename1 string, filename2 string, ignorePrivateToFile util.WildcardMap) {

	if filename1 != filename2 {
		if _, ok := featurePrivateToFile.privateToFileDecl[n.Name]; ok {
			if _, ok2 := ignorePrivateToFile.Find(n.Name); ok2 {
				if util.IsDebug() {
					util.DebugPrintf("Ignoring private to file: %s when used in %s (from %s)", n.Name, filename1, filename2)
				}
			} else {
				util.NotPass("Cannot use %s in %s, declared as private to file in %s", n.Name, filename1, filename2)
			}
		}
	}
	return
}

//------------------------------------------------------------------------------
