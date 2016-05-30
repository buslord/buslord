# Buslord

London Bus ETAs nearby 


## Dev environment

### Init

 * install go
 * setup GOPATH

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

Why a backend? Why wont the js client talk direcly to the TFL api? 
The backend should compute. If it's just to hide the api secret maybe a better solution would be just to proxy requests to TFL adding the secret query param. 
One may say that we can do caching. Maybe they are doing the server side caching already. And maybe the client should do caching. 
One may say that the TFL certainly didn't geodistributed frontends, load balanced based on location with GeoDNS or anycast to minimize the response time. No need the big majority of requests will come from London.  
