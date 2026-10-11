package main

import (
	"bytes"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/fontscan"
)

type systemFontTheme struct {
	fyne.Theme
	regular fyne.Resource
	bold    fyne.Resource
}

func (t *systemFontTheme) Font(style fyne.TextStyle) fyne.Resource {
	if style.Symbol || style.Monospace {
		return t.Theme.Font(style)
	}
	if style.Bold {
		return t.bold
	}
	return t.regular
}

func newSystemFontTheme() fyne.Theme {
	base := theme.DefaultTheme()
	fonts := fontscan.NewFontMap(log.Default())
	if err := fonts.UseSystemFonts(""); err != nil {
		log.Printf("failed to scan system fonts: %v", err)
		return base
	}

	for _, family := range []string{"Noto Sans KR", "Malgun Gothic"} {
		if _, found := fonts.FindSystemFont(family); !found {
			continue
		}

		var resources [2]fyne.Resource
		for i, weight := range []font.Weight{font.WeightNormal, font.WeightBold} {
			fonts.SetQuery(fontscan.Query{
				Families: []string{family},
				Aspect:   font.Aspect{Style: font.StyleNormal, Weight: weight},
			})
			face := fonts.ResolveFace('가')
			if face == nil {
				break
			}
			matched, _ := fonts.FontMetadata(face.Font)
			if matched != font.NormalizeFamily(family) {
				break
			}
			path := fonts.FontLocation(face.Font).File
			resource, err := fyne.LoadResourceFromPath(path)
			if err != nil {
				log.Printf("failed to load system font %s: %v", path, err)
				break
			}
			if _, err := font.ParseTTF(bytes.NewReader(resource.Content())); err != nil {
				log.Printf("unsupported system font %s: %v", path, err)
				break
			}
			resources[i] = resource
		}
		if resources[0] != nil && resources[1] != nil {
			return &systemFontTheme{Theme: base, regular: resources[0], bold: resources[1]}
		}
	}
	log.Print("Noto Sans KR and Malgun Gothic unavailable; using the default font")
	return base
}
