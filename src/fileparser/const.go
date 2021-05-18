package fileparser

//------------------------------------------------------------------------------

type ConstValue struct {
	Name  string
	Value string
	Node  *Node
}

//------------------------------------------------------------------------------

func retrieveAllConstants(rootNode *Node) (fileConstants []ConstValue) {
	rootNode.Visit(func(n *Node) {
		if n.TypeStr == "DeclStmt" {
			if len(n.Children) == 2 && n.Children[0].TypeStr == "GenDecl" && n.Children[1].TypeStr == "ValueSpec" &&
				len(n.Children[1].Children) == 2 && n.Children[1].Children[0].TypeStr == "Ident" {
				if len(n.Bytes) > len(n.Children[1].Bytes) && n.Bytes[:len(n.Bytes)-len(n.Children[1].Bytes)] == "const " {
					// TODO ComputeBinaryExpressionStrings(fileConstants) works only for constants already found (i.e. upper in the source code)
					var constValue, incomplete = n.Children[1].Children[1].ComputeStringExpression(fileConstants)
					if !incomplete {
						fileConstants = append(fileConstants, ConstValue{
							Name:  n.Children[1].Children[0].Name,
							Value: constValue,
							Node:  n,
						})
					}
				}
			}
		}
	})
	return
}

//------------------------------------------------------------------------------
