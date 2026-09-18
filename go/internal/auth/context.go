package auth

import "github.com/gofiber/fiber/v3"

type identityKey struct{}

func SetIdentity(c fiber.Ctx, id *Identity) {
	fiber.Locals[*Identity](c, identityKey{}, id)
}

func FromContext(c fiber.Ctx) (*Identity, bool) {
	id := fiber.Locals[*Identity](c, identityKey{})
	return id, id != nil
}

func MustUser(c fiber.Ctx) User {
	id, ok := FromContext(c)
	if !ok {
		panic("auth: MustUser called on an unauthenticated request")
	}
	return id.User
}
