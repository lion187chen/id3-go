// Copyright 2013 Michael Yang. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package v2

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/lion187chen/id3-go/encodedbytes"
)

const (
	HeaderSize = 10
)

// Tag represents an ID3v2 tag
type Tag struct {
	*Header
	frames                []Framer
	padding               uint
	commonMap             map[string]FrameType
	frameHeaderSize       int
	frameConstructor      func(io.Reader) Framer
	frameBytesConstructor func(Framer) []byte
	dirty                 bool
}

// Creates a new tag
func NewTag(version byte) *Tag {
	header := &Header{version: version}

	t := &Tag{
		Header: header,
		frames: make([]Framer, 0, 5),
		dirty:  false,
	}

	switch t.version {
	case 2:
		t.commonMap = V22CommonFrame
		t.frameConstructor = ParseV22Frame
		t.frameHeaderSize = V22FrameHeaderSize
		t.frameBytesConstructor = V22Bytes
	case 3:
		t.commonMap = V23CommonFrame
		t.frameConstructor = ParseV23Frame
		t.frameHeaderSize = FrameHeaderSize
		t.frameBytesConstructor = V23Bytes
	case 4:
		t.commonMap = V24CommonFrame
		t.frameConstructor = ParseV24Frame
		t.frameHeaderSize = FrameHeaderSize
		t.frameBytesConstructor = V24Bytes
	default:
		t.commonMap = V23CommonFrame
		t.frameConstructor = ParseV23Frame
		t.frameHeaderSize = FrameHeaderSize
		t.frameBytesConstructor = V23Bytes
	}

	return t
}

// Parses a new tag
func ParseTag(readSeeker io.ReadSeeker) *Tag {
	header := ParseHeader(readSeeker)

	if header == nil {
		return nil
	}

	t := NewTag(header.version)
	t.Header = header

	var frame Framer
	size := int(t.size)
	for size > 0 {
		frame = t.frameConstructor(readSeeker)

		if frame == nil {
			break
		}

		t.frames = append(t.frames, frame)
		frame.setOwner(t)

		size -= t.frameHeaderSize + int(frame.Size())
	}

	t.padding = uint(size)
	if _, err := readSeeker.Seek(int64(HeaderSize+t.Size()), os.SEEK_SET); err != nil {
		return nil
	}

	return t
}

// Real size of the tag
func (t Tag) RealSize() int {
	size := uint(t.size) - t.padding
	return int(size)
}

func (t *Tag) changeSize(diff int) {
	if d := int(t.padding) - diff; d < 0 {
		t.padding = 0
		t.size += uint32(-d)
	} else {
		t.padding = uint(d)
	}

	t.dirty = true
}

// Modified status of the tag
func (t Tag) Dirty() bool {
	return t.dirty
}

func (t Tag) Bytes() []byte {
	data := make([]byte, t.Size())

	index := 0
	for _, f := range t.frames {
		size := t.frameHeaderSize + int(f.Size())
		copy(data[index:index+size], t.frameBytesConstructor(f))

		index += size
	}

	return append(t.Header.Bytes(), data...)
}

// The amount of padding in the tag
func (t Tag) Padding() uint {
	return t.padding
}

// All frames
func (t Tag) AllFrames() []Framer {
	// Most of the time each ID will only have one frame
	m := len(t.frames)
	frames := make([]Framer, m)

	copy(frames, t.frames)

	return frames
}

// All frames with specified ID
func (t Tag) Frames(id string) []Framer {
	rv := make([]Framer, 0, 1)

	for _, f := range t.frames {
		if f.Id() == id {
			rv = append(rv, f)
		}
	}

	return rv
}

// First frame with specified ID
func (t Tag) Frame(id string) Framer {
	if frames := t.Frames(id); len(frames) != 0 {
		return frames[0]
	}

	return nil
}

// Delete and return all frames with specified ID
func (t *Tag) DeleteFrames(id string) []Framer {
	frames := t.Frames(id)
	if frames == nil {
		return nil
	}

	diff := 0
	i := 0
	for i < len(t.frames) {
		frame := t.frames[i]
		if frame.Id() != id {
			i++
		} else {
			frame.setOwner(nil)
			diff += t.frameHeaderSize + int(frame.Size())
			t.frames = append(t.frames[:i], t.frames[i+1:]...)
		}
	}
	t.changeSize(-diff)

	return frames
}

