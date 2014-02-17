package main

import (
	"bytes"
	"io"
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
	r := bytes.NewBuffer(in)
	w := new(bytes.Buffer)

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		lspace := len(line) > 0 && unicode.IsSpace(rune(line[0]))

		var comment []byte
		if n := bytes.Index(line, []byte("//")); n != -1 {
			comment = bytes.TrimSpace(line[n+2:])
			line = line[:n]
		}

		fields := bytes.Fields(line)
		if len(fields) > 0 {
			f := fields[0]
			if f[0] == '.' && !lspace || len(fields) == 1 && f[len(f)-1] == ':' {
				// directives at line begin and alone labels
				writeEsc(w)
				for i, f := range fields {
					w.Write(f)
					if i < len(fields)-1 {
						writeSpc(w)
					}
				}
				writeEsc(w)
			} else {
				fields = fields[1:]
				if f[len(f)-1] == ':' {
					// label before instruction
					w.Write(f)
					f = fields[0]
					fields = fields[1:]
				}

				writeTab(w)
				w.Write(f)
				writeTab(w)

				for i, f := range fields {
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

			}
		}

		if comment != nil {
			if len(fields) == 0 && lspace || len(fields) > 0 {
				writeTab(w)
			}
			writeEsc(w)
			w.WriteString("// ")
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
