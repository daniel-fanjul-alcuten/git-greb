package main

import (
	"fmt"
	"sort"
	"strings"
)

type graph struct {
	direct  map[string]string
	reverse map[string]map[string]struct{}
}

func newGraph() *graph {
	direct := make(map[string]string)
	reverse := make(map[string]map[string]struct{})
	return &graph{direct, reverse}
}

func (g *graph) add(branch, tracking string) {

	g.direct[branch] = tracking
	if m, ok := g.reverse[tracking]; ok {
		m[branch] = struct{}{}
	} else {
		m = make(map[string]struct{})
		g.reverse[tracking] = m
		m[branch] = struct{}{}
	}
}

func (g *graph) sort() (branches []string) {

	direct := make(map[string]string, len(g.direct))
	for k, v := range g.direct {
		if v != "" {
			direct[k] = v
		}
	}

	for {
		l := len(direct)
		for branch, tracking := range direct {
			if _, ok := direct[tracking]; ok {
				// the branch depends on another local branch
				continue
			}
			branches = append(branches, branch)
			delete(direct, branch)
		}
		if len(direct) == l {
			break
		}
	}

	return
}

func (g *graph) remove(branch string) (tracking string, branches []string) {

	tracking = g.direct[branch]
	for b := range g.reverse[branch] {
		branches = append(branches, b)
	}

	delete(g.direct, branch)
	delete(g.reverse, branch)
	delete(g.reverse[tracking], branch)

	for _, b := range branches {
		g.add(b, tracking)
	}
	return
}

func (g *graph) toText(branch, indent string, count int) (s string) {
	branches := make([]string, 0, len(g.direct))
	for b := range g.reverse[branch] {
		branches = append(branches, b)
	}
	if branch == "" {
		for b := range g.reverse {
			if b != "" {
				if _, ok := g.direct[b]; !ok {
					branches = append(branches, b)
				}
			}
		}
	}
	sort.Strings(branches)
	for _, b := range branches {
		s += fmt.Sprintf("%v%v\n", strings.Repeat(indent, count), b)
		s += g.toText(b, indent, count+1)
	}
	return
}

func (g *graph) toDot() (s string) {
	branches := make([]string, 0, len(g.direct))
	for b := range g.direct {
		branches = append(branches, b)
	}
	sort.Strings(branches)
	s += "digraph {\n"
	for _, b := range branches {
		t := g.direct[b]
		if t == "" {
			s += fmt.Sprintf("  \"%v\";\n", b)
		} else {
			s += fmt.Sprintf("  \"%v\" -> \"%v\";\n", b, t)
		}
	}
	s += "}\n"
	return
}
