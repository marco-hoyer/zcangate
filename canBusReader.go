package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type CanBusReader struct {
	interfaceReader io.Reader
	output          chan string
}

func newCanBusReader(reader io.Reader, output chan string) *CanBusReader {
	return &CanBusReader{interfaceReader: reader, output: output}
}

func (c *CanBusReader) Read() {
	go func() {
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
			//fmt.Println(line)
			c.output <- line
		}
	}()
}
