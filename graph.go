package main

type graph struct {
	direct map[string]string
}

func newGraph() *graph {
	direct := make(map[string]string)
	return &graph{direct}
}

func (g *graph) add(branch, tracking string) {
	g.direct[branch] = tracking
}

func (g *graph) sort() (branches []string) {

	direct := make(map[string]string, len(g.direct))
	for k, v := range g.direct {
		direct[k] = v
	}

	for {
		l := len(direct)
		for branch, tracking := range direct {
			if _, ok := direct[tracking]; ok {
				// the branch depends on another local branch
				continue
			}
			branches = append(branches, branch)
			delete(direct, branch)
		}
		if len(direct) == l {
			break
		}
	}

	return
}
