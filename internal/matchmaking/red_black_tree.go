package matchmaking

// redBlackTree implements the Red-Black Tree data structure described in
// "Introduction to Algorithms 3rd Edition" by Cormen et al.
//
// It is used by the matchmaking algorithm because it allows player MMRs to be
// stored in sorted order while supporting efficient lookups.
//
// TODO: The main downside of a Red-Black Tree is poor concurrent performance.
// It may be better to use the SyncList data structure instead.
type redBlackTree struct {
	root, leaf *node
}

type node struct {
	parent, left, right *node
	id                  string
	mmr                 float64
	// Max allowed MMR gap. See package documentation for more details.
	maxGap float64
	// isRed is used to maintain the Red-Black Tree properties.
	isRed bool
}

func newRedBlackTree() *redBlackTree { leaf := &node{}; return &redBlackTree{root: leaf, leaf: leaf} }

// insert inserts the [node] and fixes any violations of the Red-Black Tree properties.
func (t *redBlackTree) insert(z *node) {
	y := t.leaf
	x := t.root

	for x != t.leaf {
		y = x
		if z.mmr < x.mmr {
			x = x.left
		} else {
			x = x.right
		}
	}

	z.parent = y
	if y == t.leaf {
		t.root = z
	} else if z.mmr < y.mmr {
		y.left = z
	} else {
		y.right = z
	}

	t.fixInsert(z)
}

// remove removes the [node] and fixes any violations of the Red-Black Tree properties.
func (t *redBlackTree) remove(z *node) {
	var x *node

	y := z
	wasRed := y.isRed

	if z.left == t.leaf {
		x = z.right
		t.transplant(z, z.right)
	} else if z.right == t.leaf {
		x = z.left
		t.transplant(z, z.left)
	} else {
		y = t.findMax(z.left)
		wasRed = y.isRed

		x = y.left
		if y.parent == z {
			x.parent = y
		} else {
			t.transplant(y, y.left)
			y.left = z.left
			y.left.parent = y
		}

		t.transplant(z, y)
		y.right = z.right
		y.right.parent = y
		y.isRed = z.isRed
	}

	if !wasRed {
		t.fixRemove(x)
	}
}

func (t *redBlackTree) fixInsert(z *node) {
	for z.parent.isRed {
		if z.parent == z.parent.parent.left {
			uncle := z.parent.parent.right
			if uncle.isRed {
				// Case 1: z's parent is the left child, and the uncle is red.
				// Perform recoloring.
				recolor(z, uncle)
				z = z.parent.parent
			} else {
				if z == z.parent.right {
					// Case 2: z is the right child, its parent is the left
					// child, and the uncle is black.
					// Perform left rotation.
					z = z.parent
					t.rotateLeft(z)
				}

				// Case 3: z is the left child, its parent is the left
				// child, and the uncle is black.
				// Perform some recoloring and right rotation.
				z.parent.isRed = false
				z.parent.parent.isRed = true
				t.rotateRight(z.parent.parent)
			}
		} else {
			uncle := z.parent.parent.left
			if uncle.isRed {
				// Case 4: z's parent is the right child, and the uncle is red.
				// Perform recoloring.
				recolor(z, uncle)
				z = z.parent.parent
			} else {
				if z == z.parent.left {
					// Case 5: z is the left child, its parent is the right
					// child, and the uncle is black.
					// Perform right rotation.
					z = z.parent
					t.rotateRight(z)
				}

				// Case 6: z is the right child, its parent is the right
				// child, and the uncle is black.
				// Perform recoloring and right and left rotation.
				z.parent.isRed = false
				z.parent.parent.isRed = true
				t.rotateLeft(z.parent.parent)
			}
		}
	}
	t.root.isRed = false
}

