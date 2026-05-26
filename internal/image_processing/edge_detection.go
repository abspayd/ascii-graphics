package image_processing

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
)

const (
	WEAK_EDGE   = 32767
	STRONG_EDGE = 65535
)

func WriteImage(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	return png.Encode(file, img)
}

// Find the sobel gradient for an image
// Returns a magnitude gradient and the direction gradient
func SobelGradient(img image.Gray16, debug_path string) ([][]float64, [][]float64, error) {
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
		return nil, nil, err
	}
	if len(debug_path) > 0 {
		if err := WriteImage(debug_path+"/gx.png", gx_image); err != nil {
			return nil, nil, err
		}
	}

	gy_image, err := matrixToImage(gy)
	if err != nil {
		return nil, nil, err
	}
	if len(debug_path) > 0 {
		if err := WriteImage(debug_path+"/gy.png", gy_image); err != nil {
			return nil, nil, err
		}
	}

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

	return g, theta, nil
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

func SuppressGradient(g [][]float64, d [][]float64) ([][]float64, error) {
	if len(g) != len(d) {
		return nil, fmt.Errorf("Invalid state: Gradient magnitude and direction do not match!")
	}

	result := make([][]float64, len(g))
	for i := range result {
		result[i] = make([]float64, len(g[0]))
		copy(result[i], g[i])
	}

	for y := range g {
		for x, value := range result[y] {
			var x1, y1 int
			var x2, y2 int

			switch d[y][x] {
			case 0:
				x1, y1 = x-1, y
				x2, y2 = x+1, y
			case math.Pi / 2:
				x1, y1 = x, y-1
				x2, y2 = x, y+1
			case math.Pi / 4:
				x1, y1 = x-1, y-1
				x2, y2 = x+1, y+1
			case (3 * math.Pi) / 4:
				x1, y1 = x-1, y+1
				x2, y2 = x+1, y-1

			default:
				continue
			}

			if x1 >= 0 && x1 < len(g[y]) && y1 >= 0 && y1 < len(g) && g[y1][x1] > value {
				result[y][x] = 0.0
			}
			if x2 >= 0 && x2 < len(g[y]) && y2 >= 0 && y2 < len(g) && g[y2][x2] > value {
				result[y][x] = 0.0
			}
		}
	}

	return result, nil
}

func DoubleThreshold(m [][]float64, lower, upper float64) [][]float64 {
	result := make([][]float64, len(m))
	for i := range result {
		result[i] = make([]float64, len(m[i]))
		copy(result[i], m[i])
	}

	for i, row := range result {
		for j, value := range row {
			if value < lower {
				result[i][j] = 0.0
			} else if value >= lower && value < upper {
				// mark weak
				result[i][j] = WEAK_EDGE
			} else {
				// mark strong
				result[i][j] = STRONG_EDGE
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

func HysteresisEdgeTracking(m [][]float64) [][]float64 {
	result := make([][]float64, len(m))
	for i := range result {
		result[i] = make([]float64, len(m[i]))
		copy(result[i], m[i])
	}

	for y := range result {
	edge_tracking:
		for x := range result[y] {
			if result[y][x] != WEAK_EDGE {
				continue
			}

			for i := -1; i <= 1; i++ {
				for j := -1; j <= 1; j++ {
					if y+i >= 0 && y+i < len(result) && x+j >= 0 && x+j < len(result[y]) && result[y+i][x+j] == 65535 {
						result[y][x] = STRONG_EDGE
						continue edge_tracking
					}
				}
			}
			result[y][x] = 0.0
		}
	}

	return result
}

func CannyEdgeDetect(img image.Gray16, debug_path string, lower_threshold, upper_threshold float64, kernel_size int, sigma float64) (image.Gray16, error) {
	img, err := GaussianFilter(img, kernel_size, sigma)
	if err != nil {
		return image.Gray16{}, nil
	}

	debug := len(debug_path) > 0

	if debug {
		err := WriteImage(debug_path+"/gaussian.png", &img)
		if err != nil {
			return image.Gray16{}, nil
		}
	}

	g, d, err := SobelGradient(img, debug_path)
	if err != nil {
		return image.Gray16{}, err
	}

	imgPtr, err := matrixToImage(g)
	if err != nil {
		return image.Gray16{}, err
	}
	if debug {
		err = WriteImage(debug_path+"/gradient.png", imgPtr)
		if err != nil {
			return image.Gray16{}, err
		}
	}

	suppressed, err := SuppressGradient(g, d)
	if err != nil {
		return image.Gray16{}, err
	}

	imgPtr, err = matrixToImage(suppressed)
	if err != nil {
		return image.Gray16{}, err
	}
	if debug {
		err = WriteImage(debug_path+"/suppressed.png", imgPtr)
		if err != nil {
			return image.Gray16{}, err
		}
	}

	threshold := DoubleThreshold(suppressed, lower_threshold, upper_threshold)
	imgPtr, err = matrixToImage(threshold)
	if err != nil {
		return image.Gray16{}, err
	}
	if debug {
		err = WriteImage(debug_path+"/doublethreshold.png", imgPtr)
		if err != nil {
			return image.Gray16{}, err
		}
	}

	edge_detect := HysteresisEdgeTracking(threshold)
	imgPtr, err = matrixToImage(edge_detect)
	if err != nil {
		return image.Gray16{}, err
	}
	if debug {
		err = WriteImage(debug_path+"/edges.png", imgPtr)
		if err != nil {
			return image.Gray16{}, err
		}
	}

	return *imgPtr, nil
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

func GaussianFilter(img image.Gray16, kernel_size int, sigma float64) (image.Gray16, error) {
	kernel, err := generateGaussianKernel(kernel_size, sigma)
	if err != nil {
		return image.Gray16{}, err
	}

	m := convolute(img, kernel)
	imgPtr, err := matrixToImage(m)
	if err != nil {
		return image.Gray16{}, err
	}
	result := *imgPtr

	return result, nil
}
