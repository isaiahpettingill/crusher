//go:build !darwin

package notification

import (
	_ "embed"
)

//go:embed crusher-icon-solo.png
var Icon []byte
