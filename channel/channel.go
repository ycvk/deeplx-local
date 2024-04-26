package channel

import "os"

var (
	Restart = make(chan os.Signal, 1)
	Quit    = make(chan os.Signal, 1)
)
