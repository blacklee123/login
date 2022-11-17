package api

import (
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// Healthz godoc
// @Summary Liveness check
// @Description used by Kubernetes liveness probe
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Router /healthz [get]
// @Success 200 {string} string "OK"
func (s *Server) healthzHandler(c *gin.Context) {
	if atomic.LoadInt32(&healthy) == 1 {
		c.JSON(http.StatusOK, map[string]string{"status": "OK"})
	} else {
		c.String(http.StatusServiceUnavailable, "")
	}
}

// Readyz godoc
// @Summary Readiness check
// @Description used by Kubernetes readiness probe
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Router /readyz [get]
// @Success 200 {string} string "OK"
func (s *Server) readyzHandler(c *gin.Context) {
	if atomic.LoadInt32(&ready) == 1 {
		c.JSON(http.StatusOK, map[string]string{"status": "OK"})
		return
	}
	c.String(http.StatusServiceUnavailable, "")
}

// EnableReady godoc
// @Summary Enable ready state
// @Description signals the Kubernetes LB that this instance is ready to receive traffic
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Router /readyz/enable [post]
// @Success 202 {string} string "OK"
func (s *Server) enableReadyHandler(c *gin.Context) {
	atomic.StoreInt32(&ready, 1)
	c.String(http.StatusAccepted, "")
}

// DisableReady godoc
// @Summary Disable ready state
// @Description signals the Kubernetes LB to stop sending requests to this instance
// @Tags Kubernetes
// @Accept json
// @Produce json
// @Router /readyz/disable [post]
// @Success 202 {string} string "OK"
func (s *Server) disableReadyHandler(c *gin.Context) {
	atomic.StoreInt32(&ready, 0)
	c.String(http.StatusAccepted, "")
}
