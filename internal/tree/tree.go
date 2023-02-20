package tree

import (
	"bytes"
	"encoding/json"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	p = message.NewPrinter(language.English)
)

type Node struct {
	Min   int   `json:"Min"`
	Max   int   `json:"Max"`
	L     *Node `json:"L"`
	R     *Node `json:"R"`
	clean bool
	d     int
}

type Gap struct {
	Min int `json:"Min"`
	Max int `json:"Max"`
}

func InsertInt(n *Node, v int) *Node {
	nn := insert(n, &Node{
		Min:   v,
		Max:   v,
		clean: true,
	})
	return nn
}

func insert(n *Node, v *Node) *Node {
	if n == nil {
		return v
	}
	n.clean = false
	if n.isLeaf() {
		if v.Min >= n.Min && v.Max <= n.Max {
			n.clean = true
			return n
		}
		// insert as new right Node, copy Node as its own left
		if v.Min >= n.Max {
			n.R = v
			n.L = &Node{
				Min:   n.Min,
				Max:   n.Max,
				clean: true,
			}
			nn := reduce(n)
			nn.d = nn.depth()
			nn.clean = true
			return nn
		}
		// insert as new left Node, copy Node as its own right
		if v.Max < n.Max {
			n.L = v
			n.R = &Node{
				Min:   n.Min,
				Max:   n.Max,
				clean: true,
			}
			nn := reduce(n)
			nn.d = nn.depth()
			nn.clean = true
			return nn
		}
	}
	if v.Max < n.Max {
		n.L = insert(n.L, v)
		nn := reduce(n)
		nn.d = nn.depth()
		nn.clean = true
		return nn
	}

	n.R = insert(n.R, v)
	nn := reduce(n)
	nn.d = nn.depth()
	nn.clean = true
	return nn
}

func find(v int, n *Node) bool {
	if n == nil {
		return false
	}
	if n.isLeaf() {
		return v >= n.Min && v <= n.Max
	}
	if v == n.Min || v == n.Max {
		return true
	}
	if v > n.Min && v < n.Max {
		return false
	}
	if v < n.Min {
		return find(v, n.L)
	}
	return find(v, n.R)
}

func reduce(n *Node) *Node {
	if n == nil {
		return nil
	}
	if n.isLeaf() || n.clean {
		n.clean = true
		return n
	}
	n.L = reduce(n.L)
	n.R = reduce(n.R)

	if n.R.isLeaf() && n.L.isLeaf() {
		nn := reduceTwoLeaves(n)
		nn.d = nn.depth()
		nn.clean = true
		return nn
	}

	n.Min, _ = n.L.FindMax()
	n.Max, _ = n.R.FindMin()

	// no gap, this node can be removed
	if n.Max-n.Min == 1 {
		if n.L.isLeaf() {
			nn := &Node{
				Min:   n.L.Min,
				Max:   n.R.Min,
				clean: true,
			}
			n.R.L = nn

			nr := n.R
			nr.clean = true
			return nr
		}
		if n.R.isLeaf() {
			nn := &Node{
				Min:   n.L.Max,
				Max:   n.R.Max,
				clean: true,
			}
			n.L.R = nn

			nl := n.L
			nl.clean = true
			return nl
		}
		return (changeParent(n))
	}
	r := balance(n)
	r.d = r.depth()
	r.clean = true
	return r
}

func balance(n *Node) *Node {
	// if n.isLeaf() {
	// 	n.clean = true
	// 	return n
	// }

	if n.R.depth() > n.L.depth()+1 {
		oldL := n.R.L
		nn := n.R
		n.R = oldL
		nn.L = n
		nn.clean = false
		return nn
	}
	if n.L.depth() > n.R.depth()+1 {
		oldR := n.L.R
		nn := n.L
		n.L = oldR
		nn.R = n
		nn.clean = false
		return nn
	}
	return n
}

func changeParent(n *Node) *Node {
	l := n.L
	r := n.R
	nn := insert(l, r)
	nn.clean = true
	nn.d = nn.depth()
	return nn
}

func reduceTwoLeaves(n *Node) *Node {
	if n.R.Min-n.L.Max == 1 {
		n.Min = n.L.Min
		n.Max = n.R.Max
		n.L = nil
		n.R = nil
		n.d = 0
		return n
	}
	n.Min = n.L.Max
	n.Max = n.R.Min
	return n
}

func (n *Node) isLeaf() bool {
	return n.L == nil && n.R == nil
}

func (n *Node) FindMax() (int, bool) {
	if n == nil {
		return 0, false
	}
	if n.isLeaf() {
		return n.Max, true
	}
	return n.R.FindMax()
}
func (n *Node) FindMin() (int, bool) {
	if n == nil {
		return 0, false
	}
	if n.isLeaf() {
		return n.Min, true
	}
	return n.L.FindMin()
}

func (n Node) String() string {
	s := ""
	g := n.Gaps()
	gm := 0
	if len(g) > 0 {
		gm = g[0].Min
	}
	var m int
	for _, gap := range g {
		m += gap.Length()
	}
	s += p.Sprintf("TotalMissing: %v ", m)
	if min, ok := n.FindMin(); ok {
		s += p.Sprintf("Min: %v ", min)
	} else {
		s += p.Sprintf("Min: %v ", "-")
	}
	if max, ok := n.FindMax(); ok {
		s += p.Sprintf("Max: %v ", max)
	} else {
		s += p.Sprintf("Max: %v ", "-")
	}
	s += p.Sprintf("Min Missing: %v ", gm)
	s += p.Sprintf("Missing: %v", g)
	return s
}

func PrettyString(str string) (string, error) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

func (n *Node) depth() int {
	if n.isLeaf() {
		return 0
	}
	if n.clean {
		return n.d
	}
	l := n.L.depth()
	r := n.R.depth()
	if l > r {
		n.d = l + 1
		return l + 1
	}
	n.d = r + 1
	return r + 1
}

func (n *Node) Gaps() []Gap {
	if n == nil || n.isLeaf() {
		return []Gap{}
	}
	g := append(n.L.Gaps(), Gap{Min: n.Min + 1, Max: n.Max - 1})
	return append(g, n.R.Gaps()...)
}

func (g Gap) Length() int {
	return g.Max - (g.Min - 1)
}

func (g Gap) String() string {
	if g.Min == g.Max {
		return p.Sprintf("[%v]", g.Min)
	}
	return p.Sprintf("[%v -> %v]", g.Min, g.Max)
}
