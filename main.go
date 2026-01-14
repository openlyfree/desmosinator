package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var (
	browser *rod.Browser
	page    *rod.Page
	chunk   int
	step    int
	id      int
	novis   bool
)

func main() {
	getConfig()

	fmt.Println("\n Desmosinator")
	fmt.Printf("   ├─ Chunk size: %d\n", chunk)
	fmt.Printf("   ├─ Step: %d\n", step)
	fmt.Printf("   └─ No visualization: %v\n", novis)

	fmt.Println("\nLaunching Desmos...")
	browser = rod.New().ControlURL(launcher.New().Leakless(false).Headless(false).MustLaunch()).MustConnect()
	page = browser.MustPage("https://www.desmos.com/calculator").MustWaitStable()

	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "http") && strings.Contains(os.Args[1], "musescore") {
			fmt.Printf("\nDownloading from MuseScore: %s\n", os.Args[1])
			wd, _ := os.Getwd()
			tempDir := filepath.Join(wd, "temp")
			exec.Command("npx", "dl-librescore@latest", "-i", os.Args[1], "-t", "midi", "-o", "temp").Run()
			files, _ := os.ReadDir(tempDir)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".mid") || strings.HasSuffix(file.Name(), ".midi") {
					os.Rename(filepath.Join(tempDir, file.Name()), filepath.Join(tempDir, "temp.mid"))
					break
				}
			}

			if novis {
				GraphMidiNoVis(filepath.Join(tempDir, "temp.mid"))
			} else {
				GraphMidi(filepath.Join(tempDir, "temp.mid"))
			}
			os.RemoveAll(tempDir)
		} else {
			if strings.Contains(os.Args[1], ".mid") {
				GraphMidi(os.Args[1])
			} else {
				GraphPhoto(os.Args[1])
			}

		}
	}
}

func getConfig() {
	chunk = func() int {
		for i, v := range os.Args {
			if strings.HasPrefix(v, "-cs") {
				var z int
				fmt.Sscanf(strings.ReplaceAll(v, "-cs", ""), "%d", &z)
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
				return z
			}
		}
		return 500
	}()

	step = func() int {
		for i, v := range os.Args {
			if strings.HasPrefix(v, "-step") {
				var z int
				fmt.Sscanf(strings.ReplaceAll(v, "-step", ""), "%d", &z)
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
				return z
			}
		}
		return 10
	}()

	novis = func() bool {
		for i, v := range os.Args {
			if v == "-nv" {
				os.Args = append(os.Args[:i], os.Args[i+1:]...)
				return true
			}
		}
		return false
	}()
}
