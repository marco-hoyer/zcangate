package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type CanBusReader struct {
	interfaceReader io.Reader
	output          chan ComfoNetMessage
}

func newCanBusReader(reader io.Reader, output chan ComfoNetMessage) *CanBusReader {
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
			if strings.HasPrefix(line, "T") {
				c.output <- toCanBusFrame(strings.Trim(line, "\r"))
			}
		}
	}()
}

type ComfoNetMessage struct {
	id             string
	binaryId       uint64
	pdu            int
	length         int
	ComfoAirHeader ComfoNetHeader
	data           string
	pingDeviceId   int
}

func getIdFromPing(binaryId uint64) int {
	// address 1000000x indicates a heartbeat from a CAN bus device with id x
	if binaryId&0xFFFFFFC0 == 0x10000000 {
		return int(binaryId & 0x3f)
	} else {
		return 0
	}
}

func toCanBusFrame(line string) ComfoNetMessage {
	address := line[1:9]
	binaryAddress, _ := strconv.ParseUint(address, 16, 32)

	length, _ := strconv.ParseInt(string(line[9]), 10, 64)
	pdu := int(binaryAddress >> 14 & 0x7ff)

	return ComfoNetMessage{
		id:             address,
		binaryId:       binaryAddress,
		pdu:            pdu,
		length:         int(length),
		ComfoAirHeader: toComfoNetHeader(int(binaryAddress)),
		data:           line[10:],
		pingDeviceId:   getIdFromPing(binaryAddress),
	}

}
