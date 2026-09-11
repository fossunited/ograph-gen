package convert

import (
	"bytes"
	"fmt"
	"os/exec"
)

type Converter struct {
	bin string
}

func New() *Converter {
	return &Converter{bin: "rsvg-convert"}
}

func (c *Converter) SetBinary(b string) {
	c.bin = b
}

func (c *Converter) Convert(svg []byte) ([]byte, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(c.bin, "--format=png")
	cmd.Stdin = bytes.NewReader(svg)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("rsvg-convert: %w\nstderr: %s", err, stderr.String())
	}

	if stdout.Len() == 0 {
		return nil, fmt.Errorf("rsvg-convert produced no output")
	}

	return stdout.Bytes(), nil
}
