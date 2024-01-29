package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/go-resty/resty/v2"
	"github.com/mosajjal/kev/pkg/server"
)

func getParamsFromArgs(args ...string) server.ProcJSON {
	id, _ := machineid.ID()
	if id == "" {
		id = "NO_ID"
	}
	cwd, _ := os.Getwd()
	if cwd == "" {
		cwd = "NO_CWD"
	}
	envs := make(map[string]string)
	oldEnv := os.Environ()
	for _, env := range oldEnv {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envs[parts[0]] = parts[1]
	}

	// get the current process
	p := server.ProcJSON{
		Cmdline:   strings.Join(args, " "),
		MachineId: id,
		Cwd:       cwd,
		Exe:       os.Args[0],
		UID:       os.Geteuid(),
		GID:       os.Getegid(),
		Env:       envs,
	}
	return p
}

func GetEnv(kevd string, args ...string) map[string]string {
	p := getParamsFromArgs(args...)

	// send the process to the server
	pData, _ := json.Marshal(p)

	resp, err := resty.New().R().SetBody(pData).Get(kevd)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// response looks like this:
	// { "key1": "value1", "key2": "value2" } or {}
	env := make(map[string]string)
	if resp.Result() == nil {
		return env
	}

	for k, v := range resp.Result().(map[string]interface{}) {
		env[k] = v.(string)
	}
	return env
}
