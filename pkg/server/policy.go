package server

import "regexp"

// a simple implementation of the policy interface which links a set of processess to a set of keys in the kv store as allowed

type Process interface {
	GetMachineId() string
	GetCmdline() string
	GetEnvs() map[string]string
	GetCwd() string
	GetExe() string
	GetUid() uint32
	GetGid() uint32
}

type PolicyEngine struct {
	kv       KV
	policies []Policy
}

// Policy is an individual logic that associates a list of keys with a process
type Policy interface {
	AllowedKeys(Process) map[string]struct{}
}

// NewPolicyEngine creates a new policy engine based on the KV and the list of policies. note that policies are allow-only
// and not allow-deny. if a key doesn't match any of the policies, it won't be allowed
func NewPolicyEngine(kv KV, policies ...Policy) PolicyEngine {
	return PolicyEngine{
		kv:       kv,
		policies: policies,
	}
}

func (p *PolicyEngine) AllowedEnv(process Process) map[string]string {
	allowedKeys := map[string]struct{}{}
	for _, policy := range p.policies {
		for key := range policy.AllowedKeys(process) {
			allowedKeys[key] = struct{}{}
		}
	}
	// grab all the values from kv and map out the environment
	allowedEnv := map[string]string{}
	for _, key := range p.kv.GetAllKeys() {
		if _, ok := allowedKeys[key]; ok {
			if value, err := p.kv.Get(key); err == nil {
				allowedEnv[key] = value
			}
			//TODO: handle error
		}
	}
	return allowedEnv
}

// a simple policy implementation based on the cmdline regex. gets a list of regex and list of keys allowed for that regex
type CmdlinePolicy struct {
	cmd         regexp.Regexp
	allowedKeys map[string]struct{}
}

func (c *CmdlinePolicy) AllowedKeys(process Process) map[string]struct{} {
	allowedKeys := map[string]struct{}{}
	for envKey := range c.allowedKeys {
		// regex match against the cmdline
		if c.cmd.MatchString(process.GetCmdline()) {
			allowedKeys[envKey] = struct{}{}
		}
	}
	return allowedKeys
}

func NewCmdlinePolicy(cmdRE string, allowedKeys ...string) Policy {
	p := CmdlinePolicy{
		allowedKeys: map[string]struct{}{},
		cmd:         *regexp.MustCompile(cmdRE),
	}
	for _, key := range allowedKeys {
		p.allowedKeys[key] = struct{}{}
	}
	return &p
}
