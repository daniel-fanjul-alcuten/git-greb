package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	quiet   bool
	verbose bool
	noop    bool

	graphtxt bool
	graphdot bool

	skip        bool
	checkout    bool
	rebase      bool
	merge       bool
	interactive bool

	remove bool
	local  bool
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
		"it shows a forest of dependencies in text and exits (text graph).")
	flag.BoolVar(&graphdot, "dot", false,
		"it shows a forest of dependencies in dot format and exits (dot graph).")

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

func main() {

	flag.Parse()
	branches := flag.Args()

	getGitColors()

	if err := greb(branches); err != nil {
		logFatal(err)
	}
}

var (
	blueColor  string
	resetColor string
)

func getGitColors() {

	cmd := NewCommand(verbose, false, "git", "config", "--get-colorbool", "color.greb")
	cmd.Stdout = os.Stdout
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
	cmd = NewCommand(verbose, false, "git", "config", "--get-color", "color.greb.branch", "blue")
	if output, err := cmd.Output(); err != nil {
		return
	} else {
		blue = string(output)
	}

	cmd = NewCommand(verbose, false, "git", "config", "--get-color", "", "reset")
	var reset string
	if output, err := cmd.Output(); err != nil {
		return
	} else {
		reset = string(output)
	}

	blueColor, resetColor = blue, reset
	return
}

func getColors(color bool) (blue, reset string) {
	if color {
		blue = blueColor
		reset = resetColor
	}
	return
}

func greb(branches []string) (err error) {

	if err := assertFlags(); err != nil {
		return err
	}

	g := newGraph()
	if len(branches) == 0 {
		if err = fillGraphForAllBranches(g); err != nil {
			return
		}
	} else {
		fillGraphForUserBranches(g, branches)
	}

	if graphtxt {
		fmt.Print(g.toText("", "  ", 0))
		return
	} else if graphdot {
		fmt.Print(g.toDot())
		return
	}

	branch, _ := getBranch()

	sort := g.sort()
	if !skip {
		for _, branch := range sort {
			if err = pullBranch(branch); err != nil {
				return
			}
		}
	}

	if remove {
		for i := len(sort) - 1; i >= 0; i-- {
			branch := sort[i]
			if err = deleteBranch(g, branch); err != nil {
				return
			}
		}
	}

	if branch != "" {
		cmd := NewCommand(!quiet, true, "git", "checkout", branch)
		if !noop {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
	}

	return
}

func getBranch() (branch string, err error) {

	cmd := NewCommand(verbose, false, "git", "symbolic-ref", "-q", "--short",
		"HEAD")
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = CmdErrOutput(cmd, err, output)
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

func fillGraphForAllBranches(g *graph) (err error) {

	var branches []string
	if branches, err = getBranches(); err != nil {
		return
	}

	for _, branch := range branches {
		tracking, _ := getTrackingBranch(branch)
		g.add(branch, tracking)
	}
	return
}

func fillGraphForUserBranches(g *graph, branches []string) {

	pending := append([]string(nil), branches...)
	processed := make(map[string]struct{}, len(branches))
	for len(pending) > 0 {
		var branch string
		branch, pending = pending[0], pending[1:]
		if _, ok := processed[branch]; !ok {
			processed[branch] = struct{}{}
			tracking, _ := getTrackingBranch(branch)
			g.add(branch, tracking)
			if tracking != "" {
				pending = append(pending, tracking)
			}
		}
	}
	return
}

func getBranches() (branches []string, err error) {

	cmd := NewCommand(verbose, false, "git", "for-each-ref", "refs/heads/",
		"--format", "%(refname:short)")
	var output io.ReadCloser
	if output, err = cmd.StdoutPipe(); err != nil {
		err = CmdError(cmd, err)
		return
	}
	if err = cmd.Start(); err != nil {
		err = CmdError(cmd, err)
		return
	}

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		text := scanner.Text()
		branches = append(branches, text)
	}
	if verbose {
		logPrintf("-> %s\n", strings.Join(branches, ", "))
	}

	if err = cmd.Wait(); err != nil {
		err = CmdError(cmd, err)
		return
	}
	return
}

func getTrackingBranch(branch string) (tracking string, err error) {

	cmd := NewCommand(verbose, false, "git", "rev-parse", "-q", "--verify",
		"--symbolic-full-name", "--abbrev-ref", branch+"@{u}")
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = CmdErrOutput(cmd, err, output)
		if verbose {
			logPrintf("-> no branch\n")
		}
		return
	}

	tracking = strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", tracking)
	}
	return
}

