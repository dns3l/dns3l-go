package client

import (
	"fmt"
	"runtime/debug"
)

const modulePath = "github.com/dns3l/dns3l-go"

func GetVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// if this is the main module
	if info.Main.Path == modulePath && info.Main.Version != "" {
		return stripV(info.Main.Version)
	}

	// if a module depends this one
	for _, dep := range info.Deps {
		if dep.Path == modulePath && dep.Version != "" {
			return stripV(dep.Version)
		}
	}

	return "devel"
}

func stripV(verStr string) string {
	if len(verStr) > 0 && verStr[0] == 'v' {
		return verStr[1:]
	}
	return verStr
}

func GetDefaultUserAgent() string {
	return fmt.Sprintf("dns3l-go/%s (https://%s)", GetVersion(), modulePath)
}
