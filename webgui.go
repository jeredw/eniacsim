package main

import (
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

	for {
		fmt.Fprintf(w, "data: %s\n\n", accumulator[0].Stat())
		time.Sleep(100 * time.Millisecond)
		flusher.Flush()
	}
}

func pushButton(w http.ResponseWriter, req *http.Request) {
}
