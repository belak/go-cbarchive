package main

import (
	"io"
	"io/ioutil"

	"github.com/nwaples/rardecode"
)

type rarArchive struct {
	filename string
	archive  *rardecode.ReadCloser
}

func openRARArchive(filename string) (Archive, error) {
	var err error
	ret := &rarArchive{filename: filename}
	ret.archive, err = rardecode.OpenReader(filename, "")
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// TODO: This breaks horribly - we either need to cache the result or
// re-read every time we need the list, both of which aren't great
// options.
func (a *rarArchive) Files() []ArchiveFile {
	var ret []ArchiveFile
	for af, err := a.archive.Next(); err != nil; af, err = a.archive.Next() {
		ret = append(ret, &rarArchiveFile{
			archive:  a,
			filename: af.Name,
		})
	}
	return ret
}

func (a *rarArchive) Comment() string {
	return ""
}

func (a *rarArchive) Close() error {
	return a.archive.Close()
}

type rarArchiveFile struct {
	archive  *rarArchive
	filename string
}

func (f *rarArchiveFile) Name() string {
	return f.filename
}

func (f *rarArchiveFile) Open() (io.ReadCloser, error) {
	// Because the only way with this library to get back to a
	// specific file, we need to reopen the file and iterate through
	// everything.
	archive, err := rardecode.OpenReader(f.archive.filename, "")
	if err != nil {
		return nil, err
	}

	for af, err := archive.Next(); err != nil; af, err = archive.Next() {
		if f.filename == af.Name {
			// We need to wrap the archive so we don't accidentally
			// close the full archive, even though it shouldn't matter
			// because we have to iterate through the whole thing
			// anyway.
			return ioutil.NopCloser(f.archive.archive), nil
		}
	}

	return nil, ErrInternalError
}

func (f *rarArchiveFile) Comment() string {
	return ""
}
