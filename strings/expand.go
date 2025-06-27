//
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const maxExpansionDepth = 64

// expand recursively expands expressions of '(prefix)key(postfix)' to their corresponding values.
// The function keeps track of the keys that were already expanded and stops if it
// detects a circular reference or a malformed expression of the form '(prefix)key'.
func expand0(s string, keys []string, prefix, postfix string, values map[string]string) (string, error) {
	if len(keys) > maxExpansionDepth {
		return "", fmt.Errorf("expansion too deep")
	}

	for {
		start := strings.Index(s, prefix)
		if start == -1 {
			return s, nil
		}

		keyStart := start + len(prefix)
		keyLen := strings.Index(s[keyStart:], postfix)
		if keyLen == -1 {
			return "", fmt.Errorf("malformed expression")
		}

		end := keyStart + keyLen + len(postfix) - 1
		key := s[keyStart : keyStart+keyLen]

		// fmt.Printf("s:%q pp:%q start:%d end:%d keyStart:%d keyLen:%d key:%q\n", s, prefix + "..." + postfix, start, end, keyStart, keyLen, key)

		for _, k := range keys {
			if key == k {
				var b bytes.Buffer
				b.WriteString("circular reference in:\n")
				for _, k1 := range keys {
					fmt.Fprintf(&b, "%s=%s\n", k1, values[k1])
				}
				return "", fmt.Errorf(b.String())
			}
		}

		val, ok := values[key]
		if !ok {
			val = os.Getenv(key)
		}
		new_val, err := expand0(val, append(keys, key), prefix, postfix, values)
		if err != nil {
			return "", err
		}
		s = s[:start] + new_val + s[end+1:]
	}
	return s, nil
}

type Resolver func(justKey, fullMatch string, logger *log.Entry) (string, error)

// warning: this is a non-recursive expander - it will not resolve expanded values
func Expand(source string, prefix, postfix string, resolver Resolver, logger *log.Entry) (string, error) {
	already := ""
	for {
		start := strings.Index(source, prefix)
		if start < 0 {
			return already + source, nil
		}

		keyStart := start + len(prefix)
		keyLen := strings.Index(source[keyStart:], postfix)

		if keyLen < 0 {
			// there is no postfix found - return as it is
			logger.Warningf("found `prefix` but not `postfix` - possible but unlikely schenario...")
			return already + source, nil
		}

		end := keyStart + keyLen + len(postfix) - 1
		key := source[keyStart : keyStart+keyLen]

		val, err := resolver(key, prefix+key+postfix, logger)
		if err != nil {
			logger.WithError(err).Errorf("the provided `resolver` failed to find a match for the key (%s)", key)
			return "", err
		}

		already += source[:start] + val
		source = source[end+1:]
	}
	// return source, nil
}

const (
	what = `
		admin.host: "${SESSIONNAME}${USERNAME}${SESSIONNAME}"
		admin.port: "${NotExistingOne}"
		admin.database: "$incomplete$$${abc${"
		admin.user: "$$$"
		admin.password: "&5mXDYW6WZyT>P$VMrY(N+-+?ZxXHrpy"
`

	prefix  = `${`
	postfix = `}`
)

func resolver(justKey, fullMatch string, logger *log.Entry) (string, error) {
	maps2, found := os.LookupEnv(justKey)
	if found {
		return maps2, nil
	}
	logger.Warningf("failed to find env var (%s) - leaving it intact", justKey)
	return fullMatch, nil
}

func main() {
	logger := log.New().WithField("what", "test")
	back, err := Expand(what, prefix, postfix, resolver, logger)
	if err != nil {
		fmt.Printf("error: %v", err)
	} else {
		if back == what {
			fmt.Printf("results are the same\n")
		}
		fmt.Printf("result: %s", back)
	}
}
