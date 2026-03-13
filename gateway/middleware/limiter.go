package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count   int
	expires time.Time
}

func RateLimit(maxRequests int, windowSeconds int) gin.HandlerFunc {
	if maxRequests <= 0 || windowSeconds <= 0 {
		return nil
	}

	var (
		mu       sync.Mutex
		visitors = make(map[string]visitor)
		window   = time.Duration(windowSeconds) * time.Second
	)

	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()

		mu.Lock()
		record, ok := visitors[key]
		if !ok || now.After(record.expires) {
			record = visitor{count: 0, expires: now.Add(window)}
		}
		record.count++
		visitors[key] = record
		mu.Unlock()

		if record.count > maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"message": "too many requests"})
			return
		}
		c.Next()
	}
}
