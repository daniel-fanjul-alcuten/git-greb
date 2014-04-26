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
	"strings"
)

var (
	quiet       bool
	verbose     bool
	noop        bool
	graphtxt    bool
	graphdot    bool
	graphxlib   bool
	skip        bool
	checkout    bool
	rebase      bool
	merge       bool
	interactive bool
	remove      bool
	local       bool
)

func init() {
	log.SetFlags(0)
	flag.BoolVar(&quiet, "q", false,
		"it does not print the command lines (quiet).")
	flag.BoolVar(&verbose, "v", false,
		"it explains intermediate steps (verbose).")
	flag.BoolVar(&noop, "n", false,
		"it does not run the commands (noop).")
	flag.BoolVar(&graphtxt, "t", false,
		"it shows the graph of dependencies in text and exits (text graph).")
	flag.BoolVar(&graphdot, "dot", false,
		"it shows the graph of dependencies in dot format and exits (dot graph).")
	flag.BoolVar(&graphxlib, "x", false,
		"it shows the graph of dependencies in xlib and exits (xlib graph).")
	flag.BoolVar(&skip, "s", false,
		"it does not pull at all (skip).")
	flag.BoolVar(&checkout, "c", false,
		"it checks out instead of pulling (checkout).")
	flag.BoolVar(&rebase, "r", false,
		"it pulls with --rebase (rebase).")
	flag.BoolVar(&merge, "m", false,
		"it pulls with --no-rebase (merge).")
	flag.BoolVar(&interactive, "i", false,
		"it rebases with --interactive (interactive).")
	flag.BoolVar(&remove, "d", false,
		"it deletes fully merged branches after pulling (delete).")
	flag.BoolVar(&local, "l", false,
		"it only pulls local tracking branches (local).")
}

func assertFlags() (err error) {
	flags := []struct {
		name  string
		value bool
	}{
		{"-t (text graph)", graphtxt},
		{"-dot (dot graph)", graphdot},
		{"-x (xlib graph)", graphxlib},
		{"-s (skip)", skip},
		{"-c (checkout)", checkout},
		{"-r (rebase)", rebase},
		{"-m (merge)", merge},
		{"-i (interactive)", interactive},
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
	blueColor  string
	resetColor string
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
	var blue string
	cmd = newCommand(verbose, false, "git", "config", "--get-color",
		"color.greb.branch", "blue")
	var output []byte
	var err error
	if output, err = cmd.CombinedOutput(); err != nil {
		return
	}
	blue = string(output)
	cmd = newCommand(verbose, false, "git", "config", "--get-color", "", "reset")
	var reset string
	if output, err = cmd.CombinedOutput(); err != nil {
		return
	}
	reset = string(output)
	blueColor, resetColor = blue, reset
	return
}

func getColors(color bool) (blue, reset string) {
	if color {
		blue, reset = blueColor, resetColor
	}
	return
}

func main() {
	flag.Parse()
	if err := assertFlags(); err != nil {
		logFatal(err)
	}
	initColors()
	if err := greb(flag.Args()); err != nil {
		logFatal(err)
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
	if graphtxt {
		fmt.Print(g.text(nil, "", "  "))
		return
	} else if graphdot {
		fmt.Print(g.dot())
		return
	} else if graphxlib {
		cmd := newCommand(!quiet, true, "dot", "-Txlib")
		if !noop {
			cmd.Stdin = bytes.NewBufferString(g.dot())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				err = cmdError(cmd, err)
				return
			}
		}
		return
	}
	branch, _ := getCurrentBranch()
	current := branch
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
	if branch != "" {
		checkoutBranchIfNeeded(branch, &current)
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
					if branch, err = getAbbrevSymbolicFullName(r); err != nil {
						return
					}
					// TODO tracking ref is assumed to be a branch
					u, _ := g.node(ref{"refs/heads/" + branch, remote})
					g.edge(n, u, r)
					pending = append(pending, branch)
				}
			} else if remote != "" {
				for _, r := range refnames {
					u, ok := g.node(ref{r, remote})
					if !ok {
						if u.branch, err = findRemoteTrackingBranch(u.ref); err != nil {
							return
						}
					}
					g.edge(n, u, r)
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

func getCurrentBranch() (branch string, err error) {
	cmd := newCommand(verbose, false, "git", "symbolic-ref", "-q", "--short",
		"HEAD")
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = cmdError(cmd, err)
		if verbose {
			logPrintf("-> no branch\n")
		}
		return
	}
	branch = strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", branch)
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
		blue, reset := getColors(color)
		logPrintf("%s%s%s\n", blue, cmdArgs(cmd), reset)
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
