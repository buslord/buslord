package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/soveran/redisurl"
)

var (
	client = &http.Client{
		Timeout: time.Second * 5,
	}

	mc         *memcache.Client
	tflBaseURL = "https://api.tfl.gov.uk"
	conn       redis.Conn
)

var runPrefetcher = flag.Bool("prefetcher", false, "run the prefetcher now")

func init() {
	mc = memcache.New(config.Cache.MemcacheServers...)
}

func main() {
	flag.Parse()

	var err error
	conn, err = redisurl.Connect()
	if err != nil {
		log.Fatal(err)
	}

	// just run the prefetcher and quit
	if *runPrefetcher == true {
		log.Println("Running prefetcher once.")

		prefetchStops()

		return
	}

	// run periodically the stop prefetcher
	go stopPrefetcher()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/stops", stopsHandler).Methods("GET")
	r.HandleFunc("/etas", etasHandler).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func stopsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	// validate query params
	vals := map[string]float64{}
	for _, key := range []string{"swLat", "swLng", "neLat", "neLng"} {
		if r.FormValue(key) == "" {
			errorHandler(w, http.StatusBadRequest, fmt.Errorf("The '%s' query param is mandatory.", key))
			return
		}
		f, err := strconv.ParseFloat(r.FormValue(key), 64)
		if err != nil {
			errorHandler(w, http.StatusBadRequest, fmt.Errorf("The '%s' query param should be a float.", key))
			return
		}
		vals[key] = f
	}

	stops, err := GetStops(
		vals["swLat"],
		vals["swLng"],
		vals["neLat"],
		vals["neLng"],
	)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(&stops); err != nil {
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}
}

func etasHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	stopID := r.FormValue("stop")
	if stopID == "" {
		errorHandler(w, http.StatusBadRequest, fmt.Errorf("The 'stop' query param is mandatory."))
		return
	}

	etas, err := GetETAs(stopID)
	if err != nil {
		// TODO 500 is not always a good answer. 404 is the right thing in cases
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(etas); err != nil {
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}
}

func errorHandler(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	data := struct {
		Error   bool
		Status  int
		Message string
	}{
		true,
		status,
		err.Error(),
	}
	// respond json
	bytes, _ := json.Marshal(data)
	json := string(bytes[:])
	fmt.Fprint(w, json)

	log.Println("Err: " + json)
}
