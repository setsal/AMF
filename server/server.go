package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
)

// Start Requests handler HTTP Server.
func Start(logDir string) error {
	e := echo.New()
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// setup access logger
	logPath := filepath.Join(logDir, "httpd.log")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if _, err := os.Create(logPath); err != nil {
			return err
		}
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Output: f,
	}))

	// Routes
	e.GET("/health", health())

	bindURL := fmt.Sprintf("%s:%s", viper.GetString("bind"), viper.GetString("port"))
	log.Infof("Listening on %s", bindURL)

	return e.Start(bindURL)
}

// Handler
func health() echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "")
	}
}
