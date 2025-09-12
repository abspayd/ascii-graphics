package image_processing

import (
	"abspayd/ascii-graphics/internal/logger"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
)

func WriteImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	return png.Encode(file, img)
}

func SobelGradient(img image.Gray16) image.Gray16 {
	output := image.NewGray16(img.Bounds())

	img_width := img.Bounds().Max.X
	img_height := img.Bounds().Max.Y

	gx_kernel := [][]float64{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	gy_kernel := [][]float64{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}
	gx := convolute(img, gx_kernel)
	gy := convolute(img, gy_kernel)

	WriteImage("resources/gx.png", &gx)
	WriteImage("resources/gy.png", &gy)

	for y := range img_height {
		for x := range img_width {
			gx_val := uint64(gx.Gray16At(x, y).Y)
			gy_val := uint64(gy.Gray16At(x, y).Y)

			gradient := math.Sqrt(float64(gx_val*gx_val + gy_val*gy_val))
			output.SetGray16(x, y, color.Gray16{Y: uint16(gradient)})
		}
	}

	return *output
}

func CannyEdgeDetect(img image.Gray16) image.Gray16 {
	img = GaussianFilter(img, 5, 2.0)
	err := WriteImage("resources/gaussian.png", &img)
	if err != nil {
		log.Fatal(err)
	}

	img = SobelGradient(img)
	err = WriteImage("resources/gradient.png", &img)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

func generateGaussianKernel(size int, sigma float64) ([][]float64, error) {
	if size%2 == 0 {
		return nil, fmt.Errorf("Invalid size: %d. Kernel size must be an odd number.", size)
	} else if size <= 0 {
		return nil, fmt.Errorf("Invalid size: %d. Kernel size must be a positive number.", size)
	}

	kernel := make([][]float64, size)
	for i := range kernel {
		kernel[i] = make([]float64, size)
	}

	sum := 0.0
	radius := size / 2

	a := 1.0 / (2.0 * math.Pi * sigma * sigma)
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			val := a * math.Exp(-1*float64(x*x+y*y)/(2*sigma*sigma))
			kernel[y+radius][x+radius] = val
			sum += val
		}
	}

	for y := range size {
		for x := range size {
			kernel[y][x] /= sum
		}
	}

	return kernel, nil
}

func convolute(img image.Gray16, kernel [][]float64) image.Gray16 {
	output := image.NewGray16(img.Bounds())

	img_width := img.Bounds().Max.X
	img_height := img.Bounds().Max.Y

	kernel_radius := len(kernel) / 2

	for y := range img_height {
		for x := range img_width {
			var sum uint16 = 0
			for j := -kernel_radius; j <= kernel_radius; j++ {
				for i := -kernel_radius; i <= kernel_radius; i++ {
					if y+j >= 0 && x+i >= 0 && y+j < img_height && x+i < img_width {
						sum += uint16(float64(kernel[i+kernel_radius][j+kernel_radius]) * float64(img.Gray16At(x+i, y+j).Y))
					}
				}
			}

			output.SetGray16(x, y, color.Gray16{Y: sum})
		}
	}

	return *output
}

// func threadedConvolute(img image.Gray16, kernel [][]float64) image.Gray16 {
// 	output := image.NewGray16(img.Bounds())
//
// 	n := 5
//
//
// 	return *output
// }

func GaussianFilter(img image.Gray16, kernel_size int, sigma float64) image.Gray16 {
	kernel, err := generateGaussianKernel(kernel_size, sigma)
	if err != nil {
		logger.Logger.Fatal(err)
	}

	return convolute(img, kernel)
}
