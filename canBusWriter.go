package main

import (
	"fmt"
	"github.com/tarm/serial"
)

type CanBusWriter struct {
	serial *serial.Port
}

func (w *CanBusWriter) write(id string, data string) {
	fmt.Println("data", data)
	length := len(data) / 2
	fmt.Println("len", length)
	if length > 8 {
		numberOfDatagrams := length / 7
		rest := length % 7
		if rest > 0 {
			numberOfDatagrams -= 1
		}

		n := 0
		for i := 0; i <= numberOfDatagrams; i++ {
			chunk := data[i*14 : i*14+14]
			fmt.Println("i", i)
			fmt.Println("chunk", chunk)
			// foobaa is going on here
			payload := fmt.Sprintf("T%s%x%02x%s", id, len(chunk)/2+1, i, chunk)
			fmt.Println("message", payload)
			w.serial.Write([]byte(payload))
			n += 1
		}

		fmt.Println("n", n)
		restLength := (length - n*7) * 2
		fmt.Println("rest len", restLength)

		chunk := data[n*14 : n*14+restLength]
		fmt.Println("chunk", chunk)
		payload := fmt.Sprintf("T%s%x%02x%s", id, len(chunk)/2+1, n|0x88, chunk)
		fmt.Println("message", payload)
		w.serial.Write([]byte(payload))

	} else {
		//w.serial.Write(f.toBytes())
		fmt.Println()
	}
}
