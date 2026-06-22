//go:build !darwin

package notification

import (
	_ "embed"
)

//go:embed chuck.png
var Icon []byte
