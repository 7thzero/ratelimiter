package ratelimiter

import (
	"sync"
	"time"
)

//
// Per-identifier access tracking
//
// Originally designed with rate-limiting calls to go HTTP services
type RateLimit struct{
	id 					string			// host/ip address or other unique identifier
	accessLog			[]time.Time		// Records each access time
	checkIntervalSec	int				// Time in seconds to check access attempts
	maxAttempts			int				// Maximum number of access attempts in the interval
	configured			bool			// Set once instantiated with target values (or defaults)
	mtx					sync.Mutex		// Ensure consistency in the event of multiple concurrent requests
}

//
// Configure monitoring interval and max access attempts in the interval
func (rl *RateLimit) Init(maxAttempts int, checkIntervalSec int){
	rl.checkIntervalSec = checkIntervalSec
	rl.maxAttempts = maxAttempts
	rl.configured = true
	rl.mtx = sync.Mutex{}
}

//
// Set the identifier for this rate limit instance
// Originally designed to be the IP address of hosts making HTTP connections
func (rl *RateLimit) SetIdentifier(id string){
	rl.mtx.Lock()
	rl.id = id

	//
	// Default to 20 access attempts in 60 seconds if no explicit settings are configured
	if !rl.configured{
		rl.checkIntervalSec = 60
		rl.maxAttempts = 20
	}
	rl.mtx.Unlock()
}

//
// Return true if a host/identifier should be rate limited
//	false is returned if the target identifier/host should not be rate limited
func (rl *RateLimit) IsRateLimited() bool{
	rl.mtx.Lock()
	// Log this access request
	accessRequest := time.Now()
	rl.accessLog = append(rl.accessLog, accessRequest)

	//
	// Only consider access attempts that fall within the interval
	var accessLogInInterval []time.Time
	for _, access := range rl.accessLog{
		diff := accessRequest.Sub(access)

		if diff < time.Duration(rl.checkIntervalSec) * time.Second{
			accessLogInInterval = append(accessLogInInterval, access)
		}
	}
	rl.accessLog = accessLogInInterval			// Clear out/ignore accesses that fall outside the interval range
	rl.mtx.Unlock()

	//
	// Check if this access request should be allowed.
	//	Has the host exceeded the rate limit for the specified time interval?
	if len(accessLogInInterval) >= rl.maxAttempts{
		// If so, signal that the connection should be rejected/limited
		return true
	}

	//
	// Otherwise, signal that the connection should be allowed
	return false
}