package progressbar

import (
	"fmt"
	"io"
	"strings"
)

type ProgressBar struct {
	progress, total, width int
	closed                 bool
	fieldCharacter         string
	emptyCharacter         string
	prefix                 string
	suffix                 string
	out                    io.Writer
	err                    error
}

func New(t int, out io.Writer) (*ProgressBar, error) {
	if t <= 0 {
		return nil, fmt.Errorf("total must be greater than 0")
	}
	b := &ProgressBar{
		progress:       0,
		total:          t,
		width:          20,
		closed:         false,
		fieldCharacter: "█",
		emptyCharacter: "░",
		prefix:         "|",
		suffix:         "|",
		out:            out,
		err:            nil,
	}
	return b, nil
}

func (b *ProgressBar) Tick() {
	if b.closed || b.err != nil {
		return
	}
	b.progress += 1
	b.render()
	if b.progress >= b.total {
		b.Close()
	}
}

func (b *ProgressBar) render() {
	progress := float64(b.progress) / float64(b.total)
	filled := int(progress * float64(b.width))
	empty := b.width - filled
	bar := strings.Repeat(b.fieldCharacter, filled) + strings.Repeat(b.emptyCharacter, empty)
	percent := progress * 100
	_, err := fmt.Fprintf(b.out, "\r%s%s%s %.1f%%", b.prefix, bar, b.suffix, percent)
	if err != nil {
		b.err = err
	}
}

func (b *ProgressBar) Close() {
	if !b.closed {
		b.closed = true
		_, err := fmt.Fprintf(b.out, "\r%s\r", strings.Repeat(" ", 10))
		if err != nil {
			b.err = err
		}
	}
}
