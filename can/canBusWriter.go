package can

import (
	"fmt"
	"github.com/tarm/serial"
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
			payload := fmt.Sprintf("T%s%x%02x%s", id, len(chunk)/2+1, i, chunk)
			fmt.Println("message", payload)
			w.Serial.Write([]byte(payload))
			n += 1
		}

		restLength := (length - n*7) * 2
		chunk := data[n*14 : n*14+restLength]

		payload := fmt.Sprintf("T%s%x%02x%s", id, len(chunk)/2+1, n|0x88, chunk)
		fmt.Println("message", payload)
		w.Serial.Write([]byte(payload))
	} else {
		//w.Serial.Write(f.toBytes())
		fmt.Println("not implemented yet")
	}
}
