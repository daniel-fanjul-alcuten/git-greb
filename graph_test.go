package main

import (
	"testing"
)

func TestGraphAddOne(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	if l := len(g.direct); l != 1 {
		t.Error(l)
	}
	if b, ok := g.direct["foo"]; !ok {
		t.Error(ok)
	} else if b != "bar" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["foo"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 1 {
		t.Fatal(l)
	}
	if b := s[0]; b != "foo" {
		t.Error(b)
	}
}

func TestGraphAddTwoIndependent(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("baz", "qux")
	if l := len(g.direct); l != 2 {
		t.Error(l)
	}
	if b, ok := g.direct["foo"]; !ok {
		t.Error(ok)
	} else if b != "bar" {
		t.Error(b)
	}
	if b, ok := g.direct["baz"]; !ok {
		t.Error(ok)
	} else if b != "qux" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 2 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["foo"]; !ok {
		t.Error(ok)
	}
	if r, ok := g.reverse["qux"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["baz"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 2 {
		t.Fatal(l)
	}
	if b := s[0]; b != "foo" {
		t.Error(b)
	}
	if b := s[1]; b != "baz" {
		t.Error(b)
	}
}

func TestGraphAddTwoDependent(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("baz", "bar")
	if l := len(g.direct); l != 2 {
		t.Error(l)
	}
	if b, ok := g.direct["foo"]; !ok {
		t.Error(ok)
	} else if b != "bar" {
		t.Error(b)
	}
	if b, ok := g.direct["baz"]; !ok {
		t.Error(ok)
	} else if b != "bar" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 2 {
		t.Error(l)
	} else if _, ok := r["foo"]; !ok {
		t.Error(ok)
	} else if _, ok := r["baz"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 2 {
		t.Fatal(l)
	}
	if b := s[0]; b != "foo" {
		t.Error(b)
	}
	if b := s[1]; b != "baz" {
		t.Error(b)
	}
}

func TestGraphAddTwoChain(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("bar", "baz")
	if l := len(g.direct); l != 2 {
		t.Error(l)
	}
	if b, ok := g.direct["foo"]; !ok {
		t.Error(ok)
	} else if b != "bar" {
		t.Error(b)
	}
	if b, ok := g.direct["bar"]; !ok {
		t.Error(ok)
	} else if b != "baz" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 2 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["foo"]; !ok {
		t.Error(ok)
	}
	if r, ok := g.reverse["baz"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["bar"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 2 {
		t.Fatal(l)
	}
	if b := s[0]; b != "bar" {
		t.Error(b)
	}
	if b := s[1]; b != "foo" {
		t.Error(b)
	}
}

func TestGraphDeleteOne(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	tr, br := g.remove("foo")
	if tr != "bar" {
		t.Error(tr)
	}
	if l := len(br); l != 0 {
		t.Fatal(l)
	}
	if l := len(g.direct); l != 0 {
		t.Error(l)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 0 {
		t.Error(l)
	}
	s := g.sort()
	if l := len(s); l != 0 {
		t.Fatal(l)
	}
}

func TestGraphAddOneDelete(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	tr, br := g.remove("foo")
	if tr != "bar" {
		t.Error(tr)
	}
	if l := len(br); l != 0 {
		t.Fatal(l)
	}
	if l := len(g.direct); l != 0 {
		t.Error(l)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 0 {
		t.Error(l)
	}
	s := g.sort()
	if l := len(s); l != 0 {
		t.Fatal(l)
	}
}

func TestGraphAddTwoIndependentDelete(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("baz", "qux")
	tr, br := g.remove("foo")
	if tr != "bar" {
		t.Error(tr)
	}
	if l := len(br); l != 0 {
		t.Fatal(l)
	}
	if l := len(g.direct); l != 1 {
		t.Error(l)
	}
	if b, ok := g.direct["baz"]; !ok {
		t.Error(ok)
	} else if b != "qux" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 2 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 0 {
		t.Error(l)
	}
	if r, ok := g.reverse["qux"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["baz"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 1 {
		t.Fatal(l)
	}
	if b := s[0]; b != "baz" {
		t.Error(b)
	}
}

func TestGraphAddTwoDependentDelete(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("baz", "bar")
	tr, br := g.remove("foo")
	if tr != "bar" {
		t.Error(tr)
	}
	if l := len(br); l != 0 {
		t.Fatal(l)
	}
	if l := len(g.direct); l != 1 {
		t.Error(l)
	}
	if b, ok := g.direct["baz"]; !ok {
		t.Error(ok)
	} else if b != "bar" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["baz"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 1 {
		t.Fatal(l)
	}
	if b := s[0]; b != "baz" {
		t.Error(b)
	}
}

func TestGraphAddTwoChainDeleteFirst(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("bar", "baz")
	tr, br := g.remove("foo")
	if tr != "bar" {
		t.Error(tr)
	}
	if l := len(br); l != 0 {
		t.Fatal(l)
	}
	if l := len(g.direct); l != 1 {
		t.Error(l)
	}
	if b, ok := g.direct["bar"]; !ok {
		t.Error(ok)
	} else if b != "baz" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 2 {
		t.Error(l)
	}
	if r, ok := g.reverse["bar"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 0 {
		t.Error(l)
	}
	if r, ok := g.reverse["baz"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["bar"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 1 {
		t.Fatal(l)
	}
	if b := s[0]; b != "bar" {
		t.Error(b)
	}
}

func TestGraphAddTwoChainDeleteLast(t *testing.T) {
	g := newGraph()
	g.add("foo", "bar")
	g.add("bar", "baz")
	tr, br := g.remove("bar")
	if tr != "baz" {
		t.Error(tr)
	}
	if l := len(br); l != 1 {
		t.Fatal(l)
	}
	if b := br[0]; b != "foo" {
		t.Error(b)
	}
	if l := len(g.direct); l != 1 {
		t.Error(l)
	}
	if b, ok := g.direct["foo"]; !ok {
		t.Error(ok)
	} else if b != "baz" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse["baz"]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["foo"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 1 {
		t.Fatal(l)
	}
	if b := s[0]; b != "foo" {
		t.Error(b)
	}
}

func TestGraphAddEmptyTracking(t *testing.T) {
	g := newGraph()
	g.add("foo", "")
	if l := len(g.direct); l != 1 {
		t.Error(l)
	}
	if b, ok := g.direct["foo"]; !ok {
		t.Error(ok)
	} else if b != "" {
		t.Error(b)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse[""]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 1 {
		t.Error(l)
	} else if _, ok := r["foo"]; !ok {
		t.Error(ok)
	}
	s := g.sort()
	if l := len(s); l != 0 {
		t.Fatal(l)
	}
}

func TestGraphAddEmptyTrackingDelete(t *testing.T) {
	g := newGraph()
	g.add("foo", "")
	tr, br := g.remove("foo")
	if tr != "" {
		t.Error(tr)
	}
	if l := len(br); l != 0 {
		t.Fatal(l)
	}
	if l := len(g.direct); l != 0 {
		t.Error(l)
	}
	if l := len(g.reverse); l != 1 {
		t.Error(l)
	}
	if r, ok := g.reverse[""]; !ok {
		t.Error(ok)
	} else if l := len(r); l != 0 {
		t.Error(l)
	}
	s := g.sort()
	if l := len(s); l != 0 {
		t.Fatal(l)
	}
}

func TestGraphToText(t *testing.T) {
	g := newGraph()
	g.add("foo", "")
	g.add("bar", "baz")
	g.add("baz", "qux")
	g.add("quux", "qux")
	if str := g.toText("", " ", 0); str != "foo\nqux\n baz\n  bar\n quux\n" {
		t.Error(str)
	}
}
