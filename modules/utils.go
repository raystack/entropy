package modules

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
)

func CloneAndMergeMaps(m1, m2 map[string]string) map[string]string {
	res := map[string]string{}
	for k, v := range m1 {
		res[k] = v
	}
	for k, v := range m2 {
		res[k] = v
	}
	return res
}

func MustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func SafeName(name string, suffix string, maxLen int) string {
	const randomHashLen = 6
	// remove suffix if already there.
	name = strings.TrimSuffix(name, suffix)

	if len(name) <= maxLen-len(suffix) {
		return name + suffix
	}

	val := sha256.Sum256([]byte(name))
	hash := fmt.Sprintf("%x", val)
	suffix = fmt.Sprintf("-%s%s", hash[:randomHashLen], suffix)

	// truncate and make room for the suffix. also trim any leading, trailing
	// hyphens to prevent '--' (not allowed in deployment names).
	truncLen := maxLen - len(suffix)
	truncated := name[0:truncLen]
	truncated = strings.Trim(truncated, "-")
	return truncated + suffix
}
