package main

import (
	"image"
	"io"

	jpegnative "image/jpeg"

	jpegturbo "github.com/pixiv/go-libjpeg/jpeg"
)

type JpegBE interface {
	Encode(w io.Writer, src image.Image, opts interface{}) error
	Decode(r io.Reader) (image.Image, error)
}

type JpegTurbo struct{}

func (JpegTurbo) Encode(w io.Writer, src image.Image, opts interface{}) error {
	return jpegturbo.Encode(w, src, &jpegturbo.EncoderOptions{Quality: 90})
}

func (JpegTurbo) Decode(r io.Reader) (image.Image, error) {
	return jpegturbo.Decode(r, &jpegturbo.DecoderOptions{})
}

type JpegNative struct{}

func (JpegNative) Encode(w io.Writer, src image.Image, opts interface{}) error {
	return jpegnative.Encode(w, src, &jpegnative.Options{Quality: 90})
}

func (JpegNative) Decode(r io.Reader) (image.Image, error) {
	return jpegnative.Decode(r)
}
