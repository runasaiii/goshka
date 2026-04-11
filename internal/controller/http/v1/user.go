package v1
import (
	"errors"
	"net/http"

	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/pkg/logger"
	"practice-7/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)


type userRoutes struct {
	t usecase.UserInterface
	l logger.Interface
}

func newUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface) {
	r := &userRoutes{t, l}
	h := handler.Group("/users")

	h.POST("/login", r.Login)

	lim := h.Group("")
	lim.Use(DefaultRateLimit())
	lim.POST("/", r.RegisterUser)
	lim.POST("/verify-email", r.VerifyEmail)
	lim.POST("/refresh", r.Refresh)

	protected := lim.Group("")
	protected.Use(JWTAuthMiddleware())
	protected.GET("/me", r.GetMe)

	admin := lim.Group("")
	admin.Use(JWTAuthMiddleware(), RoleMiddleware("admin"))
	admin.PATCH("/promote/:id", r.PromoteUser)
}

func (r *userRoutes) RegisterUser(c *gin.Context) {
	var createUserDTO entity.CreateUserDTO
	if err := c.ShouldBindJSON(&createUserDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, err := utils.HashPassword(createUserDTO.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}
	user := entity.User{
		Username: createUserDTO.Username,
		Email:    createUserDTO.Email,
		Password: hashedPassword,
		Role:     "user",
	}
	createdUser, sessionID, err := r.t.RegisterUser(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":    "User registered successfully. Please check your email for verification code.",
		"session_id": sessionID,
		"user":       createdUser,
	})
}

func (r *userRoutes) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, access, refresh, err := r.t.Login(input.Username, input.Password)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		if errors.Is(err, usecase.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Please verify your email before logging in"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          user,
	})
}

func (r *userRoutes) VerifyEmail(c *gin.Context) {
	var dto entity.VerifyEmailDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := r.t.VerifyEmail(dto.Email, dto.Code); err != nil {
		if errors.Is(err, usecase.ErrInvalidVerification) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or verification code"})
			return
		}
		if errors.Is(err, usecase.ErrVerificationExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code has expired"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully. You can log in now"})
}

func (r *userRoutes) Refresh(c *gin.Context) {
	var dto entity.RefreshTokenDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	access, err := r.t.RefreshAccessToken(dto.RefreshToken)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidRefreshToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}
		if errors.Is(err, usecase.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is not verified"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": access})
}

func (r *userRoutes) GetMe(c *gin.Context) {
	v, ok := c.Get(CtxJWTUserID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	idStr, _ := v.(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id in token"})
		return
	}
	user, err := r.t.GetUserByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"verified": user.Verified,
	})
}

func (r *userRoutes) PromoteUser(c *gin.Context) {
	idStr := c.Param("id")
	targetID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := r.t.PromoteUserToAdmin(targetID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "user promoted to admin",
		"id":      targetID.String(),
		"role":    "admin",
	})
}
