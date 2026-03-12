package browser

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
)

func Open(raw string) error {

	_, err := url.ParseRequestURI(raw)
	if err != nil {
		return errors.New("invalid url")
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {

	case "windows":
		//exec.Command("cmd", "/c", "start", url)
		fmt.Println("Windows 系统")
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", raw)

	case "darwin":
		cmd = exec.Command("open", raw)

	case "linux":
		cmd = exec.Command("xdg-open", raw)

	default:
		return errors.New("unsupported platform")
	}

	return cmd.Start()
}
