package main

import (
	"errors"
	"io"
	"path"
)

var (
	ErrUnknownFiletype = errors.New("cbarchive: unknown file extension")
	ErrInternalError   = errors.New("cbarchive: internal error")
)

var archiveFormatMap = map[string]func(string) (Archive, error){
	".cbz": openZipArchive,
	".cbr": openRARArchive,
}

// Archive represents a full archive file. Only a small subset of
// functions is available to allow for a common interface between
// different formats.
type Archive interface {
	Files() []ArchiveFile
	Comment() string
	Close() error
}

// ArchiveFile represents a file header in an archive file. Only a
// small subset of functions is available to allow for a common
// interface between different formats.
type ArchiveFile interface {
	Name() string
	Comment() string
	Open() (io.ReadCloser, error)
}

func OpenArchive(filename string) (Archive, error) {
	ext := path.Ext(filename)
	constructor, ok := archiveFormatMap[ext]
	if !ok {
		return nil, ErrUnknownFiletype
	}
	return constructor(filename)
}
