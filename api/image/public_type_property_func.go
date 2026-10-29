package image

import (
	"fmt"
	"path/filepath"
)

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
