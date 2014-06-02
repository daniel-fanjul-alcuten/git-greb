package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
)

var (
	bash        string
	graphtxt    bool
	graphdot    bool
	graphxlib   bool
	change      string
	rebase      bool
	merge       bool
	interactive bool
	checkout    bool
	skip        bool
	remove      bool
	local       bool
	quiet       bool
	verbose     bool
	noop        bool
)

func init() {
	log.SetFlags(0)
	flag.StringVar(&bash, "bash", "",
		"name of the bash function to use with complete (bash completion)")
	flag.BoolVar(&graphtxt, "t", false,
		"it uses a custom text format (text graph).")
	flag.BoolVar(&graphdot, "dot", false,
		"it uses the dot format (dot graph).")
	flag.BoolVar(&graphxlib, "x", false,
		"it draws the dot format in an xlib window (xlib graph).")
	flag.StringVar(&change, "C", "HEAD",
		"it checks out the given branch before exit (change branch).")
	flag.BoolVar(&rebase, "r", false,
		"it pulls every branch with --rebase (rebase).")
	flag.BoolVar(&merge, "m", false,
		"it pulls every branch with --no-rebase (merge).")
	flag.BoolVar(&interactive, "i", false,
		"it rebases with --interactive (interactive).")
	flag.BoolVar(&checkout, "c", false,
		"it checks out instead of pulling (checkout).")
	flag.BoolVar(&skip, "s", false,
		"it does not pull at all (skip).")
	flag.BoolVar(&remove, "d", false,
		"it deletes fully merged branches after pulling (delete).")
	flag.BoolVar(&local, "l", false,
		"it only pulls local tracking branches (local).")
	flag.BoolVar(&quiet, "q", false,
		"it does not print the command lines (quiet).")
	flag.BoolVar(&verbose, "v", false,
		"it explains intermediate steps (verbose).")
	flag.BoolVar(&noop, "n", false,
		"it does not run any command (noop).")
}

func assertFlags() (err error) {
	flags := []struct {
		name  string
		value bool
	}{
		{"-bash (bash completion)", bash != ""},
		{"-t (text graph)", graphtxt},
		{"-dot (dot graph)", graphdot},
		{"-x (xlib graph)", graphxlib},
		{"-r (rebase)", rebase},
		{"-m (merge)", merge},
		{"-i (interactive)", interactive},
		{"-c (checkout)", checkout},
		{"-s (skip)", skip},
	}
	var found []string
	for _, f := range flags {
		if f.value {
			found = append(found, f.name)
		}
	}
	if len(found) > 1 {
		err = fmt.Errorf("incompatible flags: %s", strings.Join(found, ", "))
		return
	}
	return
}

var (
	commandColorName string
	commandColorCode string
	currentColorName string
	currentColorCode string
	remoteColorName  string
	remoteColorCode  string
	resetColorCode   string
)

func initColors() {
	cmd := newCommand(verbose, false, "git", "config", "--get-colorbool",
		"color.greb")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if verbose {
			logPrintf("-> false\n")
		}
		return
	}
	if verbose {
		logPrintf("-> true\n")
	}
	cmd = newCommand(verbose, false, "git", "config", "--get-color", "", "reset")
	if output, err := cmd.CombinedOutput(); err == nil {
		resetColorCode = string(output)
	}
	commandColorName, commandColorCode = initColor("command", "blue")
	currentColorName, currentColorCode = initColor("current", "green")
	remoteColorName, remoteColorCode = initColor("remote", "red")
	return
}

func initColor(slot, defaultName string) (name, code string) {
	cmd := newCommand(verbose, false, "git", "config",
		"color.greb."+slot)
	if output, err := cmd.CombinedOutput(); err != nil {
		name = defaultName
	} else {
		name = strings.TrimSpace(string(output))
	}
	if verbose {
		logPrintf("-> %v\n", name)
	}
	cmd = newCommand(verbose, false, "git", "config", "--get-color",
		"color.greb."+slot, name)
	if output, err := cmd.CombinedOutput(); err != nil {
		if verbose && resetColorCode != "" {
			logPrintf("-> %v\n", err)
		}
	} else {
		code = string(output)
		if verbose && resetColorCode != "" {
			logPrintf("-> %v%v%v\n", code, name, resetColorCode)
		}
	}
	return
}

