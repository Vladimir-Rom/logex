package colors

import (
	"fmt"

	"github.com/fatih/color"
)

type Colorize func(s string) string

type Colors struct {
	Enabled   bool
	Err       Colorize
	Warn      Colorize
	Info      Colorize
	Debug     Colorize
	Property  Colorize
	Highlight Colorize
	Timestamp Colorize
	def       Colorize
}

func NewColors() *Colors {
	toColorize := func(cf func(a ...any) string) Colorize {
		return func(s string) string {
			return cf(s)
		}
	}
	gray := toColorize(color.New(90).SprintFunc())

	return &Colors{
		Enabled:   !color.NoColor,
		Err:       toColorize(color.New(color.FgRed).SprintFunc()),
		Warn:      toColorize(color.New(color.FgYellow).SprintFunc()),
		Info:      toColorize(color.New(color.FgBlue).SprintFunc()),
		Debug:     gray,
		Property:  gray,
		Highlight: toColorize(color.New(color.FgCyan).SprintFunc()),
		Timestamp: gray,
		def:       toColorize(fmt.Sprint),
	}
}

func (c *Colors) ForProperty(property string) Colorize {
	switch property {
	case "level":
		return c.colorizeLogLevel
	case "ts", "timestamp":
		return c.Timestamp
	default:
		return c.def
	}
}

func (c *Colors) colorizeLogLevel(loglevel string) string {
	switch loglevel {
	case "err", "error":
		return c.Err(loglevel)
	case "warn", "warning":
		return c.Warn(loglevel)
	case "info":
		return c.Info(loglevel)
	case "deb", "debug":
		return c.Debug(loglevel)
	default:
		return c.def(loglevel)
	}
}
