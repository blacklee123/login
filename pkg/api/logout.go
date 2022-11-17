package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/viper"

	"github.com/gin-gonic/gin"
)

// LoginOut godoc
// @Summary LoginOut
// @Description LoginOut
// @Success 302
// @Router /web/logout [get]
func (s *Server) logout(c *gin.Context) {
	next := c.DefaultQuery("next", "/")
	host := c.Request.Host
	host = strings.Split(host, ":")[0]
	s.oauth.DelJWTCookie(c, host)
	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf(viper.GetString("github.url.authorize"), url.QueryEscape(next)))
}
