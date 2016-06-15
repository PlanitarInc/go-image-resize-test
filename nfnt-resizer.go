package main

import (
	"io"

	nfnt "github.com/nfnt/resize"
)

func NewNfnt(intr nfnt.InterpolationFunction) *NfntResizer {
	return &NfntResizer{Interpolation: intr}
}

type NfntResizer struct {
	JpegTurbo     bool
	Interpolation nfnt.InterpolationFunction
}

func (r NfntResizer) Resize(
	dst io.Writer,
	src io.Reader,
	maxWidth int,
	maxHeight int,
) error {
	srcImage, err := jpegBackend.Decode(src)
	if err != nil {
		return err
	}

	w := uint(maxWidth)
	if w == 0 {
		w = 10000
	}
	h := uint(maxHeight)
	if h == 0 {
		h = 10000
	}
	dstImage := nfnt.Thumbnail(w, h, srcImage, r.Interpolation)

	return jpegBackend.Encode(dst, dstImage, nil)
}

func init() {
	RegisterResizer(
		ResizerDesc{
			Name:     "nfnt__bilinear",
			Library:  "nfnt",
			Url:      "github.com/nfnt/resize",
			Filter:   "bilinear",
			Instance: NewNfnt(nfnt.Bilinear),
		},
		ResizerDesc{
			Name:     "nfnt__bicubic",
			Library:  "nfnt",
			Url:      "github.com/nfnt/resize",
			Filter:   "bicubic",
			Instance: NewNfnt(nfnt.Bicubic),
		},
		ResizerDesc{
			Name:     "nfnt__MitchellNetravali",
			Library:  "nfnt",
			Url:      "github.com/nfnt/resize",
			Filter:   "MitchellNetravali",
			Instance: NewNfnt(nfnt.MitchellNetravali),
		},
		ResizerDesc{
			Name:     "nfnt__lanczos2",
			Library:  "nfnt",
			Url:      "github.com/nfnt/resize",
			Filter:   "lanczos2",
			Instance: NewNfnt(nfnt.Lanczos2),
		},
		ResizerDesc{
			Name:     "nfnt__lanczos3",
			Library:  "nfnt",
			Url:      "github.com/nfnt/resize",
			Filter:   "lanczos3",
			Instance: NewNfnt(nfnt.Lanczos3),
		},
	)
}
