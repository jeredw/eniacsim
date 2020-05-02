package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func webGui() {
	http.HandleFunc("/", serveFile("webgui.html"))
	http.HandleFunc("/webgui.js", serveFile("webgui.js"))
	http.HandleFunc("/webgui.css", serveFile("webgui.css"))
	http.HandleFunc("/eniac.svg", serveFile("eniac.svg"))
	http.HandleFunc("/neons.json", serveFile("neons.json"))
	http.HandleFunc("/panels.json", serveFile("panels.json"))
	http.HandleFunc("/switches.json", serveFile("switches.json"))
	http.HandleFunc("/events", streamEvents)
	http.HandleFunc("/button", pushButton)
	http.ListenAndServe(":8000", nil)
}

func serveFile(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		s := string(data)
		http.ServeContent(w, req, path, time.Now(), strings.NewReader(s))
	}
}

func cacheAndServeFile(path string) func(http.ResponseWriter, *http.Request) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	s := string(data)
	return func(w http.ResponseWriter, req *http.Request) {
		http.ServeContent(w, req, path, time.Now(), strings.NewReader(s))
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

func pushButton(w http.ResponseWriter, req *http.Request) {
}
