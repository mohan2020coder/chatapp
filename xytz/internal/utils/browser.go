package utils

import (
	"log"
	"os/exec"
	"runtime"
)

func OpenURL(url string) {
	go func() {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		case "darwin":
			cmd = exec.Command("open", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to open URL: %v", err)
			return
		}
		if err := cmd.Wait(); err != nil {
			log.Printf("Failed to open URL: %v", err)
		}
	}()
}
