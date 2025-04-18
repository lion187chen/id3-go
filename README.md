# id3

[![build status](https://travis-ci.org/mikkyang/id3-go.svg)](https://travis-ci.org/mikkyang/id3-go)

ID3 library for Go.

Supported formats:

- ID3v1
- ID3v2.2
- ID3v2.3

This project is patched from the following projects:

- [alimoeeny/id3-go](https://github.com/alimoeeny/id3-go)
- [jcs/id3-go](https://github.com/jcs/id3-go)
- [ghenry22/id3-go](https://github.com/ghenry22/id3-go)
- [arkhipovkm/id3-go](https://github.com/arkhipovkm/id3-go)
- [fancybits/id3-go](https://github.com/fancybits/id3-go)
- [richdawe/id3-go](https://github.com/richdawe/id3-go/tree/master)

## Install

The platform ($GOROOT/bin) "go get" tool is the best method to install.

```bash
go get github.com/lion187chen/id3-go
```

This downloads and installs the package into your $GOPATH. If you only want to
recompile, use "go install".

```bash
go install github.com/lion187chen/id3-go
```

## Usage

An import allows access to the package.

```go
import (
    id3 "github.com/lion187chen/id3-go"
)
```

Version specific details can be accessed through the subpackages.

```go
import (
    "github.com/lion187chen/id3-go/v1"
    "github.com/lion187chen/id3-go/v2"
)
```

## Quick Start

To access the tag of a file, first open the file using the package's `Open`
function.

```go
mp3File, err := id3.Open("All-In.mp3")
```

It's also a good idea to ensure that the file is closed using `defer`.

```go
defer mp3File.Close()
```

### Accessing Information

Some commonly used data have methods in the tag for easier access. These
methods are for `Title`, `Artist`, `Album`, `Year`, `Genre`, and `Comments`.

```go
mp3File.SetArtist("Okasian")
fmt.Println(mp3File.Artist())
```

## ID3v2 Frames

v2 Frames can be accessed directly by using the `Frame` or `Frames` method
of the file, which return the first frame or a slice of frames as `Framer`
interfaces. These interfaces allow read access to general details of the file.

```go
lyricsFrame := mp3File.Frame("USLT")
lyrics := lyricsFrame.String()
```

If more specific information is needed, or frame-specific write access is
needed, then the interface must be cast into the appropriate underlying type.
The example provided does not check for errors, but it is recommended to do
so.

```go
lyricsFrame := mp3File.Frame("USLT").(*v2.UnsynchTextFrame)
```

### Adding Frames

For common fields, a frame will automatically be created with the `Set` method.
For other frames or more fine-grained control, frames can be created with the
corresponding constructor, usually prefixed by `New`. These constructors require
the first argument to be a FrameType struct, which are global variables named by
version.

```go
ft := V23FrameTypeMap["TIT2"]
text := "Hello"
textFrame := NewTextFrame(ft, text)
mp3File.AddFrames(textFrame)
```