// Delete the specified frame
func (t *Tag) DeleteFrame(delFrame Framer) []Framer {
	frames := t.AllFrames()
	if frames == nil {
		return nil
	}

	diff := 0
	i := 0
	for i < len(t.frames) {
		frame := t.frames[i]
		if frame == delFrame {
			frame.setOwner(nil)
			diff += t.frameHeaderSize + int(frame.Size())
			t.frames = append(t.frames[:i], t.frames[i+1:]...)
		} else {
			i++
		}
	}
	t.changeSize(-diff)

	return frames
}

// Add frames
func (t *Tag) AddFrames(frames ...Framer) {
	for _, frame := range frames {
		t.changeSize(t.frameHeaderSize + int(frame.Size()))

		t.frames = append(t.frames, frame)
		frame.setOwner(t)
	}
}

func (t Tag) Title() string {
	return t.textFrameText(t.commonMap["Title"])
}

func (t Tag) Artist() string {
	return t.textFrameText(t.commonMap["Artist"])
}

func (t Tag) Album() string {
	return t.textFrameText(t.commonMap["Album"])
}

func (t Tag) Year() string {
	return t.textFrameText(t.commonMap["Year"])
}

func (t Tag) Genre() string {
	return t.textFrameText(t.commonMap["Genre"])
}

func (t Tag) Length() int {
	length, err := strconv.ParseInt(t.textFrameText(t.commonMap["Length"]), 10, 32)
	if err != nil {
		return -1
	}
	return int(length)
}

func (t Tag) Comments() []string {
	frames := t.Frames(t.commonMap["Comments"].Id())
	if frames == nil {
		return nil
	}

	comments := make([]string, len(frames))
	for i, frame := range frames {
		comments[i] = frame.String()
	}

	return comments
}

func (t *Tag) SetTitle(text string) {
	t.setTextFrameText(t.commonMap["Title"], text)
}

func (t *Tag) SetArtist(text string) {
	t.setTextFrameText(t.commonMap["Artist"], text)
}

func (t *Tag) SetAlbum(text string) {
	t.setTextFrameText(t.commonMap["Album"], text)
}

func (t *Tag) SetYear(text string) {
	t.setTextFrameText(t.commonMap["Year"], text)
}

func (t *Tag) SetGenre(text string) {
	t.setTextFrameText(t.commonMap["Genre"], text)
}

func (t *Tag) SetLength(length int) {
	t.setTextFrameText(t.commonMap["Length"], fmt.Sprintf("%d", length))
}

func (t *Tag) textFrame(ft FrameType) TextFramer {
	if frame := t.Frame(ft.Id()); frame != nil {
		if textFramer, ok := frame.(TextFramer); ok {
			return textFramer
		}
	}

	return nil
}

func (t Tag) textFrameText(ft FrameType) string {
	if frame := t.textFrame(ft); frame != nil {
		return frame.Text()
	}

	return ""
}

func (t *Tag) setTextFrameText(ft FrameType, text string) {
	if frame := t.textFrame(ft); frame != nil {
		frame.SetEncoding("UTF-8")
		frame.SetText(text)
	} else {
		f := NewTextFrame(ft, text, "UTF-8")
		t.AddFrames(f)
	}
}

func ParseHeader(reader io.Reader) *Header {
	data := make([]byte, HeaderSize)
	n, err := io.ReadFull(reader, data)
	if n < HeaderSize || err != nil || string(data[:3]) != "ID3" {
		return nil
	}

	size, err := encodedbytes.SynchInt(data[6:])
	if err != nil {
		return nil
	}

	header := &Header{
		version:  data[3],
		revision: data[4],
		flags:    data[5],
		size:     size,
	}

	switch header.version {
	case 2:
		header.unsynchronization = isBitSet(header.flags, 7)
		header.compression = isBitSet(header.flags, 6)
	case 3:
		header.unsynchronization = isBitSet(header.flags, 7)
		header.extendedHeader = isBitSet(header.flags, 6)
		header.experimental = isBitSet(header.flags, 5)
	case 4:
		header.unsynchronization = isBitSet(header.flags, 7)
		header.extendedHeader = isBitSet(header.flags, 6)
		header.experimental = isBitSet(header.flags, 5)
	}

	return header
}

// Header represents the data of the header of the entire tag
type Header struct {
	version, revision byte
	flags             byte
	unsynchronization bool
	compression       bool
	experimental      bool
	extendedHeader    bool
	size              uint32
}

func (h Header) Version() string {
	return fmt.Sprintf("2.%d.%d", h.version, h.revision)
}

func (h Header) Size() int {
	return int(h.size)
}

func (h Header) Bytes() []byte {
	data := make([]byte, 0, HeaderSize)

	data = append(data, "ID3"...)
	data = append(data, h.version, h.revision, h.flags)
	data = append(data, encodedbytes.SynchBytes(h.size)...)

	return data
}
