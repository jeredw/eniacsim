package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

func webGui(dir string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", streamEvents)
	mux.HandleFunc("/command", postCommand)
	mux.Handle("/", http.FileServer(http.Dir(dir)))
	err := http.ListenAndServe(":8000", mux)
	log.Fatal(err)
}

func streamEvents(w http.ResponseWriter, req *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	status := make(map[string]json.RawMessage)
	for {
		status["initiate"], _ = json.Marshal(u.Initiate.Stat())
		status["cycling"], _ = json.Marshal(cycle.Stat())
		status["mp"] = u.Mp.State()
		ftState := []json.RawMessage{u.Ft[0].State(), u.Ft[1].State(), u.Ft[2].State()}
		status["ft"], _ = json.Marshal(ftState)
		accState := [20]json.RawMessage{}
		for i := range u.Accumulator {
			accState[i] = u.Accumulator[i].State()
		}
		status["acc"], _ = json.Marshal(accState)
		status["div"] = u.Divsr.State()
		status["mult"] = u.Multiplier.State()
		status["constant"], _ = json.Marshal(u.Constant.Stat())
		message, _ := json.Marshal(status)
		fmt.Fprintf(w, "data: %s\n\n", message)
		time.Sleep(100 * time.Millisecond)
		flusher.Flush()
	}
}

type commandRequest struct {
	Commands []string `json:"commands"`
}

type commandResponse struct {
	Outputs []string `json:"outputs"`
}

func postCommand(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Invalid HTTP method")
		return
	}

	reqData := commandRequest{}
	err := json.NewDecoder(req.Body).Decode(&reqData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	respData := commandResponse{Outputs: make([]string, 0, len(reqData.Commands))}
	for i := range reqData.Commands {
		var buf bytes.Buffer
		doCommand(&buf, reqData.Commands[i])
		respData.Outputs = append(respData.Outputs, buf.String())
	}

	output, _ := json.Marshal(respData)
	w.Write(output)
}
