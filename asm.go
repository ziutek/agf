package main

import (
	"bytes"
	"text/tabwriter"
	"unicode"
)

func writeSpc(w *bytes.Buffer) {
	w.WriteByte(' ')
}

func writeTab(w *bytes.Buffer) {
	w.WriteByte('\t')
}

func writeLF(w *bytes.Buffer) {
	w.WriteByte('\n')
}

func writeEsc(w *bytes.Buffer) {
	w.WriteByte('\xff')
}

func formatAsm(in []byte) ([]byte, error) {
	w := new(bytes.Buffer)

	lines := bytes.Split(in, []byte{'\n'})
	for i := len(lines) - 1; i >= 0 && len(bytes.TrimSpace(lines[i])) == 0; i-- {
		lines = lines[:i]
	}
	for _, line := range lines {
		lspace := len(line) > 0 && unicode.IsSpace(rune(line[0]))

		var comment []byte
		if n := bytes.IndexByte(line, '/'); n != -1 && n+1 < len(line) && (line[n+1] == '/' || line[n+1] == '*') {
			comment = bytes.TrimRightFunc(line[n:], unicode.IsSpace)
			line = line[:n]
			if comment[1] == '/' && len(bytes.TrimSpace(line)) == 0 {
				if len(line) != 0 {
					writeTab(w)
				}
				writeEsc(w)
				w.Write(comment)
				writeEsc(w)
				writeLF(w)
				continue
			}
		}

		fields := bytes.Fields(line)
		if len(fields) > 0 {
			f := string(fields[0])
			if f[0] == '.' && !lspace || len(fields) == 1 && f[len(f)-1] == ':' ||
				f == "TEXT" || f == "DATA" || f == "GLOBL" || f[0] == '#' {
				// directives at line begin and alone labels
				writeEsc(w)
				for i, f := range fields {
					w.Write(f)
					if i < len(fields)-1 {
						writeSpc(w)
					}
				}
				if comment != nil {
					writeSpc(w)
					w.Write(comment)
				}
				writeEsc(w)
			} else {
				fields = fields[1:]
				if f[len(f)-1] == ':' {
					// label before instruction
					w.WriteString(f)
					f = string(fields[0])
					fields = fields[1:]
				}

				writeTab(w)
				w.WriteString(f)

				for i, f := range fields {
					if i == 0 {
						writeTab(w)
					}
					if f[0] == ',' {
						w.WriteByte(',')
						writeSpc(w)
						if len(f) == 1 {
							continue
						}
						f = f[1:]
					}
					w.Write(f)
					if i < len(fields)-1 && fields[i+1][0] != ',' {
						writeSpc(w)
					}
				}
				if comment != nil {
					writeTab(w)
					writeEsc(w)
					w.Write(comment)
					writeEsc(w)
				}
			}
		} else if comment != nil {
			writeEsc(w)
			w.Write(comment)
			writeEsc(w)
		}
		writeLF(w)
	}
	out := new(bytes.Buffer)
	tw := tabwriter.NewWriter(
		out, 4, 4, 2, ' ',
		tabwriter.StripEscape|tabwriter.TabIndent,
	)
	w.WriteTo(tw)
	tw.Flush()
	return out.Bytes(), nil
}
