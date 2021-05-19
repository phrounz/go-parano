/*
	https://medium.com/justforfunc/understanding-go-programs-with-go-parser-c4e88a6edb87
*/
package fileparser

import (
	"go/ast"
	"go/token"
	"reflect"
	"strings"
)

//------------------------------------------------------------------------------

type visitor struct {
	father     *visitor
	node       *Node
	pkgDecl    map[*ast.GenDecl]bool
	depthLevel int
}

//------------------------------------------------------------------------------

func newVisitor(f *ast.File) visitor {
	decls := make(map[*ast.GenDecl]bool)
	for _, decl := range f.Decls {
		if v, ok := decl.(*ast.GenDecl); ok {
			decls[v] = true
		}
	}
	return visitor{
		pkgDecl:    decls,
		depthLevel: 0,
		node:       &Node{DepthLevel: 0},
	}
}

//------------------------------------------------------------------------------

func (v visitor) Visit(nodeObj ast.Node) ast.Visitor {

	if nodeObj == nil {
		return nil
	}

	var n = &Node{
		DepthLevel:      v.depthLevel + 1,
		Bytes:           string(fileBytes[nodeObj.Pos()-1 : nodeObj.End()-1]),
		BytesIndexBegin: int(nodeObj.Pos() - 1),
		BytesIndexEnd:   int(nodeObj.End() - 1),
		nodeObj:         &nodeObj,
		Father:          v.node,
		TypeStr:         "(unknown)",
	}

	v.node.Children = append(v.node.Children, n)

	// if PrintDebug {
	// 	var line = ""
	// 	line += fmt.Sprintf("%s%T [", strings.Repeat("\t", n.DepthLevel), nodeObj)
	// 	for i, b := range n.Bytes {
	// 		if b == '\n' {
	// 			if i < len(n.Bytes) {
	// 				line += fmt.Sprintf("[...]")
	// 			}
	// 			break
	// 		}
	// 		line += fmt.Sprintf("%c", b)
	// 	}
	// 	line += fmt.Sprintf("]")
	// 	util.DebugPrintf(line)
	// }

	n.TypeStr = reflect.TypeOf(nodeObj).Elem().Name()
	//fmt.Println(name)

	switch d := nodeObj.(type) {
	case *ast.CallExpr:
		n.Name = n.Bytes[0:strings.Index(n.Bytes, "(")]
	case *ast.FuncDecl:
		n.Name = d.Name.Name
	case *ast.TypeSpec:
		n.Name = d.Name.Name
	case *ast.GenDecl:
		if d.Tok != token.VAR {
			return v
		}
	case *ast.BasicLit:
		var nodeCasted = nodeObj.(*ast.BasicLit)
		switch nodeCasted.Kind {
		case token.STRING:
			n.Name = removeQuotes(nodeCasted.Value)
		}
		n.TypeStr = "BasicLit"
		// for _, spec := range d.Specs {
		// 	if value, ok := spec.(*ast.ValueSpec); ok {
		// 		for _, name := range value.Names {
		// 			if name.Name == "_" {
		// 				continue
		// 			}
		// 		}
		// 	}
		// }
	case *ast.Ident:
		n.Name = n.Bytes
		if n.Name == "" {
			panic("empty bytes for *ast.Ident ?")
		}
		n.TypeStr = "Ident"
	}

	return &visitor{
		father:     &v,
		depthLevel: n.DepthLevel,
		node:       n,
	}
}

//------------------------------------------------------------------------------

// remove '"' or '`'
func removeQuotes(value string) string {
	var useQuotes = (value[0] == '"')      // false if '`'
	var output = (value[1 : len(value)-1]) // remove begin/end quotes '"' or '`'
	if useQuotes {
		var parts1 = strings.Split(output, "\\\\") // replace escaped characters
		var parts2 []string
		for _, part := range parts1 {
			parts2 = append(parts2, strings.Replace(part, "\\", "", -1)) // replace escaped characters
		}
		output = strings.Join(parts2, "\\")
	}
	return output
}

//------------------------------------------------------------------------------
