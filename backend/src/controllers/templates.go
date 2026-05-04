package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phishguard/backend/src/models"
	"gorm.io/gorm"
)

type TemplateController struct {
	DB *gorm.DB
}

func NewTemplateController(db *gorm.DB) *TemplateController {
	return &TemplateController{DB: db}
}

func (ctrl *TemplateController) List(c *gin.Context) {
	var templates []models.EmailTemplate
	query := ctrl.DB

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR subject ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Order("created_at DESC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

func (ctrl *TemplateController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var template models.EmailTemplate
	if err := ctrl.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

type CreateTemplateRequest struct {
	Name            string  `json:"name" binding:"required"`
	Subject         string  `json:"subject" binding:"required"`
	BodyHTML        string  `json:"body_html" binding:"required"`
	DifficultyScore float64 `json:"difficulty_score"`
	Category        string  `json:"category"`
	FromName        string  `json:"from_name"`
}

func (ctrl *TemplateController) Create(c *gin.Context) {
	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := models.EmailTemplate{
		Name:            req.Name,
		Subject:         req.Subject,
		BodyHTML:        req.BodyHTML,
		DifficultyScore: req.DifficultyScore,
		Category:        req.Category,
		FromName:        req.FromName,
	}

	if err := ctrl.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

type UpdateTemplateRequest struct {
	Name            string  `json:"name"`
	Subject         string  `json:"subject"`
	BodyHTML        string  `json:"body_html"`
	DifficultyScore float64 `json:"difficulty_score"`
	Category        string  `json:"category"`
	FromName        string  `json:"from_name"`
}

func (ctrl *TemplateController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req UpdateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var template models.EmailTemplate
	if err := ctrl.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Subject != "" {
		template.Subject = req.Subject
	}
	if req.BodyHTML != "" {
		template.BodyHTML = req.BodyHTML
	}
	if req.DifficultyScore > 0 {
		template.DifficultyScore = req.DifficultyScore
	}
	if req.Category != "" {
		template.Category = req.Category
	}
	if req.FromName != "" {
		template.FromName = req.FromName
	}

	if err := ctrl.DB.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

func (ctrl *TemplateController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	if err := ctrl.DB.Delete(&models.EmailTemplate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted"})
}

func (ctrl *TemplateController) Preview(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var template models.EmailTemplate
	if err := ctrl.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	previewBody := template.BodyHTML
	previewBody = replaceVariables(previewBody, map[string]string{
		"Name":      "John Doe",
		"TrackLink": "http://localhost:8081/track/click/preview",
	})

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, previewBody)
}

func replaceVariables(html string, vars map[string]string) string {
	result := html
	for key, value := range vars {
		result = replaceAllStrings(result, "{{."+key+"}}", value)
	}
	return result
}

func replaceAllStrings(s, old, new string) string {
	result := s
	for {
		i := findSubstringIndex(result, old)
		if i == -1 {
			break
		}
		result = result[:i] + new + result[i+len(old):]
	}
	return result
}

func findSubstringIndex(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
