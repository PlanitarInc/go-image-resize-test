package main

import (
	"fmt"
	"image"
	"io"

	"github.com/bamiaux/rez"
)

func NewRez(filter rez.Filter) *RezResizer {
	return &RezResizer{Filter: filter}
}

type RezResizer struct {
	Filter rez.Filter
}

func (r RezResizer) Resize(
	w io.Writer,
	src io.Reader,
	width int,
	height int,
) error {
	srcImage, err := jpegBackend.Decode(src)
	if err != nil {
		return err
	}

	im, ok := srcImage.(*image.YCbCr)
	if !ok {
		return fmt.Errorf("input picture is not ycbcr")
	}

	wDst, hDst := getImageDimensions(srcImage, width, height)
	resized := image.NewYCbCr(image.Rect(0, 0, wDst, hDst), im.SubsampleRatio)
	if err := rez.Convert(resized, srcImage, r.Filter); err != nil {
		return fmt.Errorf("unable to convert picture: %s", err)
	}

	return jpegBackend.Encode(w, resized, nil)
}

func init() {
	RegisterResizer(
		ResizerDesc{
			Name:     "rez__bilinear",
			Library:  "rez",
			Url:      "github.com/bamiaux/rez",
			Filter:   "bilinear",
			Instance: NewRez(rez.NewBilinearFilter()),
		},
		ResizerDesc{
			Name:     "rez__bicubic",
			Library:  "rez",
			Url:      "github.com/bamiaux/rez",
			Filter:   "bicubic",
			Instance: NewRez(rez.NewBicubicFilter()),
		},
		ResizerDesc{
			Name:     "rez__lanczos2",
			Library:  "rez",
			Url:      "github.com/bamiaux/rez",
			Filter:   "lanczos2",
			Instance: NewRez(rez.NewLanczosFilter(2)),
		},
		ResizerDesc{
			Name:     "rez__lanczos3",
			Library:  "rez",
			Url:      "github.com/bamiaux/rez",
			Filter:   "lanczos3",
			Instance: NewRez(rez.NewLanczosFilter(3)),
		},
	)
}
