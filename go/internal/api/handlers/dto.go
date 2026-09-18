package handlers

import (
	"time"

	"evertonbez/better-chat/internal/auth"
)

type userResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Image         string    `json:"image,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type sessionResponse struct {
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress,omitempty"`
	UserAgent string    `json:"userAgent,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type identityResponse struct {
	User    userResponse    `json:"user"`
	Session sessionResponse `json:"session"`
}

type sessionHistoryResponse struct {
	Active    bool       `json:"active"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt"`
	IPAddress string     `json:"ipAddress,omitempty"`
	UserAgent string     `json:"userAgent,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

func newUserResponse(u auth.User) userResponse {
	return userResponse{
		ID:            u.UID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
	}
}

func newSessionResponse(s auth.Session) sessionResponse {
	return sessionResponse{
		ExpiresAt: s.ExpiresAt,
		IPAddress: s.IPAddress,
		UserAgent: s.UserAgent,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func newIdentityResponse(id *auth.Identity) identityResponse {
	return identityResponse{
		User:    newUserResponse(id.User),
		Session: newSessionResponse(id.Session),
	}
}

func newSessionHistoryResponse(sessions []auth.Session, now time.Time) []sessionHistoryResponse {
	out := make([]sessionHistoryResponse, 0, len(sessions))

	for _, s := range sessions {
		item := sessionHistoryResponse{
			Active:    !s.Revoked() && !s.Expired(now),
			ExpiresAt: s.ExpiresAt,
			IPAddress: s.IPAddress,
			UserAgent: s.UserAgent,
			CreatedAt: s.CreatedAt,
		}
		if s.Revoked() {
			revokedAt := s.RevokedAt
			item.RevokedAt = &revokedAt
		}
		out = append(out, item)
	}

	return out
}
