package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const usage = `pidshooter - First Person PID Shoot

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
  pidshooter sleep --confirm
  pidshooter node --speed=2.5 --time=60
  pidshooter node --time=0`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse arguments
	patterns, confirmMode, speed, timeLimit, err := parseArgs(os.Args[1:])
	if err != nil {
		return err
	}

	// Find matching processes
	processes, err := FindProcesses(patterns)
	if err != nil {
		return fmt.Errorf("process search failed: %w", err)
	}

	if len(processes) == 0 {
		fmt.Printf("No processes found matching %v\n", patterns)
		return nil
	}

	fmt.Printf("Found %d process(es) matching %v. Starting game...\n", len(processes), patterns)

	// Create game
	scoreBoard := LoadScores()
	game := NewGame(processes, confirmMode, speed, timeLimit)
	game.highScore = scoreBoard.HighScore()

	// Initialize screen
	if err := game.Init(); err != nil {
		return fmt.Errorf("screen initialization failed: %w", err)
	}
	defer game.Cleanup() // Safety net for panics/signals

	// Populate entities after screen is ready
	game.PopulateEntities(processes)

	// Handle OS signals for clean exit
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	go func() {
		<-sigCh
		game.running = false
	}()

	// Run game loop
	game.Run()

	// Game over — save score
	duration := time.Since(game.startTime).Seconds()
	entry := ScoreEntry{
		Kills:    game.kills,
		FreedMem: game.freedMem,
		Speed:    speed,
		Time:     timeLimit,
		Duration: duration,
		Date:     time.Now(),
	}

	scoreBoard.Add(entry)
	_ = scoreBoard.Save()

	// Print results after terminal is restored
	game.Cleanup()
	fmt.Printf("\n  Game Over! Kills: %d | Freed: %s | Time: %.1fs\n",
		game.kills, formatBytes(game.freedMem), duration)
	if game.kills > 0 && game.kills >= scoreBoard.HighScore() {
		fmt.Println("  🏆 New high score!")
	}
	scoreBoard.PrintScores()

	return nil
}

// parseArgs parses command-line arguments and returns the patterns, confirm mode, speed, and time limit.
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
			// Validate pattern length (SECURITY-05: input validation)
			if len(arg) > maxPatternLength {
				return nil, false, 0, 0, fmt.Errorf("search pattern '%s' exceeds maximum length of %d characters", arg, maxPatternLength)
			}
			patterns = append(patterns, arg)
		}
	}

	if len(patterns) == 0 {
		return nil, false, 0, 0, fmt.Errorf("at least one search pattern is required\nRun 'pidshooter --help' for usage")
	}

	return patterns, confirmMode, speed, timeLimit, nil
}
