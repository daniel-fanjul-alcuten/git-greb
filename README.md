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
