package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/mosajjal/kev/pkg/client"
)

func main() {
	// environment variables for the client
	// KEVD_URI=http://localhost:8080

	kevdURI := os.Getenv("KEVD_URI")
	if kevdURI == "" {
		log.Fatalln("KEVD_URI is not set")
	}
	// create a new client
	newEnvs := client.GetEnv(kevdURI)
	// inject the new environment variables
	oldEnv := os.Environ()
	for _, env := range oldEnv {
		parts := strings.SplitN(env, "=", 2)
		if _, ok := newEnvs[parts[0]]; ok {
			os.Setenv(parts[0], newEnvs[parts[0]])
		}
	}

	cmdenv := make([]string, 0)
	for k, v := range newEnvs {
		cmdenv = append(cmdenv, fmt.Sprintf("%s=%s", k, v))
	}

	if len(os.Args) < 2 {
		log.Fatalln("No command provided")
	}

	// run everything else as os.exec with the new env
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Env = append(os.Environ(), cmdenv...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to run command: %v", err)
	}

}
