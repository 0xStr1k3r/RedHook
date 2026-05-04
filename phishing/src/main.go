package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/phishguard/backend/src/config"
	"github.com/phishguard/backend/src/controllers"
	"github.com/phishguard/backend/src/db"
	"github.com/phishguard/backend/src/models"
	"github.com/phishguard/backend/src/services"
)

func main() {
	log.Println("Starting PhishGuard Phishing Server...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	conn, err := db.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	handlers := controllers.NewPhishingHandler(conn, cfg)

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/track/open/:token", handlers.TrackOpen)
	router.GET("/track/click/:token", handlers.TrackClick)
	router.POST("/track/submit/:token", handlers.TrackSubmit)
	router.GET("/landing/:token", handlers.ShowLanding)
	router.POST("/landing/:token", handlers.ShowLanding)
	router.GET("/train/:token", handlers.ShowTraining)
	router.GET("/report", handlers.ReportPhishing)
	router.GET("/:path/report", handlers.ReportPhishing)

	srv := &http.Server{
		Addr:         ":" + cfg.Server.PortPhishing,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	go func() {
		log.Printf("Phishing server starting on port %s", cfg.Server.PortPhishing)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down phishing server...")
	srv.Close()
	log.Println("Phishing server exited")
}
