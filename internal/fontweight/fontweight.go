package fontweight

import (
	"fyne.io/fyne/v2"
	internalapp "fyne.io/fyne/v2/internal/app"
	"fyne.io/fyne/v2/internal/cache"
	"fyne.io/fyne/v2/internal/painter"
	"github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype"
)

// Enable must run after SetTheme and before creating widgets.
func Enable(app fyne.App) {
	applyWeights()
	update := func(settings fyne.Settings) {
		applyWeights()
		cache.ResetThemeCaches()
		internalapp.ApplySettings(settings, app)
	}
	// Fyne registers its cache-reset listener when the driver starts.
	app.Lifecycle().SetOnStarted(func() {
		app.Settings().AddListener(update)
		update(app.Settings())
	})
}

func applyWeights() {
	// shortcut: Fyne 2.8.1의 내부 캐시에 의존하며, 가변 폰트 공개 API가 생기면 대체한다.
	for _, style := range []fyne.TextStyle{{}, {Bold: true}, {Italic: true}, {Bold: true, Italic: true}} {
		weight := font.WeightNormal
		if style.Bold {
			weight = font.WeightBold
		}
		// Select the primary Latin font without changing the default theme's Korean fallback.
		face := painter.CachedFontFace(style, nil, nil).Fonts.ResolveFace('A')
		face.SetVariations([]font.Variation{{Tag: opentype.NewTag('w', 'g', 'h', 't'), Value: float32(weight)}})
	}
}
