package main

import (
	"fmt"
	"strings"
)

type ComfoNetHeader struct {
	src            int
	dst            int
	A3000          int
	multiMessage   int
	A8000          int
	A10000         int
	SequenceNumber int
}

func newComfoNetHeader(sourceId int, destinationId int, multiMessageFlag int, sequenceNumber int) ComfoNetHeader {
	return ComfoNetHeader{
		src:            sourceId,
		dst:            destinationId,
		A3000:          0x1,
		multiMessage:   multiMessageFlag,
		A8000:          0x0,
		A10000:         0x1,
		SequenceNumber: sequenceNumber,
	}
}

func toComfoNetHeader(binaryCanId int) ComfoNetHeader {
	return ComfoNetHeader{
		src:            (binaryCanId >> 0) & 0x3f,
		dst:            (binaryCanId >> 6) & 0x3f,
		A3000:          (binaryCanId >> 12) & 0x03,
		multiMessage:   (binaryCanId >> 14) & 0x01,
		A8000:          (binaryCanId >> 15) & 0x01,
		A10000:         (binaryCanId >> 16) & 0x01,
		SequenceNumber: (binaryCanId >> 17) & 0x03,
	}
}

func (c *ComfoNetHeader) toId() string {
	addr := 0x0
	addr |= c.src << 0
	addr |= c.dst << 6
	addr |= c.A3000 << 12
	addr |= c.multiMessage << 14
	addr |= c.A8000 << 15
	addr |= c.A10000 << 16
	addr |= c.SequenceNumber << 17
	addr |= 0x1F << 24

	return strings.ToUpper(fmt.Sprintf("%x", addr))

}