func getCommandColor(color bool) (command, reset string) {
	if color {
		command, reset = commandColorCode, resetColorCode
	}
	return
}

const usage = `Usage of %[1]s [<options>] [<branches>]:

%[2]s builds a graph of dependencies of the local branches. They usually depend
on remote branches but they also can track other local branches. The graph is
defined using the usual git config options branch.<name>.remote and
branch.<name>.merge.

%[2]s supports branches that track multiple branches at the same time. In git
there is only one value of the remote option for one or more merge options;
%[2]s takes this into account.

The arguments are processed with git rev-parse to discover the abbreviated name
of the branches. If there are no arguments, all local branches are retrieved
instead. Then %[2]s queries recursively the remote and merge tracking options
to discover the dependencies and build the graph.

The first set of options makes %[2]s dump the graph in the standard output in
different formats and then exit:

%[3]s
%[4]s
%[5]s

The second set of options makes %[2]s traverse the graph visiting the branches
in order from the downstreams to the upstreams and running some variant of 'git
pull' on them. If no option is provided 'git pull' merges or rebases depending
on the usual git options.

%[6]s
%[7]s
%[8]s
%[9]s
%[10]s

The option %[11]s makes %[2]s delete branches that don't create new history
over their tracking branches. None is deleted until all of them have been
successfuly updated in the previous step. A branch is deleted if it points to
the same commit as at least one of its upstream branches. They are deleted in
the opposite order.

The tracking configuration is updated accordingly: if a set of branches A
depends on a branch B, B depends on a set of branches C, and B is eligible for
deletion: all branches A stop tracking B and start tracking the branches C.

%[12]s

The option %[13]s makes %[2]s pull only those branches that don't have any
upstream branch in a different repository from the local one.

%[14]s

%[2]s checks out every branch before pulling and stops when a command doesn't
finish with exit status 0. If all pulls finish successfully %[2]s tries to
return to the original branch. The option %[18]s may be used to return to a
different one. As the branch may have been deleted or any command may have
failed, the user should expect that the current branch may be different or even
that the HEAD may become detached. %[2]s tries to minimize the number of
checkouts.

%[19]s

Other options:

%[15]s
%[16]s
%[17]s
%[20]s

%[2]s checks the following options in the usual git configuration files:

  greb.local:         If the option -l is false, this bool option is used
                      instead.
  color.greb:         It enables or disables color in %[2]s. See color.ui for
                      more information.
  color.greb.command: The color of the git commands that the user needs to know
                      that have been run. Blue by default.
  color.greb.current: The color of the current branch. Green by default.
  color.greb.remote:  The color of the branches in other repositories. Red by
                      default.
`

func main() {
	flag.Usage = func() {
		f := func(name string) string {
			f := flag.Lookup(name)
			format := "  %5s=%s: %s"
			return fmt.Sprintf(format, "-"+name, f.DefValue, f.Usage)
		}
		fmt.Fprintf(os.Stderr, usage, os.Args[0], filepath.Base(os.Args[0]),
			f("t"), f("dot"), f("x"),
			f("r"), f("m"), f("i"), f("c"), f("s"),
			"-d", f("d"),
			"-l", f("l"),
			f("q"), f("v"), f("n"),
			"-C", f("C"),
			f("bash"),
		)
	}
	flag.Parse()
	initColors()
	updateFlagsWithOptions()
	if err := assertFlags(); err != nil {
		logFatal(err)
	}
	if bash != "" {
		fmt.Println(bashCompletion(bash))
	} else if err := greb(flag.Args()); err != nil {
		logFatal(err)
	}
}

