package main

import (
	"fmt"
	"os"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "init":
		err = cmdInit(os.Args[2:])
	case "ss":
		err = cmdScreenshot(os.Args[2:])
	case "note":
		err = cmdNote(os.Args[2:])
	case "finding":
		err = cmdFinding(os.Args[2:])
	case "export":
		err = cmdExport(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Println("pencap", version)
	case "help", "-h", "--help":
		usage()
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "pencap:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, banner())
	fmt.Fprintln(os.Stderr, `Usage:
  pencap init <engagement-name>				scaffold a new engagement folder
  pencap ss <type> [note...]				capture a screenshot into evidence/<type>/
  pencap ss file <src-path> [note...]		copy a file into evidence/files/
  pencap note <type> <text...>				append a timestamped note
  pencap finding add <title>				[--severity crit|high|med|low|info]
  pencap finding link <id> <evidence-path>
  pencap finding list
  pencap export								write findings + evidence to report/draft/report.md`)
}
