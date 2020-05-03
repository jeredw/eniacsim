package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func webGui() {
	http.HandleFunc("/", serveFile("webgui/webgui.html"))
	http.HandleFunc("/webgui.js", serveFile("webgui/webgui.js"))
	http.HandleFunc("/webgui.css", serveFile("webgui/webgui.css"))
	http.HandleFunc("/eniac.svg", serveFile("webgui/eniac.svg"))
	http.HandleFunc("/neons.json", serveFile("webgui/neons.json"))
	http.HandleFunc("/panels.json", serveFile("webgui/panels.json"))
	http.HandleFunc("/switches.json", serveFile("webgui/switches.json"))
	http.HandleFunc("/events", streamEvents)
	http.HandleFunc("/command", postCommand)
	http.ListenAndServe(":8000", nil)
}

func serveFile(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, path)
	}
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
		status["initiate"], _ = json.Marshal(initiate.Stat())
		status["cycling"], _ = json.Marshal(cycle.Stat())
		status["mp"] = mp.State()
		ftState := []json.RawMessage{ft[0].State(), ft[1].State(), ft[2].State()}
		status["ft"], _ = json.Marshal(ftState)
		accState := [20]json.RawMessage{}
		for i := range accumulator {
			accState[i] = accumulator[i].State()
		}
		status["acc"], _ = json.Marshal(accState)
		status["div"] = divsr.State()
		status["mult"] = multiplier.State()
		status["constant"], _ = json.Marshal(constant.Stat())
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
