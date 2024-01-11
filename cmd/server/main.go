package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/mosajjal/kev/pkg/server"
	"github.com/rs/zerolog"

	"github.com/spf13/cobra"
)

var nocolorLog = strings.ToLower(os.Getenv("NO_COLOR")) == "true"
var logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: nocolorLog})

var (
	version string = "UNKNOWN"
	commit  string = "NOT_PROVIDED"
)

//go:embed config.defaults.yaml
var defaultConfig []byte

func main() {

	cmd := &cobra.Command{
		Use:   "kevd",
		Short: "kevd is awesome",
		Long:  `kevd is the best CLI ever!`,
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	flags := cmd.Flags()

	logLevel := flags.StringP("loglevel", "v", "info", "log level (debug, info, warn, error, fatal, panic)")
	config := flags.StringP("config", "c", "$HOME/.kevd.yaml", "path to YAML configuration file")
	_ = flags.BoolP("defaultconfig", "d", false, "write default config to $HOME/.kevd.yaml")

	if err := cmd.Execute(); err != nil {
		logger.Error().Msgf("failed to execute command: %s", err)
		return
	}

	// set up log level
	lvl, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		logger.Fatal().Msgf("failed to parse log level: %s", err)
	}
	zerolog.SetGlobalLevel(lvl)

	if !flags.Changed("config") {
		home, err := os.UserHomeDir()
		if err != nil {
			logger.Fatal().Msgf("failed to get user home directory: %s", err)
		}
		*config = filepath.Join(home, ".kevd.yaml")
	}
	if flags.Changed("help") {
		return
	}
	if flags.Changed("version") {
		fmt.Printf("kevd version %s, commit %s\n", version, commit)
		return
	}

	// load the default config
	if flags.Changed("defaultconfig") {
		if err := os.WriteFile(*config, defaultConfig, 0644); err != nil {
			logger.Fatal().Msgf("failed to write default config: %s", err)
		}
		logger.Info().Msgf("wrote default config to %s", *config)
		return
	}

	k := koanf.New(".")
	// load the defaults first, so if the config file is missing some values, we can fall back to the defaults
	if err := k.Load(rawbytes.Provider(defaultConfig), yaml.Parser()); err != nil {
		logger.Fatal().Msgf("failed to load default config: %s", err)
	}

	if err := k.Load(file.Provider(*config), yaml.Parser()); err != nil {
		logger.Fatal().Msgf("failed to load config file: %s", err)
	}

	// cut the KV config
	KVKoanf := k.Cut("kv")
	// check for badger since only badger is implemented
	if KVKoanf.MustString("type") != "badger" {
		logger.Fatal().Msgf("unsupported kv type: %s", KVKoanf.Get("type"))
	}
	// set up badger kv
	badgerOpts := badger.DefaultOptions(KVKoanf.MustString("settings.path"))
	if KVKoanf.Bool("settings.encryption") {
		badgerOpts.WithEncryptionKey([]byte(KVKoanf.MustString("settings.encryption_key")))
	}
	KV, err := server.NewBadgerKV(badgerOpts)
	if err != nil {
		logger.Fatal().Msgf("failed to set up badger: %s", err)
	}

	// see if kv rest is enabled
	if KVKoanf.Bool("rest.enabled") {
		parsedURL, err := url.Parse(KVKoanf.MustString("rest.listen"))
		if err != nil {
			logger.Fatal().Msgf("failed to parse listen address: %s", err)
		}

		var l net.Listener

		switch parsedURL.Scheme {
		case "unix":
			l, err = net.Listen("unix", parsedURL.Path)
			if err != nil {
				logger.Fatal().Msgf("failed to listen: %s", err)
			}
			//BUG: the unix socket doesn't get auto deleted on ctrl-c

		case "tcp":
			l, err = net.Listen("tcp", parsedURL.Host)
			if err != nil {
				logger.Fatal().Msgf("failed to listen: %s", err)
			}
		default:
			logger.Fatal().Msgf("unsupported scheme: %s", parsedURL.Scheme)

		}

		if err := server.ServeKVRest(KV, l,
			KVKoanf.MustString("rest.base_path"),
			KVKoanf.StringMap("rest.auth.users")); err != nil {
			logger.Fatal().Msgf("failed to set up rest: %s", err)
		}
	}

	// cut the policy config
	policies := []server.Policy{}
	for _, policy := range k.Slices("policies") {
		if policy.MustString("type") != "cmdline" {
			logger.Fatal().Msgf("unsupported policy type: %s", policy.String("type"))
		}
		policies = append(policies, server.NewCmdlinePolicy(
			policy.MustString("settings.cmd"),
			policy.MustStrings("settings.allowed_keys")...,
		))
	}
	engine := server.NewPolicyEngine(KV, policies...)

	// cut the rest config
	restK := k.Cut("rest")
	parsedURL, err := url.Parse(restK.MustString("listen"))
	if err != nil {
		logger.Fatal().Msgf("failed to parse listen address: %s", err)
	}
	var l net.Listener

	switch parsedURL.Scheme {
	case "unix":
		l, err = net.Listen("unix", parsedURL.Path)
		if err != nil {
			logger.Fatal().Msgf("failed to listen: %s", err)
		}
		//BUG: the unix socket doesn't get auto deleted on ctrl-c

	case "tcp":
		l, err = net.Listen("tcp", parsedURL.Host)
		if err != nil {
			logger.Fatal().Msgf("failed to listen: %s", err)
		}
	default:
		logger.Fatal().Msgf("unsupported scheme: %s", parsedURL.Scheme)

	}
	// create a http listener
	logger.Fatal().Msgf("failed to serve: %s", server.ServeRest(
		engine,
		l,
		restK.MustString("base_path"),
		restK.StringMap("auth.users"),
	))
}
