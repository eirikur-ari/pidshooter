// Package cli implements the driving adapter that parses os.Args and calls the application service.
package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/eirikur-ari/pidshooter/internal/domain/game/ports/driving"
	procdriven "github.com/eirikur-ari/pidshooter/internal/domain/process/ports/driven"
)

const usage = `pidshooter - First Person PID Shooter

Usage: pidshooter <pattern> [pattern2] [pattern3...] [--confirm] [--speed=N] [--time=N]

Arguments:
  pattern     One or more substrings to match against process names
  --confirm   Ask for confirmation before killing (optional)
  --speed=N   Speed multiplier (default: 2.0, range: 0.1-5.0)
  --time=N    Time limit in seconds (default: 30, 0 = no limit)

Controls:
  Click       Kill the process under cursor
  +/-         Speed up / slow down
  q / Escape  Quit the program

Examples:
  pidshooter firefox
  pidshooter chrome firefox node
  pidshooter firefox --confirm
  pidshooter node --speed=2.5 --time=60
  pidshooter node --time=0`

// CLI is the driving adapter that translates command-line arguments to application calls.
type CLI struct {
	service driving.GameService
}

// New returns a CLI adapter wrapping the given application service.
func New(service driving.GameService) *CLI {
	return &CLI{service: service}
}

// Run parses args and calls the application service.
func (c *CLI) Run(args []string) error {
	patterns, confirmMode, speed, timeLimit, err := parseArgs(args)
	if err != nil {
		return err
	}
	return c.service.Play(driving.Config{
		Patterns:    patterns,
		ConfirmMode: confirmMode,
		Speed:       speed,
		TimeLimit:   timeLimit,
	})
}

func parseArgs(args []string) ([]string, bool, float64, int, error) {
	if len(args) == 0 {
		fmt.Println(usage)
		os.Exit(0)
	}

	var patterns []string
	var confirmMode bool
	speed := 2.0
	timeLimit := 30

	for _, arg := range args {
		switch {
		case arg == "--confirm":
			confirmMode = true
		case strings.HasPrefix(arg, "--speed="):
			val := strings.TrimPrefix(arg, "--speed=")
			s, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return nil, false, 0, 0, fmt.Errorf("invalid speed value: %s\nRun 'pidshooter --help' for usage", val)
			}
			if s < 0.1 || s > 5.0 {
				return nil, false, 0, 0, fmt.Errorf("speed must be between 0.1 and 5.0, got: %s", val)
			}
			speed = s
		case strings.HasPrefix(arg, "--time="):
			val := strings.TrimPrefix(arg, "--time=")
			t, err := strconv.Atoi(val)
			if err != nil {
				return nil, false, 0, 0, fmt.Errorf("invalid time value: %s\nRun 'pidshooter --help' for usage", val)
			}
			if t < 0 {
				return nil, false, 0, 0, fmt.Errorf("time must be 0 or positive, got: %s", val)
			}
			timeLimit = t
		case arg == "--help" || arg == "-h":
			fmt.Println(usage)
			os.Exit(0)
		case len(arg) > 0 && arg[0] == '-':
			return nil, false, 0, 0, fmt.Errorf("unknown flag: %s\nRun 'pidshooter --help' for usage", arg)
		default:
			if len(arg) < procdriven.MinPatternLength {
				return nil, false, 0, 0, fmt.Errorf("search pattern %q must be at least %d characters", arg, procdriven.MinPatternLength)
			}
			if len(arg) > procdriven.MaxPatternLength {
				return nil, false, 0, 0, fmt.Errorf("search pattern %q exceeds maximum length of %d characters", arg, procdriven.MaxPatternLength)
			}
			patterns = append(patterns, arg)
		}
	}

	if len(patterns) == 0 {
		return nil, false, 0, 0, fmt.Errorf("at least one search pattern is required\nRun 'pidshooter --help' for usage")
	}

	return patterns, confirmMode, speed, timeLimit, nil
}
