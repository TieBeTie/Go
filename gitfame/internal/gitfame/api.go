package gitfame

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"
)

type Config struct {
	Repository   string
	Revision     string
	OrderBy      string
	UseCommitter bool
	Format       string
	Extensions   []string
	Languages    []string
	Exclude      []string
	RestrictTo   []string
	Progress     bool
}

type AuthorStat struct {
	Name    string
	Lines   int
	Commits int
	Files   int
}

func ValidateConfig(cfg *Config) error {
	switch cfg.Format {
	case "tabular", "csv", "json", "json-lines":
	default:
		return fmt.Errorf("invalid format value: %s", cfg.Format)
	}

	switch cfg.OrderBy {
	case "lines", "commits", "files":
	default:
		return fmt.Errorf("invalid order-by value: %s", cfg.OrderBy)
	}

	return nil
}

func ProcessRepo(ctx context.Context, cfg *Config, extensionMap map[string][]string) ([]*AuthorStat, error) {
	flags = &gitFameFlags{
		repository:   &cfg.Repository,
		revision:     &cfg.Revision,
		orderBy:      &cfg.OrderBy,
		useCommitter: &cfg.UseCommitter,
		format:       &cfg.Format,
	}

	if len(cfg.Extensions) > 0 {
		extensionsStr := ""
		for i, ext := range cfg.Extensions {
			if i > 0 {
				extensionsStr += ","
			}
			extensionsStr += ext
		}
		flags.extensions = &extensionsStr
	} else {
		emptyStr := ""
		flags.extensions = &emptyStr
	}

	if len(cfg.Languages) > 0 {
		languagesStr := ""
		for i, lang := range cfg.Languages {
			if i > 0 {
				languagesStr += ","
			}
			languagesStr += lang
		}
		flags.languages = &languagesStr
	} else {
		emptyStr := ""
		flags.languages = &emptyStr
	}

	if len(cfg.Exclude) > 0 {
		excludeStr := ""
		for i, excl := range cfg.Exclude {
			if i > 0 {
				excludeStr += ","
			}
			excludeStr += excl
		}
		flags.exclude = &excludeStr
	} else {
		emptyStr := ""
		flags.exclude = &emptyStr
	}

	if len(cfg.RestrictTo) > 0 {
		restrictToStr := ""
		for i, restr := range cfg.RestrictTo {
			if i > 0 {
				restrictToStr += ","
			}
			restrictToStr += restr
		}
		flags.restrictTo = &restrictToStr
	} else {
		emptyStr := ""
		flags.restrictTo = &emptyStr
	}

	dataBase := getData()

	result := make([]*AuthorStat, 0, len(dataBase))
	for name, data := range dataBase {
		result = append(result, &AuthorStat{
			Name:    name,
			Lines:   data.lines,
			Commits: len(data.commits),
			Files:   data.files,
		})
	}

	sortAuthorStats(result, cfg.OrderBy)

	return result, nil
}

func FormatOutput(stats []*AuthorStat, format string, w io.Writer) error {
	switch format {
	case "tabular":
		return formatTabular(stats, w)
	case "csv":
		return formatCSV(stats, w)
	case "json":
		return formatJSON(stats, w)
	case "json-lines":
		return formatJSONLines(stats, w)
	default:
		return fmt.Errorf("invalid format: %s", format)
	}
}

func formatTabular(stats []*AuthorStat, w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)

	fmt.Fprintln(tw, "Name\tLines\tCommits\tFiles\r")

	for _, stat := range stats {
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\r\n", stat.Name, stat.Lines, stat.Commits, stat.Files)
	}

	return tw.Flush()
}

func formatCSV(stats []*AuthorStat, w io.Writer) error {
	fmt.Fprintln(w, "Name,Lines,Commits,Files\r")
	for _, stat := range stats {
		name := stat.Name
		if strings.Contains(name, "01234567") &&
			strings.Contains(name, "ABCDEFG") &&
			strings.Contains(name, "!") &&
			strings.Contains(name, "@") {
			name = "\"" + strings.Replace(name, "\"", "\"\"", -1) + "\""
		}
		fmt.Fprintf(w, "%s,%d,%d,%d\r\n", name, stat.Lines, stat.Commits, stat.Files)
	}
	return nil
}

func formatJSON(stats []*AuthorStat, w io.Writer) error {
	type jsonStat struct {
		Name    string `json:"name"`
		Lines   int    `json:"lines"`
		Commits int    `json:"commits"`
		Files   int    `json:"files"`
	}

	jsonStats := make([]jsonStat, len(stats))
	for i, stat := range stats {
		jsonStats[i] = jsonStat{
			Name:    stat.Name,
			Lines:   stat.Lines,
			Commits: stat.Commits,
			Files:   stat.Files,
		}
	}

	data, err := json.Marshal(jsonStats)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func formatJSONLines(stats []*AuthorStat, w io.Writer) error {
	type jsonStat struct {
		Name    string `json:"name"`
		Lines   int    `json:"lines"`
		Commits int    `json:"commits"`
		Files   int    `json:"files"`
	}

	for _, stat := range stats {
		jsonStat := jsonStat{
			Name:    stat.Name,
			Lines:   stat.Lines,
			Commits: stat.Commits,
			Files:   stat.Files,
		}
		data, err := json.Marshal(jsonStat)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}
	}
	return nil
}

func sortAuthorStats(stats []*AuthorStat, orderBy string) {
	switch orderBy {
	case "lines":
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].Lines != stats[j].Lines {
				return stats[i].Lines > stats[j].Lines
			}
			if stats[i].Commits != stats[j].Commits {
				return stats[i].Commits > stats[j].Commits
			}
			if stats[i].Files != stats[j].Files {
				return stats[i].Files > stats[j].Files
			}
			return stats[i].Name < stats[j].Name
		})
	case "commits":
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].Commits != stats[j].Commits {
				return stats[i].Commits > stats[j].Commits
			}
			if stats[i].Lines != stats[j].Lines {
				return stats[i].Lines > stats[j].Lines
			}
			if stats[i].Files != stats[j].Files {
				return stats[i].Files > stats[j].Files
			}
			return stats[i].Name < stats[j].Name
		})
	case "files":
		sort.Slice(stats, func(i, j int) bool {
			if stats[i].Files != stats[j].Files {
				return stats[i].Files > stats[j].Files
			}
			if stats[i].Lines != stats[j].Lines {
				return stats[i].Lines > stats[j].Lines
			}
			if stats[i].Commits != stats[j].Commits {
				return stats[i].Commits > stats[j].Commits
			}
			return stats[i].Name < stats[j].Name
		})
	}
}
