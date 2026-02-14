package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load parses a dotenv-style file and sets values into process environment.
func Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid .env format at line %d", lineNumber)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return fmt.Errorf("empty key in .env at line %d", lineNumber)
		}

		value = strings.Trim(value, "\"")
		value = strings.Trim(value, "'")

		if setErr := os.Setenv(key, value); setErr != nil {
			return fmt.Errorf("failed to set env %s: %w", key, setErr)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}
