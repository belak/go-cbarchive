package main

import (
	"errors"
	"fmt"
	"image"
	"path"
	"sync"

	"github.com/nfnt/resize"
)

type Book struct {
	lock           sync.Mutex
	archive        Archive
	currentIndex   int
	images         []*BookImage
	currentMaxSize image.Point
}

func OpenBook(filename string) (*Book, error) {
	var err error

	ret := &Book{}

	ret.archive, err = OpenArchive(filename)
	if err != nil {
		return nil, err
	}

	for _, f := range ret.archive.Files() {
		switch path.Ext(f.Name()) {
		case ".jpg", ".jpeg", ".png", ".gif":
			ret.images = append(ret.images, &BookImage{book: ret, file: f})
		default:
			continue
		}
	}

	if len(ret.images) == 0 {
		return nil, errors.New("No valid images in book")
	}

	return ret, nil
}

func (b *Book) Close() error {
	return b.archive.Close()
}

func (b *Book) Current() *BookImage {
	return b.images[b.currentIndex]
}

func (b *Book) PrevPage() {
	b.currentIndex--
	if b.currentIndex < 0 {
		b.currentIndex = 0
	}

	b.refill()
}

func (b *Book) NextPage() {
	b.currentIndex++
	if b.currentIndex >= len(b.images) {
		b.currentIndex = len(b.images) - 1
	}

	b.refill()
}

func (b *Book) SetSize(maxWidth, maxHeight int) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.currentMaxSize = image.Point{maxWidth, maxHeight}

	go b.refill()
}

func (b *Book) refill() {
	// We want to be sure to load the current image first
	b.images[b.currentIndex].Image()

	for i, bi := range b.images {
		if b.currentIndex-i >= -5 && b.currentIndex-i <= 5 {
			fmt.Println("Preloading", i)
			go bi.Image()
		} else {
			bi.Clear()
		}
	}
}

type BookImage struct {
	book           *Book
	lock           sync.Mutex
	file           ArchiveFile
	img            image.Image
	resized        image.Image
	currentMaxSize image.Point
}

func (bi *BookImage) Image() (image.Image, error) {
	bi.lock.Lock()
	defer bi.lock.Unlock()

	var err error

	if bi.img == nil {
		bi.img, err = decode(bi.file)
		if err != nil {
			return nil, err
		}
	}

	bi.book.lock.Lock()
	bookMaxSize := bi.book.currentMaxSize
	bi.book.lock.Unlock()

	if bi.resized == nil || bookMaxSize != bi.currentMaxSize {
		bi.resized = resize.Thumbnail(
			uint(bookMaxSize.X), uint(bookMaxSize.Y),
			bi.img, resize.Bilinear,
		)
		bi.currentMaxSize = bookMaxSize
	}

	return bi.resized, nil
}

func (bi *BookImage) Clear() {
	bi.lock.Lock()
	defer bi.lock.Unlock()

	bi.img = nil
	bi.resized = nil
	bi.currentMaxSize = image.Point{0, 0}
}
