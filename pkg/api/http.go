package api

import (
	"github.com/gin-gonic/gin"
	"github.com/qaq-public/login/pkg/version"
)

func versionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Set("X-API-Version", version.VERSION)
		c.Request.Header.Set("X-API-Revision", version.REVISION)
		c.Next()
	}
}
