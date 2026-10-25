package gitfame

import (
	"log"
	"strings"
)

type dataWithName struct {
	data
	name string
}

type comparator func(f, s dataWithName) int

func cmpLines(f, s dataWithName) int {
	fl, sl := f.lines, s.lines
	return sl - fl
}

func cmpCommits(f, s dataWithName) int {
	fc, sc := len(f.commits), len(s.commits)
	return sc - fc
}

func cmpFiles(f, s dataWithName) int {
	ff, sf := f.files, s.files
	return sf - ff
}

func cmpName(f, s dataWithName) int {
	return strings.Compare(f.name, s.name)
}

func cmpData(f, s dataWithName) int {
	var cmps []comparator
	switch {
	case strings.Compare(*flags.orderBy, "lines") == 0:
		cmps = []comparator{cmpLines, cmpCommits, cmpFiles, cmpName}
	case strings.Compare(*flags.orderBy, "commits") == 0:
		cmps = []comparator{cmpCommits, cmpLines, cmpFiles, cmpName}
	case strings.Compare(*flags.orderBy, "files") == 0:
		cmps = []comparator{cmpFiles, cmpLines, cmpCommits, cmpName}
	default:
		log.Fatal("order-by must be a lines/commits/files")
	}

	for _, cmp := range cmps {
		if res := cmp(f, s); res != 0 {
			return res
		}
	}
	return 0
}
