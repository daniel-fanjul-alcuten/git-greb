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

    Usage of git-greb [<options>] [<branches>]:
    
    git-greb builds a graph of dependencies of the local branches. They usually depend
    on remote branches but they also can track other local branches. The graph is
    defined using the usual git config options branch.<name>.remote and
    branch.<name>.merge.
    
    git-greb supports branches that track multiple branches at the same time. In git
    there is only one value of the remote option for one or more merge options;
    git-greb takes this into account.
    
    The arguments are processed with git rev-parse to discover the abbreviated name
    of the branches. If there are no arguments, all local branches are retrieved
    instead. Then git-greb queries recursively the remote and merge tracking options
    to discover the dependencies and build the graph.
    
    The first set of options makes git-greb dump the graph in the standard output in
    different formats and then exit:
    
        -t=false: it uses a custom text format (text graph).
      -dot=false: it uses the dot format (dot graph).
        -x=false: it draws the dot format in an xlib window (xlib graph).
    
    The second set of options makes git-greb traverse the graph visiting the branches
    in order from the downstreams to the upstreams and running some variant of 'git
    pull' on them. If no option is provided 'git pull' merges or rebases depending
    on the usual git options.
    
        -r=false: it pulls every branch with --rebase (rebase).
        -m=false: it pulls every branch with --no-rebase (merge).
        -i=false: it rebases with --interactive (interactive).
        -c=false: it checks out instead of pulling (checkout).
        -s=false: it does not pull at all (skip).
    
    The option -d makes git-greb delete branches that don't create new history
    over their tracking branches. None is deleted until all of them have been
    successfuly updated in the previous step. A branch is deleted if it points to
    the same commit as at least one of its upstream branches. They are deleted in
    the opposite order.
    
    The tracking configuration is updated accordingly: if a set of branches A
    depends on a branch B, B depends on a set of branches C, and B is eligible for
    deletion: all branches A stop tracking B and start tracking the branches C.
    
        -d=false: it deletes fully merged branches after pulling (delete).
    
    The option -l makes git-greb pull only those branches that don't have any
    upstream branch in a different repository from the local one.
    
        -l=false: it only pulls local tracking branches (local).
    
    git-greb checks out every branch before pulling it and stops when a command
    doesn't finish with exit status 0. If this happens, the current branch may not
    be the same than the original one. If all of them finish successfully git-greb
    tries to return to the original branch, but it may have been deleted. As a
    consequence, the user should expect that the current branch may be different or
    even that the head may become detached. git-greb tries to minimize the number of
    check outs.
    
    Other options:
    
        -q=false: it does not print the command lines (quiet).
        -v=false: it explains intermediate steps (verbose).
        -n=false: it does not run any command (noop).
    
    git-greb checks the following options in the usual git configuration files:
    
      greb.local:         If the option -l is false, this bool option is used
                          instead.
      color.greb:         It enables or disables color in git-greb. See color.ui for
                          more information.
      color.greb.command: The color of the git commands that the user needs to know
                          that have been run. Blue by default.
      color.greb.current: The color of the current branch. Green by default.

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
+ git config greb.local false
+ touch master
+ git add master
+ git commit -m file master
[master (root-commit) a8edcf4] file master
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
[foo 25b82aa] file foo
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
[bar 48d1af5] file bar
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
[baz 9fd9857] file baz
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
Trying simple merge with 6b8c18672704a102cf590eeaf2dd206df6cd4ae2
Trying simple merge with c5981cea2f3a68fe470a6e99ea03fb8158a677af
Merge made by the 'octopus' strategy.
greb: git branch -D master
Deleted branch master (was a8edcf4).
greb: git config --unset branch.foo.merge ^refs/heads/master$
greb: git config branch.foo.remote origin
greb: git config --add branch.foo.merge refs/heads/master
+ git push origin bar:master
To /tmp/git-greb1.git
   a8edcf4..6b8c186  bar -> master
+ /home/dfanjul/lib/go/bin/git-greb -d
greb: git checkout foo
Switched to branch 'foo'
Your branch is behind 'origin/master' by 1 commit, and can be fast-forwarded.
greb: git pull
First, rewinding head to replay your work on top of it...
Fast-forwarded foo to 6b8c18672704a102cf590eeaf2dd206df6cd4ae2.
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
Deleted branch bar (was 6b8c186).
greb: git config --unset branch.qux.merge ^refs/heads/bar$
greb: git config --add branch.qux.merge refs/heads/foo
greb: git branch -D foo
Deleted branch foo (was 6b8c186).
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
HEAD is now at ae95dcf... file baz
greb: git branch -D qux
Deleted branch qux (was ae95dcf).

</pre>
