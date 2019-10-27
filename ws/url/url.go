package url

import (
	"strings"
)

func ModifyUrl(u string) string {
	if len(u) == 0 {
		return u
	}

	if len(u) > 0 && u[0] == ':' {
		return "ws://127.0.0.1" + u
	}

	if len(u) > 0 && u[0] == '/' {
		return "ws://127.0.0.1:80" + u
	}

	if !strings.HasPrefix(u, "ws") {
		return "ws://" + u
	}

	return u
}
