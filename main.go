package main

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"strings"
	"time"
)

func readSerial(s *serial.Port) <-chan string {
	out := make(chan string)
	go func() {
		buf := make([]byte, 128)
		var readCount int
		for {
			n, err := s.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			readCount++
			//log.Printf("Read %v %v bytes: % 02x %s", readCount, n, buf[:n], buf[:n])
			//log.Printf("% 02x", buf[:n])

			parts := strings.Split(string(buf[:n]), "\r")
			length := len(parts)
			fmt.Println("LENGTH: ", length)
			fmt.Println(parts)

			//for index <= maxIndex {
			//	fmt.Println("index:", index)
			//
			//	if index % 2 == 1 {
			//		if index < maxIndex {
			//
			//		}
			//	} else {
			//		println("even index ignored")
			//	}
			//
			//	if index == maxIndex {
			//		println("index == length")
			//		println("part found at the end", parts[index])
			//		break
			//	} else if index+2 > length {
			//		println("index+2 > length")
			//		println("part found near the end", parts[index])
			//		break
			//	} else {
			//		println("part found in between", parts[index])
			//		out <- parts[index]
			//	}
			//	index += 1
			//	fmt.Println("index increased:", index)
			//}
		}
	}()
	return out
}

func process(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for b := range in {
			//log.Println(b)
			out <- b
		}
	}()
	return out
}

func logLines(in <-chan string) {
	go func() {
		for b := range in {
			//toType(b)
			fmt.Println(b)
		}
	}()
}

func main() {
	c := &serial.Config{Name: "/tmp/ttyACM0", Baud: 115200, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})
	defer close(done)

	lines := readSerial(s)
	messages := process(lines)
	logLines(messages)
	time.Sleep(10 * time.Second)
}

//func main() {
//	c := &serial.Config{Name: "/tmp/ttyACM0", Baud: 115200, ReadTimeout: time.Second * 5}
//	s, err := serial.OpenPort(c)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println("before")
//	ch := make(chan int, 1)
//	go func() {
//		log.Println("inside")
//		buf := make([]byte, 128)
//		var readCount int
//		for {
//			n, err := s.Read(buf)
//			if err != nil {
//				log.Fatal(err)
//			}
//			readCount++
//			log.Printf("Read %v %v bytes: % 02x %s", readCount, n, buf[:n], buf[:n])
//			select {
//			case <-ch:
//				ch <- readCount
//				close(ch)
//			default:
//			}
//		}
//	}()
//
//	fmt.Println(<-ch)
//	log.Println("end")
//}
