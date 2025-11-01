package image_processing

import (
	"abspayd/ascii-graphics/internal/logger"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
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

// Find the sobel gradient for an image
// Returns a magnitude gradient and the direction gradient
func SobelGradient(img image.Gray16) ([][]float64, [][]float64) {
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

	gx_image, err := matrixToImage(gx)
	if err != nil {
		log.Fatal(err)
	}
	WriteImage("resources/gx.png", gx_image)
	gy_image, err := matrixToImage(gy)
	if err != nil {
		log.Fatal(err)
	}
	WriteImage("resources/gy.png", gy_image)

	g := make([][]float64, img_height)
	theta := make([][]float64, img_height)
	for i := range g {
		g[i] = make([]float64, img_width)
		theta[i] = make([]float64, img_width)
	}

	for y := range img_height {
		for x := range img_width {
			gradient := math.Sqrt(math.Pow(gx[y][x], 2) + math.Pow(gy[y][x], 2))
			direction := math.Atan2(gy[y][x], gx[y][x])

			g[y][x] = gradient
			theta[y][x] = roundAngle(direction)
		}
	}

	return g, theta
}

// Round an angle to a multiple of a quarter of pi
func roundAngle(angle float64) float64 {
	const quarterPi = math.Pi / 4

	multiple := angle / quarterPi

	rounded := math.Round(multiple)

	// normalize the result such that 0 <= result < PI
	result := math.Mod(rounded*quarterPi, 2.0*math.Pi)

	if result < 0 {
		result += 2.0 * math.Pi
	}

	if result >= math.Pi {
		result -= math.Pi
	}

	return result
}

func SuppressGradient(g [][]float64, d [][]float64) [][]float64 {
	if len(g) != len(d) {
		log.Fatal("Invalid state: Gradient magnitude and direction do not match!")
	}

	result := make([][]float64, len(g))
	for i := range result {
		result[i] = make([]float64, len(g[0]))
		copy(result[i], g[i])
	}

	for y := range g {
		for x, value := range result[y] {
			switch d[y][x] {
			case 0:
				if x-1 >= 0 {
					if g[y][x-1] > value {
						result[y][x] = 0.0
					}
				}
				if x+1 < len(g[y]) {
					if g[y][x+1] > value {
						result[y][x] = 0.0
					}
				}
			case math.Pi / 2:
				if y-1 >= 0 {
					if g[y-1][x] > value {
						result[y][x] = 0.0
					}
				}
				if y+1 < len(g) {
					if g[y+1][x] > value {
						result[y][x] = 0.0
					}
				}
			case math.Pi / 4:
				if y-1 >= 0 && x-1 >= 0 {
					if g[y-1][x-1] > value {
						result[y][x] = 0
					}
				}

				if y+1 < len(g) && x+1 < len(g[y]) {
					if g[y+1][x+1] > value {
						result[y][x] = 0
					}
				}
			case (3 * math.Pi) / 4:
				if x-1 >= 0 && y+1 < len(g) {
					if g[y+1][x-1] > value {
						result[y][x] = 0
					}
				}
				if x+1 < len(g[y]) && y-1 >= 0 {
					if g[y-1][x+1] > value {
						result[y][x] = 0
					}
				}

			}
		}
	}

	return result
}

func Suppress(m [][]float64) [][]float64 {

	result := make([][]float64, len(m))
	for i := range len(result) {
		result[i] = make([]float64, len(m[0]))
		copy(result[i], m[i])
	}

	mask := [][]float64{
		{0.5, 0.75, 0.5},
		{0.75, 0, 0.75},
		{0.5, 0.75, 0.5},
	}
	mask_radius := len(mask) / 2

	for y := range m {
		for x := range m[y] {
			for i := -mask_radius; i <= mask_radius; i++ {
				for j := -mask_radius; j <= mask_radius; j++ {
					if y+j >= 0 && x+i >= 0 && y+j < len(m) && x+i < len(m[y]) {
						compare := mask[i+mask_radius][j+mask_radius] * m[y+j][x+i]
						if result[y][x] < compare {
							result[y][x] = 0
						}
					}
				}
			}
		}
	}

	return result
}

func CannyEdgeDetect(img image.Gray16) image.Gray16 {
	// img = GaussianFilter(img, 15, 2)
	img = GaussianFilter(img, 5, 1.4)
	err := WriteImage("resources/gaussian.png", &img)
	if err != nil {
		log.Fatal(err)
	}

	g, d := SobelGradient(img)
	imgPtr, err := matrixToImage(g)
	if err != nil {
		log.Fatal(err)
	}
	err = WriteImage("resources/gradient.png", imgPtr)
	if err != nil {
		log.Fatal(err)
	}

	suppressed := SuppressGradient(g, d)
	imgPtr, err = matrixToImage(suppressed)
	if err != nil {
		log.Fatal(err)
	}
	err = WriteImage("resources/suppressed2.png", imgPtr)
	if err != nil {
		log.Fatal(err)
	}

	return *imgPtr
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

func GaussianFilter(img image.Gray16, kernel_size int, sigma float64) image.Gray16 {
	kernel, err := generateGaussianKernel(kernel_size, sigma)
	if err != nil {
		logger.Logger.Fatal(err)
	}

	m := convolute(img, kernel)
	imgPtr, err := matrixToImage(m)
	if err != nil {
		log.Fatal(err)
	}
	result := *imgPtr

	return result
}
