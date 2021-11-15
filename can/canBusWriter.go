package can

import (
	"fmt"
	"github.com/marco-hoyer/zcangate/common"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/tarm/serial"
	"log"
	"strings"
	"time"
)

const sequenceNumberStateKey = "canCommandSequenceNumber"
const ComfoAirId = "canComfoAirId"

type CanBusWriter struct {
	Serial   *serial.Port
	StateDao dao.StateDao
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

	addr |= 0x0 << 12
	addr |= fragmentation << 14
	addr |= 0x0 << 15
	addr |= 0x1 << 16
	addr |= sequenceNumber << 17

	return fmt.Sprintf("%X", addr)
}

func (w *CanBusWriter) WriteCommand(command common.Command) {
	frames := CommandToFrames(command)
	w.Send(frames)
}

func CommandToFrames(command common.Command) []string {
	data := command.Code
	address := GenerateAddress(5, 1, command.Fragmentation, 1)
	length := len(data) / 2

	var result []string

	if length > 8 {
		numberOfDatagrams := length / 7
		rest := length % 7
		if rest > 0 {
			numberOfDatagrams -= 1
		}

		n := 0
		for i := 0; i <= numberOfDatagrams; i++ {
			chunk := data[i*14 : i*14+14]
			payload := fmt.Sprintf("T%s%x%02x%s\r", address, len(chunk)/2+1, i, chunk)
			result = append(result, payload)
			n += 1
		}

		restLength := (length - n*7) * 2
		chunk := data[n*14 : n*14+restLength]

		payload := fmt.Sprintf("T%s%x%02x%s\r", address, len(chunk)/2+1, n|0x80, chunk)
		result = append(result, payload)
	} else {
		payload := fmt.Sprintf("T%s%x%s\r", address, len(data)/2, data)
		result = append(result, payload)
	}

	return result
}

func (w *CanBusWriter) Send(frames []string) {
	for _, frame := range frames {
		w.writeAndWait(frame)
	}
}

func (w *CanBusWriter) Write(id string, data string) {
	length := len(data) / 2
	log.Println("Length", length)
	if length > 8 {
		numberOfDatagrams := length / 7
		rest := length % 7
		if rest > 0 {
			numberOfDatagrams -= 1
		}

		n := 0
		for i := 0; i <= numberOfDatagrams; i++ {
			chunk := data[i*14 : i*14+14]
			payload := fmt.Sprintf("T%s%x%02x%s\r", id, len(chunk)/2+1, i, chunk)
			w.writeAndWait(payload)
			n += 1
		}

		restLength := (length - n*7) * 2
		chunk := data[n*14 : n*14+restLength]

		payload := fmt.Sprintf("T%s%x%02x%s\r", id, len(chunk)/2+1, n|0x80, chunk)
		w.writeAndWait(payload)
	} else {
		//w.Serial.Write(f.toBytes())
		payload := fmt.Sprintf("T%s%x%s\r", id, len(data)/2, data)
		w.writeAndWait(payload)
	}
}

func (w *CanBusWriter) writeAndWait(payload string) {
	fmt.Println("command string: ", payload)
	fmt.Println("command ascii: ", []byte(payload))

	w.Serial.Write([]byte(payload))
	time.Sleep(500 * time.Millisecond)
}

func (w *CanBusWriter) writeAndWait2(payload string) {
	fmt.Println("command string: ", payload)
	fmt.Println("command ascii: ", []byte(payload))

	w.Serial.Flush()
	w.Serial.Write([]byte(payload))

	response := make([]byte, 128)
	var retries = 50
	for {
		w.Serial.Read(response)
		//log.Println("raw response: ", response)
		if strings.Contains(string(response), "Z\r") {
			log.Printf("COMMAND FINISHED SUCCESSFULLY")
			break
		}

		retries--
		if retries < 1 {
			log.Println("COMMAND TIMED OUT")
			break

		}
	}

}
