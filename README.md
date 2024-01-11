# KEV 

Key Environment Variable 


A WIP tool to dynamically inject (secret) environment variables into programs.

Components:

KEV is implemented as a client/server tool. `kevd` is the engine that maintains the policies as well as the KV store of the environment variables. and `kev` is the binary used regularly to dynamically inject proper env variables to a running process. 

```bash
# will return an error saying you're not authenticated
$ aws iam list-roles
## login failure
# will dynamically inject AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY then runs the command
$ kev aws iam list-roles
## successful login 
```

## KEVD

- an encrypted KV store for all environment variables. currently, BadgerDB is chosen but the design is modular enough to add your own
- a ReST APIs endpoint to add new environment variables to the KV (can listen on unix socket or tcp with optional auth)
- a policy engine that maps process attributes to a list of env variables they can access. currently, the `cmdline`, which is the full executable path is implemented.

an example of the policy could be a regex of your script's full path and `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` as allowed environments. refer to the `config.defaults.yaml` in `cmd/server` for more info. 

## KEV

a simple CLI application that acts as a wrapper for other Linux commands. it accepts an environment variables (`KEVD_URI`) as a configuration. when a command is run with `kev`, `kev` will curate a list of attributes associated with the process that's about to be run as well as `kev`'s own environemnt. the policy engine will make a decision on which environment variables needs to be added to the running command, `kev` will add them into the env variables and proceeds to execute it. 