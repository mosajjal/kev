package client

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/mosajjal/kev/pkg/server"
)

func Run() {
	fmt.Println("Hello World")

}

func GetEnv(kevd string) map[string]string {
	p := server.ProcJSON{
		Cmdline: "string",
		Cwd:     "string",
		Exe:     "string",
		UID:     0,
		GID:     0,
		Env:     map[string]string{},
	}
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
