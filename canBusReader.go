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
	id           string
	binaryId     uint64
	pdu          int
	length       int
	data         string
	CN1FAddress  CN1FAddress
	pingDeviceId int
}

func (f CanBusFrame) toBytes() []byte {
	return []byte("")
}

type CN1FAddress struct {
	src            uint64
	dst            uint64
	unknownCounter uint64
	multiMessage   uint64
	A8000          uint64
	A10000         uint64
	SequenceNumber uint64
}

func CN1FAddressFromBinaryAddress(a uint64) CN1FAddress {
	return CN1FAddress{
		src:            (a >> 0) & 0x3f,
		dst:            (a >> 6) & 0x3f,
		unknownCounter: (a >> 12) & 0x03,
		multiMessage:   (a >> 14) & 0x01,
		A8000:          (a >> 15) & 0x01,
		A10000:         (a >> 16) & 0x01,
		SequenceNumber: (a >> 17) & 0x03,
	}
}

func getIdFromPing(binaryId uint64) int {
	// address 1000000x indicates a heartbeat from a CAN bus device with id x
	if binaryId&0xFFFFFFC0 == 0x10000000 {
		return int(binaryId & 0x3f)
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
		address := value[1:9]
		binaryAddress, _ := strconv.ParseUint(address, 16, 32)

		length, _ := strconv.ParseInt(string(value[9]), 10, 64)
		pdu := int(binaryAddress >> 14 & 0x7ff)

		var cn1fAddress CN1FAddress
		if strings.HasPrefix(address, "1F") {
			cn1fAddress = CN1FAddressFromBinaryAddress(binaryAddress)
		}

		return CanBusFrame{
			id:           address,
			binaryId:     binaryAddress,
			pdu:          pdu,
			length:       int(length),
			data:         value[10:],
			CN1FAddress:  cn1fAddress,
			pingDeviceId: getIdFromPing(binaryAddress),
		}
	}
}
