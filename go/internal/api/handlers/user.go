package handlers

import (
	"time"

	"evertonbez/better-chat/internal/api/services"
	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/auth"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	users    *services.UserService
	sessions *services.SessionService
}

func NewUserHandler(users *services.UserService, sessions *services.SessionService) *UserHandler {
	return &UserHandler{users: users, sessions: sessions}
}

type updateProfileRequest struct {
	Name  *string `json:"name"`
	Image *string `json:"image"`
}

func (h *UserHandler) Me(c fiber.Ctx) error {
	identity, ok := auth.FromContext(c)
	if !ok {
		return apperr.Unauthorized
	}

	return c.JSON(newUserResponse(identity.User))
}

func (h *UserHandler) UpdateMe(c fiber.Ctx) error {
	var req updateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperr.InvalidRequest.Cause(err)
	}

	identity, ok := auth.FromContext(c)
	if !ok {
		return apperr.Unauthorized
	}

	user, err := h.users.UpdateProfile(c, identity.User.ID, services.UpdateProfileParams{
		Name:  req.Name,
		Image: req.Image,
	})
	if err != nil {
		return err
	}

	return c.JSON(newUserResponse(*user))
}

func (h *UserHandler) Sessions(c fiber.Ctx) error {
	identity, ok := auth.FromContext(c)
	if !ok {
		return apperr.Unauthorized
	}

	sessions, err := h.sessions.History(c, identity.User.ID)
	if err != nil {
		return err
	}

	return c.JSON(newSessionHistoryResponse(sessions, time.Now()))
}
