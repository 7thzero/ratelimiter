package samplerouter

import (
	"log"
	"net/http"
	"github.com/7thzero/ratelimiter"
	"strings"
	"sync"
	"time"
)

type SampleRouter struct{
	throttler map[string]ratelimiter.RateLimit
	throttlerMtx sync.Mutex						// Ensures we don't have a race condition when accessing throttler
}

//
// Initialize the sample HTTP Router
func (sr *SampleRouter) Init(){
	// Guard against re-initialization. Only initialize a new map on creation
	if !sr.isInitted(){
		sr.throttler = make(map[string]ratelimiter.RateLimit)
		sr.throttlerMtx = sync.Mutex{}
	}
}

//
// Checks if the router has been successfully initialized
func (sr *SampleRouter) isInitted() bool{
	if sr.throttler != nil{
		return true
	}
	return false
}

//
// Route traffic to specified endpoints
func (sr *SampleRouter) Route(writer http.ResponseWriter, request *http.Request){
	if !sr.isInitted(){
		log.Println("Router is uninitialized. Ensure .Init() is called prior to use")
		return
	}

	// Simple request log
	log.Println(request.RemoteAddr, request.Method, request.URL.Path, request.UserAgent())

	// Check if we should limit/block the connection attempt
	host := strings.Split(request.RemoteAddr, ":")[0]
	if sr.dosShield(host){
		writer.WriteHeader(http.StatusTooManyRequests)
		writer.Write([]byte(`{"message": "Rate limit exceeded"}`))
		return
	}

	// Wide open CORS for this example
	writer.Header().Set("Access-Control-Allow-Origin", "*")

	// Ignore requests for favicon.ico
	if strings.HasPrefix(strings.ToUpper(request.URL.Path), "/FAVICON.ICO"){
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	//
	// Write out the sample response
	now := time.Now().Format(time.RFC3339)
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("<html><head></head><body><h1>Successful request!</h1><br>"+now+"</body></html>"))
}

//
// leverages the rate limiter to check if DOS is suspected
func (sr *SampleRouter) dosShield(host string) bool{
	sr.throttlerMtx.Lock()
	// Ensure we have an entry for the host if it doesn't already exist
	if _, exists := sr.throttler[host]; !exists{
		limiter := ratelimiter.RateLimit{}
		limiter.Init(15, 65)
		sr.throttler[host] = limiter
	}

	hostLimit := sr.throttler[host]				// Get the throttler for this host
	isRateLimited := hostLimit.IsRateLimited()	// Records the access attempt and checks if this connection should be throttled
	sr.throttler[host] = hostLimit				// Re-assign the RateLimiter to record the updated access attempt
	sr.throttlerMtx.Unlock()
	return isRateLimited
}