package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"log"
	"github.com/tarm/serial"
	"github.com/marco-hoyer/zcangate/can"
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

	s.CanBusWriter.Write("1F015051", "8415010100000000FFFFFFFF03")
	s.CanBusWriter.Write("1F035051", "8415010100000000FFFFFFFF03")
	s.CanBusWriter.Write("1F055051", "8415010100000000FFFFFFFF03")
	s.CanBusWriter.Write("1F075051", "8415010100000000FFFFFFFF03")

	w.Write([]byte("OK"))
}

func (s *WebServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/commands", commandsIndexHandler)
	router.HandleFunc("/commands/{command}", s.commandHandler)
	http.ListenAndServe(":8080", router)
}
