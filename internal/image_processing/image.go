package image_processing

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"os"

	"golang.org/x/term"
)

const (
	CTRL_C = '\x03'
	ESC    = '\x1b'
)

var (
	palette = []byte{' ', '.', ':', ';', '-', '=', '+', '*', '#', '%', '@'}

	// More detailed text gradient
	// palette = []byte{'$', '@', 'B', '%', '8', '&', 'W', 'M', '#', '*', 'o', 'a', 'h', 'k', 'b', 'd', 'p', 'q', 'w', 'm', 'Z', 'O', '0', 'Q', 'L', 'C', 'J', 'U', 'Y', 'X', 'z', 'c', 'v', 'u', 'n', 'x', 'r', 'j', 'f', 't', '/', '\\', '|', '(', ')', '1', '{', '}', '[', ']', '?', '-', '_', '+', '~', '<', '>', 'i', '!', 'l', 'I', ';', ':', ',', '"', '^', '`', '\'', '.', ' '}
)

func Generate_Image(source_path, output_path, debug_path string, lower_threshold, upper_threshold int64, kernel_size int, sigma float64) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	term_width, term_height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	fmt.Print("\x1B[?1049h")       // Enter alternate screen buffer
	defer fmt.Print("\x1B[?1049l") // Exit alternate screen buffer

	image_reader, err := os.Open(source_path)
	if err != nil {
		return err
	}
	defer image_reader.Close()

	img, _, err := image.Decode(image_reader)
	if err != nil {
		return err
	}

	gray := image.NewGray16(img.Bounds())
	draw.Draw(gray, gray.Bounds(), img, img.Bounds().Min, draw.Src)

	if len(debug_path) > 0 {
		err = WriteImage(debug_path+"/gray.png", gray)
		if err != nil {
			return err
		}
	}

	output := image.NewGray16(gray.Bounds())
	copy(output.Pix, gray.Pix)
	*output, err = CannyEdgeDetect(*output, debug_path, lower_threshold, upper_threshold, kernel_size, sigma)
	if err != nil {
		return err
	}

	if len(output_path) > 0 {
		file, err := os.Create(output_path)
		if err != nil {
			return err
		}
		for y := range gray.Bounds().Max.Y {
			for x := range gray.Bounds().Max.X {
				output_color := color.Black
				if output.Gray16At(x, y).Y > 0 {
					output_color = gray.Gray16At(x, y)
				}
				fmt.Fprint(file, string(palette[int(output_color.Y)%len(palette)]))
			}
			fmt.Fprintln(file)
		}
		file.Close()
	}

	var buf bytes.Buffer
	fmt.Fprint(&buf, "\x1B[2J\x1B[H") // Erase screen and home cursor

	for y := range term_height {
		for x := range term_width {
			output_color := color.Black
			if output.Gray16At(x, y).Y > 0 {
				output_color = gray.Gray16At(x, y)
			}
			fmt.Fprint(&buf, string(palette[int(output_color.Y)%len(palette)]))
		}
	}

	_, err = os.Stdout.Write(buf.Bytes())
	if err != nil {
		return err
	}

	reader := bufio.NewReader(os.Stdin)
	for true {
		char, _, err := reader.ReadRune()
		if err != nil {
			return err
		}

		if char == 'q' || char == CTRL_C || char == ESC {
			break
		}
	}

	return nil
}
