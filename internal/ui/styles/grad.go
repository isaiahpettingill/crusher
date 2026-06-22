package styles

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/rivo/uniseg"
)

// ForegroundGrad returns a red-white-blue foreground ramp.
func ForegroundGrad(base lipgloss.Style, input string, bold bool, color1, color2 color.Color) []string {
	if input == "" {
		return []string{""}
	}
	_ = color1
	_ = color2
	var clusters []string
	gr := uniseg.NewGraphemes(input)
	for gr.Next() {
		clusters = append(clusters, string(gr.Runes()))
	}
	palette := []color.Color{lipgloss.Color("#bf0a30"), lipgloss.Color("#ffffff"), lipgloss.Color("#3c5aa6")}
	for i := range clusters {
		style := base.Foreground(palette[i%len(palette)])
		if bold {
			style = style.Bold(true)
		}
		clusters[i] = style.Render(clusters[i])
	}
	return clusters
}

// ApplyForegroundGrad renders a given string with a horizontal gradient
// foreground.
func ApplyForegroundGrad(base lipgloss.Style, input string, color1, color2 color.Color) string {
	if input == "" {
		return ""
	}
	var o strings.Builder
	for _, c := range ForegroundGrad(base, input, false, color1, color2) {
		fmt.Fprint(&o, c)
	}
	return o.String()
}

// ApplyBoldForegroundGrad renders a given string with a horizontal gradient
// foreground.
func ApplyBoldForegroundGrad(base lipgloss.Style, input string, color1, color2 color.Color) string {
	if input == "" {
		return ""
	}
	var o strings.Builder
	for _, c := range ForegroundGrad(base, input, true, color1, color2) {
		fmt.Fprint(&o, c)
	}
	return o.String()
}