func bashCompletion(funcname string) string {
	return fmt.Sprintf(`%s() {
	local cur=${COMP_WORDS[COMP_CWORD]}
	if [ $COMP_CWORD -gt 1 ]; then
		local prev=${COMP_WORDS[$(($COMP_CWORD-1))]}
		case $prev in
			-C|--C)
				local branches=$(git for-each-ref refs/heads --format '%%(refname:short)' -s)
				COMPREPLY=( $(compgen -W "$branches" -- "$cur") )
				return
				;;
			-bash|--bash)
				COMPREPLY=()
				return
				;;
		esac
	fi
	case $cur in
		--*)
			local opts="--bash --t --dot --x --C --r --m --i --c --s --d --l --q --v --n"
			COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
			return
			;;
	esac
	local opts="-bash -t -dot -x -C -r -m -i -c -s -d -l -q -v -n"
	COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
}`, funcname)
}

func updateFlagsWithOptions() {
	if !local {
		cmd := newCommand(verbose, false, "git", "config", "--bool", "greb.local")
		if output, err := cmd.CombinedOutput(); err != nil {
			if verbose {
				logPrintf("-> no config\n")
			}
		} else {
			l := strings.TrimSpace(string(output))
			if verbose {
				logPrintf("-> %s\n", l)
			}
			local = l == "true"
		}
	}
}

func greb(branches []string) (err error) {
	var g *graph
	if len(branches) == 0 {
		if g, err = fillGraphForAllBranches(); err != nil {
			return
		}
	} else {
		branches = filterBranches(branches)
		if g, err = fillGraphForBranches(branches); err != nil {
			return
		}
	}
	current, _ := getAbbrevSymbolicFullName("HEAD")
	if graphtxt {
		fmt.Print(g.text(nil, "", "  ", current, currentColorCode, remoteColorCode,
			resetColorCode))
		return
	} else if graphdot {
		fmt.Print(g.dot(current, currentColorName, remoteColorName))
		return
	} else if graphxlib {
		cmd := newCommand(!quiet, true, "dot", "-Txlib")
		if !noop {
			cmd.Stdin = bytes.NewBufferString(g.dot(current, currentColorName,
				remoteColorName))
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				err = cmdError(cmd, err)
				return
			}
		}
		return
	}
	branch, _ := getAbbrevSymbolicFullName(change)
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	defer func() {
		signal.Stop(s)
		// There is a race condition: the signal may not be sent to the channel
		// before we reach this point. The channel cannot be used.
		if err != nil {
			// The string comparision is ugly, but the race condition is too much
			// uncertain.
			if !strings.HasSuffix(err.Error(), "signal: interrupt") {
				return
			}
		}
		if branch != "" {
			checkoutBranchIfNeeded(branch, &current)
		}
	}()
	sort := g.sort()
	if !skip {
		for _, n := range sort {
			if err = pullBranch(n, &current); err != nil {
				return
			}
		}
	}
	if remove {
		for i := len(sort) - 1; i >= 0; i-- {
			n := sort[i]
			if err = deleteBranchIfMerged(g, n, &branch, &current); err != nil {
				return
			}
		}
	}
	return
}

func filterBranches(branches []string) []string {
	fb := make([]string, 0, len(branches))
	for _, b := range branches {
		if n, err := getAbbrevSymbolicFullName(b); err == nil {
			fb = append(fb, n)
		}
	}
	return fb
}

