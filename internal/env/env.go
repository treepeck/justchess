package env

import (
	"log"
	"os"
	"strings"
)

// Load parses the specified file and sets environment variables for the current process.
// Accepted format for a variable: KEY=VALUE
// Comments which begin with '#' and empty lines are ignored.
func Load(envPath string) {
	env, err := os.ReadFile(envPath)
	if err != nil {
		log.Fatalf("%s file cannot be read %v", envPath, err)
	}

	for line := range strings.SplitSeq(string(env), "\n") {
		// Skip empty lines and comments.
		if len(line) < 3 || line[0] == '#' {
			continue
		}

		pair := strings.SplitN(line, "=", 2)
		// Skip malformed variable.
		if len(pair) != 2 {
			continue
		}

		os.Setenv(pair[0], pair[1])
	}
}
