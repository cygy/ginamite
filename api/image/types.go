package image

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Property : property of the version of an image.
type Property struct {
	Width        int
	Height       int
	Name         string
	RelativePath string
	AbsolutePath string
}

// Sizes : different sizes of an image.
type Sizes struct {
	Size1x Property
	Size2x Property
}

// Versions : different versions of an image.
type Versions struct {
	Full  Sizes
	Small Sizes
	Thumb Sizes
}

// DefineNameAndPaths : defines the name and the paths of the image.
func (p *Property) DefineNameAndPaths(filename, fileExtension, suffixVersion, suffixSize, directoryDest string) {
	p.Name = filename + suffixVersion + suffixSize + fileExtension
	p.RelativePath = filepath.Clean(fmt.Sprintf("%c/%c/%s", p.Name[0], p.Name[1], p.Name))
	p.AbsolutePath = filepath.Clean(directoryDest + "/" + p.RelativePath)
}

// IsDefined : returns true if the property is defined.
func (p Property) IsDefined() bool {
	return len(p.Name) > 0
}

// IsDefined : returns true if the properties are defined.
func (s Sizes) IsDefined() bool {
	return s.Size1x.IsDefined() && s.Size2x.IsDefined()
}

// All : returns all the properties of the image.
func (v Versions) All() []Property {
	properties := []Property{v.Full.Size1x, v.Full.Size2x}

	if v.Small.IsDefined() {
		properties = append(properties, v.Small.Size1x, v.Small.Size2x)
	}

	if v.Thumb.IsDefined() {
		properties = append(properties, v.Thumb.Size1x, v.Thumb.Size2x)
	}

	return properties
}

// NewVersions : returns a new NewVersions struct.
func NewVersions(src string, width, height int, directoryDest string, saveSmallVersion, saveThumbVersion bool) *Versions {
	ext := strings.ToLower(filepath.Ext(src))
	filename := strings.Replace(strings.ToLower(filepath.Base(src)), ext, "", 1)

	versions := &Versions{}
	versions.Full.Size1x.Width = width
	versions.Full.Size1x.Height = height
	versions.Full.Size1x.DefineNameAndPaths(filename, ext, suffixFull, suffix1x, directoryDest)
	versions.Full.Size2x.Width = width * 2
	versions.Full.Size2x.Height = height * 2
	versions.Full.Size2x.DefineNameAndPaths(filename, ext, suffixFull, suffix2x, directoryDest)

	if saveSmallVersion {
		baseWidth := width / 3
		baseHeight := height / 3

		versions.Small.Size1x.Width = baseWidth
		versions.Small.Size1x.Height = baseHeight
		versions.Small.Size1x.DefineNameAndPaths(filename, ext, suffixSmall, suffix1x, directoryDest)
		versions.Small.Size2x.Width = baseWidth * 2
		versions.Small.Size2x.Height = baseHeight * 2
		versions.Small.Size2x.DefineNameAndPaths(filename, ext, suffixSmall, suffix2x, directoryDest)
	}

	if saveThumbVersion {
		baseWidth := width / 6
		baseHeight := height / 6

		versions.Thumb.Size1x.Width = baseWidth
		versions.Thumb.Size1x.Height = baseHeight
		versions.Thumb.Size1x.DefineNameAndPaths(filename, ext, suffixThumb, suffix1x, directoryDest)
		versions.Thumb.Size2x.Width = baseWidth * 2
		versions.Thumb.Size2x.Height = baseHeight * 2
		versions.Thumb.Size2x.DefineNameAndPaths(filename, ext, suffixThumb, suffix2x, directoryDest)
	}

	return versions
}
