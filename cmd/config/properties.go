package config

type (
	Properties map[string]Property
	Property   struct {
		Colors
	}
	Colors []Color
	Color  struct {
		Color       PColor
		CustomColor []int
		Value       string
		Pattern     string
	}
	PColor string
)

const (
	PColorBlack   PColor = "black"
	PColorRed     PColor = "red"
	PColorGreen   PColor = "green"
	PColorYellow  PColor = "yellow"
	PColorBlue    PColor = "blue"
	PColorMagenta PColor = "magenta"
	PColorCyan    PColor = "cyan"
	PColorWhite   PColor = "white"
)
