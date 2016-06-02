package main

import (
	"log"
	"net/http"
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

	cli := &http.Client{}

	v, err := conn.Do("GET", "prefetch_run")
	if err != nil {
		log.Fatal("Prefetch err: " + err.Error())
	}
	previousRun := int64(0)
	if v != nil {
		previousRun, err = strconv.ParseInt(string(v.([]byte)), 10, 64)
		if err != nil {
			log.Fatal("Prefetch err: " + err.Error())
		}
	}
	run := previousRun + 1

	// previousKey := "geostops_" + strconv.FormatInt(previousRun, 10)
	key := "geostops_" + strconv.FormatInt(run, 10)

	log.Println("Prefetcher: using key: " + key)

	// the region we'll iterate
	swLat := 51.452206
	swLng := -0.263672
	neLat := 51.616222
	neLng := 0.052185

	/*
	   (51.599820, 0.336456) ne: (51.616222, 0.368042)
	   (51.452206, 0.052185) ne: (51.468608, 0.083771)
	*/

	// total of requests will be latIterations*lngIterations
	latIterations := 10
	lngIterations := 10

	latStep := (neLat - swLat) / float64(latIterations)
	lngStep := (neLng - swLng) / float64(lngIterations)

	log.Printf("Prefetcher: latStep=%f lngStep=%f", latStep, lngStep)

	// we start iterating from nw
	nwLat := swLat
	nwLng := neLng

	// we go from north to south
	for i := 0; i < latIterations; i++ {
		// and west to east
		for j := 0; j < lngIterations; j++ {
			// we iterate the whole reagion with a window
			swWinLat := nwLat - float64(i+1)*latStep
			swWinLng := nwLng + float64(j)*lngStep
			neWinLat := nwLat - float64(i)*latStep
			neWinLng := nwLng + float64(j+1)*lngStep

			log.Printf("Prefetcher: fetching sw: (%f, %f) ne: (%f, %f)", swWinLat, swWinLng, neWinLat, neWinLng)

			stops, err := FetchStops(cli, swWinLat, swWinLng, neWinLat, neWinLng)
			if err != nil {
				log.Println("Prefetcher: err: " + err.Error())
				continue
			}

			log.Printf("Prefetcher: got %d stops", len(stops))
			for _, stop := range stops {
				sLng := strconv.FormatFloat(stop.LatLng.Lng, 'f', -1, 64)
				sLat := strconv.FormatFloat(stop.LatLng.Lat, 'f', -1, 64)
				// GEOADD key longitude latitude member
				log.Printf("GEOADD %s %s %s %s", key, sLng, sLat, "stop_"+stop.ID)
				conn.Do("GEOADD", key, sLng, sLat, "stop_"+stop.ID)

				bs, err := stop.Encode()
				if err != nil {
					log.Println("Prefetcher err: " + err.Error())
					continue
				}
				conn.Do("SET", "stop_"+stop.ID, string(bs))
			}

			// slowly
			time.Sleep(500 * time.Millisecond)
		}
	}

	// this is ready. incoming requests should use it
	_, err = conn.Do("SET", "prefetch_run", strconv.FormatInt(run, 10))
	if err != nil {
		log.Println("Prefetcher err: " + err.Error())
	}

	// delete the old geo key
	//_, err = conn.Do("DEL", previousKey)
	//if err != nil {
	//	log.Println("Prefetcher err: " + err.Error())
	//}
}
