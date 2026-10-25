package gitfame

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"text/tabwriter"
)

type data struct {
	lines   int
	commits map[string]struct{}
	files   int
}

type dataJSON struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

type dataMap map[string]*data

func (dm dataMap) parseBlame(blameData string) {
	var (
		name   string
		commit string
		lines  int
	)
	commitName := make(map[string]string)
	names := make(map[string]struct{})

	blames := strings.Split(strings.TrimSuffix(blameData, "\n"), "\n")
	for _, blame := range blames {
		info := strings.Split(blame, " ")
		key := info[0]
		switch {
		case isHash(key):
			if len(info) != 4 {
				continue
			}
			commit = key
			lines, _ = strconv.Atoi(info[3])
			if v, ok := commitName[commit]; ok {
				name = v
				dm.addData(name, lines, commit, 0)
			}
		case (key == "author" && !*flags.useCommitter) || (key == "committer" && *flags.useCommitter):
			var buildName strings.Builder
			for i := range info {
				if i == 0 {
					continue
				}
				if i > 1 {
					buildName.Write([]byte(" "))
				}
				buildName.Write([]byte(info[i]))
			}
			name = buildName.String()
			dm.addData(name, lines, commit, 0)
			commitName[commit] = name
			names[name] = struct{}{}
		}
	}

	for k := range names {
		dm.addData(k, 0, "", 1)
	}
}

func (dm dataMap) blameFile(filePath string) {
	var out strings.Builder

	cmd := exec.Command("git", "-C", *flags.repository, "blame", "--porcelain", *flags.revision, "--", filePath)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	if out.String() == "" {
		dm.blameEmpty(filePath)
	}
	dm.parseBlame(out.String())
}

func (dm dataMap) blameEmpty(filePath string) {
	var format, out, commitHash, name strings.Builder

	format.Write([]byte("--pretty=format:%H "))
	if *flags.useCommitter {
		format.Write([]byte("%cn"))
	} else {
		format.Write([]byte("%an"))
	}

	cmd := exec.Command("git", "-C", *flags.repository, "log", "-n", "1", "--pretty=format:%H %an", *flags.revision, "--", filePath)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	logData := strings.Split(out.String(), " ")
	commitHash.Write([]byte(logData[0]))
	for i, str := range logData {
		if i == 0 {
			continue
		}
		if i > 1 {
			name.Write([]byte(" "))
		}
		name.Write([]byte(str))
	}
	dm.addData(name.String(), 0, commitHash.String(), 1)
}

func (dm dataMap) addData(name string, lines int, commit string, files int) {
	if _, ok := dm[name]; !ok {
		dm[name] = &data{commits: make(map[string]struct{})}
	}
	dm[name].lines += lines
	if commit != "" {
		dm[name].commits[commit] = struct{}{}
	}
	dm[name].files += files

}

func (dm dataMap) print() {
	switch *flags.format {
	case "tabular":
		dm.printTabular()
	case "csv":
		dm.printCSV()
	case "json":
		dm.printJSON()
	case "json-lines":
		dm.printJSONLines()
	default:
		log.Fatal("wrong argument for --format. It must be tabular/csv/json/json-lines")
	}
}

func (dm dataMap) printTabular() {
	records := dm.sorted()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	for _, record := range records {
		var str strings.Builder
		str.Write([]byte(record[0]))
		str.Write([]byte("\t"))
		str.Write([]byte(record[1]))
		str.Write([]byte("\t"))
		str.Write([]byte(record[2]))
		str.Write([]byte("\t"))
		str.Write([]byte(record[3]))

		fmt.Fprintln(w, str.String())
	}
	w.Flush()
}

func (dm dataMap) printCSV() {
	records := dm.sorted()

	w := csv.NewWriter(os.Stdout)
	err := w.WriteAll(records)
	if err != nil {
		log.Fatal(err)
	}
}

func (dm dataMap) printJSON() {
	records := dm.sorted()[1:]

	recordsJSON := make([]dataJSON, len(records))
	for i, record := range records {
		recordsJSON[i] = dataJSON{Name: record[0]}
		recordsJSON[i].Lines, _ = strconv.Atoi(record[1])
		recordsJSON[i].Commits, _ = strconv.Atoi(record[2])
		recordsJSON[i].Files, _ = strconv.Atoi(record[3])
	}
	out, err := json.Marshal(recordsJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))
}

func (dm dataMap) printJSONLines() {
	records := dm.sorted()[1:]
	var output strings.Builder

	for _, record := range records {
		recordJSON := dataJSON{Name: record[0]}
		recordJSON.Lines, _ = strconv.Atoi(record[1])
		recordJSON.Commits, _ = strconv.Atoi(record[2])
		recordJSON.Files, _ = strconv.Atoi(record[3])

		out, err := json.Marshal(recordJSON)
		if err != nil {
			log.Fatal(err)
		}
		output.Write(out)
		output.Write([]byte("\n"))
	}
	fmt.Print(output.String())
}

func (dm dataMap) sorted() [][]string {
	datas := make([]dataWithName, 0)
	for k, v := range dm {
		datas = append(datas, dataWithName{*v, k})
	}

	slices.SortFunc(datas, cmpData)

	records := [][]string{{"Name", "Lines", "Commits", "Files"}}
	for _, data := range datas {
		records = append(records, []string{data.name, strconv.Itoa(data.lines), strconv.Itoa(len(data.commits)), strconv.Itoa(data.files)})
	}
	return records
}
