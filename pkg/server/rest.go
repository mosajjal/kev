package server

import (
	"net"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
)

// exposes a rest API for the client to connect and grab the environment variable based on the policy
// the calls:
// GET /env
// with the data:
// {
// "process":{
// "cmdline":"string",
// "cwd":"string",
// "exe":"string",
// "uid":"int",
// "gid":"int",
// "env":{
// 		"key":"value"
//     }
//   }
// }

type ProcJSON struct {
	Cmdline string            `json:"cmdline"`
	Cwd     string            `json:"cwd"`
	Exe     string            `json:"exe"`
	UID     int               `json:"uid"`
	GID     int               `json:"gid"`
	Env     map[string]string `json:"env"`
}

// implement the process interface for procJSON
func (p ProcJSON) GetCmdline() string {
	return p.Cmdline
}
func (p ProcJSON) GetEnvs() map[string]string {
	return p.Env
}
func (p ProcJSON) GetCwd() string {
	return p.Cwd
}
func (p ProcJSON) GetExe() string {
	return p.Exe
}
func (p ProcJSON) GetUid() uint32 {
	return uint32(p.UID)
}
func (p ProcJSON) GetGid() uint32 {
	return uint32(p.GID)
}

func ServeRest(p PolicyEngine, l net.Listener, basePath string, users map[string]string) error {
	defer l.Close()
	fiberApp := fiber.New()

	if users != nil {
		if len(users) > 0 {
			fiberApp.Use(basicauth.New(basicauth.Config{
				Users: users,
			}))
		}
	}

	fiberApp.Get(basePath, func(c *fiber.Ctx) error {
		proc := new(ProcJSON)
		if c.BodyParser(&proc); proc == nil {
			return c.SendStatus(500)
		}
		allowedEnv := p.AllowedEnv(proc)
		return c.JSON(allowedEnv)
	})
	return fiberApp.Listener(l)
}
