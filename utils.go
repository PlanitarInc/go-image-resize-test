package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
)

func getImageDimensions(im image.Image, wMax int, hMax int) (int, int) {
	src := im.Bounds().Size()
	wSrc, hSrc := src.X, src.Y

	if wMax == 0 {
		return rescale(wSrc, hMax, hSrc), hMax
	}

	if hMax == 0 {
		return wMax, rescale(hSrc, wMax, wSrc)
	}

	ratioSrc := float32(wSrc) / float32(hSrc)
	ratioDst := float32(wMax) / float32(hMax)

	if ratioSrc < ratioDst {
		return rescale(wSrc, hMax, hSrc), hMax
	}

	return wMax, rescale(hSrc, wMax, wSrc)
}

func rescale(val int, src, dst int) int {
	return int(math.Floor(.5 + float64(val)*float64(src)/float64(dst)))
}

func imageDims(rs io.ReadSeeker) (string, int, int) {
	offset, _ := rs.Seek(0, 1)
	config, format, err := image.DecodeConfig(rs)
	if err != nil {
		fmt.Printf("  imageDims: %#v, %s, %s\n", config, format, err)
		rs.Seek(offset, 0)
		config, err = jpeg.DecodeConfig(rs)
		if err != nil {
			fmt.Printf("  imageDims: %#v, jpeg, %s\n", config, err)
			rs.Seek(offset, 0)
			config, err = png.DecodeConfig(rs)
			if err != nil {
				fmt.Printf("  imageDims: %#v, png, %s\n", config, err)
				format = "unknown"
			} else {
				fmt.Println("  imageDims: png")
				format = "manual-png"
			}
		} else {
			fmt.Println("  imageDims: jpeg")
			format = "manual-jpeg"
		}
	}
	defer rs.Seek(offset, 0)
	return format, config.Width, config.Height
}

func seekerLen(s io.Seeker) int64 {
	offset, _ := s.Seek(0, 1)
	end, _ := s.Seek(0, 2)
	length := end - offset
	s.Seek(offset, 0)
	return length
}

func b2s(bytes int64) string {
	const (
		KB = 1024
		MB = 1014 * KB
		GB = 1014 * MB
	)

	if bytes > GB {
		return fmt.Sprintf("%.2fGB", float32(bytes)/float32(GB))
	}
	if bytes > MB {
		return fmt.Sprintf("%.2fMB", float32(bytes)/float32(MB))
	}
	if bytes > KB {
		return fmt.Sprintf("%.2fKB", float32(bytes)/float32(KB))
	}
	return fmt.Sprintf("%dB", bytes)
}

func percent(part, total int64) string {
	return fmt.Sprintf("%.2f%%", 100*float64(part)/float64(total))
}

type DiscardCount struct {
	N int64
}

func (c *DiscardCount) Write(p []byte) (int, error) {
	c.N += int64(len(p))
	return len(p), nil
}

func (c *DiscardCount) WriteString(s string) (int, error) {
	c.N += int64(len(s))
	return len(s), nil
}
