package api

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/BlackLee123/handlers_gin"
	"github.com/gin-gonic/gin"
	_ "github.com/qaq-public/login/pkg/api/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"go.uber.org/zap"
)

var (
	healthy int32
	ready   int32
	// watcher *fscache.Watcher
)

type Config struct {
	HttpClientTimeout     time.Duration `mapstructure:"http-client-timeout"`
	HttpServerTimeout     time.Duration `mapstructure:"http-server-timeout"`
	ServerShutdownTimeout time.Duration `mapstructure:"server-shutdown-timeout"`
	Host                  string        `mapstructure:"host"`
	Port                  string        `mapstructure:"port"`
	Hostname              string        `mapstructure:"hostname"`
	UIPath                string        `mapstructure:"ui-path"`
}

type Server struct {
	router  *gin.Engine
	logger  *zap.Logger
	oauth   *OAuth
	config  *Config
	handler http.Handler
}

func NewServer(config *Config, logger *zap.Logger) (*Server, error) {
	srv := &Server{
		router: gin.Default(),
		logger: logger,
		oauth:  NewOAuth(logger),
		config: config,
	}

	return srv, nil
}

func (s *Server) registerHandlers() {
	s.router.Static("/assets", "./assets")
	s.router.GET("/", s.indexHandler)
	s.router.GET("/web/login", s.login)
	s.router.GET("/web/login/callback", s.loginCallback)
	s.router.GET("/web/logout", s.logout)
	s.router.GET("/healthz", s.healthzHandler)
	s.router.GET("/readyz", s.readyzHandler)
	s.router.POST("/readyz/enable", s.enableReadyHandler)
	s.router.POST("/readyz/disable", s.disableReadyHandler)
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

func (s *Server) registerMiddlewares() {
	s.router.Use(handlers_gin.ProxyHeaders())
	s.router.Use(versionMiddleware())
}

func (s *Server) ListenAndServe() (*http.Server, *int32, *int32) {
	s.router.Delims("[[", "]]")
	s.router.LoadHTMLGlob("templates/*")
	s.registerMiddlewares()
	s.registerHandlers()
	s.handler = s.router

	// create the http server
	srv := s.startServer()

	atomic.StoreInt32(&healthy, 1)
	atomic.StoreInt32(&ready, 1)

	return srv, &healthy, &ready
}

func (s *Server) startServer() *http.Server {

	// determine if the port is specified
	if s.config.Port == "0" {

		// move on immediately
		return nil
	}

	srv := &http.Server{
		Addr:         s.config.Host + ":" + s.config.Port,
		WriteTimeout: s.config.HttpServerTimeout,
		ReadTimeout:  s.config.HttpServerTimeout,
		IdleTimeout:  2 * s.config.HttpServerTimeout,
		Handler:      s.handler,
	}

	// start the server in the background
	go func() {
		s.logger.Info("Starting HTTP Server.", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Fatal("HTTP server crashed", zap.Error(err))
		}
	}()

	// return the server and routine
	return srv
}
