package steps

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/samber/lo"
	"github.com/vladimir-rom/logex/colors"
	"github.com/vladimir-rom/logex/pipeline"
)

func JsonToText(
	opts pipeline.PipelineOptions,
	props []string,
	noNewLine,
	noProp bool,
	textDelim string,
	highlights []string) pipeline.Step[JSON, string] {
	propsMap := make(map[string]struct{})
	for _, p := range props {
		propsMap[p] = struct{}{}
	}

	c := colors.NewColors()
	highlighter := getHighlighter(highlights, c)

	return pipeline.NewStep[JSON, string](opts, func(obj pipeline.Item[JSON], yield pipeline.Yield[string]) bool {
		res := strings.Builder{}
		for i, p := range props {
			delim := textDelim
			if i == len(props)-1 {
				delim = ""
			}

			if v, ok := obj.Value[p]; ok {
				vstr := fmt.Sprint(v)
				res.Write([]byte(c.ForProperty(p)(vstr) + delim))
			}
		}
		if !noProp {
			for k, v := range obj.Value.SortedByKey {
				if _, ok := propsMap[k]; ok {
					continue
				}
				vstr := fmt.Sprint(v)
				res.Write([]byte(fmt.Sprintf(" %s:%v", c.Property(k), c.ForProperty(k)(vstr))))
			}
		}

		outStr := strings.TrimSpace(res.String())

		if len(outStr) == 0 {
			return true
		}

		if !noNewLine {
			outStr += "\n"
		}

		return yield(pipeline.ToItem[JSON, string](obj, highlighter(outStr)), nil)
	})
}

func getHighlighter(subs []string, c *colors.Colors) func(string) string {
	if !c.Enabled || len(subs) == 0 {
		return func(s string) string {
			return s
		}
	}

	escapedSubs := lo.Map(subs, func(ss string, _ int) string {
		return regexp.QuoteMeta(ss)
	})
	reg := regexp.MustCompile("(?i)" + strings.Join(escapedSubs, "|"))
	return func(s string) string {
		return reg.ReplaceAllStringFunc(s, c.Highlight)
	}
}
