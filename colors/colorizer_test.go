package colors

import (
	"fmt"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/require"
	"github.com/vladimir-rom/logex/cmd/config"
)

func TestColorizer(t *testing.T) {
	testee, err := NewColorizer(
		config.Properties{
			"p1": config.Property{
				Colors: []config.Color{
					{
						Color: config.PColorRed,
						Value: "v1",
					},
					{
						Color: config.PColorGreen,
						Value: "v2",
					},
					{
						Color:   config.PColorBlue,
						Pattern: ".3",
					},
				},
			},
			"p2": config.Property{
				Colors: []config.Color{
					{
						Color: config.PColorMagenta,
					},
				},
			},
		},
		colorBuilder,
	)

	require.NoError(t, err)
	checkPropertyColor(t, testee, "p1", "v1", color.FgRed)
	checkPropertyColor(t, testee, "p1", "v2", color.FgGreen)
	checkPropertyColor(t, testee, "p1", "r3", color.FgBlue)
	checkPropertyColor(t, testee, "p1", "unknown value")
	checkPropertyColor(t, testee, "p2", "p2", color.FgMagenta)
	checkPropertyColor(t, testee, "unknown property", "unknown value")
}

func checkPropertyColor(t *testing.T, testee *Colorizer, propName, propValue string, expectedColor ...color.Attribute) {
	t.Helper()
	res := testee.ForProperty(propName)(propValue)
	if len(expectedColor) == 0 {
		require.Equal(t, propValue, res)
	} else {
		require.Equal(t, fmt.Sprintf("%v", expectedColor), res)
	}
}

func colorBuilder(value ...color.Attribute) StrColorizer {
	return func(s string) string {
		return fmt.Sprintf("%v", value)
	}
}
