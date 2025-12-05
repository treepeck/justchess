package matchmaking

// node represents the Red-Black Tree node.
type node struct {
	parent, left, right *node
	// Node's value.
	leaf int
	// Node's color.
	isRed bool
}

// insert inserts the node with the specified leaf following the Red-Black Tree
// properties and returns the updated root node.
func insert(root *node, leaf int) *node {
	// Shortcut: empty tree.
	if root == nil {
		return &node{leaf: leaf, isRed: false}
	}

	// Find parent.
	curr := root
	var parent *node
	for curr != nil {
		parent = curr
		if leaf < curr.leaf {
			// Go to the left subtree.
			curr = curr.left
		} else {
			// Go to the right subtree.
			curr = curr.right
		}
	}

	// Insert node.
	n := &node{leaf: leaf, isRed: true}
	n.parent = parent
	if leaf < parent.leaf {
		parent.left = n
	} else {
		parent.right = n
	}

	// Exit when parent is root node.
	if parent.parent == nil {
		return root
	}

	// Fix violations of Red-Black Tree properties after BST insertion.
	curr = n
	for curr != nil && curr != root && curr.isRed && curr.parent.isRed {
		if curr.parent == curr.parent.parent.left {
			uncle := curr.parent.parent.right

			if uncle != nil && uncle.isRed {
				curr.parent.parent.isRed = true
				curr.parent.isRed = false
				uncle.isRed = false
				curr = curr.parent
			} else {
				if curr == curr.parent.left {
					rotateRight(curr.parent.parent)

					if root.parent != nil {
						root = curr.parent
					}

					curr.parent.isRed = !curr.parent.isRed
					curr.parent.right.isRed = !curr.parent.right.isRed
				} else {
					rotateLeft(curr.parent)
					rotateRight(curr.parent)

					if root.parent != nil {
						root = curr
					}

					curr.isRed = !curr.isRed
					curr.right.isRed = !curr.right.isRed
				}
			}
		} else {
			uncle := curr.parent.parent.left

			if uncle != nil && uncle.isRed {
				// If the uncle is red, only recoloring required.
				curr.parent.parent.isRed = true
				curr.parent.isRed = false
				uncle.isRed = false
				curr = curr.parent
			} else {
				if curr == curr.parent.left {
					rotateRight(curr.parent)
					rotateLeft(curr.parent)

					curr.isRed = !curr.isRed
					curr.left.isRed = !curr.left.isRed

					if root.parent != nil {
						root = curr
					}
				} else {
					rotateLeft(curr.parent.parent)

					if root.parent != nil {
						root = curr.parent
					}

					curr.parent.isRed = !curr.parent.isRed
					curr.parent.left.isRed = !curr.parent.left.isRed
				}
			}
		}
		curr = curr.parent
	}
	root.isRed = false

	return root
}

// b - black node, r - red node.
//
// Before:
//
//	    (2 b)
//	   /     \
//	(1 b)   (3 r)
//	             \
//	            (4 r)
//
// After:
//
//	 	    (3 b)
//	 	   /     \
//	 	(2 r)   (4 r)
//	   /
//	(1 b)
func rotateLeft(n *node) {
	head := n.right
	head.parent = n.parent
	if n.parent != nil {
		if n == n.parent.left {
			n.parent.left = head
		} else {
			n.parent.right = head
		}
	}

	n.right = head.left
	if n.right != nil {
		n.right.parent = n
	}

	head.left = n
	n.parent = head
}

// b - black node, r - red node.
//
// Before:
//
//	 	    (3 b)
//	 	   /     \
//	 	(2 r)   (4 b)
//	   /
//	(1 r)
//
// After:
//
//	    (2 b)
//	   /     \
//	(1 r)   (3 r)
//	             \
//	            (4 b)
func rotateRight(n *node) {
	head := n.left
	head.parent = n.parent
	if n.parent != nil {
		if n == n.parent.left {
			n.parent.left = head
		} else {
			n.parent.right = head
		}
	}

	n.left = head.right
	if n.left != nil {
		n.left.parent = n
	}

	head.right = n
	n.parent = head
}
