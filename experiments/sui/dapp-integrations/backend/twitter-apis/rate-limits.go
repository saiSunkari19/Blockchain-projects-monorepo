package twitterapis

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"` // Time when the limit window resets
	mu        sync.RWMutex
}

type RateLimitManager struct {
	User15Min      *RateLimitInfo // related to x-rate-limit-*
	User24HourPost *RateLimitInfo // related to x-user-limit-24hour-*
	App24HourPost  *RateLimitInfo // related to x-app-limit-24hour-*
}

func (rli *RateLimitInfo) Update(limit, remaining int, resetUnix int64) {
	rli.mu.Lock()
	defer rli.mu.Unlock()

	rli.Limit = limit
	rli.Remaining = remaining
	rli.ResetAt = time.Unix(resetUnix, 0)
}

func (rli *RateLimitInfo) CanMakeRequest() (bool, time.Duration) {
	rli.mu.RLock()
	defer rli.mu.RUnlock()

	if rli.Remaining > 0 {
		return true, 0
	}

	waitUntil := rli.ResetAt
	if time.Now().After(waitUntil) {
		// if reset time has passed, we can assume that we can make request
		return true, 0
	}

	return false, time.Until(waitUntil)

}

func NewRateLimitManager() *RateLimitManager {

	initialRemaining := 9999
	initialReset := time.Now().Add(-1 * time.Hour) // a time in past

	return &RateLimitManager{
		User15Min: &RateLimitInfo{
			Limit: initialRemaining, Remaining: initialRemaining,
			ResetAt: initialReset,
		},
		User24HourPost: &RateLimitInfo{
			Limit: initialRemaining, Remaining: initialRemaining,
			ResetAt: initialReset,
		},
		App24HourPost: &RateLimitInfo{
			Limit: initialRemaining, Remaining: initialRemaining,
			ResetAt: initialReset,
		},
	}
}

func (rlm *RateLimitManager) UpdateFromHeader(header http.Header) {

	parseAndUpdate := func(rli *RateLimitInfo, limitH, remainingH, resetH string) {

		limitStr := header.Get(limitH)
		remainingStr := header.Get(remainingH)
		resetStr := header.Get(resetH)

		limit, errLimit := strconv.Atoi(limitStr)
		if errLimit != nil && limitStr != "" { // Log error if limitStr is not empty but fails to parse
			log.Printf("Error parsing rate limit limit value '%s' for header %s: %v", limitStr, limitH, errLimit)
			limit = 0 // Default to 0 or some other sensible default if parsing fails
		}

		remaining, errRemaining := strconv.Atoi(remainingStr)
		if errRemaining != nil && remainingStr != "" {
			log.Printf("Error parsing rate limit remaining value '%s' for header %s: %v", remainingStr, remainingH, errRemaining)
			remaining = 0
		}

		resetUnix, errReset := strconv.ParseInt(resetStr, 10, 64)
		if errReset != nil && resetStr != "" {
			log.Printf("Error parsing rate limit reset value '%s' for header %s: %v", resetStr, resetH, errReset)
			resetUnix = time.Now().Unix() // Default to 0 if parsing fails, leading to epoch time (needs careful consideration)
		}

		var resetAt time.Time
		if errReset == nil {
			resetAt = time.Unix(resetUnix, 0)
			now := time.Now().Unix()

			waitSecs := resetUnix - now
			if waitSecs < 0 {
				waitSecs = 60
			}

			fmt.Println("Twitter 429 - waiting until reset", "wait_secs", waitSecs, "reset_header", resetStr)

		} else {
			resetAt = time.Now()
		}

		if remainingStr != "" && resetStr != "" {
			rli.Update(limit, remaining, resetAt.Unix())

			log.Printf("Updated rate limit: %s - Limit: %d, Remaining: %d, ResetAt: %s",
				remainingH, rli.Limit, rli.Remaining, rli.ResetAt.Local().Format(time.RFC1123))
		}

	}

	// Standard request rate limits (often per 15min for a user token)
	parseAndUpdate(rlm.User15Min, "x-rate-limit-limit", "x-rate-limit-remaining", "x-rate-limit-reset")

	// User-Specific 24-hours limits
	parseAndUpdate(rlm.User24HourPost, "x-user-limit-24hour-limit", "x-user-limit-24hour-remaining", "x-user-limit-24hour-reset")

	// App-Specific
	parseAndUpdate(rlm.App24HourPost, "x-app-limit-24hour-limit", "x-app-limit-24hour-remaining", "x-app-limit-24hour-reset")

}

// CheckOverallRateLimits check if the post can be made and how long wait if not.
// It checks all relevant limits and returns the longest wait time
func (rlm *RateLimitManager) CheckOverallRateLimits() (bool, time.Duration) {
	var canPostOverall = true
	var maxDurataion time.Duration = 0

	checkers := []*RateLimitInfo{
		rlm.User15Min,
		rlm.User24HourPost,
		rlm.App24HourPost,
	}

	for _, checker := range checkers {
		canPost, waitDuration := checker.CanMakeRequest()
		if !canPost {
			canPostOverall = false
			if waitDuration > maxDurataion {
				maxDurataion = waitDuration
			}
		}
	}

	return canPostOverall, maxDurataion
}
