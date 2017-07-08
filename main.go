package main

import (
	"image"
	"log"
	"os"

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

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("You must enter a file name")
	}

	book, err := OpenBook(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}
	defer book.Close()

	var width, height int

	driver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title: "CBArchiveViewer",
		})
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer w.Release()

		// Create a buffer for drawing which matches the window size
		buff, err := s.NewBuffer(image.Point{width, height})
		if err != nil {
			log.Fatalln(err)
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
				if e.Direction != key.DirPress {
					continue
				}

				switch e.Code {
				case key.CodeRightArrow:
					book.NextPage()
					w.Send(paint.Event{})
				case key.CodeLeftArrow:
					book.PrevPage()
					w.Send(paint.Event{})
				}
			case paint.Event:
				// Clear the background
				draw.Draw(
					buff.RGBA(),
					buff.Bounds(),
					&image.Uniform{&color.RGBA{0, 0, 0, 1}},
					image.ZP,
					draw.Src,
				)

				img, err := book.Current().Image()
				if err != nil {
					log.Fatalln(err)
					return
				}
				drawImgSize := img.Bounds().Size()

				startX := (width - drawImgSize.X) / 2
				startY := (height - drawImgSize.Y) / 2

				draw.Draw(
					buff.RGBA(),
					image.Rect(startX, startY, startX+drawImgSize.X, startY+drawImgSize.Y),
					img,
					image.Point{0, 0},
					draw.Over,
				)

				// Drop our render texture to the screen
				w.Upload(image.Point{0, 0}, buff, buff.Bounds())

				// Publish flips the buffers so we actually display to
				// the screen.
				w.Publish()
			case size.Event:
				width = e.WidthPx
				height = e.HeightPx

				buff.Release()
				buff, err = s.NewBuffer(image.Point{width, height})
				if err != nil {
					return
				}

				book.SetSize(width, height)
			}
		}
	})
}
