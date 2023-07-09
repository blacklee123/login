package api

import (
	"github.com/blacklee123/login/pkg/version"
	"github.com/gin-gonic/gin"
)

func versionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Header.Set("X-API-Version", version.VERSION)
		c.Request.Header.Set("X-API-Revision", version.REVISION)
		c.Next()
	}
}
