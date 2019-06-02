package can

import (
	"fmt"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/tarm/serial"
	"log"
	"strings"
	"time"
)

const sequenceNumberStateKey = "canCommandSequenceNumber"

type CanBusWriter struct {
	Serial   *serial.Port
	StateDao dao.StateDao
}

func GenerateAddress(source int, destination int, fragmentation int, sequenceNumber int) string {
	addr := 0x0
	addr |= source << 0
	addr |= destination << 6

	addr |= 0x1 << 12
	addr |= fragmentation << 14
	addr |= 0x0 << 15
	addr |= 0x1 << 16
	addr |= sequenceNumber << 17
	addr |= 0x1F << 24

	return fmt.Sprintf("%X", addr)
}

func (w *CanBusWriter) WriteCommand(source int, destination int, fragmentation int, data string) {
	oldSequenceNumber := w.StateDao.GetInt(sequenceNumberStateKey)
	sequenceNumber := (oldSequenceNumber + 1) & 0x3
	w.StateDao.Set(sequenceNumberStateKey, sequenceNumber)
	log.Println("using sequence number: ", sequenceNumber)

	address := GenerateAddress(source, destination, fragmentation, sequenceNumber)
	log.Println("Generated address: ", address)
	w.Write(address, data)
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
