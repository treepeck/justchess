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
		{
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			&node{
				left: &node{
					left: &node{
						left: &node{
							leaf:  1,
							isRed: false,
						},
						right: &node{
							leaf:  3,
							isRed: false,
						},
						leaf:  2,
						isRed: false,
					},
					right: &node{
						left: &node{
							leaf:  5,
							isRed: false,
						},
						right: &node{
							leaf:  7,
							isRed: false,
						},
						leaf:  6,
						isRed: false,
					},
					leaf:  4,
					isRed: true,
				},
				right: &node{
					left: &node{
						left: &node{
							leaf:  9,
							isRed: false,
						},
						right: &node{
							leaf:  11,
							isRed: false,
						},
						leaf:  10,
						isRed: false,
					},
					right: &node{
						left: &node{
							left: &node{
								leaf:  13,
								isRed: false,
							},
							right: &node{
								leaf:  15,
								isRed: false,
							},
							leaf:  14,
							isRed: true,
						},
						right: &node{
							left: &node{
								leaf:  17,
								isRed: false,
							},
							right: &node{
								leaf:  19,
								isRed: false,
							},
							leaf:  18,
							isRed: true,
						},
						leaf:  16,
						isRed: false,
					},
					leaf:  12,
					isRed: true,
				},
				leaf:  8,
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
			t.Fatalf("case %d failed: %v", i+1, err)
		}

		t.Logf("case %d passed", i+1)
	}
}

func TestSearch(t *testing.T) {
	testcases := []struct {
		root    *node
		leaf    int
		canFind bool
	}{
		{
			&node{
				left: &node{
					leaf: 15,
				},
				right: &node{
					leaf: 20,
				},
				leaf: 19,
			},
			100,
			false,
		},
		{
			&node{
				left: &node{
					leaf: 15,
				},
				right: &node{
					leaf: 20,
				},
				leaf: 19,
			},
			15,
			true,
		},
	}

	for _, tc := range testcases {
		got := search(tc.root, tc.leaf)
		wasFound := got != nil

		if tc.canFind != wasFound {
			t.Fatalf("expected: %t, got: %t", tc.canFind, wasFound)
		} else if tc.canFind && tc.leaf != got.leaf {
			t.Fatalf("expected: %d, got: %d", tc.leaf, got.leaf)
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

func BenchmarkSearch(b *testing.B) {
	var root *node

	// Create a moderately big tree.
	for leaf := 10; leaf <= 10000000; leaf += 10 {
		root = insert(root, leaf)
	}

	for b.Loop() {
		search(root, 10000000)
	}
}
