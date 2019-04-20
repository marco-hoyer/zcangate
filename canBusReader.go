package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type CanBusReader struct {
	interfaceReader io.Reader
}

func newCanBusReader(src io.Reader) *CanBusReader {
	return &CanBusReader{interfaceReader: src}
}

func (c *CanBusReader) Read(p []byte) (int, error) {
	reader := bufio.NewReader(c.interfaceReader)

	for {
		line, err := reader.ReadString('\r')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		fmt.Print(line)
	}

	return 0, nil
}
