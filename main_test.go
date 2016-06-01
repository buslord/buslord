package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopsHandler(t *testing.T) {

	reqParams := url.Values{}
	reqParams.Set("swLat", "51.47450678007974")
	reqParams.Add("swLng", "0.14333535461423708")
	reqParams.Set("neLat", "51.49450678007974")
	reqParams.Add("neLng", "0.14533535461423708")

	tflStopPoints := []TFLStopPoint{
		{
			ID:         "1234567a",
			CommonName: "Praça Eufrásio Correia",
			Lat:        51.47450678007974,
			Lng:        0.14333535461423708,
		},
		{
			ID:         "ba567a",
			CommonName: "Catedral da Fé",
			Lat:        52.47450678007974,
			Lng:        0.18333535461423708,
		},
	}

	handlerCalled := false

	// setup a test server that mocks the tfl server
	//    we test the params the handler uses and that it returns what we expect
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// ensure the handler requests the right path
		assert.Equal(t, "/StopPoint", r.URL.Path)

		// be sure all bound params are forwarded
		for _, key := range []string{"swLat", "swLng", "neLat", "neLng"} {
			tflKey := strings.Replace(key, "Lng", "Lon", -1) // google => Lng. tfl => Lon
			assert.Equal(t, reqParams.Get(key), r.FormValue(tflKey))
		}

		// we are looking for bus stops
		assert.Equal(t, "NaptanPublicBusCoachTram", r.FormValue("stopTypes"))

		// answer stub data
		enc := json.NewEncoder(w)
		err := enc.Encode(&tflStopPoints)
		assert.Nil(t, err)
	}))
	defer ts.Close()

	// change the TFL base url to the mock server's url
	tflBaseURL = ts.URL

	req, err := http.NewRequest("GET", "http://buslord.com/stops?"+reqParams.Encode(), nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	stopsHandler(w, req)

	assert.Equal(t, true, handlerCalled)

	assert.Equal(t, http.StatusOK, w.Code)

	dec := json.NewDecoder(w.Body)
	var stops []Stop
	err = dec.Decode(&stops)
	assert.Nil(t, err)
	assert.Equal(t, len(tflStopPoints), len(stops))

	for i, tflSP := range tflStopPoints {
		assert.Equal(t, tflSP.ID, stops[i].ID)
		assert.Equal(t, tflSP.CommonName, stops[i].Name)
		assert.Equal(t, tflSP.Lat, stops[i].LatLng.Lat)
		assert.Equal(t, tflSP.Lng, stops[i].LatLng.Lng)
	}

}
