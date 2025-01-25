package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	addr, ok := os.LookupEnv("ADDR")
	if !ok {
		a := ":8080"
		log.Printf("no environment variable ADDR, default=%s\n", a)
		addr = a
	}
	log.Printf("ADDR is %v\n", addr)

	st := os.Getenv("SLEEP_TIME")
	sleepTime, err := time.ParseDuration(st)
	if err != nil {
		log.Println("no environment variable SLEEP_TIME, default=10s")
		sleepTime = time.Second * 10
	}
	log.Printf("SLEEP_TIME is %v sec\n", sleepTime)

	rf := os.Getenv("RANDOM_FAIL")
	var fail bool
	if rf == "true" {
		fail = true
	}
	log.Printf("RANDOM_FAIL is %v\n", fail)

	m := http.NewServeMux()

	m.HandleFunc("GET /config", func(w http.ResponseWriter, r *http.Request) {
		config := struct {
			Name     string         `json:"name"`
			Desc     string         `json:"description"`
			Props    map[string]any `json:"properties"`
			Required []string
		}{
			Name: "get_weather",
			Desc: "Get weather at the given location",
			Props: map[string]any{
				"location": map[string]any{
					"type": "string",
				},
			},
			Required: []string{"location"},
		}

		bb, err := json.Marshal(config)
		if err != nil {
			http.Error(w, "marshal response", http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bb)
		if err != nil {
			log.Println("Error: write response")
		}
	})

	m.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil || r.ContentLength == 0 {
			http.Error(w, "body empty", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		bb, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		location := struct {
			Loc string `json:"location"`
		}{}

		err = json.Unmarshal(bb, &location)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		log.Printf("received request for location=%s", location.Loc)

		fmt.Fprintf(w, "30 degree in %s", location.Loc)
		log.Println("finished processing")
	})

	log.Printf("start mock weatherTool %s ...\n", addr)
	err = http.ListenAndServe(addr, m)
	log.Fatal(err)
}
