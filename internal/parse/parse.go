package parse

import (
	"encoding/json"
	"github.com/tangyatsu/gitfame/configs"
	"github.com/tangyatsu/gitfame/internal/gitreq"
	"path/filepath"
	"sort"
	"strings"
)

func FilterExtensions(files []string, exts []string) []string {
	if len(exts) == 0 {
		return files
	}
	res := make([]string, 0)

	for _, file := range files {
		for _, ext := range exts {
			if v := filepath.Ext(file); v == ext {
				res = append(res, file)
			}
		}
	}
	return res
}

type Language struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Extensions []string `json:"extensions"`
}

func LoadLanguages(exts *[]string, langs []string) {
	if len(langs) == 0 {
		return
	}

	var l []Language
	err := json.Unmarshal(configs.MappingFile, &l)
	if err != nil {
		panic(err)
	}

	set := make(map[string]int, len(l))
	for i, name := range l {
		set[strings.ToLower(name.Name)] = i
	}

	for _, i := range langs {
		if itr, ok := set[strings.ToLower(i)]; ok {
			*exts = append(*exts, l[itr].Extensions...)
		}
	}
}

func FilterExclude(files []string, exclude []string) []string {
	if len(exclude) == 0 {
		return files
	}
	res := make([]string, 0)

Fileloop:
	for _, file := range files {
		for _, excl := range exclude {
			if match, _ := filepath.Match(excl, file); match {
				continue Fileloop
			}
		}
		res = append(res, file)
	}
	return res
}

func FilterRestrict(files []string, restrict []string) []string {
	if len(restrict) == 0 {
		return files
	}
	res := make([]string, 0)

	for _, file := range files {
		for _, excl := range restrict {
			if match, _ := filepath.Match(excl, file); match {
				res = append(res, file)
			}
		}
	}
	return res
}

func FilterFiles(repo string, rev string, exts []string, langs []string, exclude []string, restrict []string) []string {
	files := gitreq.GetFiles(repo, rev)
	LoadLanguages(&exts, langs)
	files = FilterExtensions(files, exts)
	files = FilterExclude(files, exclude)
	files = FilterRestrict(files, restrict)

	return files
}

type Stats struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

type AuthorSorter struct {
	Authors []Stats
	OrderBy string
}

func (a AuthorSorter) Len() int {
	return len(a.Authors)
}

func (a AuthorSorter) Swap(i, j int) {
	a.Authors[i], a.Authors[j] = a.Authors[j], a.Authors[i]
}

func (a AuthorSorter) Less(i, j int) bool {
	var val1 []int
	var val2 []int

	switch a.OrderBy {
	case "lines":
		val1 = []int{a.Authors[i].Lines, a.Authors[i].Commits, a.Authors[i].Files}
		val2 = []int{a.Authors[j].Lines, a.Authors[j].Commits, a.Authors[j].Files}

	case "commits":
		val1 = []int{a.Authors[i].Commits, a.Authors[i].Lines, a.Authors[i].Files}
		val2 = []int{a.Authors[j].Commits, a.Authors[j].Lines, a.Authors[j].Files}

	case "files":
		val1 = []int{a.Authors[i].Files, a.Authors[i].Lines, a.Authors[i].Commits}
		val2 = []int{a.Authors[j].Files, a.Authors[j].Lines, a.Authors[j].Commits}
	default:
		panic("DEFAULT OCCURED")
	}

	if val1[0] == val2[0] {
		if val1[1] == val2[1] {
			if val1[2] == val2[2] {
				return strings.ToLower(a.Authors[i].Name) < strings.ToLower(a.Authors[j].Name)
			}
			return val1[2] > val2[2]
		}
		return val1[1] > val2[1]
	}
	return val1[0] > val2[0]
}

func Sort(authors map[string]*Stats, orderBy string) AuthorSorter {
	var a AuthorSorter
	a.OrderBy = orderBy
	for _, stats := range authors {
		a.Authors = append(a.Authors, *stats)
	}

	sort.Sort(a)
	return a
}
