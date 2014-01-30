Description
===========

It gets a list of all local git branches, gets their tracking branches -local or remote-, builds a graph of dependencies and run 'git pull' for them in order.

It is a different approach to [TopGit](http://repo.or.cz/w/topgit.git?a=blob;f=README) commands: update and export.

Installation
============

Use 'go get':
<pre>
go get github.com/daniel-fanjul-alcuten/git-greb
</pre>

Example
=======

These git commands:

    cd /tmp
    dir=git-greb-example.git
    rm -rf $dir
    git init $dir
    cd $dir

    touch master
    git add master
    git commit -m "file master"

    git checkout -b foo -t master
    touch foo
    git add foo
    git commit -m "file foo"
    git push

    git checkout -b bar -t master
    touch bar
    git add bar
    git commit -m "file bar"

    git checkout -b baz -t master
    touch baz
    git add baz
    git commit -m "file baz"

    git branch -u foo bar
    git branch -u foo baz
    git checkout -b qux -t baz

    git greb -d

would make greb run these commands:

    git checkout foo
    git pull
    git checkout bar
    git pull
    git checkout baz
    git pull
    git checkout qux
    git pull
    git checkout qux@{u}
    git branch -d qux
    git checkout foo@{u}
    git branch -d foo
    git branch -u master bar
    git branch -u master baz

with this output:

    greb: git checkout foo
    Switched to branch 'foo'
    Your branch is up-to-date with 'master'.
    greb: git pull
    From .
     * branch            master     -> FETCH_HEAD
    Current branch foo is up to date.
    greb: git checkout bar
    Switched to branch 'bar'
    Your branch is ahead of 'foo' by 1 commit.
    greb: git pull
    From .
     * branch            foo        -> FETCH_HEAD
    Current branch bar is up to date.
    greb: git checkout baz
    Switched to branch 'baz'
    Your branch is ahead of 'foo' by 1 commit.
    greb: git pull
    From .
     * branch            foo        -> FETCH_HEAD
    Current branch baz is up to date.
    greb: git checkout qux
    Switched to branch 'qux'
    Your branch is up-to-date with 'baz'.
    greb: git pull
    From .
     * branch            baz        -> FETCH_HEAD
    Current branch qux is up to date.
    greb: git checkout qux@{u}
    Switched to branch 'baz'
    Your branch is ahead of 'foo' by 1 commit.
    greb: git branch -d qux
    Deleted branch qux (was 7de7306).
    greb: git checkout foo@{u}
    Switched to branch 'master'
    greb: git branch -d foo
    Deleted branch foo (was 5bebe43).
    greb: git branch -u master bar
    Branch bar set up to track local branch master.
    greb: git branch -u master baz
    Branch baz set up to track local branch master.
