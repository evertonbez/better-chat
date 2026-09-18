package handlers

import (
	"evertonbez/better-chat/internal/api/services"
	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	auth     *services.AuthService
	sessions *services.SessionService
	cookie   auth.CookieConfig
	secret   string
}

func NewAuthHandler(authService *services.AuthService, sessions *services.SessionService, cookie auth.CookieConfig, secret string) *AuthHandler {
	return &AuthHandler{
		auth:     authService,
		sessions: sessions,
		cookie:   cookie,
		secret:   secret,
	}
}

type signUpRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (h *AuthHandler) SignUp(c fiber.Ctx) error {
	var req signUpRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperr.InvalidRequest.Cause(err)
	}

	identity, err := h.auth.SignUp(c, services.SignUpParams{
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		return err
	}

	h.issue(c, identity)

	return c.Status(fiber.StatusCreated).JSON(newIdentityResponse(identity))
}

func (h *AuthHandler) SignIn(c fiber.Ctx) error {
	var req signInRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperr.InvalidRequest.Cause(err)
	}

	identity, err := h.auth.SignIn(c, services.SignInParams{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		return err
	}

	h.issue(c, identity)

	return c.JSON(newIdentityResponse(identity))
}

func (h *AuthHandler) SignOut(c fiber.Ctx) error {
	defer h.cookie.Clear(c)

	identity, ok := auth.FromContext(c)
	if !ok {
		return c.SendStatus(fiber.StatusNoContent)
	}

	if err := h.auth.SignOut(c, identity.Session.Token); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) SignOutAll(c fiber.Ctx) error {
	identity, ok := auth.FromContext(c)
	if !ok {
		return apperr.Unauthorized
	}

	if err := h.auth.SignOutAll(c, identity.User.ID); err != nil {
		return err
	}

	h.cookie.Clear(c)

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) GetSession(c fiber.Ctx) error {
	identity, ok := auth.FromContext(c)
	if !ok {
		return c.JSON(nil)
	}

	return c.JSON(newIdentityResponse(identity))
}

func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	var req changePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperr.InvalidRequest.Cause(err)
	}

	identity, ok := auth.FromContext(c)
	if !ok {
		return apperr.Unauthorized
	}

	if err := h.auth.ChangePassword(c, identity.User.ID, req.CurrentPassword, req.NewPassword); err != nil {
		return err
	}

	h.cookie.Clear(c)

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AuthHandler) issue(c fiber.Ctx, identity *auth.Identity) {
	h.cookie.Write(c, auth.SignToken(identity.Session.Token, h.secret), identity.Session.ExpiresAt)
}
