package matchmaking

import (
	"testing"
	"strconv"
)

func TestMatchmaking(t *testing.T) {
	cases := []struct {
		mmrs     []float64
		expected [][2]string
	}{
		{
			[]float64{38, 19, 120, 8, 31, 86, 140, 55, 89, 130, 150, 56, 160},
			[][2]string{
				{"5", "8"}, {"11", "7"}, {"0", "4"}, {"1", "3"}, {"6", "10"}, {"9", "2"},
			},
		},
		{
			[]float64{1500, 3000, 2900, 2300, 500, 780, 6000, 200},
			[][2]string{{"2", "1"}, {"4", "5"}},
		},
	}

	for i, tc := range cases {
		pool := NewPool()

		for i, mmr := range tc.mmrs {
			pool.nodes.insert(pool.nodes.spawn(mmr, strconv.Itoa(i)))
		}

		got := make([][2]string, 0)
		for match := range pool.Matchmaking() {
			got = append(got, match)
		}

		if len(tc.expected) != len(got) {
			t.Fatalf("case %d: expected: %v, got: %v", i, tc.expected, got)
		}

		for j, pair := range tc.expected {
			if pair[0] != got[j][0] || pair[1] != got[j][1] {
				t.Fatalf("case %d: expected: %v, got: %v", i, tc.expected, got)
			}
		}
	}
}

func TestExpandMMRGaps(t *testing.T) {
	cases := []struct {
		gaps     []float64
		expected []float64
	}{
		{
			[]float64{3000, 1400, 1500, 500, 200, 12},
			[]float64{2000, 1000, 3000, 700, 1900, 512},
		},
	}

	for i, tc := range cases {
		pool := NewPool()

		for j := range len(tc.gaps) {
			n := pool.nodes.spawn(tc.gaps[j], "")
			n.maxGap = tc.gaps[j]
			pool.nodes.insert(n)
			pool.size++
		}

		pool.ExpandMMRGaps()
		got := bfs(pool.nodes)

		for j, maxGap := range tc.expected {
			if maxGap != got[j].maxGap {
				t.Fatalf("case %d: expected: %v, got: %v", i, tc.expected[j], got[j])
			}
		}
	}
}

func BenchmarkMatchmaking(b *testing.B) {
	pool := NewPool()

	i := float64(0)
	for ; i <= 10000; i++ {
		pool.nodes.insert(pool.nodes.spawn(i, ""))
	}

	cnt := 0
	for b.Loop() {
		for range pool.Matchmaking() {
			cnt++
		}
	}

	b.Logf("matches counter: %d", cnt)
}
