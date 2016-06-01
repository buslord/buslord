# Buslord

London Bus ETAs nearby 


## Dev environment

### Init

 * install go
 * setup GOPATH
 * ```export REDIS_URL="redis://127.0.0.1:6379"```

### Server

```
go get github.com/codegangsta/gin
gin -p 3000
```
 
### Test 

```
go get github.com/stretchr/testify/assert
go test
```


## Things I don't intend to do 

 * dependency management
 * very good frontent

## Overal comments

**Why a backend?** Why wont the js client talk direcly to the TFL api? 
The backend should compute. If it's just to hide the api secret maybe a better solution would be just to proxy requests to TFL adding the secret query param. 
One may say that we can do caching. Maybe they are doing the server side caching already. And maybe the client should do caching. 
One may say that the TFL certainly didn't geodistributed frontends, load balanced based on location with GeoDNS or anycast to minimize the response time. No need the big majority of requests will come from London.  


Over the code I used the **names** "Stop" and "ETA". TFL uses "StopPoint" and "Arrival" respectively. I regret my decision of starting with these names because now developers maintaining this code will have to learn both and their mapping to each other. 
It happened because I took the top down approach and started at what the rider would see. So I was mostly ignorant of the TFL API naming despite giving it a look before starting to code. I created a ticket to use the TFL names (#3) but I decided not to do it for this excercise. 

The frontend js code will refresh a bus stop ETA list only when one of the busses should be arriving (ETA=0). A better implementation should use the smallest TimeToLive like the backend does (#7). I didn't do this because the frontend is not the focus. 

 
