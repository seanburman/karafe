package gui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func OpenURL(url string) error {
	fmt.Printf("Opening %s...\n", url)
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func OpenDesktop() {
	fmt.Printf("Opening desktop...\n")
	cmnd := exec.Command("cmd/gui/kaw")
	cmnd.Start()
}

func ListenKeyBoard(callback func(string)) {
	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		callback(input)
	}
}

func ListenCommands() {
	// commands := make(map[string]func())
	commands := map[string]func(){
		"w": func() { OpenURL("http://localhost:8080/store") },
		"d": func() { OpenDesktop() },
	}

	callback := func(input string) {
		for command, fn := range commands {
			if string([]rune(input)[:len(command)]) == command {
				fn()
			}
		}
	}

	go ListenKeyBoard(callback)
}
