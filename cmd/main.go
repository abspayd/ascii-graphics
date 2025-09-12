package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"log"
	"os"

	"golang.org/x/term"

	"abspayd/ascii-graphics/internal/image_processing"
	"abspayd/ascii-graphics/internal/logger"
)

const (
	CTRL_C = '\x03'
	ESC    = '\x1b'
)

var (
	palette = []byte{' ', '.', ':', ';', '-', '=', '+', '*', '#', '%', '@'}

	//	palette = []byte{
	//		'@',
	//		'%',
	//		'#',
	//		'*',
	//		'+',
	//		'=',
	//		'-',
	//		';',
	//		':',
	//		'.',
	//		' ',
	//	}
	//

	// palette = []byte{'$', '@', 'B', '%', '8', '&', 'W', 'M', '#', '*', 'o', 'a', 'h', 'k', 'b', 'd', 'p', 'q', 'w', 'm', 'Z', 'O', '0', 'Q', 'L', 'C', 'J', 'U', 'Y', 'X', 'z', 'c', 'v', 'u', 'n', 'x', 'r', 'j', 'f', 't', '/', '\\', '|', '(', ')', '1', '{', '}', '[', ']', '?', '-', '_', '+', '~', '<', '>', 'i', '!', 'l', 'I', ';', ':', ',', '"', '^', '`', '\'', '.', ' '}
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func fitImageToTerminal(img image.Image, termSize image.Rectangle) image.Image {
	img_bounds := img.Bounds()
	img_ratio := float64(img_bounds.Dx()) / float64(img_bounds.Dy())

	//output_height := min(img_bounds.Dy(), termSize.Dy())
	// output_width := min(int(float64(output_height)*img_ratio), termSize.Dx())
	output_height := termSize.Dy()
	output_width := termSize.Dx()
	output_bounds := image.Rectangle{
		Max: image.Point{X: output_width, Y: output_height},
	}

	logger.Logger.Printf("Image ratio: %f\n", img_ratio)
	logger.Logger.Printf("Image: width:%d, height:%d\n", img_bounds.Dx(), img_bounds.Dy())
	logger.Logger.Printf("Terminal: width:%d, height:%d\n", termSize.Dx(), termSize.Dy())
	logger.Logger.Printf("Output: width:%d, height:%d\n", output_width, output_height)

	output := image.NewRGBA(output_bounds)
	x_factor := img_bounds.Dx() / output_width
	y_factor := img_bounds.Dy() / output_height

	for x := range output_bounds.Max.X {
		for y := range output_bounds.Max.Y {
			color := img.At(x*x_factor, y*y_factor)
			output.Set(x, y, color)
		}
	}

	return output
}

func main() {
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	logger.Logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	logger.Logger.Println("========")

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	term_width, term_height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%d, %d\n", term_width, term_height)

	fmt.Print("\x1B[?1049h")       // Enter alternate screen buffer
	defer fmt.Print("\x1B[?1049l") // Exit alternate screen buffer

	// var buf bytes.Buffer
	// fmt.Fprint(&buf, "\x1B[2J\x1B[H") // Erase screen and home cursor

	// for range term_height {
	// 	for range term_width {
	// 		fmt.Fprint(&buf, "#")
	// 	}
	// 	fmt.Fprint(&buf, "\r\n")
	// }

	// _, err = os.Stdout.Write(buf.Bytes())
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// path := "resources/johann-siemens-EPy0gBJzzZU-unsplash.jpg"
	path := "resources/circle.png"
	// path := "resources/tree-1798062137.jpg"
	// path := "resources/Bikesgray.jpg"
	image_reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer image_reader.Close()

	img, _, err := image.Decode(image_reader)
	if err != nil {
		log.Fatal(err)
	}

	img = fitImageToTerminal(img, image.Rect(0, 0, term_width, term_height))

	gray := image.NewGray16(img.Bounds())
	draw.Draw(gray, gray.Bounds(), img, img.Bounds().Min, draw.Src)

	err = image_processing.WriteImage("resources/gray.png", gray)
	if err != nil {
		log.Fatal(err)
	}

	*gray = image_processing.CannyEdgeDetect(*gray)

	// scale_x := min(1, (gray.Bounds().Max.X)/(term_width))
	// scale_y := min(1, (gray.Bounds().Max.Y)/(term_height*2))

	var buf bytes.Buffer

	fmt.Fprint(&buf, "\x1B[2J\x1B[H") // Erase screen and home cursor

	// slices.Reverse(palette)

	for y := range term_height {
		for x := range term_width {
			color := gray.Gray16At(x, y)
			fmt.Fprint(&buf, string(palette[int(color.Y)%len(palette)]))
		}
	}

	// for y := 0; y <= gray.Bounds().Max.Y && (y/scale_y) <= term_height; y += scale_y {
	// 	for x := 0; x <= gray.Bounds().Max.X && (x/scale_x) <= term_width; x += scale_x {
	// 		// color := gray.GrayAt(x, y)
	// 		// fmt.Fprint(&buf, string(palette[int(color.Y)%len(palette)]))
	// 		fmt.Fprint(&buf, "o")
	// 	}
	// }

	_, err = os.Stdout.Write(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)
	for true {
		char, _, err := reader.ReadRune()
		if err != nil {
			log.Fatal(err)
		}

		if char == 'q' || char == CTRL_C || char == ESC {
			break
		}
	}
}
