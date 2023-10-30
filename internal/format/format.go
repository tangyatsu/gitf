package format

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"gitlab.com/slon/shad-go/gitfame/internal/parse"
	"os"
	"strconv"
	"text/tabwriter"
)

func Print(authors parse.AuthorSorter, format string) {
	switch format {
	case "tabular":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")

		for _, author := range authors.Authors {
			fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", author.Name, author.Lines, author.Commits, author.Files)
		}
		w.Flush()

	case "csv":
		w := csv.NewWriter(os.Stdout)
		w.Write([]string{"Name", "Lines", "Commits", "Files"})
		for _, author := range authors.Authors {
			w.Write([]string{author.Name, strconv.Itoa(author.Lines), strconv.Itoa(author.Commits), strconv.Itoa(author.Files)})
		}
		w.Flush()

	case "json":
		data, _ := json.Marshal(authors.Authors)
		fmt.Println(string(data))

	case "json-lines":
		for _, author := range authors.Authors {
			data, _ := json.Marshal(author)
			fmt.Println(string(data))
		}
	}
}
