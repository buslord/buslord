package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopsHandler(t *testing.T) {
	v := url.Values{}
	v.Set("lat", "51.47450678007974")
	v.Add("lng", "0.14333535461423708")
	req, err := http.NewRequest("GET", "http://buslord.com/stops?"+v.Encode(), nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	stopsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	dec := json.NewDecoder(w.Body)
	var stops []Stop
	err = dec.Decode(&stops)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(stops))
}

func TestEtasHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://buslord.com/etas?stop=334", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	etasHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	dec := json.NewDecoder(w.Body)
	var etas []ETA
	err = dec.Decode(&etas)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(etas))
}