func fillGraphForAllBranches() (g *graph, err error) {
	cmd := newCommand(verbose, false, "git", "for-each-ref", "refs/heads/",
		"--format", "%(refname:short)")
	var outpipe io.ReadCloser
	if outpipe, err = cmd.StdoutPipe(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	if err = cmd.Start(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	scanner := bufio.NewScanner(outpipe)
	var branches []string
	for scanner.Scan() {
		branches = append(branches, scanner.Text())
	}
	if verbose {
		logPrintf("-> %s\n", strings.Join(branches, ", "))
	}
	if err = cmd.Wait(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	return fillGraphForBranches(branches)
}

func fillGraphForBranches(branches []string) (g *graph, err error) {
	g = newGraph()
	pending := append([]string(nil), branches...)
	processed := make(map[string]struct{}, len(branches))
	for len(pending) > 0 {
		var branch string
		branch, pending = pending[0], pending[1:]
		if _, ok := processed[branch]; !ok {
			processed[branch] = struct{}{}
			n, _ := g.node(ref{"refs/heads/" + branch, "."})
			n.branch = branch
			var remote string
			var refnames []string
			if remote, refnames, err = getTrackingInfo(branch); err != nil {
				return
			}
			if remote == "." {
				for _, r := range refnames {
					if branch, err := getAbbrevSymbolicFullName(r); err == nil {
						// TODO tracking ref is assumed to be a branch
						u, _ := g.node(ref{"refs/heads/" + branch, remote})
						g.edge(n, u, r)
						pending = append(pending, branch)
					}
				}
			} else if remote != "" {
				for _, r := range refnames {
					u, ok := g.node(ref{r, remote})
					if ok {
						g.edge(n, u, r)
					} else if branch, err := findRemoteTrackingBranch(u.ref); err == nil {
						u.branch = branch
						g.edge(n, u, r)
					} else {
						delete(g.nodes, u.ref)
					}
				}
			}
		}
	}
	return
}

func getTrackingInfo(branch string) (remote string, refnames []string, err error) {
	cmd := newCommand(verbose, false, "git", "config", "branch."+branch+".remote")
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = nil
		if verbose {
			logPrintf("-> no config\n")
		}
		return
	}
	remote = strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", remote)
	}
	cmd = newCommand(verbose, false, "git", "config", "--get-all",
		"branch."+branch+".merge")
	var outpipe io.ReadCloser
	if outpipe, err = cmd.StdoutPipe(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	if err = cmd.Start(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	scanner := bufio.NewScanner(outpipe)
	for scanner.Scan() {
		refnames = append(refnames, scanner.Text())
	}
	if verbose {
		logPrintf("-> %s\n", strings.Join(refnames, ", "))
	}
	if err = cmd.Wait(); err != nil {
		err = nil
		return
	}
	return
}

func findRemoteTrackingBranch(r ref) (branch string, err error) {
	// there is no git command to retrieve it, remote.<remote>.fetch is parsed
	cmd := newCommand(verbose, false, "git", "config", "--get-all",
		"remote."+r.remote+".fetch")
	var outpipe io.ReadCloser
	if outpipe, err = cmd.StdoutPipe(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	if err = cmd.Start(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	scanner := bufio.NewScanner(outpipe)
	var fetchspecs []string
	for scanner.Scan() {
		fetchspecs = append(fetchspecs, scanner.Text())
	}
	if verbose {
		logPrintf("-> %s\n", strings.Join(fetchspecs, ", "))
	}
	if err = cmd.Wait(); err != nil {
		err = cmdError(cmd, err)
		return
	}
	for _, s := range fetchspecs {
		if strings.HasPrefix(s, "+") {
			s = s[1:]
		}
		p := strings.SplitN(s, ":", 2)
		f, l := p[0], p[1]
		if strings.HasSuffix(f, "*") && strings.HasSuffix(l, "*") {
			f, l = f[:len(f)-1], l[:len(l)-1]
		}
		if strings.HasPrefix(r.name, f) {
			b := l + r.name[len(f):]
			if b, err = getAbbrevSymbolicFullName(b); err != nil {
				return
			}
			branch = b
			return
		}
	}
	err = fmt.Errorf("remote %v does not fetch ref %v", r.remote, r.name)
	return
}

func getAbbrevSymbolicFullName(refname string) (fullname string, err error) {
	cmd := newCommand(verbose, false, "git", "rev-parse", "--symbolic-full-name",
		"--abbrev-ref", refname)
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = cmdError(cmd, err)
		if verbose {
			logPrintf("-> no name\n")
		}
		return
	}
	fullname = strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", fullname)
	}
	return
}

func pullBranch(n *node, current *string) (err error) {
	if local {
		for u := range n.upstreams {
			if u.remote != "." {
				return
			}
		}
	}
	if err = checkoutBranchIfNeeded(n.branch, current); err != nil {
		return
	}
	args := []string{"pull"}
	if checkout {
		return
	} else if rebase {
		args = append(args, "--rebase")
	} else if merge {
		args = append(args, "--no-rebase")
	} else if interactive {
		args = []string{"rebase", "--interactive"}
	} else if len(n.upstreams) > 1 {
		args = append(args, "--no-rebase")
	}
	cmd := newCommand(!quiet, true, "git", args...)
	if !noop {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			err = cmdError(cmd, err)
			return
		}
	}
	return
}

func deleteBranchIfMerged(g *graph, n *node, branch, current *string) (err error) {
	cmd := newCommand(verbose, false, "git", "rev-parse", "-q", "--verify",
		n.branch)
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = cmdError(cmd, err)
		if verbose {
			logPrintf("-> no hash\n")
		}
		return
	}
	hash := strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", hash)
	}
	for u := range n.upstreams {
		cmd := newCommand(verbose, false, "git", "rev-parse", "-q", "--verify",
			u.branch)
		var output []byte
		if output, err = cmd.CombinedOutput(); err != nil {
			err = cmdError(cmd, err)
			if verbose {
				logPrintf("-> no hash\n")
			}
			return
		}
		uhash := strings.TrimSpace(string(output))
		if verbose {
			logPrintf("-> %s\n", uhash)
		}
		if uhash == hash {
			return deleteBranch(g, n, branch, current)
		}
	}
	return
}

