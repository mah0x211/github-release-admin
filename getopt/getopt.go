package getopt

import (
	"strings"
)

type Option interface {
	SetArg(v string) bool
	SetFlag(v string) bool
	SetKeyValue(k, v, arg string) bool
}

func Parse(o Option, args []string) []string {
	for i, arg := range args {
		var next bool
		arg = strings.TrimSpace(arg)
		if !strings.HasPrefix(arg, "--") {
			next = o.SetArg(arg)
		} else if kv := strings.SplitN(arg, "=", 2); len(kv) == 2 {
			next = o.SetKeyValue(kv[0], strings.TrimSpace(kv[1]), arg)
		} else {
			next = o.SetFlag(arg)
		}

		if !next {
			return args[i+1:]
		}
	}

	return nil
}
