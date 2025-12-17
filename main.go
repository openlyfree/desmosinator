package main

import (
	"fmt"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)
var browser *rod.Browser
var page *rod.Page
var id int

func main() {
	browser = rod.New().ControlURL(launcher.New().Leakless(false).Headless(false).MustLaunch()).MustConnect()
	page = browser.MustPage("https://www.desmos.com/calculator").MustWaitStable()

}

func graph(latex string) {

	page.MustEval(`(id, latex) => {
		Calc.setExpression({ id: id, latex: latex , pointSize: 2});
	}`, func() string {
		id++
		return fmt.Sprint(id)
	}(), latex)
}