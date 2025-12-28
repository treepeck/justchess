package matchmaking

import (
	"testing"
)

func TestInsertNode(t *testing.T) {
	testcases := []struct {
		values      []int
		expectedBFS []int
	}{
		{
			[]int{41, 38, 31, 12, 19, 8, 7, 40, 45, 49, 48},
			[]int{38, 19, 41, 8, 31, 40, 48, 7, 12, 45, 49},
		},
		{
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			[]int{8, 4, 12, 2, 6, 10, 16, 1, 3, 5, 7, 9, 11, 14, 18, 13, 15, 17, 19, 20},
		},
		{
			[]int{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			[]int{13, 9, 17, 5, 11, 15, 19, 3, 7, 10, 12, 14, 16, 18, 20, 2, 4, 6, 8, 1},
		},
	}

	for i, tc := range testcases {
		tree := newRedBlackTree()

		for _, val := range tc.values {
			tree.insertNode(tree.spawn(val))
		}

		got := bfs(tree)

		if len(tc.expectedBFS) != len(got) {
			t.Fatalf("case %d: expected: %v, got: %v", i, tc.expectedBFS, got)
		}

		for j, val := range tc.expectedBFS {
			if val != got[j] {
				t.Fatalf("case %d: expected: %v, got: %v", i, tc.expectedBFS, got)
			}
		}
	}
}

func TestRemoveNode(t *testing.T) {
	testcases := []struct {
		values      []int
		value       int
		expectedBFS []int
	}{
		{
			[]int{41, 38, 31, 12, 19, 8},
			41,
			[]int{19, 12, 38, 8, 31},
		},
		{
			[]int{60, 30, 80, 20, 50, 70, 90, 10, 100},
			60,
			[]int{50, 20, 80, 10, 30, 70, 90, 100},
		},
		{
			[]int{50, 20, 80, 10, 30, 70, 90, 100},
			80,
			[]int{50, 20, 90, 10, 30, 70, 100},
		},
		{
			[]int{60, 30, 80, 20, 50, 70, 90, 10, 100},
			30,
			[]int{60, 20, 80, 10, 50, 70, 90, 100},
		},
		{
			[]int{38, 19, 86, 10, 31, 55, 89, 8, 56, 120},
			10,
			[]int{38, 19, 86, 8, 31, 55, 89, 56, 120},
		},
		{
			[]int{38, 19, 120, 8, 31, 86, 140, 55, 89, 130, 150, 56, 160},
			120,
			[]int{86, 38, 140, 19, 55, 89, 150, 8, 31, 56, 130, 160},
		},
		{
			[]int{20, 10, 30, 25},
			10,
			[]int{25, 20, 30},
		},
		{
			[]int{20, 10, 30, 15},
			30,
			[]int{15, 10, 20},
		},
	}

	for i, tc := range testcases {
		tree := newRedBlackTree()

		for _, val := range tc.values {
			tree.insertNode(tree.spawn(val))
		}

		tree.removeNode(tree.search(tc.value))

		got := bfs(tree)

		if len(tc.expectedBFS) != len(got) {
			t.Fatalf("case %d: expected: %v, got: %v", i, tc.expectedBFS, got)
		}

		if len(tc.expectedBFS) > 0 {
			for j, val := range tc.expectedBFS {
				if val != got[j] {
					t.Fatalf("case %d: expected: %v, got: %v", i, tc.expectedBFS, got)
				}
			}
		}
	}
}

func bfs(t *redBlackTree) []int {
	res := make([]int, 0)
	if t.root == t.leaf {
		return res
	}

	visit := make([]*redBlackNode, 1)
	visit[0] = t.root

	for len(visit) > 0 {
		node := visit[0]
		visit = visit[1:]

		res = append(res, node.value)

		if node.left != t.leaf {
			visit = append(visit, node.left)
		}

		if node.right != t.leaf {
			visit = append(visit, node.right)
		}
	}

	return res
}

func BenchmarkInsertNode(b *testing.B) {
	tree := newRedBlackTree()

	i := 10
	for b.Loop() {
		tree.insertNode(tree.spawn(i))
		i += 10
	}
}

func BenchmarkRemoveNode(b *testing.B) {
	tree := newRedBlackTree()

	i := 0
	for ; i <= 10000000; i++ {
		tree.insertNode(tree.spawn(i))
	}

	for b.Loop() {
		i -= 1
		n := tree.search(i)
		if n != nil {
			tree.removeNode(tree.search(i))
		}
	}
}
