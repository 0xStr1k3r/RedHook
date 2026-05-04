package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/redhook/backend/src/models"
	"gorm.io/gorm"
)

type LandingPageController struct {
	DB *gorm.DB
}

func NewLandingPageController(db *gorm.DB) *LandingPageController {
	return &LandingPageController{DB: db}
}

func (ctrl *LandingPageController) List(c *gin.Context) {
	var pages []models.LandingPage
	query := ctrl.DB

	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Order("created_at DESC").Find(&pages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch landing pages"})
		return
	}

	c.JSON(http.StatusOK, pages)
}

func (ctrl *LandingPageController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landing page ID"})
		return
	}

	var page models.LandingPage
	if err := ctrl.DB.First(&page, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landing page not found"})
		return
	}

	c.JSON(http.StatusOK, page)
}

type CreateLandingPageRequest struct {
	Name               string `json:"name" binding:"required"`
	HTMLContent        string `json:"html_content" binding:"required"`
	RedirectURL        string `json:"redirect_url"`
	CaptureCredentials bool   `json:"capture_credentials"`
	TrainingContent    string `json:"training_content"`
}

func (ctrl *LandingPageController) Create(c *gin.Context) {
	var req CreateLandingPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page := models.LandingPage{
		Name:               req.Name,
		HTMLContent:        req.HTMLContent,
		RedirectURL:        req.RedirectURL,
		CaptureCredentials: req.CaptureCredentials,
		TrainingContent:    req.TrainingContent,
	}

	if err := ctrl.DB.Create(&page).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create landing page"})
		return
	}

	c.JSON(http.StatusCreated, page)
}

type UpdateLandingPageRequest struct {
	Name               string `json:"name"`
	HTMLContent        string `json:"html_content"`
	RedirectURL        string `json:"redirect_url"`
	CaptureCredentials bool   `json:"capture_credentials"`
	TrainingContent    string `json:"training_content"`
}

func (ctrl *LandingPageController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landing page ID"})
		return
	}

	var req UpdateLandingPageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var page models.LandingPage
	if err := ctrl.DB.First(&page, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Landing page not found"})
		return
	}

	if req.Name != "" {
		page.Name = req.Name
	}
	if req.HTMLContent != "" {
		page.HTMLContent = req.HTMLContent
	}
	if req.RedirectURL != "" {
		page.RedirectURL = req.RedirectURL
	}
	if req.TrainingContent != "" {
		page.TrainingContent = req.TrainingContent
	}
	page.CaptureCredentials = req.CaptureCredentials

	if err := ctrl.DB.Save(&page).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update landing page"})
		return
	}

	c.JSON(http.StatusOK, page)
}

func (ctrl *LandingPageController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid landing page ID"})
		return
	}

	if err := ctrl.DB.Delete(&models.LandingPage{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete landing page"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Landing page deleted"})
}
