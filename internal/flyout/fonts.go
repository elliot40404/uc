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
	regular, bold, icons *opentype.Font
	faces                map[faceKey]font.Face
}

type faceKey struct {
	src  *opentype.Font
	size float64
}

func LoadFonts() *Fonts {
	dir := filepath.Join(os.Getenv("WINDIR"), "Fonts")
	return &Fonts{
		regular: load(filepath.Join(dir, "segoeui.ttf"), goregular.TTF),
		bold:    load(filepath.Join(dir, "seguisb.ttf"), gomedium.TTF),
		icons:   loadFirst(filepath.Join(dir, "SegoeIcons.ttf"), filepath.Join(dir, "segmdl2.ttf")),
		faces:   map[faceKey]font.Face{},
	}
}

func load(path string, fallback []byte) *opentype.Font {
	if f := loadFirst(path); f != nil {
		return f
	}
	f, _ := opentype.Parse(fallback)
	return f
}

func loadFirst(paths ...string) *opentype.Font {
	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			if f, err := opentype.Parse(data); err == nil {
				return f
			}
		}
	}
	return nil
}

func (f *Fonts) face(bold bool, size float64) font.Face {
	if bold {
		return f.of(f.bold, size)
	}
	return f.of(f.regular, size)
}

func (f *Fonts) of(src *opentype.Font, size float64) font.Face {
	key := faceKey{src, size}
	if face, ok := f.faces[key]; ok {
		return face
	}
	face, _ := opentype.NewFace(src, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	f.faces[key] = face
	return face
}
