package tree

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_node_insert(t *testing.T) {
	type args struct {
		v int
	}
	tests := []struct {
		name    string
		node    *Node
		args    args
		expects *Node
	}{
		{
			name: "add right without gap",
			node: &Node{},
			args: args{v: 1},
			expects: &Node{
				Min:   0,
				Max:   1,
				clean: true,
			},
		},
		{
			name: "add left without gap",
			node: &Node{Min: 1, Max: 1},
			args: args{v: 0},
			expects: &Node{
				Min:   0,
				Max:   1,
				clean: true,
			},
		},
		{
			name: "add with gap",
			node: &Node{},
			args: args{v: 2},
			expects: func() *Node {
				n := &Node{
					Min:   0,
					Max:   2,
					clean: true,
					d:     1,
				}
				left := &Node{
					Min: 0,
					Max: 0,
					// parent: n,
					clean: true,
				}
				right := &Node{
					Min: 2,
					Max: 2,
					// parent: n,
					clean: true,
				}
				n.L = left
				n.R = right

				return n
			}(),
		},

		{
			name: "add into gap",
			node: func() *Node {
				n := &Node{
					Min: 1,
					Max: 5,
				}
				left := &Node{
					Min: 1,
					Max: 1,
					// parent: n,
				}
				right := &Node{
					Min: 5,
					Max: 5,
					// parent: n,
				}
				n.L = left
				n.R = right

				return n
			}(),
			args: args{v: 3},
			expects: func() *Node {
				n := &Node{
					Min:   3,
					Max:   5,
					clean: true,
					d:     2,
				}
				l := &Node{
					Min: 1,
					Max: 3,
					// parent: n,
					clean: true,
					d:     1,
				}
				r := &Node{
					Min: 5,
					Max: 5,
					// parent: n,
					clean: true,
				}
				ll := &Node{
					Min: 1,
					Max: 1,
					// parent: l,
					clean: true,
				}
				lr := &Node{
					Min: 3,
					Max: 3,
					// parent: l,
					clean: true,
				}
				l.L = ll
				l.R = lr
				n.L = l
				n.R = r

				return n
			}(),
		},
		{
			name: "add into gap and close",
			node: func() *Node {
				n := &Node{
					Min: 3,
					Max: 5,
				}
				l := &Node{
					Min: 1,
					Max: 3,
					// parent: n,
				}
				r := &Node{
					Min: 5,
					Max: 5,
					// parent: n,
				}
				ll := &Node{
					Min: 1,
					Max: 1,
					// parent: l,
				}
				lr := &Node{
					Min: 3,
					Max: 3,
					// parent: l,
				}
				l.L = ll
				l.R = lr
				n.L = l
				n.R = r

				return n
			}(),
			args: args{v: 4},
			expects: func() *Node {
				n := &Node{
					Min:   1,
					Max:   3,
					d:     1,
					clean: true,
				}
				l := &Node{
					Min: 1,
					Max: 1,
					// parent: n,
					clean: true,
				}
				r := &Node{
					Min: 3,
					Max: 5,
					// parent: n,
					clean: true,
				}
				n.L = l
				n.R = r

				return n
			}(),
		},
		{
			name: "close gap",
			node: func() *Node {
				n := &Node{
					Min: 1,
					Max: 3,
				}
				left := &Node{
					Min: 1,
					Max: 1,
					// parent: n,
				}
				right := &Node{
					Min: 3,
					Max: 3,
					// parent: n,
				}
				n.L = left
				n.R = right

				return n
			}(),
			args: args{v: 2},
			expects: func() *Node {
				n := &Node{
					Min:   1,
					Max:   3,
					clean: true,
				}
				return n
			}(),
		},
		{
			name: "close gap Node left",
			node: func() *Node {
				n := &Node{
					Min: 3,
					Max: 5,
				}
				left := &Node{
					Min: 0,
					Max: 3,
					// parent: n,
				}
				right := &Node{
					Min: 5,
					Max: 5,
					// parent: n,
				}
				ll := &Node{
					Min: 0,
					Max: 0,
					// parent: left,
				}
				lr := &Node{
					Min: 3,
					Max: 3,
					// parent: left,
				}
				n.L = left
				n.R = right
				left.L = ll
				left.R = lr

				return n
			}(),
			args: args{v: 4},
			expects: func() *Node {
				n := &Node{
					Min:   0,
					Max:   3,
					clean: true,
					d:     1,
				}
				l := &Node{
					Min:   0,
					Max:   0,
					clean: true,
					// parent: n,
				}
				r := &Node{
					Min:   3,
					Max:   5,
					clean: true,
					// parent: n,
				}
				n.L = l
				n.R = r
				return n
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := InsertInt(tt.node, tt.args.v)
			assert.Equal(t, tt.expects, n)
		})
	}
}

