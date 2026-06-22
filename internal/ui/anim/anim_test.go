package anim

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestAnimCyclesThroughFoundingDocumentText(t *testing.T) {
	t.Parallel()

	a := New(Settings{ID: "docs", Size: 12})
	require.NotEmpty(t, foundingDocumentRunes)

	for range maxBirthSteps {
		a.Animate(StepMsg{ID: "docs"})
	}

	got := ansi.Strip(a.Render())[:12]
	want := string(foundingDocumentRunes[maxBirthSteps : maxBirthSteps+12])
	require.Equal(t, want, got)

	for range len(foundingDocumentRunes) {
		a.Animate(StepMsg{ID: "docs"})
	}

	cycled := ansi.Strip(a.Render())[:12]
	require.Equal(t, want, cycled)
}
