package main

import (
	"code.google.com/p/goplan9/plan9/acme"
	"go/format"
	"io"
	"os"
	"strconv"
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

	body, err := win.ReadAll("body")
	checkErr(err)
	body, err = format.Source(body)
	checkErr(err)

	// Read current dot addr
	_, _, err = win.ReadAddr() // only for open 'addr' file before write to 'ctl'
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
