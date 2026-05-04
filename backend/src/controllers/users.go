package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/phishguard/backend/src/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

func (ctrl *UserController) List(c *gin.Context) {
	var users []models.User
	query := ctrl.DB

	if department := c.Query("department"); department != "" {
		query = query.Where("department = ?", department)
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (ctrl *UserController) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	if err := ctrl.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

type CreateUserRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Name       string `json:"name" binding:"required"`
	Password   string `json:"password" binding:"required,min=6"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

func (ctrl *UserController) Create(c *gin.Context) {
	var req CreateUserRequest
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

	c.JSON(http.StatusCreated, user)
}

type UpdateUserRequest struct {
	Name       string `json:"name"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

func (ctrl *UserController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ctrl.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Department != "" {
		user.Department = req.Department
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := ctrl.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (ctrl *UserController) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := ctrl.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

type BulkUserRequest struct {
	Users []CreateUserRequest `json:"users" binding:"required"`
}

func (ctrl *UserController) BulkCreate(c *gin.Context) {
	var req BulkUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var createdUsers []models.User
	for _, u := range req.Users {
		var existingUser models.User
		if err := ctrl.DB.Where("email = ?", u.Email).First(&existingUser).Error; err == nil {
			continue
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		user := models.User{
			Email:        u.Email,
			Name:         u.Name,
			Department:   u.Department,
			Role:         "user",
			PasswordHash: string(hash),
		}
		ctrl.DB.Create(&user)
		createdUsers = append(createdUsers, user)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Users created",
		"count":   len(createdUsers),
		"users":   createdUsers,
	})
}

func (ctrl *UserController) GetDepartments(c *gin.Context) {
	var departments []string
	ctrl.DB.Model(&models.User{}).Distinct("department").Pluck("department", &departments)

	c.JSON(http.StatusOK, departments)
}
