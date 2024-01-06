package can

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type BusReader struct {
	interfaceReader io.Reader
	output          chan BusFrame
}

func NewCanBusReader(reader io.Reader, output chan BusFrame) *BusReader {
	return &BusReader{interfaceReader: reader, output: output}
}

func (c *BusReader) Read() {
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

type BusFrame struct {
	Id          string
	BinaryId    uint64
	Pdu         int
	Length      int
	Data        string
	CN1FAddress CN1FAddress
}

func (f BusFrame) toBytes() []byte {
	return []byte("")
}

type CN1FAddress struct {
	src            uint64
	dst            uint64
	unknownCounter uint64
	multiMessage   uint64
	errorOccured   uint64
	isRequest      uint64
	SequenceNumber uint64
}

func CN1FAddressFromBinaryAddress(a uint64) CN1FAddress {
	return CN1FAddress{
		src:            (a >> 0) & 0x3f,
		dst:            (a >> 6) & 0x3f,
		unknownCounter: (a >> 12) & 0x03,
		multiMessage:   (a >> 14) & 0x01,
		errorOccured:   (a >> 15) & 0x01,
		isRequest:      (a >> 16) & 0x01,
		SequenceNumber: (a >> 17) & 0x03,
	}
}

func toCanBusFrame(line string) BusFrame {
	address := line[1:9]
	binaryAddress, _ := strconv.ParseUint(address, 16, 32)

	length, _ := strconv.ParseInt(string(line[9]), 10, 64)
	pdu := int(binaryAddress >> 14 & 0x7ff)

	var cn1fAddress CN1FAddress
	if strings.HasPrefix(address, "1F") {
		cn1fAddress = CN1FAddressFromBinaryAddress(binaryAddress)
	}

	return BusFrame{
		Id:          address,
		BinaryId:    binaryAddress,
		Pdu:         pdu,
		Length:      int(length),
		Data:        line[10:],
		CN1FAddress: cn1fAddress,
	}

}
