package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/marco-hoyer/zcangate/dao"
	"github.com/tarm/serial"
	"log"
	"net/http"
)

type WebServer struct {
	SerialInterface *serial.Port
	CanBusWriter    *can.BusWriter
	State           *dao.StateDao
}

type MeasurementResponse struct {
	Value float64
}

func commandsIndexHandler(w http.ResponseWriter, _ *http.Request) {
	commands := make([]string, len(can.Commands))

	i := 0
	for k := range can.Commands {
		commands[i] = k
		i++
	}

	enc := json.NewEncoder(w)
	_ = enc.Encode(commands)
}

func (s *WebServer) commandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commandQueryParam := vars["command"]
	log.Println("command query param ", commandQueryParam)
	command, found := can.Commands[commandQueryParam]

	if found {
		log.Println("executing command: ", commandQueryParam)
		s.CanBusWriter.WriteCommand(command)
		_, _ = w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("ERROR, command not found"))
	}
}

func (s *WebServer) measurementsIndexHandler(w http.ResponseWriter, _ *http.Request) {
	data := s.State.GetJson()
	_, _ = w.Write([]byte(data))
	w.Header().Set("Content-Type", "application/json")
}

func (s *WebServer) measurementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	measurementName := vars["measurement"]
	log.Println("requested measurement ", measurementName)

	value := s.State.GetFloat64(measurementName)
	log.Println("float32 value from cache:", value)
	data := MeasurementResponse{value}
	log.Println(data)

	_ = json.NewEncoder(w).Encode(data)
	w.Header().Set("Content-Type", "application/json")
}

func (s *WebServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/measurements", s.measurementsIndexHandler)
	router.HandleFunc("/measurements/{measurement}", s.measurementHandler)
	router.HandleFunc("/commands", commandsIndexHandler)
	router.HandleFunc("/commands/{command}", s.commandHandler)
	_ = http.ListenAndServe(":8080", router)
}
