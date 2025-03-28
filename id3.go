// Copyright 2013 Michael Yang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package id3

import (
	"bytes"
	"errors"
	"os"

	v1 "github.com/lion187chen/id3-go/v1"
	v2 "github.com/lion187chen/id3-go/v2"
)

const (
	LatestVersion = 3
)

// Tagger represents the metadata of a tag
type Tagger interface {
	Title() string
	Artist() string
	Album() string
	Year() string
	Genre() string
	Length() int
	Comments() []string
	SetTitle(string)
	SetArtist(string)
	SetAlbum(string)
	SetYear(string)
	SetGenre(string)
	SetLength(int)
	AllFrames() []v2.Framer
	Frames(string) []v2.Framer
	Frame(string) v2.Framer
	DeleteFrames(string) []v2.Framer
	DeleteFrame(v2.Framer) []v2.Framer
	AddFrames(...v2.Framer)
	Bytes() []byte
	Dirty() bool
	Padding() uint
	Size() int
	Version() string
}

// File represents the tagged file
type File struct {
	Tagger
	originalSize int
	file         *os.File
}

type Mp3Bytes struct {
	Tagger
	originalSize int
	blob         []byte
}

// Parses an open file
func Parse(file *os.File) (*File, error) {
	res := &File{file: file}

	if v2Tag := v2.ParseTag(file); v2Tag != nil {
		res.Tagger = v2Tag
		res.originalSize = v2Tag.Size()
	} else if v1Tag := v1.ParseTag(file); v1Tag != nil {
		res.Tagger = v1Tag
	} else {
		// Add a new tag if none exists
		res.Tagger = v2.NewTag(LatestVersion)
	}

	return res, nil
}

// NewMp3Bytes should match Parse above but for in memory mp3 data not on disk files
func NewMp3Bytes(blob []byte) (*Mp3Bytes, error) {
	res := &Mp3Bytes{blob: blob}

	if v2Tag := v2.ParseTag(bytes.NewReader(blob)); v2Tag != nil {
		res.Tagger = v2Tag
		res.originalSize = v2Tag.Size()
	} else if v1Tag := v1.ParseTag(bytes.NewReader(blob)); v1Tag != nil {
		res.Tagger = v1Tag
	} else {
		// Add a new tag if none exists
		res.Tagger = v2.NewTag(LatestVersion)
	}

	return res, nil
}

// Opens a new tagged file
func Open(name string) (*File, error) {
	fi, err := os.OpenFile(name, os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	file, err := Parse(fi)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Saves any edits to the tagged file
func (f *File) Close() error {
	defer f.file.Close()

	if !f.Dirty() {
		return nil
	}

	switch f.Tagger.(type) {
	case (*v1.Tag):
		if _, err := f.file.Seek(-v1.TagSize, os.SEEK_END); err != nil {
			return err
		}
	case (*v2.Tag):
		if f.Size() > f.originalSize {
			start := int64(f.originalSize + v2.HeaderSize)
			offset := int64(f.Tagger.Size() - f.originalSize)

			if err := shiftBytesBack(f.file, start, offset); err != nil {
				return err
			}
		}

		if _, err := f.file.Seek(0, os.SEEK_SET); err != nil {
			return err
		}
	default:
		return errors.New("Close: unknown tag version")
	}

	if _, err := f.file.Write(f.Tagger.Bytes()); err != nil {
		return err
	}

	return nil
}

// UpdateEditsIntoBytes is like Close above but for in memory mp3 data not on disk
func (b *Mp3Bytes) UpdateEditsIntoBytes() (*[]byte, error) {
	if !b.Dirty() {
		return &b.blob, nil
	}
	start := int64(0)
	offset := int64(0)

	switch b.Tagger.(type) {
	case (*v1.Tag):
		//unless I am much mistaken in v1 the tags are at the end of the file
		offset = int64(len(b.blob)) - v1.TagSize

	case (*v2.Tag):
		if b.Size() > b.originalSize {
			start = int64(b.originalSize + v2.HeaderSize)
			offset = int64(b.Tagger.Size() - b.originalSize)
			b.blob = shiftBytesBackInMem(b.blob, start, offset)
		}

	default:
		return nil, errors.New("Close: unknown tag version")
	}

	insert := b.Tagger.Bytes()
	copy(b.blob[0:start+offset], insert)
	return &b.blob, nil
}

func shiftBytesBackInMem(blob []byte, start, offset int64) []byte {
	out := make([]byte, int64(len(blob))+offset)
	copy(out, blob[:start])
	copy(out[start+offset:], blob[start:])
	return out
}
