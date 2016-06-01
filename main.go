package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gorilla/mux"
)

var (
	client     *http.Client
	mc         *memcache.Client
	tflBaseURL = "https://api.tfl.gov.uk"
)

func init() {
	client = &http.Client{
		Timeout: time.Second * 5,
	}
	mc = memcache.New(config.Cache.MemcacheServers...)
}

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
	ID     string `json:"id"`
	Name   string `json:"name"`
	LatLng LatLng `json:"lat_lng"`
}

type TFLStopPoint struct {
	ID         string  `json:"id"`
	CommonName string  `json:"commonName"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lon"`
}

func stopsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")

	// validate query params and put them in "vals"
	vals := map[string]string{}
	for _, key := range []string{"swLat", "swLng", "neLat", "neLng"} {
		vals[key] = r.FormValue(key)
		if vals[key] == "" {
			errorHandler(w, http.StatusBadRequest, fmt.Errorf("The '%s' query param is mandatory.", key))
			return
		}

	}

	// the query params we are going to sent TFL
	v := url.Values{}
	v.Add("app_id", config.TFL.AppID)
	v.Add("app_key", config.TFL.AppKey)
	v.Add("stopTypes", "NaptanPublicBusCoachTram")
	v.Add("includeChildren", "False")
	v.Add("returnLines", "False")
	v.Add("useStopPointHierarchy", "True")

	// forward the bound params
	for key, val := range vals {
		key = strings.Replace(key, "Lng", "Lon", -1) // google => Lng. tfl => Lon
		v.Add(key, val)
	}

	// https://api.tfl.gov.uk/StopPoint?appID=4b537b47&appKey=5173db496aaaf2f26a45dbfb587597d1&includeChildren=False&neLat=51.47450678007974&neLon=0.14333535461423708&returnLines=True&stopTypes=NaptanPublicBusCoachTram&swLat=51.47450678007974&swLon=0.14333535461423708&useStopPointHierarchy=True
	url := tflBaseURL + "/StopPoint?" + v.Encode()

	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	var tflStopPoints []TFLStopPoint
	err = decoder.Decode(&tflStopPoints)
	if err != nil {
		errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	// translate TFL response to the response we want
	stops := make([]Stop, 0, len(tflStopPoints))
	for _, sp := range tflStopPoints {
		stop := Stop{
			ID:     sp.ID,
			Name:   sp.CommonName,
			LatLng: LatLng{Lat: sp.Lat, Lng: sp.Lng}}
		stops = append(stops, stop)
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
