package steps

import (
	"fmt"
	"regexp"

	"github.com/vladimir-rom/logex/pipeline"
)

func IncludeRegexp(opts pipeline.PipelineOptions, regexps []string) (pipeline.Step[string, string], error) {
	if len(regexps) == 0 {
		return Noop[string](), nil
	}

	rs := make([]*regexp.Regexp, len(regexps))
	for i := range regexps {
		r, err := regexp.Compile(regexps[i])
		if err != nil {
			return nil, fmt.Errorf("invalid regular expression %s: %w", regexps[i], err)
		}
		rs[i] = r
	}

	return pipeline.NewStep(opts, func(line pipeline.Item[string], yield pipeline.Yield[string]) bool {
		if line.Metadata.Removed {
			return yield(line, nil)
		}

		for _, r := range rs {
			if r.MatchString(line.Value) {
				return yield(line, nil)
			}
		}
		line.Metadata.Removed = true
		return yield(line, nil)
	}), nil
}

func ExcludeRegexp(opts pipeline.PipelineOptions, regexps []string) (pipeline.Step[string, string], error) {
	if len(regexps) == 0 {
		return Noop[string](), nil
	}

	rs := make([]*regexp.Regexp, len(regexps))
	for i := range regexps {
		r, err := regexp.Compile(regexps[i])
		if err != nil {
			return nil, fmt.Errorf("invalid regular expression %s: %w", regexps[i], err)
		}
		rs[i] = r
	}

	return pipeline.NewStep(opts, func(line pipeline.Item[string], yield pipeline.Yield[string]) bool {
		if line.Metadata.Removed {
			return yield(line, nil)
		}

		for _, r := range rs {
			if r.MatchString(line.Value) {
				line.Metadata.Removed = true
				return yield(line, nil)
			}

		}
		return yield(line, nil)
	}), nil
}
