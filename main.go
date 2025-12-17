package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

var browser *rod.Browser
var page *rod.Page
var id int
var chunkSize int

func main() {
	chunkSize = func() int {
		for _, v := range os.Args {
			if strings.HasPrefix(v, "-cs") {
				var i int
				fmt.Sscanf(strings.ReplaceAll(v, "-nv", ""), "%d", &i)
				return i
			}
		}
		return 500
	}()

	browser = rod.New().ControlURL(launcher.New().Leakless(false).Headless(false).MustLaunch()).MustConnect()
	page = browser.MustPage("https://www.desmos.com/calculator").MustWaitStable()
	nv := func() bool {
		for _, v := range os.Args {
			if v == "-nv" {
				return true
			}
		}
		return false
	}()

	if len(os.Args) < 1 {
		if strings.Contains(os.Args[1], "http") {
			exec.Command("npx", "dl-librescore@latest", "-i", os.Args[1], "-t", "midi", "-o", "temp.mid").Run()
			if nv {
				GraphMidiNoVis("temp.mid")
			} else {
				GraphMidi("temp.mid")
			}
			os.Remove("temp.mid")
		} else {
			GraphMidi(os.Args[1])
		}
	}
}

func graph(latex string) {

	page.MustEval(`(id, latex) => {
		Calc.setExpression({ id: id, latex: latex , pointSize: 2});
	}`, func() string {
		id++
		return fmt.Sprint(id)
	}(), latex)
}