func Test_reduce(t *testing.T) {
	tests := []struct {
		name string
		node *Node
		want *Node
	}{
		{
			name: "leaf",
			node: &Node{Min: 1, Max: 1},
			want: &Node{Min: 1, Max: 1,
				clean: true,
			},
		},
		{
			name: "two leafs, squash",
			node: func() *Node {
				n := &Node{Min: 0, Max: 0}
				l := &Node{Min: 0, Max: 0} // parent: n,

				r := &Node{Min: 1, Max: 1} // parent: n

				n.L = l
				n.R = r
				return n
			}(),
			want: &Node{Min: 0, Max: 1,
				clean: true,
			},
		},
		{
			name: "two leafs, no squash",
			node: func() *Node {
				n := &Node{Min: 0, Max: 0}
				l := &Node{Min: 0, Max: 0} // parent: n

				r := &Node{Min: 2, Max: 2} // parent: n

				n.L = l
				n.R = r
				return n
			}(),
			want: func() *Node {
				n := &Node{Min: 0, Max: 2, clean: true, d: 1}
				l := &Node{Min: 0, Max: 0, clean: true}
				r := &Node{Min: 2, Max: 2, clean: true}
				n.L = l
				n.R = r
				return n
			}(),
		},
		{
			name: "L leaf, R Node, no squash",
			node: func() *Node {
				n := &Node{Min: 0, Max: 0}
				l := &Node{Min: 0, Max: 0}
				r := &Node{Min: 2, Max: 5}
				rl := &Node{Min: 2, Max: 2}
				rr := &Node{Min: 5, Max: 5}
				n.L = l
				n.R = r
				r.L = rl
				r.R = rr
				return n
			}(),
			want: func() *Node {
				n := &Node{Min: 0, Max: 2, clean: true, d: 2}
				l := &Node{Min: 0, Max: 0, clean: true}
				r := &Node{Min: 2, Max: 5, clean: true, d: 1}
				rl := &Node{Min: 2, Max: 2, clean: true}
				rr := &Node{Min: 5, Max: 5, clean: true}
				n.L = l
				n.R = r
				r.L = rl
				r.R = rr
				return n
			}(),
		},
		{
			name: "L leaf, R Node, squash",
			node: func() *Node {
				n := &Node{Min: 0, Max: 0}
				l := &Node{Min: 1, Max: 3}
				r := &Node{Min: 5, Max: 8}
				ll := &Node{Min: 1, Max: 1}
				lr := &Node{Min: 3, Max: 4}
				rl := &Node{Min: 5, Max: 5}
				rr := &Node{Min: 8, Max: 8}
				n.L = l
				n.R = r
				l.L = ll
				l.R = lr
				r.L = rl
				r.R = rr
				return n
			}(),
			want: func() *Node {
				n := &Node{Min: 1, Max: 3, clean: true, d: 2}
				l := &Node{Min: 1, Max: 1, clean: true}
				r := &Node{Min: 5, Max: 8, clean: true, d: 1}
				rl := &Node{Min: 3, Max: 5, clean: true}
				rr := &Node{Min: 8, Max: 8, clean: true}
				n.L = l
				n.R = r
				r.L = rl
				r.R = rr
				return n
			}(),
		},
		{
			name: "L Node, R leaf, no squash",
			node: func() *Node {
				n := &Node{Min: 0, Max: 0}
				l := &Node{Min: 0, Max: 2}
				r := &Node{Min: 5, Max: 5}
				ll := &Node{Min: 0, Max: 0}
				lr := &Node{Min: 2, Max: 2}
				n.L = l
				n.R = r
				l.L = ll
				l.R = lr
				return n
			}(),
			want: func() *Node {
				n := &Node{Min: 2, Max: 5, clean: true, d: 2}
				l := &Node{Min: 0, Max: 2, clean: true, d: 1}
				r := &Node{Min: 5, Max: 5, clean: true}
				ll := &Node{Min: 0, Max: 0, clean: true}
				lr := &Node{Min: 2, Max: 2, clean: true}
				n.L = l
				n.R = r
				l.L = ll
				l.R = lr
				return n
			}(),
		},
		{
			name: "L Node, R leaf, squash",
			node: func() *Node {
				n := &Node{Min: 0, Max: 0}
				l := &Node{Min: 1, Max: 3}
				r := &Node{Min: 5, Max: 5}
				ll := &Node{Min: 1, Max: 1}
				lr := &Node{Min: 3, Max: 4}
				n.L = l
				n.R = r
				l.L = ll
				l.R = lr
				return n
			}(),
			want: func() *Node {
				n := &Node{Min: 1, Max: 3, clean: true, d: 1}
				l := &Node{Min: 1, Max: 1, clean: true}
				r := &Node{Min: 3, Max: 5, clean: true}
				n.L = l
				n.R = r
				return n
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reduce(tt.node)
			t.Log(r)
			assert.Equalf(t, tt.want, r, "reduce(%v)", tt.node)
		})
	}
}

func TestInserts(t *testing.T) {
	n := &Node{}
	max := 1_000_000
	for i := 0; i < max; i = i + 2 {
		n = InsertInt(n, i)
	}
	t.Log(n.depth())
	for i := 1; i < max; i = i + 2 {
		n = InsertInt(n, i)
	}

	assert.Equal(t, &Node{Min: 0, Max: max - 1, clean: true}, n)
}

func TestRandomInserts(t *testing.T) {
	n := &Node{}
	max := 1_000_000
	bla := prepareArray(t, max)
	for i := 0; i < max; i++ {
		n = InsertInt(n, bla[i])
	}
	assert.Equal(t, &Node{Min: 0, Max: max - 1, clean: true}, n)
}

func prepareArray(t *testing.T, max int) []int {
	t.Helper()
	bla := make([]int, max)
	for i := 0; i < max; i++ {
		bla[i] = i
	}
	rand.Shuffle(len(bla), func(i, j int) {
		bla[i], bla[j] = bla[j], bla[i]
	})
	return bla
}

func TestInsert_FromFailed(t *testing.T) {
	n := &Node{}
	n = InsertInt(n, 37)
	t.Log(n)
	n = InsertInt(n, 7)
	t.Log(n)
	n = InsertInt(n, 12)
	t.Log(n)
	n = InsertInt(n, 25)
	t.Log(n)
	n = InsertInt(n, 8)
	t.Log(n)
	n = InsertInt(n, 6)
	t.Log(n)
	t.Log(n.depth())
	t.Log(n.Gaps())
}
