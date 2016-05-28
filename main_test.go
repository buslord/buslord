package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://buslord.com/stops", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	stopsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	dec := json.NewDecoder(w.Body)
	var p struct {
		Todo string `json:"todo"`
	}
	err = dec.Decode(&p)
	assert.Nil(t, err)
}

func TestEtasHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "http://buslord.com/etas", nil)
	assert.Nil(t, err)

	w := httptest.NewRecorder()
	etasHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	dec := json.NewDecoder(w.Body)
	var p struct {
		Todo string `json:"todo"`
	}
	err = dec.Decode(&p)
	assert.Nil(t, err)
}
