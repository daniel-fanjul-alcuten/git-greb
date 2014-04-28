Description
===========

It gets a list of all local git branches, gets their tracking branches -local or remote-, builds a graph of dependencies and runs 'git pull' for them in order.

It is a different approach to [TopGit](http://repo.or.cz/w/topgit.git?a=blob;f=README) commands: update and export.

Installation
============

Use 'go get':
<pre>
go get github.com/daniel-fanjul-alcuten/git-greb
</pre>

Update
======

Use 'go get -u'
<pre>
go get -u github.com/daniel-fanjul-alcuten/git-greb
</pre>

Usage
=====

esyscmd(go install && ${GOPATH%%:*}/bin/git-greb -h 2>&1 | sed -e 's/Usage of .*git-greb/Usage of git-greb/' -e 's/^/    /')dnl

Example
=======

These git and git-greb commands:

define(rungreb1, `dnl
dir1=git-greb1.git
rm -rf /tmp/$dir1
git init --bare /tmp/$dir1
cd /tmp/$dir1

dir2=git-greb2.git
rm -rf /tmp/$dir2
git clone /tmp/$dir1 /tmp/$dir2
cd /tmp/$dir2
git config pull.rebase true
git config greb.local false

touch master
git add master
git commit -m "file master"
git push -u origin master

git checkout -b foo -t master
touch foo; git add foo; git commit -m "file foo"
git status --short --branch
## foo...master [ahead 1]

git checkout -b bar -t master
touch bar; git add bar; git commit -m "file bar"
git branch -u foo
git status --short --branch
## bar...foo [ahead 1, behind 1]

git checkout -b baz -t master
touch baz; git add baz; git commit -m "file baz"
git branch -u foo
git status --short --branch
## baz...foo [ahead 1, behind 1]

git checkout -b qux -t bar
git config --add branch.qux.merge baz
git status --short --branch
## qux...bar
## qux...baz')dnl
define(rungreb2, `dnl
git push origin bar:master')dnl
define(rungreb3, `dnl
git rebase -f baz qux')dnl
<pre>
rungreb1()

git-greb -d
# it pulls master (noop)
# it pulls foo (noop)
# it pulls bar (rebase onto foo)
# it pulls baz (rebase onto foo)
# it pulls qux (octopus of bar and baz)
# it deletes master (making foo follow origin/master)

rungreb2()
# it pushes foo and bar

git-greb -d
# it pulls foo (fast forward)
# it pulls bar (noop)
# it pulls baz (rebase onto foo)
# it pulls qux (octopus of bar and baz)
# it deletes bar (making qux follow foo)
# it deletes foo (making baz follow origin/master)

rungreb3()
# it kills the merge

git-greb -d
# it pulls baz (noop)
# it pulls qux (noop)
# it deletes qux
</pre>

generate this output:

<pre>
esyscmd({
  set -x
  rungreb1()
  ${GOPATH%%:*}/bin/git-greb -d
  rungreb2()
  ${GOPATH%%:*}/bin/git-greb -d
  rungreb3()
  ${GOPATH%%:*}/bin/git-greb -d
} 2>&1)
</pre>
