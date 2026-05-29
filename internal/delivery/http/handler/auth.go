package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/predicta/predicta/internal/delivery/http/dto"
	"github.com/predicta/predicta/internal/delivery/http/middleware"
	"github.com/predicta/predicta/internal/domain/port"
	"github.com/predicta/predicta/internal/usecase"
)

type AuthHandler struct {
	auth usecase.AuthRegisterer
	login usecase.AuthLogin
	profile usecase.AuthProfileGetter
}

func NewAuthHandler(
	register usecase.AuthRegisterer,
	login usecase.AuthLogin,
	profile usecase.AuthProfileGetter,
) *AuthHandler {
	return &AuthHandler{
		auth:    register,
		login:   login,
		profile: profile,
	}
}

func (h *AuthHandler) RegisterPublic(r gin.IRoutes) {
	r.POST("/api/auth/register", h.register)
	r.POST("/api/auth/login", h.loginHandler)
}

func (h *AuthHandler) RegisterProtected(r *gin.RouterGroup) {
	r.GET("/api/auth/me", h.me)
}

func (h *AuthHandler) register(c *gin.Context) {
	var req dto.RegisterManagerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	err := h.auth.Register(c.Request.Context(), usecase.RegisterManagerCommand{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Password:     req.Password,
		TelegramNick: req.TelegramNick,
		Phone:        req.Phone,
		AvatarURL:    req.AvatarURL,
	})
	if err != nil {
		if errors.Is(err, port.ErrManagerExists) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "manager already registered"})
			return
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.MessageResponse{Message: "manager registered"})
}

func (h *AuthHandler) loginHandler(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	token, err := h.login.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, port.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{Token: token})
}

func (h *AuthHandler) me(c *gin.Context) {
	profile, err := h.profile.GetProfile(c.Request.Context(), middleware.ManagerID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ManagerProfileResponse{
		FirstName:         profile.FirstName,
		LastName:          profile.LastName,
		Email:             profile.Email,
		TelegramNick:      profile.TelegramNick,
		Phone:             profile.Phone,
		SubordinatesCount: profile.SubordinatesCount,
		AvatarURL:         profile.AvatarURL,
		JiraDisplayName:   profile.JiraDisplayName,
		JiraEmail:         profile.JiraEmail,
	})
}
