package can

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type BusWriter struct {
	Serial     *serial.Port
	sendMu     sync.Mutex
	deviceNode int32 // atomic; 0 = not yet discovered from heartbeat
}

func (w *BusWriter) SetDeviceID(id int) {
	atomic.StoreInt32(&w.deviceNode, int32(id))
	log.Printf("discovered CAN device node ID: %d", id)
}

func GenerateAddress(source int, destination int, fragmentation int, sequenceNumber int) string {
	//  1F000000
	//    + SrcAddr        << 0 6 bits  source Node-Id
	//    + DstAddr        << 6 6 bits  destination Node-Id
	//    + AnotherCounter <<12 2 bits  we dont know what this is, set it to 0, everything else wont work
	//    + MultiMsg       <<14 1 bit   if this is a message composed of multiple CAN-frames
	//    + ErrorOccured   <<15 1 bit   When Response: If an error occured
	//    + IsRequest      <<16 1 bit   If the message is a request
	//    + SeqNr          <<17 2 bits, request counter (should be the same for each frame in a multimsg), copied over to the response

	addr := 0x1F000000
	addr |= source << 0
	addr |= destination << 6
	addr |= fragmentation << 14
	addr |= 0x1 << 16
	addr |= sequenceNumber << 17

	return fmt.Sprintf("%X", addr)
}

func (w *BusWriter) WriteCommand(command Command) {
	deviceNode := int(atomic.LoadInt32(&w.deviceNode))
	if deviceNode == 0 {
		log.Println("command ignored: CAN device node not yet discovered (waiting for heartbeat)")
		return
	}
	frames := CommandToFrames(command, deviceNode)
	w.Send(frames)
}

func CommandToFrames(command Command, deviceNode int) []string {
	data := command.Code
	// src=1: our registered node ID (device sends responses to dst=1)
	address := GenerateAddress(1, deviceNode, command.Fragmentation, 1)
	length := len(data) / 2

	var result []string

	if length > 8 {
		numberOfDatagrams := length / 7
		if length%7 > 0 {
			numberOfDatagrams -= 1
		}

		for i := 0; i <= numberOfDatagrams; i++ {
			chunk := data[i*14 : i*14+14]
			payload := fmt.Sprintf("T%s%x%02x%s\r", address, len(chunk)/2+1, i, chunk)
			result = append(result, payload)
		}

		tail := numberOfDatagrams + 1
		restLength := (length - tail*7) * 2
		chunk := data[tail*14 : tail*14+restLength]
		payload := fmt.Sprintf("T%s%x%02x%s\r", address, len(chunk)/2+1, tail|0x80, chunk)
		result = append(result, payload)
	} else {
		payload := fmt.Sprintf("T%s%x%s\r", address, len(data)/2, data)
		result = append(result, payload)
	}

	return result
}

func (w *BusWriter) Send(frames []string) {
	w.sendMu.Lock()
	defer w.sendMu.Unlock()
	for _, frame := range frames {
		w.writeAndWait(frame)
	}
}

func (w *BusWriter) Write(id string, data string) {
	length := len(data) / 2
	log.Println("Length", length)
	if length > 8 {
		numberOfDatagrams := length / 7
		if length%7 > 0 {
			numberOfDatagrams -= 1
		}

		for i := 0; i <= numberOfDatagrams; i++ {
			chunk := data[i*14 : i*14+14]
			w.writeAndWait(fmt.Sprintf("T%s%x%02x%s\r", id, len(chunk)/2+1, i, chunk))
		}

		tail := numberOfDatagrams + 1
		restLength := (length - tail*7) * 2
		chunk := data[tail*14 : tail*14+restLength]
		w.writeAndWait(fmt.Sprintf("T%s%x%02x%s\r", id, len(chunk)/2+1, tail|0x80, chunk))
	} else {
		w.writeAndWait(fmt.Sprintf("T%s%x%s\r", id, len(data)/2, data))
	}
}

func (w *BusWriter) writeAndWait(payload string) {
	if debugMode {
		fmt.Println("command string: ", payload)
		fmt.Println("command ascii: ", []byte(payload))
	}

	n, err := w.Serial.Write([]byte(payload))
	if err != nil {
		log.Printf("serial write error: %v", err)
	} else {
		log.Printf("serial write ok: %d bytes", n)
	}
	time.Sleep(500 * time.Millisecond)
}
