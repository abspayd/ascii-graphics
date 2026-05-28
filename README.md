# ascii-graphics

A terminal ASCII art generator using Canny's edge detection algorithm.

## Overview

`ascii-graphics` loads a raster image, applies the Canny edge detection algorithm, and renders the result as ASCII art in your terminal.

This project is written in Go without the use of image processing libraries.

**Stages of edge detection:**

1. **Grayscale conversion** - input image is converted to grayscale
2. **Gaussian blur** - smooth noise before gradient computation
3. **Sobel gradient** - compute per-pixel gradient magnitude and direction using 3×3 Sobel kernels
4. **Gradient magnitude thresholding** - thin edges to single-pixel width by zeroing non-local-maxima along the gradient direction
5. **Double thresholding** - classify pixels as strong edges, weak edges, or suppressed
6. **Hysteresis edge tracking** - promote weak edges connected to strong edges; discard the rest
7. **ASCII rendering** - map pixel intensity to an ASCII shade gradient palette and writes to an alternate buffer in the terminal

<!-- Add screenshots here -->

## Installation & Compilation

**Prerequisites:**
- Go 1.24 or later ([golang.org/dl](https://go.dev/dl/))

**Clone and build:**

```sh
git clone https://github.com/abspayd/ascii-graphics.git
cd ascii-graphics
go build -o ascii-graphics ./cmd/main.go
```

**Run:**

```sh
./ascii-graphics
```

The binary must be run from the project root - it resolves the input image and debug output paths relative to the working directory.

**Run without building:**

```sh
go run ./cmd/main.go
```

**Controls:**

| Key | Action |
|-----|--------|
| `q` | Quit |
| `Ctrl+C` | Quit |
| `ESC` | Quit |

**Changing the input image:**

The input image path is set in `cmd/main.go`. Edit the `path` variable to point to a different file:

```go
// cmd/main.go
path := "resources/your-image.jpg" // or "resources/your-image.png"
```

**Tuning the algorithm:**

The Canny parameters are also set directly in `cmd/main.go` and `internal/image_processing/edge_detection.go`:

| Parameter | Location | Default | Description |
|-----------|----------|---------|--------|
| Gaussian kernel size | `edge_detection.go` | `5` | Kernel size for blur convolutions. This can be used to increase / decrease area of blur. Larger areas have significantly more intense computation. |
| Gaussian sigma (σ) | `edge_detection.go` | `1.4` | Standard deviations of Guassian blur. This can be used to increase blur intensity. |
| Lower threshold | `edge_detection.go` | `5000` | Intensity threshold to cut off weak edges. |
| Upper threshold | `edge_detection.go` | `15000` | Intensity threshold to be considered a "strong" edge. |

**Debug output:**

Each pipeline stage writes a PNG to `resources/`:

| File | Stage |
|------|-------|
| `gray.png` | After grayscale conversion |
| `gaussian.png` | After Gaussian blur |
| `gx.png` / `gy.png` | Sobel X and Y components |
| `gradient.png` | Gradient magnitude |
| `suppressed.png` | After non-maximum suppression |
| `doublethreshold.png` | After double thresholding |
| `edges.png` | Final edge map |

## Sources
 - https://en.wikipedia.org/wiki/Gaussian_filter
 - https://en.wikipedia.org/wiki/Canny_edge_detector
 - https://en.wikipedia.org/wiki/Sobel_operator
