package main

import (
	"fmt"
	"github.com/tarm/serial"
)

type CanBusWriter struct {
	serial *serial.Port
}

func (w *CanBusWriter) write(id string, data string) {
	length := len(data) / 2
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
			fmt.Println("message", payload)
			w.serial.Write([]byte(payload))
			n += 1
		}

		restLength := (length - n*7) * 2
		chunk := data[n*14 : n*14+restLength]
		payload := fmt.Sprintf("T%s%x%02x%s\r", id, len(chunk)/2+1, n|0x88, chunk)

		fmt.Println("message", payload)
		w.serial.Write([]byte(payload))
	} else {
		//w.serial.Write(f.toBytes())
		fmt.Println("not implemented yet")
	}
}
