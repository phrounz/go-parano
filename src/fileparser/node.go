package fileparser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/phrounz/go-parano/src/util"
)

//------------------------------------------------------------------------------

type Node struct {
	Bytes           string
	Name            string
	BytesIndexBegin int
	BytesIndexEnd   int
	TypeStr         string
	nodeObj         *ast.Node
	DepthLevel      int
	Father          *Node
	Children        []*Node
	Index           int // index of this node among the children of Father
}

//------------------------------------------------------------------------------

func (n *Node) NextNode() *Node {
	var j = -1
	for i, n2 := range n.Father.Children {
		if n2 == n {
			j = i
		} else if j != -1 && i == j+1 {
			return n2
		}
	}
	return nil
}

//------------------------------------------------------------------------------

func (n *Node) Display() {
	var line = ""
	line += fmt.Sprintf("%s%s [", strings.Repeat("\t", n.DepthLevel), n.TypeStr)
	if n.Name != "" {
		line += fmt.Sprintf("[name=%s]", n.Name)
	}
	for i, b := range n.Bytes {
		if b == '\n' {
			if i < len(n.Bytes) {
				line += fmt.Sprintf("[...]")
			}
			break
		}
		line += fmt.Sprintf("%c", b)
	}
	line += fmt.Sprintf("]")
	util.DebugPrintf(line)
}

//------------------------------------------------------------------------------

func (n *Node) Visit(fnCall func(*Node)) {
	fnCall(n)
	for _, child := range n.Children {
		child.Visit(fnCall)
	}
}

//------------------------------------------------------------------------------

// IsCommentGroupWithComment returns true if this is a CommentGroup with the comment in it.
func (n *Node) IsCommentGroupWithComment(comment string) bool {
	if n.TypeStr == "CommentGroup" {
		for _, child := range n.Children {
			if child.Bytes == comment {
				return true
			}
		}
	}
	return false
}

//------------------------------------------------------------------------------

// ComputeStringExpression compute/concatenate (recursively) a constant string expression,
// e.g. "foo"+"bar"+string("baz") will return ("foobarbaz", false).
//
// This is a work in process, trying to compute all constants string expressions that we can know
// at compile-time.
//
// If it contains something which is not processable, it will return incomplete=true and all
// unprocessable parts will be replaced by "???" in str.
func (n *Node) ComputeStringExpression(fileConstants []ConstValue) (str string, incomplete bool) {

	switch n.TypeStr {

	case "BasicLit": // "foobar"
		str = n.Name
		return

	case "BinaryExpr": // "foo"+"bar"
		if len(n.Children) == 2 && len(n.Bytes) > len(n.Children[0].Bytes)+len(n.Children[1].Bytes) {
			var len0 = len(n.Children[0].Bytes)
			var len1 = strings.Index(n.Bytes[len0:], n.Children[1].Bytes) + len0
			var operator = n.Bytes[len0:len1]
			operator = strings.Replace(operator, " ", "", -1)
			operator = strings.Replace(operator, "\n", "", -1)
			if strings.Index(operator, "+") != -1 && strings.Index(operator, "+=") == -1 {
				var subStr1, incompleteSub1 = n.Children[0].ComputeStringExpression(fileConstants)
				var subStr2, incompleteSub2 = n.Children[1].ComputeStringExpression(fileConstants)
				str = subStr1 + subStr2
				incomplete = (incompleteSub1 || incompleteSub2)
				return
			}
		}

	case "CallExpr":
		if len(n.Children) == 2 && n.Children[0].TypeStr == "Ident" && n.Children[0].Name == "string" && n.Children[1].TypeStr == "BasicLit" { // string("foobar")
			str = n.Children[1].Name
			return
		}

	case "Ident":
		for _, constValue := range fileConstants {
			if n.Name == constValue.Name && (constValue.IsTopLevel || n.IsInScope(constValue.Node)) {
				str = constValue.Value // constFoo (being declared elsewhere in the file as 'const constFoo="foo"')
				return
			}
		}
	}

	str = "???"
	incomplete = true

	return
}

//------------------------------------------------------------------------------

// IsInScope returns true if "n" can access "other" i.e. the father of "other"
// is a father or grand-father or grand-XXX-father or "n".
func (n *Node) IsInScope(other *Node) bool {
	if other.Father != nil {
		var nF = n.Father
		for nF != nil {
			if nF == other.Father {
				return true
			}
			nF = nF.Father
		}
	}
	return false
}

//------------------------------------------------------------------------------
