package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

type ETA struct {
	ID              string    `json:"id"`
	LineName        string    `json:"line_name"`
	DestinationName string    `json:"destination_name"`
	ETA             int64     `json:"eta"`
	ModeName        string    `json:"mode_name"`
	TimeToLive      time.Time `json:"time_to_live"`
}

type TFLArrival struct {
	ID              string `json:"id"`
	LineName        string `json:"lineName"`
	DestinationName string `json:"destinationName"`
	TimeToStation   int64  `json:"timeToStation"`
	ModeName        string `json:"modeName"`
	TimeToLive      string `json:"timeToLive"`
}

type ETAs []ETA

func (etas *ETAs) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(etas)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

func (etas *ETAs) Decode(bs []byte) (err error) {
	err = gob.NewDecoder(bytes.NewBuffer(bs)).Decode(&etas)
	return
}

// implements sort.Interface
func (a ETAs) Len() int           { return len(a) }
func (a ETAs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ETAs) Less(i, j int) bool { return a[i].ETA < a[j].ETA }

func GetETAs(stopID string) (etas ETAs, err error) {

	if config.Cache.ETAsEnabled == false {
		return FetchEtas(stopID)
	}

	key := "etas_" + stopID

	it, err := mc.Get(key)
	if err == memcache.ErrCacheMiss {
		log.Println("miss cache")
		etas, err = FetchEtas(stopID)
		if err != nil {
			return
		}
		var bs []byte
		bs, err = etas.Encode()
		if err != nil {
			return
		}

		// take the smallest TimeToLive of the predictions
		now := time.Now()
		expiration := now.Add(2 * time.Minute)
		for _, eta := range etas {
			if eta.TimeToLive.Before(now) {
				// discard passed times
				continue
			}
			if eta.TimeToLive.Before(expiration) {
				expiration = eta.TimeToLive
			}
		}
		expirationSecs := int32(expiration.Sub(now).Seconds())

		mc.Set(&memcache.Item{Key: key, Value: bs, Expiration: expirationSecs})
	} else if err != nil {
		log.Println("cache errrr")
		return
	} else {
		log.Println("cache hit")
		etas = ETAs{}
		err = etas.Decode(it.Value)
		if err != nil {
			return
		}
	}
	return
}

func FetchEtas(stopID string) (ETAs, error) {
	// https://api.tfl.gov.uk/StopPoint/490005183E/arrivals
	v := url.Values{}
	v.Add("app_id", config.TFL.AppID)
	v.Add("app_key", config.TFL.AppKey)

	url := tflBaseURL + "/StopPoint/" + stopID + "/arrivals?" + v.Encode()
	log.Println(url)

	req, err := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	var tflArrivals []TFLArrival
	err = decoder.Decode(&tflArrivals)
	if err != nil {
		return nil, err
	}

	// translate TFL response to the response we want
	etas := make(ETAs, 0, len(tflArrivals))
	for _, a := range tflArrivals {
		var timeToLive time.Time
		timeToLive, err = time.Parse(time.RFC3339, a.TimeToLive)
		if err != nil {
			return nil, err
		}
		eta := ETA{
			ID:              a.ID,
			LineName:        a.LineName,
			DestinationName: a.DestinationName,
			ETA:             a.TimeToStation,
			ModeName:        a.ModeName,
			TimeToLive:      timeToLive,
		}
		etas = append(etas, eta)
	}

	// sort
	sort.Sort(etas)

	return etas, nil
}
