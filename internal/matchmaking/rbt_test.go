package matchmaking

import (
	"fmt"
	"testing"
)

func compare(root1, root2 *node) error {
	if root1.leaf != root2.leaf || root1.isRed != root2.isRed {
		return fmt.Errorf("expected: %v, got: %v", *root1, *root2)
	}

	if root1.left != nil {
		return compare(root1.left, root2.left)
	} else if root1.right != nil {
		return compare(root1.right, root2.right)
	}

	return nil
}

// Validated with https://www.cs.usfca.edu/~galles/visualization/RedBlack.html
func TestInsert(t *testing.T) {
	testcases := []struct {
		leaves   []int
		expected *node
	}{
		{
			[]int{10, 20, 19},
			&node{
				left: &node{
					leaf:  10,
					isRed: true,
				},
				right: &node{
					leaf:  20,
					isRed: true,
				},
				leaf:  19,
				isRed: false,
			},
		},
		{
			[]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
			&node{
				left: &node{
					left: &node{
						leaf:  10,
						isRed: false,
					},
					right: &node{
						leaf:  30,
						isRed: false,
					},
					leaf:  20,
					isRed: false,
				},
				right: &node{
					left: &node{
						leaf:  50,
						isRed: false,
					},
					right: &node{
						left: &node{
							leaf:  70,
							isRed: false,
						},
						right: &node{
							right: &node{
								leaf:  100,
								isRed: true,
							},
							leaf:  90,
							isRed: false,
						},
						leaf:  80,
						isRed: true,
					},
					leaf:  60,
					isRed: false,
				},
				leaf:  40,
				isRed: false,
			},
		},
		{
			[]int{10, 20, 30, 40, 50, 60, 70, 65, 45, 44, 68, 71, 72},
			&node{
				left: &node{
					left: &node{
						leaf:  10,
						isRed: false,
					},
					right: &node{
						leaf:  30,
						isRed: false,
					},
					leaf:  20,
					isRed: false,
				},
				right: &node{
					left: &node{
						left: &node{
							leaf:  44,
							isRed: true,
						},
						right: &node{
							leaf:  50,
							isRed: true,
						},
						leaf:  45,
						isRed: false,
					},
					right: &node{
						left: &node{
							leaf:  65,
							isRed: false,
						},
						right: &node{
							left: &node{
								leaf:  70,
								isRed: true,
							},
							right: &node{
								leaf:  72,
								isRed: true,
							},
							leaf:  71,
							isRed: false,
						},
						leaf:  68,
						isRed: true,
					},
					leaf:  60,
					isRed: false,
				},
				leaf:  40,
				isRed: false,
			},
		},
		{
			[]int{50, 40, 55, 30},
			&node{
				left: &node{
					left: &node{
						leaf:  30,
						isRed: true,
					},
					leaf:  40,
					isRed: false,
				},
				right: &node{
					leaf:  55,
					isRed: false,
				},
				leaf:  50,
				isRed: false,
			},
		},
		{
			[]int{50, 40, 30},
			&node{
				left: &node{
					leaf:  30,
					isRed: true,
				},
				right: &node{
					leaf:  50,
					isRed: true,
				},
				leaf:  40,
				isRed: false,
			},
		},
		{
			[]int{50, 40, 45},
			&node{
				left: &node{
					leaf:  40,
					isRed: true,
				},
				right: &node{
					leaf:  50,
					isRed: true,
				},
				leaf:  45,
				isRed: false,
			},
		},
	}

	for i, tc := range testcases {
		var root *node

		for i := range tc.leaves {
			root = insert(root, tc.leaves[i])
		}

		if err := compare(tc.expected, root); err != nil {
			t.Fatalf("case %d failed: %v", i, err)
		}
	}
}

func TestRotateLeft(t *testing.T) {
	root := &node{leaf: 10}
	right := &node{
		parent: root,
		leaf:   20,
		isRed:  false,
	}
	root.right = right
	right.right = &node{
		parent: right,
		leaf:   30,
		isRed:  false,
	}

	rotateLeft(root)
	root = root.parent

	expected := &node{
		left: &node{
			leaf:  10,
			isRed: false,
		},
		right: &node{
			leaf:  30,
			isRed: false,
		},
		leaf:  20,
		isRed: false,
	}

	if err := compare(expected, root); err != nil {
		t.Fatal(err)
	}
}

func TestRotateRight(t *testing.T) {
	root := &node{leaf: 30}
	left := &node{
		parent: root,
		leaf:   20,
		isRed:  false,
	}
	root.left = left
	left.left = &node{
		parent: left,
		leaf:   10,
		isRed:  false,
	}

	rotateRight(root)
	root = root.parent

	expected := &node{
		left: &node{
			leaf:  10,
			isRed: false,
		},
		right: &node{
			leaf:  30,
			isRed: false,
		},
		leaf:  20,
		isRed: false,
	}

	if err := compare(expected, root); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkInsert(b *testing.B) {
	var root *node

	leaf := 0
	for b.Loop() {
		leaf += 10
		root = insert(root, leaf)
	}
}
