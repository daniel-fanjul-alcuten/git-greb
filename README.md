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

Usage
=====

<pre>
Usage of git-greb:
  -c=false: it checks out instead of pulling (checkout).
  -d=false: it deletes fully merged branches after pulling (delete).
  -dot=false: it shows the graph of dependencies in dot format and exits (dot graph).
  -i=false: it rebases with --interactive (interactive).
  -l=false: it only pulls local tracking branches (local).
  -m=false: it pulls with --no-rebase (merge).
  -n=false: it does not run the commands (noop).
  -q=false: it does not print the command lines (quiet).
  -r=false: it pulls with --rebase (rebase).
  -s=false: it does not pull at all (skip).
  -t=false: it shows the graph of dependencies in text and exits (text graph).
  -v=false: it explains intermediate steps (verbose).
  -x=false: it shows the graph of dependencies in xlib and exits (xlib graph).
</pre>

Example
=======

These git and git-greb commands:

<pre>
dir1=git-greb1.git
rm -rf /tmp/$dir1
git init --bare /tmp/$dir1
cd /tmp/$dir1

dir2=git-greb2.git
rm -rf /tmp/$dir2
git clone /tmp/$dir1 /tmp/$dir2
cd /tmp/$dir2
git config pull.rebase true

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
## qux...baz

git-greb -d
# it pulls master (noop)
# it pulls foo (noop)
# it pulls bar (rebase onto foo)
# it pulls baz (rebase onto foo)
# it pulls qux (octopus of bar and baz)
# it deletes master (making foo follow origin/master)

git push origin bar:master
# it pushes foo and bar

git-greb -d
# it pulls foo (fast forward)
# it pulls bar (noop)
# it pulls baz (rebase onto foo)
# it pulls qux (octopus of bar and baz)
# it deletes bar (making qux follow foo)
# it deletes foo (making baz follow origin/master)

git rebase -f baz qux
# it kills the merge

git-greb -d
# it pulls baz (noop)
# it pulls qux (noop)
# it deletes qux
</pre>

generate this output:

<pre>
+ dir1=git-greb1.git
+ rm -rf /tmp/git-greb1.git
+ git init --bare /tmp/git-greb1.git
Initialized empty Git repository in /tmp/git-greb1.git/
+ cd /tmp/git-greb1.git
+ dir2=git-greb2.git
+ rm -rf /tmp/git-greb2.git
+ git clone /tmp/git-greb1.git /tmp/git-greb2.git
Cloning into '/tmp/git-greb2.git'...
warning: You appear to have cloned an empty repository.
done.
+ cd /tmp/git-greb2.git
+ git config pull.rebase true
+ touch master
+ git add master
+ git commit -m file master
[master (root-commit) 0f81a47] file master
 1 file changed, 0 insertions(+), 0 deletions(-)
 create mode 100644 master
+ git push -u origin master
To /tmp/git-greb1.git
 * [new branch]      master -> master
Branch master set up to track remote branch master from origin.
+ git checkout -b foo -t master
Switched to a new branch 'foo'
Branch foo set up to track local branch master.
+ touch foo
+ git add foo
+ git commit -m file foo
[foo 2eca399] file foo
 1 file changed, 0 insertions(+), 0 deletions(-)
 create mode 100644 foo
+ git status --short --branch
## foo...master [ahead 1]
+ git checkout -b bar -t master
Switched to a new branch 'bar'
Branch bar set up to track local branch master.
+ touch bar
+ git add bar
+ git commit -m file bar
[bar 35eaadb] file bar
 1 file changed, 0 insertions(+), 0 deletions(-)
 create mode 100644 bar
+ git branch -u foo
Branch bar set up to track local branch foo.
+ git status --short --branch
## bar...foo [ahead 1, behind 1]
+ git checkout -b baz -t master
Switched to a new branch 'baz'
Branch baz set up to track local branch master.
+ touch baz
+ git add baz
+ git commit -m file baz
[baz 96c2a76] file baz
 1 file changed, 0 insertions(+), 0 deletions(-)
 create mode 100644 baz
+ git branch -u foo
Branch baz set up to track local branch foo.
+ git status --short --branch
## baz...foo [ahead 1, behind 1]
+ git checkout -b qux -t bar
Switched to a new branch 'qux'
Branch qux set up to track local branch bar.
+ git config --add branch.qux.merge baz
+ git status --short --branch
## qux...bar
+ /home/dfanjul/lib/go/bin/git-greb -d
greb: git checkout master
Switched to branch 'master'
Your branch is up-to-date with 'origin/master'.
greb: git pull
Current branch master is up to date.
greb: git checkout foo
Switched to branch 'foo'
Your branch is ahead of 'master' by 1 commit.
greb: git pull
From .
 * branch            master     -> FETCH_HEAD
