package api

import (
	"github.com/gorilla/mux"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/tarm/serial"
	"log"
	"net/http"
	"github.com/marco-hoyer/zcangate/common"
	"encoding/json"
)

type WebServer struct {
	SerialInterface *serial.Port
	CanBusWriter    *can.CanBusWriter
}

func commandsIndexHandler(w http.ResponseWriter, r *http.Request) {
	commands := make([]string, len(common.Commands))

	i := 0
	for k := range common.Commands {
		commands[i] = k
		i++
	}

	enc := json.NewEncoder(w)
	enc.Encode(commands)
}

func (s *WebServer) commandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commandQueryParam := vars["command"]
	log.Println("command query param ", commandQueryParam)
	command, found := common.Commands[commandQueryParam]

	if found {
		log.Println("executing command: ", commandQueryParam)
		s.CanBusWriter.WriteCommand(11, 1, command.Fragmentation, command.Code)
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("ERROR, command not found"))
	}
}

func (s *WebServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/commands", commandsIndexHandler)
	router.HandleFunc("/commands/{command}", s.commandHandler)
	http.ListenAndServe(":8080", router)
}
