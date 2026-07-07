// Package cli implements the delivery adapter that parses os.Args and calls the application service.
package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/eirikur-ari/pidshooter/internal/application/contract/inbound"
	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

var errUsage = errors.New("usage")

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

// CLI is the delivery adapter that translates command-line arguments to application calls.
type CLI struct {
	service inbound.GamePlay
}

// New returns a CLI adapter wrapping the given application service.
func New(service inbound.GamePlay) *CLI {
	return &CLI{service: service}
}

// Run parses args and calls the application service.
func (c *CLI) Run(args []string) error {
	cfg, err := parseArgs(args)
	if errors.Is(err, errUsage) {
		fmt.Println(usage)
		return nil
	}
	if err != nil {
		return err
	}
	return c.service.Play(cfg)
}

func parseArgs(args []string) (inbound.GamePlayConfig, error) {
	if len(args) == 0 {
		return inbound.GamePlayConfig{}, errUsage
	}

	cfg := inbound.GamePlayConfig{Speed: 2.0, TimeLimit: 30}

	for _, arg := range args {
		switch {
		case arg == "--confirm":
			cfg.ConfirmMode = true
		case strings.HasPrefix(arg, "--speed="):
			val := strings.TrimPrefix(arg, "--speed=")
			s, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return inbound.GamePlayConfig{}, fmt.Errorf("invalid speed value: %s\nRun 'pidshooter --help' for usage", val)
			}
			if s < 0.1 || s > 5.0 {
				return inbound.GamePlayConfig{}, fmt.Errorf("speed must be between 0.1 and 5.0, got: %s", val)
			}
			cfg.Speed = s
		case strings.HasPrefix(arg, "--time="):
			val := strings.TrimPrefix(arg, "--time=")
			t, err := strconv.Atoi(val)
			if err != nil {
				return inbound.GamePlayConfig{}, fmt.Errorf("invalid time value: %s\nRun 'pidshooter --help' for usage", val)
			}
			if t < 0 {
				return inbound.GamePlayConfig{}, fmt.Errorf("time must be 0 or positive, got: %s", val)
			}
			cfg.TimeLimit = t
		case arg == "--help" || arg == "-h":
			return inbound.GamePlayConfig{}, errUsage
		case len(arg) > 0 && arg[0] == '-':
			return inbound.GamePlayConfig{}, fmt.Errorf("unknown flag: %s\nRun 'pidshooter --help' for usage", arg)
		default:
			if len(arg) < process.MinPatternLength {
				return inbound.GamePlayConfig{}, fmt.Errorf("search pattern %q must be at least %d characters", arg, process.MinPatternLength)
			}
			if len(arg) > process.MaxPatternLength {
				return inbound.GamePlayConfig{}, fmt.Errorf("search pattern %q exceeds maximum length of %d characters", arg, process.MaxPatternLength)
			}
			cfg.Patterns = append(cfg.Patterns, arg)
		}
	}

	if len(cfg.Patterns) == 0 {
		return inbound.GamePlayConfig{}, fmt.Errorf("at least one search pattern is required\nRun 'pidshooter --help' for usage")
	}

	return cfg, nil
}
