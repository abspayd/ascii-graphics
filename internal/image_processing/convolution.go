package image_processing

import (
	"image"
)

func convoluteImage(img image.Gray16, kernel [][]float64) [][]float64 {
	image_width := img.Bounds().Max.X
	image_height := img.Bounds().Max.Y

	result := make([][]float64, image_height)
	for i := range result {
		result[i] = make([]float64, image_width)
	}

	return nil
}

func convolutionWorker() {
}
