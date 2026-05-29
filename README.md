# ascii-graphics

A terminal ASCII art generator

`ascii-graphics` loads an image, applies the Canny edge detection algorithm, and renders the result as ASCII art in your terminal.

<!-- TODO: Add screenshot or demo GIF -->

## Features

- Canny edge detection pipeline implemented from scratch in Go
- Supports any image of either PNG or JPG format as an input
- Renders to an alternate terminal buffer (similar to `less`)
- Optional text file output for generated ASCII art
- Debug mode saves intermediate images at each pipeline stage and enables logging

## Install

**Prerequisites:** Go 1.24 or later ([golang.org/dl](https://go.dev/dl/))

```sh
git clone https://github.com/abspayd/ascii-graphics.git
cd ascii-graphics
go build -o ascii-graphics .
```

## Usage

```sh
./ascii-graphics -i path/to/image.jpg
```

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --input` | *(required)* | Input image path (PNG or JPG) |
| `-o, --output` | | Output file path for ASCII text |
| `--debug-path` | | Directory to save intermediate pipeline images |
| `-l, --lower-threshold` | `5000` | Lower edge detection threshold (0-65535) |
| `-u, --upper-threshold` | `15000` | Upper edge detection threshold (0-65535) |
| `-k, --blur-kernel-size` | `5` | Gaussian kernel size (must be odd) |
| `-s, --blur-standard-deviation` | `1.4` | Gaussian blur sigma |

**Controls:**

| Key | Action |
|-----|--------|
| `q` | Quit |
| `Ctrl+C` | Quit |
| `ESC` | Quit |

## How It Works

The image is processed through the stages of the Canny edge detection algorithm:

1. **Grayscale conversion** -- input image is converted to 16-bit grayscale
2. **Gaussian blur** -- smooth noise before gradient computation
3. **Sobel gradient** -- compute per-pixel gradient magnitude and direction using 3x3 Sobel kernels
4. **Non-maximum suppression** -- thin edges to single-pixel width by zeroing non-local-maxima along the gradient direction
5. **Double thresholding** -- classify pixels as strong edges, weak edges, or suppressed
6. **Hysteresis edge tracking** -- promote weak edges connected to strong edges; discard the rest
7. **ASCII rendering** -- map pixel intensity to a character palette and render to an alternate terminal buffer

## Debug Output

When `--debug-path` is provided, PNG images are saved at each stage of the image processing steps:

| File | Stage |
|------|-------|
| `gray.png` | Grayscale conversion |
| `gaussian.png` | Gaussian blur |
| `gx.png` / `gy.png` | Sobel X and Y components |
| `gradient.png` | Gradient magnitude |
| `suppressed.png` | Non-maximum suppression |
| `doublethreshold.png` | Double thresholding |
| `edges.png` | Final edge map |

## References

- https://en.wikipedia.org/wiki/Gaussian_filter
- https://en.wikipedia.org/wiki/Canny_edge_detector
- https://en.wikipedia.org/wiki/Sobel_operator
