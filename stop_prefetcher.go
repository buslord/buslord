package main

import (
	"log"
	"strconv"
	"time"

	"github.com/soveran/redisurl"
)

func stopPrefetcher() {
	c := time.Tick(24 * time.Hour)
	for range c {
		prefetchStops()
	}
}

func prefetchStops() {
	conn, _ := redisurl.Connect()
	defer conn.Close()

	v, err := conn.Do("GET", "prefetch_run")
	if err != nil {
		log.Fatal("Prefetch err: " + err.Error())
	}
	previousRun, err := strconv.ParseInt(v.(string), 10, 64)
	if err != nil {
		log.Fatal("Prefetch err: " + err.Error())
	}
	run := previousRun + 1

	previousKey := "geostops_" + strconv.FormatInt(run, 10)
	key := "geostops_" + strconv.FormatInt(run, 10)

	swStartLat := 51.452206
	swStartLng := -0.263672
	neStartLat := 51.616222
	neStartLng := 0.052185

	// total of requests will be latIterations*lngIterations
	latIterations := 10
	lngIterations := 10

	latStep := (neStartLat - swStartLat) / float64(latIterations)
	lngStep := (neStartLng - swStartLng) / float64(lngIterations)

	for i := 0; i < latIterations; i++ {
		for j := 0; j < lngIterations; j++ {
			swLat := swStartLat + float64(i)*latStep
			swLng := swStartLng + float64(j)*lngStep
			neLat := neStartLat + float64(i)*latStep
			neLng := neStartLng + float64(j)*lngStep

			stops, err := FetchStops(swLat, swLng, neLat, neLng)
			if err != nil {
				log.Println("Prefetch err: " + err.Error())
				continue
			}

			for _, stop := range stops {

				sLng := strconv.FormatFloat(stop.LatLng.Lng, 'f', -1, 64)
				sLat := strconv.FormatFloat(stop.LatLng.Lat, 'f', -1, 64)
				// GEOADD key longitude latitude member
				conn.Do("GEOADD", key, sLng, sLat, "stop_"+stop.ID)

				bs, err := stop.Encode()
				if err != nil {
					log.Println("Prefetch err: " + err.Error())
					continue
				}
				conn.Do("SET", "stop_"+stop.ID, string(bs))
			}

			// slowly
			timer := time.NewTimer(5 * time.Second)
			<-timer.C
		}
	}

	// this is ready. incoming requests should use it
	_, err = conn.Do("SET", "prefetch_run", strconv.FormatInt(run, 10))
	if err != nil {
		log.Println("Prefetch err: " + err.Error())
	}

	// delete the old geo key
	_, err = conn.Do("DEL", previousKey)
	if err != nil {
		log.Println("Prefetch err: " + err.Error())
	}
}