func (t *redBlackTree) fixRemove(x *node) {
	for x != t.root && !x.isRed {
		if x == x.parent.left {
			sibling := x.parent.right
			if sibling.isRed {
				// Case 1: x is the left child and its sibling is red.
				// Perform recoloring and left rotation.
				sibling.isRed = false
				x.parent.isRed = true
				t.rotateLeft(x.parent)
				sibling = x.parent.right
			}

			if !sibling.left.isRed && !sibling.right.isRed {
				// Case 2: sibling is black, and both its children are black.
				// Perform recoloring.
				sibling.isRed = true
				x = x.parent
			} else {
				if !sibling.right.isRed {
					// Case 3: sibling is black, its left child is red, and
					// right child is black.
					// Perform recoloring and right rotation.
					sibling.left.isRed = false
					sibling.isRed = true
					t.rotateRight(sibling)
					sibling = x.parent.right
				}

				// Case 4: sibling is black, and its right child is red.
				// Perform recoloring and left rotation.
				sibling.isRed = x.parent.isRed
				x.parent.isRed = false
				sibling.right.isRed = false
				t.rotateLeft(x.parent)
				x = t.root
			}
		} else {
			sibling := x.parent.left
			if sibling.isRed {
				// Case 5: x is the right child and its sibling is red.
				// Perform recoloring and right rotation.
				sibling.isRed = false
				x.parent.isRed = true
				t.rotateRight(x.parent)
				sibling = x.parent.left
			}

			if !sibling.left.isRed && !sibling.right.isRed {
				// Case 6: sibling is black, and both its children are black.
				// Perform recoloring.
				sibling.isRed = true
				x = x.parent
			} else {
				if !sibling.left.isRed {
					// Case 7: sibling is black, its right child is black, and
					// left child is black.
					// Perform recoloring and left rotation.
					sibling.right.isRed = false
					sibling.isRed = true
					t.rotateLeft(sibling)
					sibling = x.parent.left
				}

				// Case 8: sibling is black, and its left child is red.
				// Perform recoloring and right rotation.
				sibling.isRed = x.parent.isRed
				x.parent.isRed = false
				sibling.left.isRed = false
				t.rotateRight(x.parent)
				x = t.root
			}
		}
	}
	x.isRed = false
}

// Performs left rotation.  Assumes that x.right != t.leaf.
// The letters a, b, and g represent arbitrary subtrees.
//
// Before:
//
//			(x)
//	       /   \
//	      a    (y)
//	          /   \
//	     	 b     g
//
// After:
//
//		    (y)
//	       /   \
//	     (x)    g
//	    /   \
//	   a     b
func (t *redBlackTree) rotateLeft(x *node) {
	y := x.right

	// Turn x's right subtree into y's left subtree.
	x.right = y.left
	if x.right != t.leaf {
		x.right.parent = x
	}

	// Link y to x's parent.
	y.parent = x.parent
	switch x {
	case t.root:
		t.root = y
	case y.parent.left:
		y.parent.left = y
	default:
		y.parent.right = y
	}

	// Put x on y's left subtree.
	y.left = x
	x.parent = y
}

// Performs right rotation.  Assumes that x.left != t.leaf.
// The letters a, b, and g represent arbitrary subtrees.
//
// Before:
//
//	     (x)
//	    /   \
//	  (y)    g
//	 /   \
//	a     b
//
// After:
//
//	  (y)
//	 /   \
//	a    (x)
//	    /   \
//	   b     g
func (t *redBlackTree) rotateRight(x *node) {
	y := x.left

	// Turn x's left subtree into y's right subtree.
	x.left = y.right
	if x.left != t.leaf {
		x.left.parent = x
	}

	// Link y to x's parent.
	y.parent = x.parent
	switch x {
	case t.root:
		t.root = y
	case y.parent.left:
		y.parent.left = y
	default:
		y.parent.right = y
	}

	// Put x on y's right subtree.
	y.right = x
	x.parent = y
}

// Finds the node with the biggest value in the specified tree.
func (t *redBlackTree) findMax(z *node) *node {
	for z.right != t.leaf {
		z = z.right
	}
	return z
}

// Finds the node with the smallest value in the specified tree.
func (t *redBlackTree) findMin(z *node) *node {
	for z.left != t.leaf {
		z = z.left
	}
	return z
}

// Recolors nodes to resolve the cases 1 and 4 of the [fixInsert] function.
func recolor(z, uncle *node) {
	z.parent.isRed = false
	uncle.isRed = false
	z.parent.parent.isRed = true
}

// Copies v into u's position.
func (t *redBlackTree) transplant(u, v *node) {
	if u.parent == t.leaf {
		t.root = v
	} else if u == u.parent.left {
		u.parent.left = v
	} else {
		u.parent.right = v
	}
	v.parent = u.parent
}

// creates a new node with specified value and default fields.
func (t *redBlackTree) spawn(mmr float64, id string) *node {
	return &node{
		mmr:    mmr,
		id:     id,
		maxGap: DefaultMaxGap,
		isRed:  true,
		parent: t.leaf,
		left:   t.leaf,
		right:  t.leaf,
	}
}

// search searches for the [node] in the subtree.
func search(n *node, mmr float64, id string) *node {
	// While n is not leaf.
	for n.left != nil && n.right != nil {
		if n.mmr > mmr {
			n = n.left
		} else if n.mmr < mmr {
			n = n.right
		} else if n.id == id {
			return n
		} else {
			right := search(n.right, mmr, id)
			if right != nil {
				return right
			}
			left := search(n.left, mmr, id)
			if left != nil {
				return left
			}
			break
		}
	}
	return nil
}
