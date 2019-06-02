package api

import (
	"github.com/gorilla/mux"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/tarm/serial"
	"log"
	"net/http"
)

type WebServer struct {
	SerialInterface *serial.Port
	CanBusWriter    *can.CanBusWriter
}

func commandsIndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("c1, c2, c3"))
}

func (s *WebServer) commandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	command := vars["command"]
	log.Println("Received command: ", command)

	switch command {
	case "auto":
		log.Println("Putting unit into auto mode")
		s.CanBusWriter.WriteCommand(0x11, 0x1, 0x0, "85150801")
	case "manual":
		log.Println("Putting unit into manual mode")
		s.CanBusWriter.WriteCommand(0x11, 0x1, 0x1, "84150801000000000100000001")
	case "3":
		log.Println("Setting ventilation level to 3")
		s.CanBusWriter.WriteCommand(0x11, 0x1, 0x1, "8415010100000000FFFFFFFF03")
	case "2":
		log.Println("Setting ventilation level to 0")
		s.CanBusWriter.WriteCommand(0x11, 0x1, 0x1, "8415010100000000FFFFFFFF02")
	case "1":
		log.Println("Setting ventilation level to 0")
		s.CanBusWriter.WriteCommand(0x11, 0x1, 0x1, "8415010100000000FFFFFFFF01")
	case "0":
		log.Println("Setting ventilation level to 0")
		s.CanBusWriter.WriteCommand(0x11, 0x1, 0x1, "8415010100000000FFFFFFFF00")
	}

	//s.CanBusWriter.Write("1F015074", "84150101000000000100000003")
	//s.CanBusWriter.Write("1F035074", "84150101000000000100000003")
	//s.CanBusWriter.Write("1F055074", "84150101000000000100000003")
	//s.CanBusWriter.Write("1F075074", "84150101000000000100000003")

	//s.CanBusWriter.Write("1F011074", "85150801")
	//s.CanBusWriter.Write("1F031074", "85150801")
	//s.CanBusWriter.Write("1F051074", "85150801")
	//s.CanBusWriter.Write("1F071074", "85150801")

	w.Write([]byte("OK"))
}

func (s *WebServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/commands", commandsIndexHandler)
	router.HandleFunc("/commands/{command}", s.commandHandler)
	http.ListenAndServe(":8080", router)
}
