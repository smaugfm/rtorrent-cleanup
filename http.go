package main

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type ThrottledTransport struct {
	roundTripperWrap http.RoundTripper
	ratelimiter      *rate.Limiter
}

func (c *ThrottledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	err := c.ratelimiter.Wait(r.Context()) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	return c.roundTripperWrap.RoundTrip(r)
}

func NewThrottledTransport(limitPeriod time.Duration, requestCount int, transportWrap http.RoundTripper) http.RoundTripper {
	return &ThrottledTransport{
		roundTripperWrap: transportWrap,
		ratelimiter:      rate.NewLimiter(rate.Every(limitPeriod), requestCount),
	}
}
