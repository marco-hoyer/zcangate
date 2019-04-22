package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type CanBusReader struct {
	interfaceReader io.Reader
	output          chan CanBusFrame
}

func newCanBusReader(reader io.Reader, output chan CanBusFrame) *CanBusReader {
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
			c.output <- toCanBusFrame(line)
		}
	}()
}

type CanBusFrame struct {
	id       string
	binaryId uint64
	pdu      int
	length   int
	data     string
}

func getComfoAirId(frame CanBusFrame) int {
	// address 1000000x indicates a heartbeat from the ComfoAir device with id x
	if frame.binaryId&0xFFFFFFC0 == 0x10000000 {
		return int(frame.binaryId & 0x3f)
	} else {
		return 0
	}
}

func toCanBusFrame(line string) CanBusFrame {
	value := strings.Trim(line, "\r")

	if !strings.HasPrefix(value, "T") || len(value) < 10 {
		log.Println("found broken message", value)
		return CanBusFrame{}
	} else {
		//prefix := string(value[0])
		address := value[1:9]
		binaryAddress, _ := strconv.ParseUint(address, 16, 32)

		length, _ := strconv.ParseInt(string(value[9]), 10, 64)

		pdu := int(binaryAddress >> 14 & 0x7ff)
		println("PDU", pdu)
		return CanBusFrame{
			id:       address,
			binaryId: binaryAddress,
			pdu:      pdu,
			length:   int(length),
			data:     value[10:],
		}
	}
}
