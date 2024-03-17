package commands

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	"github.com/vladimir-rom/logex/cmd/config"
	"github.com/vladimir-rom/logex/pipeline"
	"github.com/vladimir-rom/logex/steps"
)

func Execute() {
	var rootCmd = createRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type filterParams struct {
	fileName string
	config   string

	// filters
	kqlFilter     func() string
	jq            func() string
	include       func() []string
	exclude       func() []string
	includeRegexp func() []string
	excludeRegexp func() []string

	// properties
	selectProps func() []string
	hideProps   func() []string
	expandProps func() []string
	durationMs  func() []string
	metadata    func() string

	// text formatting
	outputFormat  func() string
	textFormat    func() []string
	textNoNewLine func() bool
	textDelim     func() string
	textNoProp    func() bool
	highlights    func() []string

	// post processing
	distinctBy func() string
	first      func() int
	last       func() int
	context    func() int

	// debug
	showErrors func() bool
}

func createRootCmd() *cobra.Command {
	var params filterParams
	var k = koanf.New(".")

	filterCmd := &cobra.Command{
		Use:   "logex [flags] file-name",
		Short: "logex is a tool for filtering and formatting structured log files",
		Run: func(cmd *cobra.Command, args []string) {
			params.fileName = args[0]
			err := loadConfiguration(&params, k, cmd)
			if err != nil {
				log.Fatal(err)
			}

			err = doFilter(&params, cmd)
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	defineFlags(
		config.NewRegistry(k, filterCmd.Flags()),
		&params)

	filterCmd.Flags().StringVar(
		&params.config,
		"config",
		"",
		"configuration file name")

	return filterCmd
}

func loadConfiguration(params *filterParams, k *koanf.Koanf, cmd *cobra.Command) error {
	if len(params.config) == 0 {
		params.config = os.Getenv("LOGEX_CONFIG")
	}
	if len(params.config) > 0 {
		if err := k.Load(file.Provider(params.config), yaml.Parser()); err != nil {
			return fmt.Errorf("error loading config file: %v", err)
		}
	}
	if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
		return fmt.Errorf("error loading command line config: %v", err)
	}

	return nil
}

func defineFlags(reg *config.Registry, params *filterParams) {
	params.kqlFilter = reg.StringP(
		"kql",
		"f",
		"",
		"Filter in the Kibana Query Language format. Example: 'level:(error OR warn)'")

	params.jq = reg.String(
		"jq",
		"",
		"specify a jq expression for filtering or transformation. Example: '.level==\"info\" or .level==\"warn\"'")

	params.include = reg.StringsP(
		"include",
		"i",
		nil,
		"include only records containing any of the specified substrings")

	params.exclude = reg.StringsP(
		"exclude",
		"e",
		nil,
		"exclude records containing any of the specified substrings")

	params.includeRegexp = reg.Strings(
		"include-regexp",
		nil,
		"include only records that match any of the specified regular expressions")

	params.excludeRegexp = reg.Strings(
		"exclude-regexp",
		nil,
		"exclude records that match any of the specified regular expressions")

	params.durationMs = reg.Strings(
		"duration-ms",
		nil,
		"treat specified fields as duration strings and convert them to milliseconds (useful for filtering)")

	params.selectProps = reg.Strings(
		"select",
		nil,
		"property names to output")

	params.hideProps = reg.Strings(
		"hide",
		nil,
		"property names to hide")

	params.expandProps = reg.Strings(
		"expand",
		nil,
		"parse property names with string values as JSON objects for use in filters and other operations")

	params.showErrors = reg.Bool(
		"show-errors",
		false,
		"show processing errors")

	params.textFormat = reg.StringsP(
		"txt-format",
		"t",
		nil,
		"property names to be printed first in plain text format")

	params.outputFormat = reg.String(
		"format",
		"text",
		"output format, can be \"text\" or \"json\"")

	params.textNoNewLine = reg.Bool(
		"txt-nonl",
		false,
		"do not add new lines after each record")

	params.textNoProp = reg.Bool(
		"txt-noprop",
		false,
		"do not print properties except those selected in the format string (txt-format)")

	params.textDelim = reg.String(
		"txt-delim",
		"|",
		"delimiter between text properties")

	params.distinctBy = reg.String(
		"distinct-by",
		"",
		"return distinct records based on the specified property names")

	params.highlights = reg.StringsP(
		"highlight",
		"l",
		nil,
		"highlight substrings in the output")

	params.first = reg.Int(
		"first",
		0,
		"print only the first N matched records",
	)

	params.last = reg.Int(
		"last",
		0,
		"print only the last N matched records",
	)

	params.context = reg.Int(
		"context",
		0,
		"print N additional records before and after matches",
	)

	params.metadata = reg.StringP(
		"metadata",
		"m",
		"rnum",
		"add metadata fields. Format: name[:property-name]. \nExamples:\n"+
			"'rnum' - adds an rnum field with the record number\n"+
			"'rnum:r1 file:f1' - adds field r1 with the record number and f1 with the name of the logfile",
	)
}

func (p *filterParams) Validate() error {
	switch f := p.outputFormat(); f {
	case "text":
	case "json":
		break
	default:
		return fmt.Errorf("Unknown output format: %s", f)
	}
	return nil
}

func doFilter(params *filterParams, cmd *cobra.Command) error {
	if err := params.Validate(); err != nil {
		return err
	}

	var reader io.Reader
	var fileName string
	if params.fileName == "-" {
		reader = cmd.InOrStdin()
		fileName = "stdin"
	} else {
		fileName = params.fileName
		close, r, err := steps.OpenFile(params.fileName)
		if err != nil {
			return err
		}
		reader = r
		defer close()
	}

	return runPipeline(params, fileName, reader, cmd.OutOrStdout())
}

func runPipeline(params *filterParams, filename string, r io.Reader, w io.Writer) error {
	opts := pipeline.PipelineOptions{
		ContextEnabled: params.context() > 0,
	}
	filterByKQL, err := steps.FilterByKQL(opts, params.kqlFilter())
	if err != nil {
		return err
	}

	filterByJq, err := steps.FilterByJq(opts, params.jq())
	if err != nil {
		return err
	}

	addMeta, err := steps.AddMeta(opts, params.metadata())
	if err != nil {
		return err
	}

	input := steps.ReadByLines(filename, r)

	var formatJSONToText pipeline.Step[steps.JSON, string]

	if len(params.textFormat()) > 0 || params.outputFormat() == "text" {
		formatJSONToText = steps.JsonToText(
			opts,
			params.textFormat(),
			params.textNoNewLine(),
			params.textNoProp(),
			params.textDelim(),
			slices.Concat(params.include(), params.highlights()))
	} else {
		formatJSONToText = steps.JsonToStr(opts)
	}

	includeRegexp, err := steps.IncludeRegexp(opts, params.includeRegexp())
	if err != nil {
		return err
	}
	excludeRegexp, err := steps.ExcludeRegexp(opts, params.excludeRegexp())
	if err != nil {
		return err
	}

	processStringInput := pipeline.Combine(
		steps.RemovePrefix(opts),
		steps.ExcludeSubstringsAny(opts, params.exclude()),
		steps.IncludeSubstringsAny(opts, params.include()),
		includeRegexp,
		excludeRegexp,
	)

	processJSON := pipeline.Combine(
		addMeta,
		steps.Expand(opts, params.expandProps()),
		filterByKQL,
		filterByJq,
		steps.DistinctBy(opts, params.distinctBy()),
		steps.Hide(opts, params.hideProps()),
		steps.Select(opts, params.selectProps()),
		steps.Context(opts, params.context(), params.context()),
		steps.First(opts, params.first()),
		steps.Last(opts, params.last()),
	)

	return steps.WriteLines(
		w,
		params.showErrors(),
		formatJSONToText(
			processJSON(
				steps.StrToJson(opts, params.durationMs())(
					processStringInput(input)))))
}
