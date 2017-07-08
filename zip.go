package main

import (
	"archive/zip"
	"io"
)

type zipArchive struct {
	archive *zip.ReadCloser
}

func openZipArchive(filename string) (Archive, error) {
	var err error
	ret := &zipArchive{}
	ret.archive, err = zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (a *zipArchive) Files() []ArchiveFile {
	var ret []ArchiveFile
	for _, f := range a.archive.File {
		ret = append(ret, &zipArchiveFile{file: f})
	}
	return ret
}

func (a *zipArchive) Comment() string {
	return a.archive.Comment
}

func (a *zipArchive) Close() error {
	return a.archive.Close()
}

type zipArchiveFile struct {
	file *zip.File
}

func (f *zipArchiveFile) Name() string {
	return f.file.Name
}

func (f *zipArchiveFile) Open() (io.ReadCloser, error) {
	return f.file.Open()
}

func (f *zipArchiveFile) Comment() string {
	return f.file.Comment
}
