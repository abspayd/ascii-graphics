package image_processing

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"
)

func convolute(img image.Gray16, kernel [][]float64) [][]float64 {
	image_width := img.Bounds().Max.X
	image_height := img.Bounds().Max.Y

	result := make([][]float64, image_height)
	for i := range result {
		result[i] = make([]float64, image_width)
	}

	var wg sync.WaitGroup
	n_threads := min(100, image_height)

	rows_per_thread := image_height / n_threads

	for t := range n_threads {
		wg.Add(1)

		start_row := t * rows_per_thread
		end_row := (t + 1) * rows_per_thread

		if t == n_threads-1 {
			end_row = image_height
		}

		go worker(&img, &result, kernel, start_row, end_row, &wg)
	}

	wg.Wait()

	return result
}

func worker(img *image.Gray16, result *[][]float64, kernel [][]float64, start_row, end_row int, wg *sync.WaitGroup) {
	defer wg.Done()

	kernel_radius := len(kernel) / 2

	for y := start_row; y < end_row; y++ {
		for x := range img.Bounds().Max.X {
			var sum float64 = 0.0
			for j := -kernel_radius; j <= kernel_radius; j++ {
				for i := -kernel_radius; i <= kernel_radius; i++ {
					if y+j >= 0 && x+i >= 0 && y+j < img.Bounds().Max.Y && x+i < img.Bounds().Max.X {
						sum += kernel[i+kernel_radius][j+kernel_radius] * float64(img.Gray16At(x+i, y+j).Y)
					}
				}
			}

			(*result)[y][x] = sum
		}
	}
}

func matrixToImage(m [][]float64) (*image.Gray16, error) {
	if len(m) == 0 {
		return nil, fmt.Errorf("Cannot convert empty matrix to an image!")
	}

	img := image.NewGray16(image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: len(m[0]), Y: len(m)},
	})

	for i, row := range m {
		for j, value := range row {
			value = math.Max(value, 0)
			img.SetGray16(j, i, color.Gray16{
				Y: uint16(value),
			})
		}
	}

	return img, nil
}
