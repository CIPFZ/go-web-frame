package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/CIPFZ/gowebframe/internal/core/config"
	"github.com/CIPFZ/gowebframe/internal/core/db"
	corelog "github.com/CIPFZ/gowebframe/internal/core/log"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/api"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/repository"
	pluginmarketrouter "github.com/CIPFZ/gowebframe/internal/modules/plugin_market/router"
	"github.com/CIPFZ/gowebframe/internal/modules/plugin_market/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const defaultConfigPath = "./configs/plugin-market.yaml"

func main() {
	configPath := flag.String("f", defaultConfigPath, "config file path")
	flag.Parse()

	cfg, v, logger, svc, err := bootstrap(*configPath)
	if err != nil {
		panic(err)
	}
	defer corelog.Sync(logger)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"name":   cfg.System.Name,
		})
	})
	pluginmarketrouter.Init(engine, api.New(svc, syncToken(v)))

	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.System.Port),
		Handler:        engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("plugin market server starting", zap.String("addr", server.Addr))
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			errCh <- listenErr
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-errCh:
		logger.Fatal("plugin market server exited with error", zap.Error(err))
	case sig := <-quit:
		logger.Info("plugin market server shutting down", zap.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("plugin market shutdown failed", zap.Error(err))
	}
}

func bootstrap(path string) (*config.Config, *viper.Viper, *zap.Logger, *service.Service, error) {
	cfg, v, err := config.Load(path)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	logger, err := corelog.NewLogger(&cfg.Logger, nil)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	gormDB, err := db.InitDatabase(cfg.Database, logger)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	repo := repository.New(gormDB)
	svc := service.New(repo)
	if err := svc.AutoMigrate(); err != nil {
		return nil, nil, nil, nil, err
	}
	if err := svc.SeedDemoDataIfEmpty(context.Background()); err != nil {
		return nil, nil, nil, nil, err
	}

	return cfg, v, logger, svc, nil
}

func syncToken(v *viper.Viper) string {
	if token := strings.TrimSpace(os.Getenv("MARKET_SYNC_TOKEN")); token != "" {
		return token
	}
	return strings.TrimSpace(v.GetString("plugin_market.sync_token"))
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-Market-Sync-Token")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
