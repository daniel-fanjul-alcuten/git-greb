package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestGraphSort1(t *testing.T) {
	g := newGraph()
	a, _ := g.node(ref{"a", "."})
	b, _ := g.node(ref{"b", "."})
	g.edge(a, b, "ab")
	nodes := g.sort()
	if l := len(nodes); l != 1 {
		t.Error(l)
	} else if n := nodes[0]; n != a {
		t.Error(n)
	}
}

func TestGraphSort2(t *testing.T) {
	g := newGraph()
	a, _ := g.node(ref{"a", "."})
	b, _ := g.node(ref{"b", "."})
	c, _ := g.node(ref{"c", "."})
	g.edge(a, b, "ab")
	g.edge(b, c, "bc")
	nodes := g.sort()
	if l := len(nodes); l != 2 {
		t.Error(l)
	} else if n := nodes[0]; n != b {
		t.Error(n)
	} else if n := nodes[1]; n != a {
		t.Error(n)
	}
}

func TestGraphRemove(t *testing.T) {
	g := newGraph()
	a, _ := g.node(ref{"a", "."})
	b, _ := g.node(ref{"b", "."})
	c, _ := g.node(ref{"c", "."})
	d, _ := g.node(ref{"d", "origin"})
	e, _ := g.node(ref{"e", "origin"})
	for _, n := range []*node{a, b, c, d} {
		n.branch = strings.Repeat(n.name, 2)
	}
	g.edge(a, c, "ac")
	g.edge(b, c, "bc")
	g.edge(c, d, "cd")
	g.edge(c, e, "cd")
	updates := g.remove(c)
	if l := len(updates); l != 8 {
		t.Error(l)
	} else {
		if u := updates[0]; u != (rmUpstream{"aa", "ac"}) {
			t.Error(u)
		}
		if u := updates[1]; u != (setRemote{"aa", "origin"}) {
			t.Error(u)
		}
		if u := updates[2]; u != (addUpstream{"aa", "d"}) {
			t.Error(u)
		}
		if u := updates[3]; u != (addUpstream{"aa", "e"}) {
			t.Error(u)
		}
		if u := updates[4]; u != (rmUpstream{"bb", "bc"}) {
			t.Error(u)
		}
		if u := updates[5]; u != (setRemote{"bb", "origin"}) {
			t.Error(u)
		}
		if u := updates[6]; u != (addUpstream{"bb", "d"}) {
			t.Error(u)
		}
		if u := updates[7]; u != (addUpstream{"bb", "e"}) {
			t.Error(u)
		}
	}
	if l := len(g.nodes); l != 4 {
		t.Error(l)
	}
	if _, ok := g.nodes[c.ref]; ok {
		t.Error(ok)
	}
	if l := len(a.upstreams); l != 2 {
		t.Error(l)
	}
	if _, ok := a.upstreams[d]; !ok {
		t.Error(ok)
	}
	if _, ok := a.upstreams[e]; !ok {
		t.Error(ok)
	}
	if l := len(a.downstreams); l != 0 {
		t.Error(l)
	}
	if l := len(b.upstreams); l != 2 {
		t.Error(l)
	}
	if _, ok := b.upstreams[d]; !ok {
		t.Error(ok)
	}
	if _, ok := b.upstreams[e]; !ok {
		t.Error(ok)
	}
	if l := len(b.downstreams); l != 0 {
		t.Error(l)
	}
	if l := len(c.upstreams); l != 2 {
		t.Error(l)
	}
	if _, ok := c.upstreams[d]; !ok {
		t.Error(ok)
	}
	if _, ok := c.upstreams[e]; !ok {
		t.Error(ok)
	}
	if l := len(c.downstreams); l != 2 {
		t.Error(l)
	}
	if _, ok := c.downstreams[a]; !ok {
		t.Error(ok)
	}
	if _, ok := c.downstreams[b]; !ok {
		t.Error(ok)
	}
	if l := len(d.upstreams); l != 0 {
		t.Error(l)
	}
	if l := len(d.downstreams); l != 2 {
		t.Error(l)
	}
	if _, ok := d.downstreams[a]; !ok {
		t.Error(ok)
	}
	if _, ok := d.downstreams[b]; !ok {
		t.Error(ok)
	}
	if l := len(e.upstreams); l != 0 {
		t.Error(l)
	}
	if l := len(e.downstreams); l != 2 {
		t.Error(l)
	}
	if _, ok := e.downstreams[a]; !ok {
		t.Error(ok)
	}
	if _, ok := e.downstreams[b]; !ok {
		t.Error(ok)
	}
}

func TestGraphText(t *testing.T) {
	g := newGraph()
	a, _ := g.node(ref{"a", "."})
	b, _ := g.node(ref{"b", "."})
	c, _ := g.node(ref{"c", "."})
	d, _ := g.node(ref{"d", "origin"})
	for _, n := range []*node{a, b, c, d} {
		n.branch = strings.Repeat(n.name, 2)
	}
	g.edge(a, b, "ab")
	g.edge(b, c, "bc")
	g.edge(b, d, "bd")
	s := bufio.NewScanner(bytes.NewBufferString(g.text(nil, "", "  ", "aa", "^", "$")))
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "cc" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  bb" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "    ^aa$" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "dd" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  bb" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "    ^aa$" {
		t.Error(e)
	}
	if v := s.Scan(); v {
		t.Fatal(v)
	}
}

func TestGraphDotWithoutColor(t *testing.T) {
	g := newGraph()
	a, _ := g.node(ref{"a", "."})
	b, _ := g.node(ref{"b", "."})
	c, _ := g.node(ref{"c", "origin"})
	for _, n := range []*node{a, b, c} {
		n.branch = strings.Repeat(n.name, 2)
	}
	g.edge(a, b, "ab")
	g.edge(a, c, "ac")
	s := bufio.NewScanner(bytes.NewBufferString(g.dot("bb", "")))
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "digraph {" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"aa\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"aa\" -> \"bb\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"aa\" -> \"cc\" [style=dotted];" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"bb\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"cc\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "}" {
		t.Error(e)
	}
	if v := s.Scan(); v {
		t.Fatal(v)
	}
}

func TestGraphDotWithColor(t *testing.T) {
	g := newGraph()
	a, _ := g.node(ref{"a", "."})
	b, _ := g.node(ref{"b", "."})
	c, _ := g.node(ref{"c", "origin"})
	for _, n := range []*node{a, b, c} {
		n.branch = strings.Repeat(n.name, 2)
	}
	g.edge(a, b, "ab")
	g.edge(a, c, "ac")
	s := bufio.NewScanner(bytes.NewBufferString(g.dot("bb", "green")))
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "digraph {" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"aa\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"aa\" -> \"bb\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"aa\" -> \"cc\" [style=dotted];" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"bb\" [color=\"green\", fontcolor=\"green\"];" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "  \"cc\";" {
		t.Error(e)
	}
	if v := s.Scan(); !v {
		t.Fatal(v)
	}
	if e := s.Text(); e != "}" {
		t.Error(e)
	}
	if v := s.Scan(); v {
		t.Fatal(v)
	}
}
