package styles

import "charm.land/lipgloss/v2"

// ThemeForProvider returns the Styles associated with the given provider ID.
// Unknown or empty provider IDs yield the default Stars and Stripes theme.
func ThemeForProvider(providerID string) Styles {
	switch providerID {
	case "hyper":
		return HypercrusherObsidiana()
	default:
		return StarsAndStripes()
	}
}

// StarsAndStripes returns Crusher's red, white, and blue dark theme.
func StarsAndStripes() Styles {
	return quickStyle(quickStyleOpts{
		primary:   lipgloss.Color("#3c5aa6"),
		secondary: lipgloss.Color("#b22234"),
		accent:    lipgloss.Color("#ffffff"),
		keyword:   lipgloss.Color("#bf0a30"),

		fgBase:       lipgloss.Color("#f8fafc"),
		fgMoreSubtle: lipgloss.Color("#cbd5e1"),
		fgSubtle:     lipgloss.Color("#e2e8f0"),
		fgMostSubtle: lipgloss.Color("#94a3b8"),

		onPrimary: lipgloss.Color("#ffffff"),

		bgBase:         lipgloss.Color("#07111f"),
		bgLeastVisible: lipgloss.Color("#0b1f3a"),
		bgLessVisible:  lipgloss.Color("#102a4c"),
		bgMostVisible:  lipgloss.Color("#1f3f70"),

		separator: lipgloss.Color("#3c5aa6"),

		destructive:       lipgloss.Color("#bf0a30"),
		error:             lipgloss.Color("#ff4d5d"),
		warningSubtle:     lipgloss.Color("#f5c542"),
		warning:           lipgloss.Color("#ffb703"),
		denied:            lipgloss.Color("#b22234"),
		busy:              lipgloss.Color("#ffffff"),
		info:              lipgloss.Color("#60a5fa"),
		infoMoreSubtle:    lipgloss.Color("#3c5aa6"),
		infoMostSubtle:    lipgloss.Color("#1f3f70"),
		success:           lipgloss.Color("#22c55e"),
		successMoreSubtle: lipgloss.Color("#86efac"),
		successMostSubtle: lipgloss.Color("#164e33"),

		ansiBlack:   lipgloss.Color("#07111f"),
		ansiRed:     lipgloss.Color("#bf0a30"),
		ansiGreen:   lipgloss.Color("#22c55e"),
		ansiYellow:  lipgloss.Color("#f5c542"),
		ansiBlue:    lipgloss.Color("#3c5aa6"),
		ansiMagenta: lipgloss.Color("#b22234"),
		ansiCyan:    lipgloss.Color("#60a5fa"),
		ansiWhite:   lipgloss.Color("#f8fafc"),

		ansiBrightBlack:   lipgloss.Color("#334155"),
		ansiBrightRed:     lipgloss.Color("#ff4d5d"),
		ansiBrightGreen:   lipgloss.Color("#86efac"),
		ansiBrightYellow:  lipgloss.Color("#ffb703"),
		ansiBrightBlue:    lipgloss.Color("#60a5fa"),
		ansiBrightMagenta: lipgloss.Color("#ef4444"),
		ansiBrightCyan:    lipgloss.Color("#93c5fd"),
		ansiBrightWhite:   lipgloss.Color("#ffffff"),
	})
}

// CatppuccinMocha is kept for older tests and callers inside this repo.
func CatppuccinMocha() Styles {
	return StarsAndStripes()
}

// CharmtonePantera is kept for older tests and callers inside this repo.
func CharmtonePantera() Styles {
	return StarsAndStripes()
}

// HypercrusherObsidiana returns the Hypercrusher dark theme.
func HypercrusherObsidiana() Styles {
	return StarsAndStripes()
}
