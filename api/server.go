package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/marco-hoyer/zcangate/can"
	"github.com/marco-hoyer/zcangate/dao"
	"log"
	"net/http"
)

type WebServer struct {
	CanBusWriter *can.BusWriter
	State        *dao.StateDao
	Addr         string
	AuthToken    string
}

type MeasurementResponse struct {
	Value float64
}

func (s *WebServer) requireBearer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.AuthToken != "" && r.Header.Get("Authorization") != "Bearer "+s.AuthToken {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("unauthorized"))
			return
		}
		next(w, r)
	}
}

func commandsIndexHandler(w http.ResponseWriter, _ *http.Request) {
	commands := make([]string, 0, len(can.Commands))
	for k := range can.Commands {
		commands = append(commands, k)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(commands)
}

func (s *WebServer) commandHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commandQueryParam := vars["command"]
	command, found := can.Commands[commandQueryParam]

	if found {
		log.Println("executing command:", commandQueryParam)
		s.CanBusWriter.WriteCommand(command)
		_, _ = w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("command not found"))
	}
}

func (s *WebServer) measurementsIndexHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(s.State.GetJson()))
}

func (s *WebServer) measurementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	measurementName := vars["measurement"]
	value := s.State.GetFloat64(measurementName)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(MeasurementResponse{value})
}

func (s *WebServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/measurements", s.measurementsIndexHandler)
	router.HandleFunc("/measurements/{measurement}", s.measurementHandler)
	router.HandleFunc("/commands", commandsIndexHandler).Methods(http.MethodGet)
	router.HandleFunc("/commands/{command}", s.requireBearer(s.commandHandler)).Methods(http.MethodPost)
	_ = http.ListenAndServe(s.Addr, router)
}