Current branch foo is up to date.
greb: git checkout baz
Switched to branch 'baz'
Your branch and 'foo' have diverged,
and have 1 and 1 different commit each, respectively.
greb: git pull
From .
 * branch            foo        -> FETCH_HEAD
First, rewinding head to replay your work on top of it...
Applying: file baz
greb: git checkout bar
Switched to branch 'bar'
Your branch and 'foo' have diverged,
and have 1 and 1 different commit each, respectively.
greb: git pull
From .
 * branch            foo        -> FETCH_HEAD
First, rewinding head to replay your work on top of it...
Applying: file bar
greb: git checkout qux
Switched to branch 'qux'
Your branch and 'bar' have diverged,
and have 1 and 2 different commits each, respectively.
greb: git pull --no-rebase
From .
 * branch            bar        -> FETCH_HEAD
 * branch            baz        -> FETCH_HEAD
Trying simple merge with 5ce26620118505e84d5603ed8c1bf49d61e6e620
Trying simple merge with b9ff5d695441dd0f6745187eddb72495e557de68
Merge made by the 'octopus' strategy.
greb: git branch -D master
Deleted branch master (was 0f81a47).
greb: git config --unset branch.foo.merge ^refs/heads/master$
greb: git config branch.foo.remote origin
greb: git config --add branch.foo.merge refs/heads/master
+ git push origin bar:master
To /tmp/git-greb1.git
   0f81a47..5ce2662  bar -> master
+ /home/dfanjul/lib/go/bin/git-greb -d
greb: git checkout foo
Switched to branch 'foo'
Your branch is behind 'origin/master' by 1 commit, and can be fast-forwarded.
greb: git pull
First, rewinding head to replay your work on top of it...
Fast-forwarded foo to 5ce26620118505e84d5603ed8c1bf49d61e6e620.
greb: git checkout baz
Switched to branch 'baz'
Your branch and 'foo' have diverged,
and have 1 and 1 different commit each, respectively.
greb: git pull
From .
 * branch            foo        -> FETCH_HEAD
First, rewinding head to replay your work on top of it...
Applying: file baz
greb: git checkout bar
Switched to branch 'bar'
Your branch is up-to-date with 'foo'.
greb: git pull
From .
 * branch            foo        -> FETCH_HEAD
Current branch bar is up to date.
greb: git checkout qux
Switched to branch 'qux'
Your branch is ahead of 'bar' by 3 commits.
greb: git pull --no-rebase
From .
 * branch            bar        -> FETCH_HEAD
 * branch            baz        -> FETCH_HEAD
Merge made by the 'recursive' strategy.
greb: git branch -D bar
Deleted branch bar (was 5ce2662).
greb: git config --unset branch.qux.merge ^refs/heads/bar$
greb: git config --add branch.qux.merge refs/heads/foo
greb: git branch -D foo
Deleted branch foo (was 5ce2662).
greb: git config --unset branch.baz.merge ^refs/heads/foo$
greb: git config branch.baz.remote origin
greb: git config --add branch.baz.merge refs/heads/master
greb: git config --unset branch.qux.merge ^refs/heads/foo$
+ git rebase -f baz qux
First, rewinding head to replay your work on top of it...
Applying: file baz
Using index info to reconstruct a base tree...
Falling back to patching base and 3-way merge...
No changes -- Patch already applied.
Applying: file bar
Using index info to reconstruct a base tree...
Falling back to patching base and 3-way merge...
No changes -- Patch already applied.
+ /home/dfanjul/lib/go/bin/git-greb -d
greb: git checkout baz
Switched to branch 'baz'
Your branch is ahead of 'origin/master' by 1 commit.
greb: git pull
Current branch baz is up to date.
greb: git checkout qux
Switched to branch 'qux'
Your branch is based on 'baz', but the upstream is gone.
greb: git pull
From .
 * branch            baz        -> FETCH_HEAD
Current branch qux is up to date.
greb: git checkout --detach
HEAD is now at 3f3d20f... file baz
greb: git branch -D qux
Deleted branch qux (was 3f3d20f).

</pre>
