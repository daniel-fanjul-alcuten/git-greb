package main

import (
	"fmt"
	"sort"
)

// ref identifier
type ref struct {
	// full name, i.e. refs/heads/master
	name string
	// i.e. origin
	remote string
}

type node struct {
	ref
	// abbreviated name, i.e. master, origin/master
	branch string
	// nodes that this node depends on, the original value of branch.<name>.merge
	upstreams map[*node]string
	// nodes that depend on this node
	downstreams map[*node]struct{}
}

type graph struct {
	nodes map[ref]*node
}

func newGraph() *graph {
	nodes := make(map[ref]*node)
	return &graph{nodes}
}

func (g *graph) node(r ref) (n *node, ok bool) {
	if n, ok = g.nodes[r]; !ok {
		n = &node{r, "", nil, nil}
		g.nodes[r] = n
	}
	return
}

func (g *graph) edge(from, to *node, reason string) {
	if from.upstreams == nil {
		from.upstreams = make(map[*node]string)
	}
	from.upstreams[to] = reason
	if to.downstreams == nil {
		to.downstreams = make(map[*node]struct{})
	}
	to.downstreams[from] = struct{}{}
}

// nodes n that len(n.upstreams) > 0, downstreams first
func (g *graph) sort() (nodes []*node) {
	pending := make(map[ref]*node, len(g.nodes))
	for r, n := range g.nodes {
		if len(n.upstreams) > 0 {
			pending[r] = n
		}
	}
	for {
		l := len(pending)
		for _, n := range pending {
			h := false
			for u := range n.upstreams {
				if _, ok := pending[u.ref]; ok {
					h = true
					break
				}
			}
			if h {
				continue
			}
			nodes = append(nodes, n)
			delete(pending, n.ref)
		}
		if len(pending) == l {
			break
		}
	}
	return
}

// for adding branch.<downstream>.merge = <upstream>
type addUpstream struct {
	downstream string
	upstream   string
}

// for unsetting branch.<downstream>.merge = <upstream>
type rmUpstream struct {
	downstream string
	upstream   string
}

// for setting branch.<downstream>.remote = <remote>
type setRemote struct {
	downstream string
	remote     string
}

// it returns instances of addUpstream, rmUpstream, setRemote
func (g *graph) remove(n *node) (updates []interface{}) {
	for d := range n.downstreams {
		updates = append(updates, rmUpstream{d.branch, d.upstreams[n]})
		// all upstreams share same remote
		var remote string
		for u := range d.upstreams {
			if u != n {
				remote = u.remote
				break
			}
		}
		for u := range n.upstreams {
			if _, ok := d.upstreams[u]; !ok {
				if remote == "" {
					remote = u.remote
					g.edge(d, u, u.name)
					updates = append(updates, setRemote{d.branch, remote}, addUpstream{d.branch, u.name})
				} else if remote == u.remote {
					g.edge(d, u, u.name)
					updates = append(updates, addUpstream{d.branch, u.name})
				} else {
					// there is only one branch.<downstream>.remote for all upstreams
				}
			}
		}
	}
	delete(g.nodes, n.ref)
	for d := range n.downstreams {
		delete(d.upstreams, n)
	}
	for u := range n.upstreams {
		delete(u.downstreams, n)
	}
	return
}

type nodesort []*node

func (ns nodesort) Len() int {
	return len(ns)
}

func (ns nodesort) Less(i, j int) bool {
	return ns[i].branch < ns[j].branch
}

func (ns *nodesort) Swap(i, j int) {
	(*ns)[i], (*ns)[j] = (*ns)[j], (*ns)[i]
}

func (g *graph) text(n *node, indent, i string, current,
	currentColor, remoteColor, resetColor string) (s string) {
	var nodes nodesort
	if n == nil {
		for _, n := range g.nodes {
			if len(n.upstreams) == 0 {
				nodes = append(nodes, n)
			}
		}
	} else {
		nodes = append(nodes, n)
	}
	sort.Sort(&nodes)
	for _, n := range nodes {
		if n.branch == current {
			s += fmt.Sprintf("%v%v%v%v\n", indent, currentColor, n.branch, resetColor)
		} else if n.remote != "." {
			s += fmt.Sprintf("%v%v%v%v\n", indent, remoteColor, n.branch, resetColor)
		} else {
			s += fmt.Sprintf("%v%v\n", indent, n.branch)
		}
		var downstreams nodesort
		for d := range n.downstreams {
			downstreams = append(downstreams, d)
		}
		sort.Sort(&downstreams)
		for _, d := range downstreams {
			s += g.text(d, indent+i, i, current, currentColor, remoteColor, resetColor)
		}
	}
	return
}

func (g *graph) dot(branch, currentColor, remoteColor string) (s string) {
	var nodes nodesort
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	sort.Sort(&nodes)
	s += "digraph {\n"
	for _, n := range nodes {
		if n.branch == branch && currentColor != "" {
			s += fmt.Sprintf("  \"%v\" [color=\"%[2]v\", fontcolor=\"%[2]v\"];\n",
				n.branch, currentColor)
		} else if n.remote != "." && remoteColor != "" {
			s += fmt.Sprintf("  \"%v\" [color=\"%[2]v\", fontcolor=\"%[2]v\"];\n",
				n.branch, remoteColor)
		} else {
			s += fmt.Sprintf("  \"%v\";\n", n.branch)
		}
		var upstreams nodesort
		for u := range n.upstreams {
			upstreams = append(upstreams, u)
		}
		sort.Sort(&upstreams)
		for _, u := range upstreams {
			var style string
			if u.remote != "." {
				style = " [style=dotted]"
			}
			s += fmt.Sprintf("  \"%v\" -> \"%v\"%v;\n", n.branch, u.branch, style)
		}
	}
	s += "}\n"
	return
}
