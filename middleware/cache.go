package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Cache() func(c *gin.Context) {
	return func(c *gin.Context) {
		uri := c.Request.RequestURI
		if uri == "/" || strings.HasSuffix(uri, ".html") {
			// HTML files: short cache to ensure users get latest version
			c.Header("Cache-Control", "no-cache, max-age=60")
		} else {
			// Static assets (JS/CSS/images with hash fingerprints): long cache
			c.Header("Cache-Control", "max-age=604800") // one week
		}
		c.Next()
	}
}
