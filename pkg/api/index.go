package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) indexHandler(c *gin.Context) {

	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Main website",
	})
}
