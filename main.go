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

	rebase bool
	merge  bool
)

func init() {

	log.SetFlags(0)

	flag.BoolVar(&quiet, "q", false,
		"quiet: it does not print the command lines.")
	flag.BoolVar(&verbose, "v", false,
		"verbose: it explains intermediate steps.")
	flag.BoolVar(&noop, "n", false,
		"noop: it does not run the commands.")

	flag.BoolVar(&rebase, "r", false,
		"it pulls with --rebase.")
	flag.BoolVar(&merge, "m", false,
		"it pulls with --no-rebase.")
	// TODO(dfanjul): interactive := flag.Bool("i", false, "interactive")
}

func main() {

	flag.Parse()
	if rebase && merge {
		rebase, merge = false, false
	}
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

	var graph map[string]string
	if len(branches) == 0 {
		if graph, err = getGraphForAllBranches(); err != nil {
			return
		}
	} else {
		graph = getGraphForUserBranches(branches)
	}

	for {
		l := len(graph)
		for branch, tracking := range graph {
			if _, ok := graph[tracking]; ok {
				// the branch depends on another local branch
				continue
			}
			if err = processBranch(branch); err != nil {
				return
			}
			delete(graph, branch)
		}
		if len(graph) == l {
			break
		}
	}
	return
}

func getGraphForAllBranches() (graph map[string]string, err error) {

	var branches []string
	if branches, err = getBranches(); err != nil {
		return
	}

	graph = make(map[string]string, len(branches))
	for _, branch := range branches {
		tracking, _ := getTrackingBranch(branch)
		if tracking != "" {
			graph[branch] = tracking
		}
	}
	return
}

func getGraphForUserBranches(branches []string) (graph map[string]string) {

	graph = make(map[string]string, len(branches))
	pending := append([]string(nil), branches...)
	processed := make(map[string]struct{}, len(branches))
	for len(pending) > 0 {
		var branch string
		branch, pending = pending[0], pending[1:]
		if _, ok := processed[branch]; !ok {
			processed[branch] = struct{}{}
			tracking, _ := getTrackingBranch(branch)
			if tracking != "" {
				graph[branch] = tracking
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

func processBranch(branch string) (err error) {

	cmd := NewCommand(!quiet, true, "git", "checkout", branch)
	if !noop {
		cmd.Stdout = os.Stdout
		if err = cmd.Run(); err != nil {
			err = CmdError(cmd, err)
			return
		}
	}

	args := []string{"pull"}
	if rebase {
		args = append(args, "--rebase")
	} else if merge {
		args = append(args, "--no-rebase")
	}

	cmd = NewCommand(!quiet, true, "git", args...)
	if !noop {
		cmd.Stdout = os.Stdout
		if err = cmd.Run(); err != nil {
			err = CmdError(cmd, err)
			return
		}
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
	return fmt.Errorf("%s: %s\n", CmdArgs(cmd), err)
}

func CmdErrOutput(cmd *exec.Cmd, err error, output []byte) error {
	return fmt.Errorf("%s: %s\n%s\n", CmdArgs(cmd), err, string(output))
}
