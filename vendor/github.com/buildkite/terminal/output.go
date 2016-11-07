package terminal

import (
	"bytes"
	"strings"
)

type outputBuffer struct {
	buf bytes.Buffer
}

func (b *outputBuffer) appendNodeStyle(n node) {
	b.buf.Write([]byte(`<span class="`))
	b.buf.Write([]byte(n.style.asClasses()))
	b.buf.Write([]byte(`">`))
}

func (b *outputBuffer) closeStyle() {
	b.buf.Write([]byte("</span>"))
}

// Append a character to our outputbuffer, escaping HTML bits as necessary.
func (b *outputBuffer) appendChar(char rune) {
	switch char {
	case '&':
		b.buf.WriteString("&amp;")
	case '\'':
		b.buf.WriteString("&#39;")
	case '<':
		b.buf.WriteString("&lt;")
	case '>':
		b.buf.WriteString("&gt;")
	case '"':
		b.buf.WriteString("&quot;")
	case '/':
		b.buf.WriteString("&#47;")
	default:
		b.buf.WriteRune(char)
	}
}

func outputLineAsHTML(line []node) string {
	var openStyles int
	var lineBuf outputBuffer

	for idx, node := range line {
		if idx == 0 && !node.style.isEmpty() {
			lineBuf.appendNodeStyle(node)
			openStyles++
		} else if idx > 0 {
			previous := line[idx-1]
			if !node.hasSameStyle(previous) {
				if node.style.isEmpty() {
					lineBuf.closeStyle()
					openStyles--
				} else {
					lineBuf.appendNodeStyle(node)
					openStyles++
				}
			}
		}
		if node.elem != nil {
			lineBuf.buf.WriteString(node.elem.asHTML())
		} else {
			lineBuf.appendChar(node.blob)
		}
	}
	for i := 0; i < openStyles; i++ {
		lineBuf.closeStyle()
	}
	return strings.TrimRight(lineBuf.buf.String(), " \t")
}
