package main

import (
	"go/format"
	"io"
	"io/ioutil"
	"os"
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

func readAddr(addr *os.File) (uint64, uint64) {
	buf := make([]byte, 24)
	_, err := io.ReadFull(addr, buf)
	checkErr(err)
	dot := strings.Fields(string(buf))
	if len(dot) != 2 {
		die("can't read addr")
	}

	a, err := strconv.ParseUint(dot[0], 0, 64)
	checkErr(err)
	b, err := strconv.ParseUint(dot[1], 0, 64)
	checkErr(err)

	return a, b
}

func main() {
	winid := os.Getenv("winid")
	if winid == "" {
		die("$winid not defined")
	}
	mnt := os.Getenv("acmefs")
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
	dotA, dotB := readAddr(addr)
	writeStr(addr, ",")

	_, err = data.Write(body)
	checkErr(err)

	writeStr(addr, "#"+strconv.FormatUint((dotA+dotB)/2, 10))
	writeStr(ctl, "dot=addr\nshow\nmark\n")
}
