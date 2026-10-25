package gitfame

import flag "github.com/spf13/pflag"

type gitFameFlags struct {
	repository   *string
	revision     *string
	orderBy      *string
	useCommitter *bool
	format       *string
	extensions   *string
	languages    *string
	exclude      *string
	restrictTo   *string
}

func initFlags() *gitFameFlags {
	gff := &gitFameFlags{
		repository:   flag.String("repository", "./", "Path to repository."),
		revision:     flag.String("revision", "HEAD", "Pointer to commit."),
		orderBy:      flag.String("order-by", "lines", "Ordering key: lines/commits/files."),
		useCommitter: flag.Bool("use-committer", false, "Use committers instead of authors."),
		format:       flag.String("format", "tabular", "Output format: tabular/csv/json/json-lines."),
		extensions:   flag.String("extensions", "", "Extensions of files, that will be used in calculations, separated with comma."),
		languages:    flag.String("languages", "", "Languages of files, that will be used in calculations, separated with comma. For example: go,markdown."),
		exclude:      flag.String("exclude", "", "Glob-filters to exclude files from calculations, separated with comma."),
		restrictTo:   flag.String("restrict-to", "", "Glob-filters for files, that will be used in calculations, separated with comma."),
	}
	flag.Parse()

	return gff
}
