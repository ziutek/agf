package main

import (
	"bytes"
	"code.google.com/p/goplan9/plan9/acme"
	"errors"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func die(info string) {
	io.WriteString(os.Stderr, info+"\n")
	os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		die(err.Error())
	}
}

func main() {
	winid := os.Getenv("winid")
	if winid == "" {
		die("$winid not defined")
	}
	id, err := strconv.ParseUint(winid, 10, 0)
	checkErr(err)

	win, err := acme.Open(int(id), nil)
	checkErr(err)

	// Obtain a file type
	tag, err := win.ReadAll("tag")
	checkErr(err)
	typ := ""
	if fields := bytes.Fields(tag); len(fields) > 0 {
		fname := filepath.Base(string(fields[0]))
		if i := strings.LastIndex(fname, "."); i != -1 {
			typ = fname[i+1:]
		}
	}
	if typ == "" {
		die("there is no extension in file name")
	}

	// Read and format a body
	body, err := win.ReadAll("body")
	checkErr(err)
	body, err = formatSrc(body, typ)
	checkErr(err)

	// Read current dot addr
	_, _, err = win.ReadAddr() // for open 'addr' file before write to 'ctl'
	checkErr(err)
	checkErr(win.Ctl("mark\nnomark\naddr=dot\n"))
	dotA, dotB, err := win.ReadAddr()
	checkErr(err)

	// Replace body
	checkErr(win.Addr(","))
	_, err = win.Write("data", body)
	checkErr(err)

	// Set cursor position near previous dot
	checkErr(win.Addr("#%d", (dotA+dotB)/2))
	checkErr(win.Ctl("dot=addr\nshow\nmark\n"))
}

func formatSrc(body []byte, typ string) ([]byte, error) {
	switch typ {
	case "go":
		return format.Source(body)
	case "c", "cc", "cpp", "cxx", "h":
		return indent(body)
	}

	return nil, errors.New("unknown file type: " + typ)
}

func indent(body []byte) ([]byte, error) {
	// Generally, try format C source in Go style.
	cmd := exec.Command(
		"indent",
		"--braces-on-if-line",
		"--cuddle-else",
		"--cuddle-do-while",
		"--braces-on-struct-decl-line",
		//"--braces-on-func-def-line",
		//"--dont-break-procedure-type",
		//"--blank-lines-after-declarations",
		"--blank-lines-after-procedures",
		"--dont-line-up-parentheses",
		"--no-space-after-function-call-names",
		"--no-space-after-casts",
		"--declaration-comment-column0",
		"--comment-indentation0",
		"--indent-level4",
		"--use-tabs",
		"--tab-size4",
	)
	cmd.Stdin = bytes.NewBuffer(body)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return nil, errors.New(err.Error() + "\n" + stderr.String())
	}
	return stdout.Bytes(), nil
}
