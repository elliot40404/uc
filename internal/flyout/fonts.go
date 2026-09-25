package flyout

import (
	"os"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

type Fonts struct {
	regular, bold *opentype.Font
	faces         map[faceKey]font.Face
}

type faceKey struct {
	bold bool
	size float64
}

func LoadFonts() *Fonts {
	dir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	return &Fonts{
		regular: load(filepath.Join(dir, "segoeui.ttf"), goregular.TTF),
		bold:    load(filepath.Join(dir, "seguisb.ttf"), gomedium.TTF),
		faces:   map[faceKey]font.Face{},
	}
}

func load(path string, fallback []byte) *opentype.Font {
	if data, err := os.ReadFile(path); err == nil {
		if f, err := opentype.Parse(data); err == nil {
			return f
		}
	}
	f, _ := opentype.Parse(fallback)
	return f
}

func (f *Fonts) face(bold bool, size float64) font.Face {
	key := faceKey{bold, size}
	if face, ok := f.faces[key]; ok {
		return face
	}
	src := f.regular
	if bold {
		src = f.bold
	}
	face, _ := opentype.NewFace(src, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	f.faces[key] = face
	return face
}