func pullBranch(branch string) (err error) {

	if local {
		cmd := NewCommand(verbose, false, "git", "config", "branch."+branch+".remote")
		var output []byte
		if output, err = cmd.CombinedOutput(); err != nil {
			err = CmdErrOutput(cmd, err, output)
			if verbose {
				logPrintf("-> no config\n")
			}
			return
		}

		remote := strings.TrimSpace(string(output))
		if remote == "." {
			if verbose {
				logPrintf("-> %s\n", remote)
			}
		} else {
			if verbose {
				logPrintf("-> %s (skipped)\n", remote)
			}
			return
		}
	}

	cmd := NewCommand(!quiet, true, "git", "checkout", branch)
	if !noop {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			err = CmdError(cmd, err)
			return
		}
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
	}

	cmd = NewCommand(!quiet, true, "git", args...)
	if !noop {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			err = CmdError(cmd, err)
			return
		}
	}

	return
}

func deleteBranch(g *graph, branch string) (err error) {

	cmd := NewCommand(verbose, false, "git", "rev-parse", "-q", "--verify",
		branch)
	var output []byte
	if output, err = cmd.CombinedOutput(); err != nil {
		err = CmdErrOutput(cmd, err, output)
		if verbose {
			logPrintf("-> %s\n", err)
		}
		return
	}

	hash1 := strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", hash1)
	}

	cmd = NewCommand(verbose, false, "git", "rev-parse", "-q", "--verify",
		branch+"@{u}")
	if output, err = cmd.CombinedOutput(); err != nil {
		err = CmdErrOutput(cmd, err, output)
		if verbose {
			logPrintf("-> %s\n", err)
		}
		return
	}

	hash2 := strings.TrimSpace(string(output))
	if verbose {
		logPrintf("-> %s\n", hash2)
	}

	if hash1 == hash2 {

		cmd = NewCommand(!quiet, true, "git", "checkout", branch+"@{u}")
		if !noop {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				err = CmdError(cmd, err)
				return
			}
		}

		cmd = NewCommand(!quiet, true, "git", "branch", "-d", branch)
		if !noop {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				err = CmdError(cmd, err)
				return
			}
		}

		tracking, branches := g.remove(branch)
		if tracking == "" {
			return
		}

		for _, branch := range branches {

			cmd = NewCommand(!quiet, true, "git", "branch", "-u", tracking, branch)
			if !noop {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err = cmd.Run(); err != nil {
					err = CmdError(cmd, err)
					return
				}
			}
		}
	}

	return
}

func assertFlags() (err error) {
	flags := []struct {
		name  string
		value bool
	}{
		{"-t (text graph)", graphtxt},
		{"-dot (dot graph)", graphdot},
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
	}
	return
}

func NewCommand(verbose, color bool, name string, arg ...string) (cmd *exec.Cmd) {
	cmd = exec.Command(name, arg...)
	if verbose {
		blue, reset := getColors(color)
		logPrintf("%s%s%s\n", blue, CmdArgs(cmd), reset)
	}
	return
}

func CmdArgs(cmd *exec.Cmd) string {
	args := append([]string(nil), cmd.Args...)
	for i, arg := range args {
		if arg == "" {
			args[i] = "''"
		}
	}
	return strings.Join(args, " ")
}

func logPrintf(format string, v ...interface{}) {
	log.Printf("greb: "+format, v...)
}

func logFatal(err error) {
	log.Fatalf("greb: %s\n", err)
}

func CmdError(cmd *exec.Cmd, err error) error {
	return fmt.Errorf("%s: %s", CmdArgs(cmd), err)
}

func CmdErrOutput(cmd *exec.Cmd, err error, output []byte) error {
	o := string(output)
	if o != "" {
		o = "\n" + o
	}
	return fmt.Errorf("%s: %s%s", CmdArgs(cmd), err, o)
}
