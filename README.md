# Buslord

## London Bus ETAs nearby 

Buslord is a browser app that uses [TFL's API](https://api.tfl.gov.uk) and [geolocation apis](https://developer.mozilla.org/en-US/docs/Web/API/Geolocation/Using_geolocation) to give London Bus users the latest arrival times nearby. 

You can visit the app on your browser by going to [https://buslord.advogadoitapoa.com/](https://buslord.advogadoitapoa.com/).

The client side uses [Google Maps Javascript API V3](https://developers.google.com/maps/documentation/javascript/reference) and vanilla javascript + jquery to fetch data from a server written in go. 

## A page loads

When a page loads this is what happens:

 * we request a list of stops around an specific location near the Big Ben
 * at the same time we initialize the google map on that location
 * when the bus stop request arrives we draw the stops on the map 
 * meanwhile the user is being prompted for her location
 * if she authorizes the red marker and the center of the map move to her location
 * if she doesn't then the stops will be there 
 * whenever the map changes it's bounds (also when you resize the window) a new request for stops happens and we draw them on the map
 * when the user taps on a bus stop marker we fetch the list of arrivals (ETAs) for that stop
 * the arrival times will be decremented every second 
 

## API

The server exposes 2 API endpoints: ``/stops`` and ``/etas`` that respond json to GET requests that use query parameters.

### GET /stops

Takes 4 params that define a retangular region using coordinates for two points: **ne** (north east) and **sw** (south west).  

Here is an example:

[https://buslord.advogadoitapoa.com/stops?neLat=51.50546&neLng=-0.114908346&swLat=51.49502&swLng=-0.1400996](https://buslord.advogadoitapoa.com/stops?neLat=51.50546&neLng=-0.114908346&swLat=51.49502&swLng=-0.1400996)

and the response is a list of stops on that region: 

```json
[
  {
    id: "490005646E",
    name: "St Thomas' Hospital / County Hall",
    lat_lng: {
      lat: 51.500865,
      lng: -0.118075
    }
  },
  {
    id: "490008376E",
    name: "Horse Guards Avenue",
    lat_lng: {
      lat: 51.504935,
      lng: -0.125184
  },
  ...
},
```

### GET /etas

Takes one param: the bus stop ID. Here is an example:

[https://buslord.advogadoitapoa.com/etas?stop=490014498S](https://buslord.advogadoitapoa.com/etas?stop=490014498S)

and the response is an list of ETAs for that bus stop:

```json
[
  {
    id: "303938447",
    line_name: "148",
    destination_name: "Camberwell Green",
    eta: 90,
    mode_name: "bus",
    time_to_live: "2016-06-03T21:17:24Z"
  },
  {
    id: "-1546743980",
    line_name: "211",
    destination_name: "Waterloo",
    eta: 483,
    mode_name: "bus",
    time_to_live: "2016-06-03T21:23:57Z"
},
...
```

The list is ordered by the arrival time. You have the lines arriving the earliest first. 

## The machine

The server is now deployed at a small VPS. The setup is very simple:

 * digital ocean droplet in the london region
 * 14.04 ubuntu
 * standard memcached ubuntu package
 * standard installation of redis

There is no complex (or good) deploy script or continuous integration:

On my machine I run a bash script that is nothing but a series of ``scp`` to the server. 

## Caching

 * ETAs get cached with memcached
 * ETAs get expired on the frontend whenever a bus ETA reaches the 0 seconds
 * ETAs get expired on the server side using the "TimeToLive" property
 
 * Many ETA requests will miss the cache: there are many more bus stops than users at the time. I would like to use the TFL's "Live bus and river bus arrivals API (stream)" to get the arrivals pushed to me and then I would save them on redis and serve from there. 
 
[I tried to cache the bus stops](https://github.com/marcelcorso/buslord/issues/8) but failed. 

The idea was to every 12h or 24h (bus stops don't change much) get a full list of stops, save and read from redis using it's [geospatial commands](http://redis.io/commands/geoadd).
I managed to make the user facing (reads) work. Also writing geospatial items on redis. 

The only bug left was on the ["sweeping" code](https://github.com/marcelcorso/buslord/blob/master/stop_prefetcher.go#L39). It was supposed to define a big region to cache and then iteratively move a smaller window over its area doing requests to tfl and saving on redis. 

There I found an interesting problem: redis only supports radius based geo queries and the frontend and the tfl API work with 2-point-defined-regions. A little bit of math and a library to calculate distance between two points (I could have written that too...) did the job.

If you want to try to find the bug (it's an interesting problem) you should:

1) Change the stops-enabled attribute on the config.yaml file: 

```yaml
cache:
  etas-enabled: true
  stops-enabled: true 
...
``` 

2) And run the prefetching task:

```bash
buslord -prefetcher=true
```

## Cool stuff

 * Removing markers from the map

Everytime the user moves around or zoom new bus stops have to be drawn. But also some of them are no longer being able to be seen. So they are removed from the screen to free resources. 

 * Testing APIs that use other APIs [can be fun](https://github.com/marcelcorso/buslord/blob/master/etas_test.go#L14) and the standard go libs help. 


## Other things I would like to do 

but I didn't had time:

 * At small zooms (from far) there are too many bus stops to be drawn. I would like to consistently sample from the list and draw less of them and give a ligther experience.
 * At small zooms the TFL API responds a lot of data and the request is slow. If I managed to cache the totallity of stops and respond a sampled list it would be faster. 

 * Use the ETA's "TimeToLive" to refresh a bus stop arrivals list
 * When a list of stops loads prefetch their ETAs. A request to ``/etas?stop=490005646E,490008376E`` could answer many arrival predictions in one request.

 * Monitoring: I would like to use [Prometheus](htt://prometheus.io) to monitor the app. 


## Dev environment

### Init

 * install go
 * setup $GOPATH
 * ```export REDIS_URL="redis://127.0.0.1:6379"```

### Server

```
go get github.com/codegangsta/gin
gin -p 3000
```

or 

```
PORT=300 buslord
```

and then

```
open http://localhost:3000
```
 
### Test 

```
go get github.com/stretchr/testify/assert
go test
```


## Things I didn't intend to do 

 * dependency management
 * very good frontent

## Overal comments

**Why a backend?** Why wont the js client talk direcly to the TFL api? 
The backend should compute. If it's just to hide the api secret maybe a better solution would be just to proxy requests to TFL adding the secret query param. 
One may say that we can do caching. Maybe they are doing the server side caching already. And maybe the client should do caching. 
One may say that the TFL certainly didn't geodistributed frontends, load balanced based on location with GeoDNS or anycast to minimize the response time. No need the big majority of requests will come from London.  


Over the code I used the **names** "Stop" and "ETA". TFL uses "StopPoint" and "Arrival" respectively. I regret my decision of starting with these names because now developers maintaining this code will have to learn both and their mapping to each other. 
It happened because I took the top down approach and started at what the rider would see. So I was mostly ignorant of the TFL API naming despite giving it a look before starting to code. I created a ticket to use the TFL names (#3) but I decided not to do it for this excercise. 

The frontend uses non-cool vanilla javascript and the very un-cool jquery. The initial idea of this project was to focus on the backend and do a minimal frontend. But I wanted to have everything at least working so I ended up spending more time on the frontend that I wanted. Ideally a project like this should something cooler like React or Angular. A lot of information could be cached on the client's localStorage too... 
 
