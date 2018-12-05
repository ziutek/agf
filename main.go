package main

import (
	"bytes"
	"errors"

	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/imports"

	"9fans.net/go/acme"
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

func checkFormatErr(path string, err error) {
	if err == nil {
		return
	}
	s := err.Error() + "\n"
	if !strings.HasPrefix(s, path) {
		s = path + s
	}
	io.WriteString(os.Stdout, s)
	os.Exit(2)
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

	tag, err := win.ReadAll("tag")
	checkErr(err)
	typ := ""
	fpath := ""
	if fields := bytes.Fields(tag); len(fields) > 0 {
		fpath = string(fields[0])
		fname := filepath.Base(fpath)
		i := strings.LastIndexByte(fname, '.')
		if i == -1 {
			die("file name without extension")
		}
		typ = fname[i+1:]
	}

	body, err := win.ReadAll("body")
	checkErr(err)
	body, err = formatSrc(fpath, body, typ)
	checkFormatErr(fpath, err)

	_, _, err = win.ReadAddr()
	checkErr(err)
	checkErr(win.Ctl("mark\nnomark\naddr=dot\n"))
	dot, _, err := win.ReadAddr()
	checkErr(err)

	checkErr(win.Addr(","))
	_, err = win.Write("data", body)
	checkErr(err)

	checkErr(win.Addr("#%d", dot))
	checkErr(win.Ctl("dot=addr\nshow\nmark\n"))
}

func formatSrc(fpath string, body []byte, typ string) ([]byte, error) {
	switch typ {
	case "go":
		return imports.Process(fpath, body, &imports.Options{Comments: true, TabWidth: 4})
	case "c", "cc", "cpp", "cxx", "h", "hh":
		return astyle("c", body)
	case "java":
		return astyle("java", body)
	case "s", "S":
		return formatAsm(body)
	}

	return nil, errors.New("unknown file type: " + typ)
}

func indent(body []byte) ([]byte, error) {
	// Try format C source in Go style.
	cmd := exec.Command(
		"indent",
		"--braces-on-if-line",
		"--cuddle-else",
		"--cuddle-do-while",
		"--braces-on-struct-decl-line",
		"--braces-on-func-def-line",
		//"--dont-break-procedure-type",
		//"--blank-lines-after-declarations"
		"--blank-lines-after-procedures",
		"--dont-line-up-parentheses",
		"--no-space-after-function-call-names",
		"--no-space-after-casts",
		"--no-space-after-parentheses",
		"--declaration-comment-column0",
		"--comment-indentation0",
		"--indent-level4",
		"--use-tabs",
		"--tab-size4",
		"--line-length80",
		"--comment-line-length80",
		"--honour-newlines",
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

func astyle(mode string, body []byte) ([]byte, error) {
	// Try format C source in Go style.
	cmd := exec.Command(
		"astyle",
		"--mode="+mode,
		"--style=java",
		"--indent=tab",
		"--pad-header",
		"--add-brackets",
		"--max-code-length=80",
		"--break-after-logical",
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
