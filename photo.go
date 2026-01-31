package main

import (
	"fmt"
	"image"
	"os"
	"sync"

	_ "golang.org/x/image/webp"
)

func GraphPhoto(filename string) {
	reader, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	m, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println("Error decoding image:", err)
		return
	}

	bounds := m.Bounds()

	width := (bounds.Max.X - bounds.Min.X) / step
	pixelsPerRow := (width + 4) / 5
	if pixelsPerRow == 0 {
		return
	}

	var wg sync.WaitGroup

	process := func(startY, endY int) {
		defer wg.Done()
		for y := startY; y < endY; y += step {
			for x := bounds.Min.X; x < bounds.Max.X; x += step {
				r, g, b, _ := m.At(x, y).RGBA()
				graphWithColor("("+fmt.Sprint(x)+","+fmt.Sprint(-y)+")", fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8))
			}
		}
	}

	rowsPerChunk := max(chunk/pixelsPerRow, 1)
	yChunkSize := rowsPerChunk * 5

	for y := bounds.Min.Y; y < bounds.Max.Y; y += yChunkSize {
		endY := min(y+yChunkSize, bounds.Max.Y)
		wg.Add(1)
		go process(y, endY)
	}

	wg.Wait()
}

func graphWithColor(latex string, color string) {
	page.MustEval(`(latex, color) => {
		Calc.setExpression({ id: Math.random().toString(), latex: latex, color:color });
	}`, latex, color)
}
