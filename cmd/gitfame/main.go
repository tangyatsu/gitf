//go:build !solution

package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"github.com/tangyatsu/gitfame/internal/format"
	"github.com/tangyatsu/gitfame/internal/gitreq"
	"github.com/tangyatsu/gitfame/internal/parse"
	"os"
	"sync"
)

func Update(authors *map[string]*parse.Stats, commits *map[string]map[string]struct{}, currAuth map[string][]string, lines map[string]int) {
	for name, curCommits := range currAuth {
		if _, ok := (*authors)[name]; !ok {
			(*authors)[name] = &parse.Stats{Name: name}
		}
		for _, comm := range curCommits {
			if _, ok := (*commits)[name]; !ok {
				(*commits)[name] = make(map[string]struct{})
			}
			if _, ok := (*commits)[name][comm]; !ok {
				(*commits)[name][comm] = struct{}{}
				(*authors)[name].Commits += 1
			}
		}

		(*authors)[name].Files += 1
		(*authors)[name].Lines += lines[name]
	}
}

var (
	flagRepo      = flag.String("repository", "./", "Path to Git repository")
	flagRev       = flag.String("revision", "HEAD", "Pointer to commit")
	flagOrder     = flag.String("order-by", "lines", "Order by. Default is in descending order by (lines, commits, files)")
	flagCommitter = flag.Bool("use-committer", false, "Use committer instead of author")
	flagformat    = flag.String("format", "tabular", "output format: tabular, csv, json, json-lines")
	flagExts      = flag.StringSlice("extensions", []string{}, "Set of file extensions to process, separated by commas. For example: '.go,.md'")
	flagLangs     = flag.StringSlice("languages", []string{}, "Set of languages to process, separated by commas. For example: 'go,markdown'")
	flagExclude   = flag.StringSlice("exclude", []string{}, "Set of Glob patterns for file exclusion, separated by commas. For example: 'foo/*,bar/*'")
	flagRestrict  = flag.StringSlice("restrict-to", []string{}, "Set of Glob patterns for file searching, separated by commas. For example: 'foo/*,bar/*'")
)

func main() {
	flag.Parse()

	if *flagOrder != "lines" && *flagOrder != "commits" && *flagOrder != "files" {
		fmt.Fprintf(os.Stderr, "unsupported flagOrder")
		os.Exit(1)
	}

	if *flagformat != "tabular" && *flagformat != "csv" && *flagformat != "json" && *flagformat != "json-lines" {
		fmt.Fprintf(os.Stderr, "unsupported format")
		os.Exit(1)
	}

	commits := make(map[string]map[string]struct{}) // map[name]map[commit]struct{} : set for commits
	authors := make(map[string]*parse.Stats)        // map[name]Stats

	files := parse.FilterFiles(*flagRepo, *flagRev, *flagExts, *flagLangs, *flagExclude, *flagRestrict)

	ch := make(chan string, 1)
	res := make(chan struct{}, len(files))
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		go func() {
			for file := range ch {
				out := gitreq.Blame(*flagRepo, *flagRev, file)
				currAuth, lines := gitreq.ProcessBlame(*flagRepo, *flagRev, file, out, *flagCommitter)
				mu.Lock()
				Update(&authors, &commits, currAuth, lines)
				mu.Unlock()
				res <- struct{}{}
			}
		}()
	}

	for _, file := range files {
		ch <- file
	}
	close(ch)
	for i := 0; i < len(files); i++ {
		<-res
	}

	sorted := parse.Sort(authors, *flagOrder)
	format.Print(sorted, *flagformat)
}
