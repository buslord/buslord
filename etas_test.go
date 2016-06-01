package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestETAsHandler(t *testing.T) {
	config.Cache.Enabled = false

	reqStopID := "4342abc"

	tflArrivals := []TFLArrival{
		{
			ID:              "-36453",
			LineName:        "179",
			DestinationName: "Ilford",
			TimeToStation:   1266,
			ModeName:        "bus",
			TimeToLive:      "2016-05-30T18:36:29Z",
		},
		{
			ID:              "-2221",
			LineName:        "212",
			DestinationName: "St James Street",
			TimeToStation:   1136,
			ModeName:        "bus",
			TimeToLive:      "2016-05-30T18:34:19Z",
		},
	}

	handlerCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		// ensure the handler requests the right path
		assert.Equal(t, "/StopPoint/"+reqStopID+"/arrivals", r.URL.Path)
		// answer stub data
		enc := json.NewEncoder(w)
		err := enc.Encode(&tflArrivals)
		assert.Nil(t, err)
	}))
	defer ts.Close()

	// change the TFL base url to the mock server's url
	tflBaseURL = ts.URL

	req, err := http.NewRequest("GET", "http://buslord.com/etas?stop="+reqStopID, nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	etasHandler(w, req)

	assert.Equal(t, true, handlerCalled)

	assert.Equal(t, http.StatusOK, w.Code)

	dec := json.NewDecoder(w.Body)
	var etas ETAs
	err = dec.Decode(&etas)
	assert.Nil(t, err)

	assert.Equal(t, len(tflArrivals), len(etas))

	for i, tflA := range tflArrivals {
		assert.Equal(t, tflA.ID, etas[i].ID)
		assert.Equal(t, tflA.LineName, etas[i].LineName)
		assert.Equal(t, tflA.DestinationName, etas[i].DestinationName)
		assert.Equal(t, tflA.TimeToStation, etas[i].ETA)
		assert.Equal(t, tflA.ModeName, etas[i].ModeName)

		timeToLive, err := time.Parse(time.RFC3339, tflA.TimeToLive)
		assert.Nil(t, err)
		assert.Equal(t, timeToLive, etas[i].TimeToLive)
	}
}

func TestETAsCacheHandler(t *testing.T) {
	config.Cache.Enabled = true

	handlerCalled := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		in2seconds := time.Now().Add(2 * time.Second).Format(time.RFC3339)

		tflArrivals := []TFLArrival{
			{
				ID:              "-36453",
				LineName:        "179",
				DestinationName: "Ilford",
				TimeToStation:   1266,
				ModeName:        "bus",
				TimeToLive:      time.Now().Add(20 * time.Minute).Format(time.RFC3339),
			},
			{
				ID:              "-2221",
				LineName:        "212",
				DestinationName: "St James Street",
				TimeToStation:   1136,
				ModeName:        "bus",
				TimeToLive:      in2seconds,
			},
		}
		enc := json.NewEncoder(w)
		enc.Encode(&tflArrivals)
	}))
	defer ts.Close()

	tflBaseURL = ts.URL
	reqStopID := RandStringRunes(100)
	req, err := http.NewRequest("GET", "http://buslord.com/etas?stop="+reqStopID, nil)
	assert.Nil(t, err)
	w := httptest.NewRecorder()
	etasHandler(w, req)

	// first time the API should be called
	assert.Equal(t, true, handlerCalled)

	handlerCalled = false
	etasHandler(w, req)
	// second time, cache hit and API not called
	assert.Equal(t, false, handlerCalled)

	// wait a second for the cache to expire
	timer := time.NewTimer(time.Second)
	<-timer.C
	handlerCalled = false
	etasHandler(w, req)
	// third time, cache miss because it expired, API called
	assert.Equal(t, true, handlerCalled)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
