package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Stop struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	LatLng LatLng `json:"lat_lng"`
}

type Stops []Stop

func (stop *Stop) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(stop)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (stop *Stop) Decode(bs []byte) (err error) {
	err = gob.NewDecoder(bytes.NewBuffer(bs)).Decode(&stop)
	return
}

type TFLStopPoint struct {
	ID         string  `json:"id"`
	CommonName string  `json:"commonName"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lon"`
}

func FetchStops(client *http.Client, swLat, swLng, neLat, neLng float64) (stops Stops, err error) {
	// the query params we are going to sent TFL
	v := url.Values{}
	v.Add("app_id", config.TFL.AppID)
	v.Add("app_key", config.TFL.AppKey)
	v.Add("stopTypes", "NaptanPublicBusCoachTram")
	v.Add("includeChildren", "False")
	v.Add("returnLines", "False")
	v.Add("useStopPointHierarchy", "True")

	// google => Lng. tfl => Lon
	v.Add("swLat", strconv.FormatFloat(swLat, 'f', -1, 64))
	v.Add("swLon", strconv.FormatFloat(swLng, 'f', -1, 64))
	v.Add("neLat", strconv.FormatFloat(neLat, 'f', -1, 64))
	v.Add("neLon", strconv.FormatFloat(neLng, 'f', -1, 64))

	// https://api.tfl.gov.uk/StopPoint?appID=4b537b47&appKey=5173db496aaaf2f26a45dbfb587597d1&includeChildren=False&neLat=51.47450678007974&neLon=0.14333535461423708&returnLines=True&stopTypes=NaptanPublicBusCoachTram&swLat=51.47450678007974&swLon=0.14333535461423708&useStopPointHierarchy=True
	url := tflBaseURL + "/StopPoint?" + v.Encode()

	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(resp.Body)
	var tflStopPoints []TFLStopPoint
	err = decoder.Decode(&tflStopPoints)
	if err != nil {
		return
	}

	// translate TFL response to the response we want
	stops = make(Stops, 0, len(tflStopPoints))
	for _, sp := range tflStopPoints {
		stop := Stop{
			ID:     sp.ID,
			Name:   sp.CommonName,
			LatLng: LatLng{Lat: sp.Lat, Lng: sp.Lng}}
		stops = append(stops, stop)
	}
	return
}
