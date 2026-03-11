package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Final-Year-Project-G22/backend/core/pkg/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func NewGinEngine(cfg *Config, log Logger) *gin.Engine {
	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router without default middleware
	router := gin.New()

	// Apply standard middleware stack
	router.Use(middleware.Recovery(log))
	router.Use(middleware.RequestLogger(log))
	router.Use(middleware.CORSMiddleware("*"))
	router.Use(middleware.ErrorHandler())

	log.Info("Gin engine initialized",
		String("mode", gin.Mode()),
	)

	return router
}

// NewHTTPServer creates a new HTTP server and registers lifecycle hooks for graceful shutdown.
func NewHTTPServer(cfg *Config, engine *gin.Engine, lc fx.Lifecycle, log Logger) *http.Server {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting HTTP server",
				String("addr", srv.Addr),
				Int("port", cfg.App.Port),
			)

			// Start server in a goroutine since ListenAndServe blocks
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Error("HTTP server error", Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down HTTP server")

			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Error("HTTP server shutdown error", Error(err))
				return err
			}

			log.Info("HTTP server stopped gracefully")
			return nil
		},
	})

	return srv
}
