package main

import "io"

type Resizer interface {
	Resize(w io.Writer, r io.Reader, width int, height int) error
}

type ResizerList []ResizerDesc

func (l ResizerList) Find(name string) *ResizerDesc {
	for i := range l {
		if l[i].Name == name {
			return &l[i]
		}
	}
	return nil
}

type ResizerDesc struct {
	Name     string
	Library  string
	Url      string
	Filter   string
	Instance Resizer
}
