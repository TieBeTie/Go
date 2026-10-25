package gitfame

import (
	"log"
	"os/exec"
	"strings"
)

var flags *gitFameFlags

func GitFame() {
	flags = initFlags()
	dataBase := getData()
	dataBase.print()
}

func getData() (dataBase dataMap) {
	dataBase = make(dataMap)
	rawFilePaths := getFiles()
	filePaths := filter(rawFilePaths)
	for _, filePath := range filePaths {
		dataBase.blameFile(filePath)
	}
	return
}

func getFiles() []string {
	var out strings.Builder

	cmd := exec.Command("git", "-C", *flags.repository, "ls-tree", "-r", "--name-only", *flags.revision)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	if out.String() == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSuffix(out.String(), "\n"), "\n")
}
