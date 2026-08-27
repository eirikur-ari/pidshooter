package game

import (
	"fmt"

	"github.com/eirikur-ari/pidshooter/internal/core/process"
)

func validateProcesses(processes []process.Info) error {
	if len(processes) == 0 {
		return fmt.Errorf("no processes found")
	}
	return nil
}