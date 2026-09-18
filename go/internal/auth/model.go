package auth

import (
	"time"

	"evertonbez/better-chat/gen/dbstore"
)

const ProviderCredential = "credential"

type User struct {
	ID            int64     `json:"id"`
	UID           string    `json:"uid"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Image         string    `json:"image"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type Session struct {
	ID        int64     `json:"id"`
	Token     string    `json:"-"`
	UserID    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
	IPAddress string    `json:"ipAddress"`
	UserAgent string    `json:"userAgent"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	RevokedAt time.Time `json:"revokedAt"`
}

func (s Session) Expired(now time.Time) bool { return !now.Before(s.ExpiresAt) }

func (s Session) Revoked() bool { return !s.RevokedAt.IsZero() }

func (s Session) NeedsRefresh(now time.Time, updateAge time.Duration) bool {
	return now.Sub(s.UpdatedAt) >= updateAge
}

type Identity struct {
	Session Session `json:"session"`
	User    User    `json:"user"`
}

func UserFromRow(u dbstore.User) User {
	return User{
		ID:            u.ID,
		UID:           u.Uid.String,
		Name:          u.Name.String,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image.String,
		CreatedAt:     u.CreatedAt.Time,
		UpdatedAt:     u.UpdatedAt.Time,
	}
}

func SessionFromRow(s dbstore.Session) Session {
	return Session{
		ID:        s.ID,
		Token:     s.Token,
		UserID:    s.UserID,
		ExpiresAt: s.ExpiresAt.Time,
		IPAddress: s.IpAddress.String,
		UserAgent: s.UserAgent.String,
		CreatedAt: s.CreatedAt.Time,
		UpdatedAt: s.UpdatedAt.Time,
		RevokedAt: s.RevokedAt.Time,
	}
}
