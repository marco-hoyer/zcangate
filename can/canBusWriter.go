package can

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"strings"
)

type CanBusWriter struct {
	Serial *serial.Port
}

func (w *CanBusWriter) Write(id string, data string) {
	length := len(data) / 2
	if length > 8 {
		fmt.Println("Length", length)
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

		payload := fmt.Sprintf("T%s%x%02x%s\r", id, len(chunk)/2+1, n|0x88, chunk)
		w.writeAndWait(payload)
	} else {
		//w.Serial.Write(f.toBytes())
		fmt.Println("not implemented yet")
	}
}

func (w *CanBusWriter) writeAndWait(payload string) {
	fmt.Println("message string: ", payload)
	fmt.Println("message ascii: ", []byte(payload))

	w.Serial.Flush()
	w.Serial.Write([]byte(payload))

	response := make([]byte, 128)
	retries := 50
	for {
		w.Serial.Read(response)
		//log.Println("raw response: ", response)
		if strings.Contains(string(response), "Z\r") {
			log.Printf("COMMAND FINISHED SUCCESSFULLY")
			break
		}

		retries--
		if retries < 1 {
			panic("timeout receiving command ACK")

		}
	}

}
