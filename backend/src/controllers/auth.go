package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/phishguard/backend/src/config"
	"github.com/phishguard/backend/src/middleware"
	"github.com/phishguard/backend/src/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func NewAuthController(db *gorm.DB, cfg *config.Config) *AuthController {
	return &AuthController{DB: db, Cfg: cfg}
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Name       string `json:"name" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (ctrl *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := ctrl.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	role := req.Role
	if role == "" {
		role = "user"
	}

	user := models.User{
		Email:        req.Email,
		Name:         req.Name,
		Department:   req.Department,
		Role:         role,
		PasswordHash: string(hash),
	}

	if err := ctrl.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, ctrl.Cfg.JWT.Secret, ctrl.Cfg.JWT.ExpireHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{
		Token: token,
		User:  user,
	})
}

func (ctrl *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ctrl.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Email, user.Role, ctrl.Cfg.JWT.Secret, ctrl.Cfg.JWT.ExpireHour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

func (ctrl *AuthController) GetMe(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := ctrl.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
