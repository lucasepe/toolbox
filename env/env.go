package env

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// FromReader read and parse an env file from
// an `io.Reader`, returning a map of keys and values.
func FromReader(r io.Reader) (envMap map[string]string, err error) {
	envMap = make(map[string]string)

	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return
	}

	for _, fullLine := range lines {
		if !isIgnoredLine(fullLine) {
			var key, value string
			key, value, err = parseLine(fullLine, envMap)
			if err != nil {
				return
			}
			envMap[key] = value
		}
	}
	return
}

// FromURL read and parse an env file from
// an HTTP URL, returning a map of keys and values.
func FromURL(url string) (envMap map[string]string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return FromReader(res.Body)
}

// FromFile read and parse an env file from
// a file, returning a map of keys and values.
func FromFile(filename string) (envMap map[string]string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return FromReader(file)
}

// Store store the map content into the os environment.
// if overload is true existing env variables will be overwritten.
func Store(envMap map[string]string, overload bool) {
	currentEnv := map[string]bool{}

	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	for key, value := range envMap {
		if !currentEnv[key] || overload {
			os.Setenv(key, value)
		}
	}
}

var (
	singleQuotesRegex  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRegex  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRegex        = regexp.MustCompile(`\\.`)
	unescapeCharsRegex = regexp.MustCompile(`\\([^$])`)
	exportRegex        = regexp.MustCompile(`^\s*(?:export\s+)?(.*?)\s*$`)
	expandVarRegex     = regexp.MustCompile(`(\\)?(\$)(\()?\{?([A-Z0-9_]+)?\}?`)
)

func parseLine(line string, envMap map[string]string) (string, string, error) {
	if len(line) == 0 {
		return "", "", errors.New("zero length string")
	}

	// ditch the comments (but keep quoted hashes)
	if strings.Contains(line, "#") {
		segmentsBetweenHashes := strings.Split(line, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, segment := range segmentsBetweenHashes {
			if strings.Count(segment, "\"") == 1 || strings.Count(segment, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, segment)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, segment)
			}
		}

		line = strings.Join(segmentsToKeep, "#")
	}

	firstEquals := strings.Index(line, "=")
	firstColon := strings.Index(line, ":")
	splitString := strings.SplitN(line, "=", 2)
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		//this is a yaml-style line
		splitString = strings.SplitN(line, ":", 2)
	}

	if len(splitString) != 2 {
		return "", "", errors.New("can't separate key from value")
	}

	// Parse the key
	key := splitString[0]
	key = strings.TrimPrefix(key, "export")
	key = strings.TrimSpace(key)
	key = exportRegex.ReplaceAllString(key, "$1")

	// Parse the value
	value, err := parseValue(splitString[1], envMap)

	return key, value, err
}

func parseValue(value string, envMap map[string]string) (res string, err error) {
	// trim
	res = strings.Trim(value, " ")
	if len(res) <= 1 {
		return res, nil
	}

	// check if we've got quoted values or possible escapes
	singleQuotes := singleQuotesRegex.FindStringSubmatch(res)
	doubleQuotes := doubleQuotesRegex.FindStringSubmatch(res)

	if singleQuotes != nil || doubleQuotes != nil {
		// pull the quotes off the edges
		res = res[1 : len(res)-1]
	}

	if doubleQuotes != nil {
		// expand newlines
		res = escapeRegex.ReplaceAllStringFunc(res, func(match string) string {
			c := strings.TrimPrefix(match, `\`)
			switch c {
			case "n":
				return "\n"
			case "r":
				return "\r"
			default:
				return match
			}
		})
		// unescape characters
		res = unescapeCharsRegex.ReplaceAllString(res, "$1")
	}

	if singleQuotes == nil {
		res = expandVariables(res, envMap)
	}

	return res, err
}

func expandVariables(v string, m map[string]string) string {
	return expandVarRegex.ReplaceAllStringFunc(v, func(s string) string {
		submatch := expandVarRegex.FindStringSubmatch(s)

		if submatch == nil {
			return s
		}
		if submatch[1] == "\\" || submatch[2] == "(" {
			return submatch[0][1:]
		} else if submatch[4] != "" {
			return m[submatch[4]]
		}
		return s
	})
}

func isIgnoredLine(line string) bool {
	trimmedLine := strings.TrimSpace(line)
	return len(trimmedLine) == 0 || strings.HasPrefix(trimmedLine, "#")
}
