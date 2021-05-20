package fileparser

//------------------------------------------------------------------------------

type ConstValue struct {
	Name       string
	Value      string
	Node       *Node
	IsTopLevel bool
}

//------------------------------------------------------------------------------

func retrieveAllConstants(rootNode *Node) (fileConstants []ConstValue) {
	fileConstants = []ConstValue{}
	rootNode.Visit(func(n *Node) {
		if n.TypeStr == "DeclStmt" {
			readConst(n, 0, false, &fileConstants)
		} else if n.Father != nil && n.Father.TypeStr == "File" && n.TypeStr == "GenDecl" {
			readConst(n.Father, n.Index, true, &fileConstants)
		}
	})
	return
}

//------------------------------------------------------------------------------

func readConst(n *Node, index int, isTopLevel bool, fileConstants *[]ConstValue) {
	if len(n.Children[index:]) >= 2 && n.Children[index].TypeStr == "GenDecl" && n.Children[index+1].TypeStr == "ValueSpec" {
		var genDecl = n.Children[index]
		var valueSpec = n.Children[index+1]
		if len(valueSpec.Children) == 2 && valueSpec.Children[0].TypeStr == "Ident" {
			if len(genDecl.Bytes) > len(valueSpec.Bytes) && genDecl.Bytes[:len(genDecl.Bytes)-len(valueSpec.Bytes)] == "const " {
				// TODO ComputeBinaryExpressionStrings(fileConstants) works only for constants already found (i.e. upper in the source code)
				var constValue, incomplete = valueSpec.Children[1].ComputeStringExpression(*fileConstants)
				//fmt.Printf("===readConst %s %s: %v", valueSpec.Children[0].Name, valueSpec.Children[1].Bytes, incomplete)
				if !incomplete {
					*fileConstants = append(*fileConstants, ConstValue{
						Name:       valueSpec.Children[0].Name,
						Value:      constValue,
						Node:       n,
						IsTopLevel: isTopLevel,
					})
				}
			}
		}
	}
}

//------------------------------------------------------------------------------
