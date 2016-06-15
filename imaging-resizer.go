package main

import (
	"io"

	"github.com/disintegration/imaging"
)

func NewImaging(filter imaging.ResampleFilter) *ImagingResizer {
	return &ImagingResizer{Filter: filter}
}

type ImagingResizer struct {
	Filter imaging.ResampleFilter
}

func (r ImagingResizer) Resize(
	w io.Writer,
	src io.Reader,
	width int,
	height int,
) error {
	srcImage, err := jpegBackend.Decode(src)
	if err != nil {
		return err
	}

	resized := imaging.Resize(srcImage, width, height, r.Filter)

	return jpegBackend.Encode(w, resized, nil)
}

func init() {
	RegisterResizer(
		ResizerDesc{
			Name:     "imaging__bilinear",
			Library:  "imaging",
			Url:      "github.com/disintegration/imaging",
			Filter:   "bilinear",
			Instance: NewImaging(imaging.Linear),
		},
		ResizerDesc{
			Name:     "imaging__MitchellNetravali",
			Library:  "imaging",
			Url:      "github.com/disintegration/imaging",
			Filter:   "MitchellNetravali",
			Instance: NewImaging(imaging.MitchellNetravali),
		},
		ResizerDesc{
			Name:     "imaging__CatmullRom",
			Library:  "imaging",
			Url:      "github.com/disintegration/imaging",
			Filter:   "CatmullRom",
			Instance: NewImaging(imaging.CatmullRom),
		},
		ResizerDesc{
			Name:     "imaging__lanczos",
			Library:  "imaging",
			Url:      "github.com/disintegration/imaging",
			Filter:   "lanczos",
			Instance: NewImaging(imaging.Lanczos),
		},
	)
}
