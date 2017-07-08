package main

import (
	"fmt"
	"image"
)

func decode(zf ArchiveFile) (image.Image, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()

	m, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("could not decode %s: %v", zf.Name, err)
	}
	return m, nil
}