func deleteBranch(g *graph, n *node, branch, current *string) (err error) {
	if n.branch == *branch {
		*branch = ""
	}
	if n.branch == *current {
		if err = checkoutBranchIfNeeded(*branch, current); err != nil {
			return
		}
	}
	cmd := newCommand(!quiet, true, "git", "branch", "-D", n.branch)
	if !noop {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			err = cmdError(cmd, err)
			return
		}
	}
	for _, update := range g.remove(n) {
		switch u := update.(type) {
		case rmUpstream:
			cmd = newCommand(!quiet, true, "git", "config", "--unset",
				"branch."+u.downstream+".merge", "^"+u.upstream+"$")
			if !noop {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					err = cmdError(cmd, err)
					return
				}
			}
		case addUpstream:
			cmd = newCommand(!quiet, true, "git", "config", "--add",
				"branch."+u.downstream+".merge", u.upstream)
			if !noop {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					err = cmdError(cmd, err)
					return
				}
			}
		case setRemote:
			cmd = newCommand(!quiet, true, "git",
				"config", "branch."+u.downstream+".remote", u.remote)
			if !noop {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					err = cmdError(cmd, err)
					return
				}
			}
		}
	}
	return
}

func checkoutBranchIfNeeded(branch string, current *string) (err error) {
	if branch == *current {
		return
	}
	var arg string
	if branch == "" {
		arg = "--detach"
	} else {
		arg = branch
	}
	cmd := newCommand(!quiet, true, "git", "checkout", arg)
	if !noop {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			err = cmdError(cmd, err)
			return
		}
	}
	*current = branch
	return
}

func newCommand(verbose, color bool, name string, arg ...string) (cmd *exec.Cmd) {
	cmd = exec.Command(name, arg...)
	if verbose {
		command, reset := getCommandColor(color)
		logPrintf("%s%s%s\n", command, cmdArgs(cmd), reset)
	}
	return
}

func cmdArgs(cmd *exec.Cmd) string {
	args := append([]string(nil), cmd.Args...)
	for i, arg := range args {
		if arg == "" {
			args[i] = "''"
		}
	}
	return strings.Join(args, " ")
}

func cmdError(cmd *exec.Cmd, err error) error {
	return fmt.Errorf("%s: %s", cmdArgs(cmd), err)
}

func logPrintf(format string, v ...interface{}) {
	log.Printf("greb: "+format, v...)
}

func logFatal(err error) {
	log.Fatalf("greb: %s\n", err)
}
