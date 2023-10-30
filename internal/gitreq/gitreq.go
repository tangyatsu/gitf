package gitreq

import (
	"os/exec"
	"strconv"
	"strings"
)

func GetFiles(repo string, rev string) []string {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", rev)
	cmd.Dir = repo
	out, _ := cmd.Output()

	res := strings.FieldsFunc(string(out), func(r rune) bool {
		return r == '\n'
	})
	return res

}

func Blame(repo string, rev string, fileName string) []string {
	cmd := exec.Command("git", "blame", "--porcelain", rev, fileName)
	cmd.Dir = repo
	out, _ := cmd.Output()

	res := strings.FieldsFunc(string(out), func(r rune) bool {
		return r == '\n'
	})

	return res
}

func Log(repo string, rev string, fileName string) (string, string) {
	cmd := exec.Command("git", "log", "-1", "--pretty=format:%H %an", rev, "--", fileName)
	cmd.Dir = repo
	out, _ := cmd.Output()

	before, after, _ := strings.Cut(string(out), " ")
	return before, after
}

func ProcessBlame(repo string, rev string, fileName string, data []string, useCommitter bool) (map[string][]string, map[string]int) {
	authors := make(map[string][]string) // map[name][]commits
	commits := make(map[string]int)      // map[commit]lines number
	lines := make(map[string]int)        // map[name]lines number

	if len(data) == 0 {
		h, a := Log(repo, rev, fileName)
		authors[a] = append(authors[a], h)
		lines[a] = 0
		return authors, lines
	}

	isHash := true
	var curHash string
	l := 0
	for i := 0; i < len(data); i++ {
		s := strings.Split(data[i], " ")
		if isHash {
			l, _ = strconv.Atoi(s[len(s)-1])
			commits[s[0]] += l
			curHash = s[0]
			isHash = false
		} else if s[0] == "author" {
			if !useCommitter {
				name := data[i][len("author "):]
				authors[name] = append(authors[name], curHash)
			} else {
				name := data[i+4][len("committer "):]
				authors[name] = append(authors[name], curHash)
			}
		} else if s[0][0] == '\t' {
			l--
			if l == 0 {
				isHash = true
			}
		}
	}

	for name, AuthCommits := range authors {
		for _, commit := range AuthCommits {
			lines[name] += commits[commit]
		}
	}

	return authors, lines
}
