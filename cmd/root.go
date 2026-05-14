package cmd

import (
	"fmt"
	"os"

	"abspayd/ascii-graphics/internal/image_processing"

	"github.com/spf13/cobra"
)

const (
	CTRL_C = '\x03'
	ESC    = '\x1b'
)

var (
	source_path string
	output_path string
	debug_path  string

	lower_threshold float64
	upper_threshold float64
	kernel_size     int
	sigma           float64

	palette = []byte{' ', '.', ':', ';', '-', '=', '+', '*', '#', '%', '@'}
)

var (
	rootCmd = &cobra.Command{
		Use:   "ascii-art",
		Short: "Convert images to ASCII art",
		RunE: func(cmd *cobra.Command, args []string) error {
			return image_processing.Generate_Image(source_path, output_path, debug_path, lower_threshold, upper_threshold, kernel_size, sigma)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// path configs
	rootCmd.Flags().StringVarP(&source_path, "input", "i", "", "Input image path (accepts PNG and JPG images).")
	rootCmd.Flags().StringVarP(&output_path, "output", "o", "", "Output file path.")
	rootCmd.Flags().StringVar(&debug_path, "debug-path", "", "Debug directory path.")

	// image processing variables
	rootCmd.Flags().Float64VarP(&lower_threshold, "lower-threshold", "l", 5000, "Used during edge detection. All values lower than this threshold will be removed from the output image. Must be between 0 and 65535.")
	rootCmd.Flags().Float64VarP(&upper_threshold, "upper-threshold", "u", 15000, "Used during edge detection. Determines the strength level of an edge to be considered \"strong\". Must be between 0 and 65535.")
	rootCmd.Flags().IntVarP(&kernel_size, "blur-kernel-size", "k", 5, "Gaussian filter kernel size. Increase or decrease to change the area of the blur.")
	rootCmd.Flags().Float64VarP(&sigma, "blur-standard-deviation", "s", 1.4, "Gaussian filter standard deviation. Increase or decrease to change the intensity of the blur.")

	rootCmd.MarkFlagRequired("input")
}
