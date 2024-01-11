package server

import (
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// a simple ReST API to interact with the kv store

type envJSON struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ServeKVRest exposes a rest API to set environment variables for the KV store
// for security reasons, it does not allow reading the environment
func ServeKVRest(kv KV, l net.Listener, basePath string, users map[string]string) error {
	defer l.Close()
	fiberApp := fiber.New()

	if users != nil {
		if len(users) > 0 {
			fiberApp.Use(basicauth.New(basicauth.Config{
				Users: users,
			}))
		}
	}

	// sets a new environment variable in the kv store
	// the key and value are both case-sentitive
	fiberApp.Put(basePath, func(c *fiber.Ctx) error {
		env := new(envJSON)
		if c.BodyParser(&env); env == nil {
			return c.SendStatus(500)
		}
		if err := kv.Set(env.Key, env.Value); err != nil {
			return c.SendStatus(500)
		}
		return c.SendStatus(201)
	})
	return fiberApp.Listener(l)
}
