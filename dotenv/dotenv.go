package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

func load(path string) error {
	varRegexp := regexp.MustCompile(`\${([a-zA-Z0-9_]+)}`) // Match ${VAR}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file
	scanner := bufio.NewScanner(file) // Create a scanner
	envMap := make(map[string]string) // Map to store the environment variables
	var line, k, v string             // Variables to store the line, key and value

	for scanner.Scan() {
		line = scanner.Text()
		if varRegexp.MatchString(line) {
			// Check if the variable is defined (warn if not)
			varName := varRegexp.FindString(line)[2 : len(varRegexp.FindString(line))-1]
			if _, ok := envMap[varName]; !ok {
				slog.Warn(fmt.Sprintf("Undefined variable: %s in line %s", varRegexp.FindString(line), line))
			}
			line = varRegexp.ReplaceAllStringFunc(line, func(s string) string {
				return envMap[s[2:len(s)-1]]
			})
		}
		// Parse the line
		k, v, err = parse(line)
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid line: %s", line))
		}
		if k == "" || v == "" { // Skip empty lines
			continue
		}
		envMap[k] = v
	}
	// Set the environment variables
	for k, v := range envMap {
		os.Setenv(k, v)
	}
	return nil
}

func parse(line string) (string, string, error) {
	s := strings.TrimSpace(line)              // Trim the line
	if strings.HasPrefix(s, "#") || s == "" { // Check if the line is a comment or empty
		return "", "", nil
	}
	if !strings.Contains(s, "=") { // Check if the line is a key value pair
		return "", "", errors.New("No '=' in line")
	}
	keyValuePair := regexp.MustCompile(`\s*=\s*`).Split(s, 2) // Split the line by the first '='
	if len(keyValuePair) == 2 {
		// Return key and value (remove comments if any)
		value := strings.Split(keyValuePair[1], "#")[0]
		// Remove quotes if any
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}
		return strings.TrimSpace(keyValuePair[0]), strings.TrimSpace(value), nil
	}
	return "", "", nil
}

func Load(paths ...string) error {
	if len(paths) == 0 {
		paths = append(paths, ".env")
	}
	for _, path := range paths {
		err := load(path)
		if err != nil {
			return err
		}
	}
	return nil
}
