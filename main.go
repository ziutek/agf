package main

import (
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func die(info string) {
	fmt.Fprintln(os.Stderr, info)
	os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func openRW(name string) *os.File {
	f, err := os.OpenFile(name, os.O_RDWR, 0)
	checkErr(err)
	return f
}

func openW(name string) *os.File {
	f, err := os.OpenFile(name, os.O_WRONLY, 0)
	checkErr(err)
	return f
}

func writeStr(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	checkErr(err)
}

func main() {
	winid := os.Getenv("winid")
	if winid == "" {
		die("$winid not defined")
	}
	mnt := os.Getenv("acme_mnt")
	if mnt == "" {
		mnt = "/mnt/acme"
	}
	checkErr(os.Chdir(filepath.Join(mnt, winid)))

	body, err := ioutil.ReadFile("body")
	checkErr(err)
	body, err = format.Source(body)
	checkErr(err)

	ctl := openW("ctl")
	defer ctl.Close()
	addr := openRW("addr")
	defer addr.Close()
	data := openW("data")
	defer data.Close()

	writeStr(ctl, "mark\nnomark\naddr=dot\n")

	buf := make([]byte, 24)
	_, err = io.ReadFull(addr, buf)
	checkErr(err)
	dot := strings.Fields(string(buf))
	if len(dot) != 2 {
		die("bad dot address")
	}

	dotA, err := strconv.ParseUint(dot[0], 0, 64)
	checkErr(err)
	dotB, err := strconv.ParseUint(dot[1], 0, 64)
	checkErr(err)

	writeStr(addr, ",")

	_, err = data.Write(body)
	checkErr(err)

	_, err = fmt.Fprintf(addr, "#%d", (dotA+dotB)/2)
	checkErr(err)

	writeStr(ctl, "dot=addr\nshow\nmark\n")
}
