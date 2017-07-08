package main

import (
	"fmt"
	"image"
	"log"
	"os"
	"strings"

	"github.com/nfnt/resize"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"

	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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

func findNext(files []ArchiveFile, curr int, diff int) (int, bool) {
	for i := curr; i >= 0 && i < len(files); i += diff {
		if strings.HasSuffix(files[i].Name(), ".jpg") {
			return i, true
		}
	}

	return curr, false
}

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("You must enter a file name")
	}

	archive, err := OpenArchive(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer archive.Close()

	idx, ok := findNext(archive.Files(), 0, 1)
	if !ok {
		log.Fatal("No valid images")
	}

	img, err := decode(archive.Files()[idx])
	if err != nil {
		fmt.Println(err)
		return
	}

	var width, height int

	driver.Main(func(s screen.Screen) {
		//body := newBook(archive, idx)

		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title: "CBArchiveViewer",
		})
		if err != nil {
			return
		}
		defer w.Release()

		buff, err := s.NewBuffer(image.Point{width, height})
		if err != nil {
			return
		}

		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				switch e.To {
				case lifecycle.StageDead:
					return
				}
			case key.Event:
				fmt.Printf("%T: %+v\n", e, e)
				if e.Direction != key.DirPress {
					continue
				}

				dir := 0
				switch e.Code {
				case key.CodeRightArrow:
					dir = 1
				case key.CodeLeftArrow:
					dir = -1
				}

				if dir != 0 {
					idx, _ = findNext(archive.Files(), idx+dir, dir)
					img, err = decode(archive.Files()[idx])
					if err != nil {
						fmt.Println(err)
						return
					}
					w.SendFirst(paint.Event{})
				}
			case paint.Event:
				fmt.Println("PAINT")

				// Clear the background
				draw.Draw(buff.RGBA(), buff.Bounds(), &image.Uniform{&color.RGBA{0, 0, 0, 1}}, image.ZP, draw.Src)

				// Shrink the image to fit on the screen
				drawImg := resize.Thumbnail(uint(width), uint(height), img, resize.Bilinear)

				drawImgSize := drawImg.Bounds().Size()

				startX := (width - drawImgSize.X) / 2
				startY := (height - drawImgSize.Y) / 2

				fmt.Println(startX, startY)

				draw.Draw(
					buff.RGBA(),
					image.Rect(startX, startY, startX+drawImgSize.X, startY+drawImgSize.Y),
					drawImg,
					image.Point{0, 0},
					draw.Over,
				)

				// Drop our render texture to the screen
				w.Upload(image.Point{0, 0}, buff, buff.Bounds())

				// Publish flips the buffers so we actually display to
				// the screen.
				w.Publish()
			case size.Event:
				fmt.Println("Resize!")

				width = e.WidthPx
				height = e.HeightPx

				buff.Release()
				buff, err = s.NewBuffer(image.Point{width, height})
				if err != nil {
					return
				}
			}
		}
	})
}
