package config

import (
	"os"
	"strings"
)

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = homeDir + path[1:]
		}
	}
	return os.ExpandEnv(path)
}

func ExpandConfigPaths(cfg *Config) {
	for agentName, agent := range cfg.Agents {
		for i, m := range agent.Mounts {
			agent.Mounts[i].Source = ExpandPath(m.Source)
			agent.Mounts[i].Destination = os.ExpandEnv(m.Destination)
		}

		for i, fc := range agent.FilesToCopy {
			agent.FilesToCopy[i].Source = ExpandPath(fc.Source)
			agent.FilesToCopy[i].Destination = os.ExpandEnv(fc.Destination)
		}

		for i, wp := range agent.WritablePaths {
			agent.WritablePaths[i].Destination = os.ExpandEnv(wp.Destination)
		}

		for i, tm := range agent.TmpfsMounts {
			agent.TmpfsMounts[i].Destination = os.ExpandEnv(tm.Destination)
		}

		cfg.Agents[agentName] = agent
	}
}
