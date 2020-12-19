package main

import (
	"fmt"

	"./node"
)

//------------------------------------------------------------------------------

func isCommentGroupWithComment(n *node.Node, comment string) bool {
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

func notPass(message string) {
	fmt.Printf("%sDO NOT PASS%s: %s\n", colorRed, colorDefault, message)
}

//------------------------------------------------------------------------------
