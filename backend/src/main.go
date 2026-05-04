package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redhook/backend/src/config"
	"github.com/redhook/backend/src/controllers"
	dbpkg "github.com/redhook/backend/src/db"
	"github.com/redhook/backend/src/middleware"
	"github.com/redhook/backend/src/models"
	"github.com/redhook/backend/src/services"
)

func main() {
	log.Println("Starting RedHook Backend...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	conn, err := dbpkg.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := dbpkg.Migrate(conn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err := dbpkg.Seed(conn); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	}

	authCtrl := controllers.NewAuthController(conn, cfg)
	userCtrl := controllers.NewUserController(conn)
	campaignCtrl := controllers.NewCampaignController(conn)
	templateCtrl := controllers.NewTemplateController(conn)
	pageCtrl := controllers.NewLandingPageController(conn)
	analyticsCtrl := controllers.NewAnalyticsController(conn)
	emailSvc := services.NewEmailService(conn, cfg)

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	phishHandler := controllers.NewPhishingHandler(conn)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	router.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "User-agent: *\nDisallow: /")
	})

	router.GET("/", func(c *gin.Context) {
		c.File("../frontend/index.html")
	})

	router.GET("/track/open/:token", phishHandler.TrackOpen)
	router.GET("/track/click/:token", phishHandler.TrackClick)
	router.POST("/track/submit/:token", phishHandler.TrackSubmit)
	router.GET("/landing/:token", phishHandler.ShowLanding)
	router.POST("/landing/:token", phishHandler.ShowLanding)
	router.GET("/landing/preview/:id", phishHandler.ShowLandingPreview)
	router.GET("/train/:token", phishHandler.ShowTraining)
	router.GET("/report", phishHandler.ReportPhishing)
	router.GET("/:path/report", phishHandler.ReportPhishing)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authCtrl.Register)
			auth.POST("/login", authCtrl.Login)
			auth.GET("/me", middleware.AuthMiddleware(cfg.JWT.Secret), authCtrl.GetMe)
		}

		api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
		{
			users := api.Group("/users")
			{
				users.GET("", userCtrl.List)
				users.POST("", userCtrl.Create)
				users.GET("/departments", userCtrl.GetDepartments)
				users.POST("/bulk", userCtrl.BulkCreate)
				users.GET("/:id", userCtrl.Get)
				users.PUT("/:id", userCtrl.Update)
				users.DELETE("/:id", userCtrl.Delete)
			}

			campaigns := api.Group("/campaigns")
			{
				campaigns.GET("", campaignCtrl.List)
				campaigns.POST("", campaignCtrl.Create)
				campaigns.GET("/:id", campaignCtrl.Get)
				campaigns.PUT("/:id", campaignCtrl.Update)
				campaigns.DELETE("/:id", campaignCtrl.Delete)
				campaigns.POST("/:id/launch", campaignCtrl.Launch)
				campaigns.GET("/:id/stats", campaignCtrl.GetStats)
			}

			templates := api.Group("/templates")
			{
				templates.GET("", templateCtrl.List)
				templates.POST("", templateCtrl.Create)
				templates.GET("/:id", templateCtrl.Get)
				templates.PUT("/:id", templateCtrl.Update)
				templates.DELETE("/:id", templateCtrl.Delete)
				templates.GET("/:id/preview", templateCtrl.Preview)
			}

			pages := api.Group("/pages")
			{
				pages.GET("", pageCtrl.List)
				pages.POST("", pageCtrl.Create)
				pages.GET("/:id", pageCtrl.Get)
				pages.PUT("/:id", pageCtrl.Update)
				pages.DELETE("/:id", pageCtrl.Delete)
			}

			analytics := api.Group("/analytics")
			{
				analytics.GET("/overview", analyticsCtrl.GetOverview)
				analytics.GET("/campaign/:id", analyticsCtrl.GetCampaignAnalytics)
				analytics.GET("/department/:dept", analyticsCtrl.GetDepartment)
				analytics.GET("/user/:id", analyticsCtrl.GetUserRisk)
				analytics.GET("/trends", analyticsCtrl.GetTrends)
			}

			api.GET("/submissions", func(c *gin.Context) {
				var subs []models.SubmissionLog
				conn.Preload("User").Preload("Campaign").Order("submitted_at DESC").Limit(100).Find(&subs)
				c.JSON(http.StatusOK, subs)
			})

			api.POST("/send", func(c *gin.Context) {
				var req struct {
					CampaignID uint   `json:"campaign_id" binding:"required"`
					Recipients []uint `json:"recipients" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				err := emailSvc.SendCampaign(services.SendRequest{
					CampaignID: req.CampaignID,
					Recipients: req.Recipients,
				})
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "Campaign emails sent"})
			})
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AdminOnly())
		{
			admin.GET("/stats", func(c *gin.Context) {
				var userCount, campaignCount int64
				conn.Model(&models.User{}).Count(&userCount)
				conn.Model(&models.Campaign{}).Count(&campaignCount)
				c.JSON(http.StatusOK, gin.H{
					"users":     userCount,
					"campaigns": campaignCount,
				})
			})
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
