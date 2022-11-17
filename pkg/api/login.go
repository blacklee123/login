package api

import (
	"encoding/base64"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @Summary Login
// @Description Login
// @Success 302
// @Router /web/login [get]
func (s *Server) login(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, s.oauth.authorize(c))
}

// LoginCallBack godoc
// @Summary LoginCallBack
// @Description Login
// @Success 302
// @Router /web/login/callback [get]
func (s *Server) loginCallback(c *gin.Context) {
	next := c.DefaultQuery("next", "")
	code := c.DefaultQuery("code", "")
	if code == "" {
		s.logger.Panic("The code check wrong")
	}
	s.logger.Info("The login_callback", zap.String("code", code))
	token := s.oauth.Code2Token(code)
	host := c.Request.Host
	host = strings.Split(host, ":")[0]
	s.oauth.SetJWTCookie(c, token, host)
	decoded, err := base64.StdEncoding.DecodeString(next)
	redirectUrl := string(decoded)
	if err != nil {
		redirectUrl = "/"
	}
	c.Redirect(http.StatusMovedPermanently, redirectUrl)
}
