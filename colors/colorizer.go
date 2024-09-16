package colors

import (
	"regexp"

	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/vladimir-rom/logex/cmd/config"
)

type Colorizer struct {
	defaultColors
	Enabled        bool
	def            StrColorizer
	cfg            config.Properties
	colorBuilder   ColorBuilder
	propColorizers map[string]StrColorizer
}

type ColorBuilder func(value ...color.Attribute) StrColorizer

type StrColorizer func(s string) string

type defaultColors struct {
	Err          StrColorizer
	Warn         StrColorizer
	Info         StrColorizer
	Debug        StrColorizer
	PropertyName StrColorizer
	Highlight    StrColorizer
	Timestamp    StrColorizer
}

func newDefaultColors(cb ColorBuilder) *defaultColors {
	gray := toStrColorizer(color.New(90).SprintFunc())
	return &defaultColors{
		Err:          cb(color.FgRed),
		Warn:         cb(color.FgYellow),
		Info:         cb(color.FgBlue),
		Debug:        gray,
		PropertyName: gray,
		Highlight:    cb(color.FgCyan),
		Timestamp:    gray,
	}
}

func NewColorizer(cfg config.Properties, colorBuilder ColorBuilder) (*Colorizer, error) {
	res := &Colorizer{
		Enabled:       !color.NoColor,
		defaultColors: *newDefaultColors(colorBuilder),
		colorBuilder:  colorBuilder,
		def:           func(s string) string { return s },
		cfg:           cfg,
	}
	err := res.initPropertyColorizers()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DefaultColorBuilder(value ...color.Attribute) StrColorizer {
	c := color.New(value...)
	return toStrColorizer(c.SprintFunc())
}

func toStrColorizer(cf func(a ...any) string) StrColorizer {
	return func(s string) string {
		return cf(s)
	}
}

func (c *Colorizer) ForProperty(property string) StrColorizer {
	if strCol, ok := c.propColorizers[property]; ok {
		return strCol
	}
	return c.def
}

func (c *Colorizer) initPropertyColorizers() error {
	c.propColorizers = make(map[string]StrColorizer)
	for prop, conf := range c.cfg {
		if len(conf.Colors) == 0 {
			c.propColorizers[prop] = c.defaultColorizer(prop)
		} else {
			propColorizer, err := c.properyColorizer(conf.Colors)
			if err != nil {
				return nil
			}
			if propColorizer == nil {
				continue
			}

			c.propColorizers[prop] = func(propValue string) string {
				strCol := propColorizer(propValue)
				if strCol == nil {
					return propValue
				}
				return strCol(propValue)
			}
		}
	}
	return nil
}

func (c *Colorizer) defaultColorizer(prop string) StrColorizer {
	switch prop {
	case "level":
		return c.colorizeLogLevel
	case "ts", "timestamp":
		return c.Timestamp
	default:
		return c.def
	}
}

func (c *Colorizer) colorizerForConfig(colorConfig config.Color) StrColorizer {
	if len(colorConfig.Color) > 0 {
		switch colorConfig.Color {
		case config.PColorBlack:
			return c.colorBuilder(color.FgBlack)
		case config.PColorBlue:
			return c.colorBuilder(color.FgBlue)
		case config.PColorCyan:
			return c.colorBuilder(color.FgCyan)
		case config.PColorGreen:
			return c.colorBuilder(color.FgGreen)
		case config.PColorMagenta:
			return c.colorBuilder(color.FgMagenta)
		case config.PColorRed:
			return c.colorBuilder(color.FgRed)
		case config.PColorWhite:
			return c.colorBuilder(color.FgWhite)
		case config.PColorYellow:
			return c.colorBuilder(color.FgYellow)
		default:
			return nil
		}
	}

	if len(colorConfig.CustomColor) > 0 {
		return c.colorBuilder(toAttributes(colorConfig.CustomColor)...)
	}

	return nil
}

func toAttributes(ints []int) []color.Attribute {
	return lo.Map(ints, func(c, _ int) color.Attribute { return color.Attribute(c) })
}

func (c *Colorizer) properyColorizer(colConfigs config.Colors) (func(val string) StrColorizer, error) {
	var res []func(val string) StrColorizer
	for _, colConfig := range colConfigs {
		col := c.colorizerForConfig(colConfig)
		if col == nil {
			continue
		}

		if len(colConfig.Value) > 0 {
			res = append(res, func(val string) StrColorizer {
				if colConfig.Value == val {
					return col
				}
				return nil
			})
		} else if len(colConfig.Pattern) > 0 {
			r, err := regexp.Compile(colConfig.Pattern)
			if err != nil {
				return nil, err
			}
			res = append(res, func(val string) StrColorizer {
				if r.MatchString(val) {
					return col
				}
				return nil
			})
		} else {
			res = append(res, func(val string) StrColorizer { return col })
			break
		}
	}

	if len(res) > 0 {
		return func(val string) StrColorizer {
			for _, f := range res {
				if strCol := f(val); strCol != nil {
					return strCol
				}
			}
			return nil
		}, nil
	}

	return func(val string) StrColorizer {
		return nil
	}, nil
}

func (c *Colorizer) colorizeLogLevel(loglevel string) string {
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
