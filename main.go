package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	r := mux.NewRouter()
	r.HandleFunc("/stops", stopsHandler).Methods("GET")
	r.HandleFunc("/etas", etasHandler).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Stop struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	LatLng LatLng `json:"lat_lng"`
}

type ETA struct {
	ID      int64  `json:"id"`
	BusName string `json:"bus_name"`
	ETA     int64  `json:"eta"`
}

func stopsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	lat, err := strconv.ParseFloat(r.FormValue("lat"), 64)
	if err != nil {
		errorHandler(w, http.StatusBadRequest, fmt.Errorf("The 'lat' query param is mandatory."))
		return
	}
	lng, err := strconv.ParseFloat(r.FormValue("lng"), 64)
	if err != nil {
		errorHandler(w, http.StatusBadRequest, fmt.Errorf("The 'lng' query param is mandatory."))
		return
	}

	stops := []Stop{
		{ID: 7, Name: "Alpha", LatLng: LatLng{Lat: lat + 0.005, Lng: lng + 0.005}},
		{ID: 8, Name: "Beta", LatLng: LatLng{Lat: lat - 0.007, Lng: lng - 0.002}},
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(&stops); err != nil {
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}
}

func etasHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	stop, err := strconv.ParseInt(r.FormValue("stop"), 10, 64)
	if err != nil {
		errorHandler(w, http.StatusBadRequest, fmt.Errorf("The 'stop' query param is mandatory."))
		return
	}
	log.Printf("got stop: %d", stop)

	etas := []ETA{
		{ID: 98, BusName: "Jardim Social", ETA: 63},
		{ID: 74, BusName: "Vila Sandra", ETA: 376},
		{ID: 99, BusName: "Jardim Esplanada", ETA: 476},
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(&etas); err != nil {
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
