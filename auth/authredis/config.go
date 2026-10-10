package authredis

import (
	"regexp"
	"strconv"
	"time"

	. "github.com/nayefradwi/nayef_go_common/errors"
)

var prefixPattern = regexp.MustCompile(`^[a-z][a-z0-9_:]*$`)

const maxPrefixLen = 48

func validPrefix(prefix string) error {
	if len(prefix) > maxPrefixLen || !prefixPattern.MatchString(prefix) {
		return BadRequestError("prefix must be lowercase letters, digits, underscores and colons, at most 48 bytes")
	}

	return nil
}

// server clock, so every app instance compares against the same time
const luaNow = `
local function nowMs()
	local t = redis.call('TIME')
	return tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)
end
`

// PEXPIREAT GT skips keys without a ttl, so compare by hand
const luaExtend = `
local function extend(key, at)
	if redis.call('PEXPIRETIME', key) < tonumber(at) then
		redis.call('PEXPIREAT', key, at)
	end
end
`

func byteArgs(bs [][]byte) []any {
	args := make([]any, len(bs))
	for i, b := range bs {
		args[i] = b
	}
	return args
}

func parseMs(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.UnixMilli(ms), nil
}

func formatMs(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return strconv.FormatInt(t.UnixMilli(), 10)
}
