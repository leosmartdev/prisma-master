package rule

// TreeExpression is a structure represents a tree for calculating a boolean logic
type TreeExpression struct {
	Result      bool
	Left, Right *TreeExpression
	Cond        string
}

func (root *TreeExpression) Calculate() bool {
	tree := root
	result := tree.Result
	for {
		switch {
		case tree.Right == nil && tree.Left == nil:
			return result
		case tree.Right != nil:
			tree = tree.Right
			result = tree.Result && result
		default:
			tree = tree.Left
			result = tree.Result || result
		}
	}
}
