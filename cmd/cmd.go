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
	"github.com/samber/lo"
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
	fileNames []string
	config    string

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
	headProps     func() []string
	orderProps    func() []string
	textNoNewLine func() bool
	textDelim     func() string
	textNoProp    func() bool
	highlights    func() []string

	// post processing
	distinctBy func() string
	mergeBy    func() []string
	first      func() int
	last       func() int
	context    func() int

	// debug
	showErrors func() bool

	propertiesConfig config.Properties
}

type fileDescr struct {
	fileName string
	r        io.Reader
	close    func() error
	err      error
}

func createRootCmd() *cobra.Command {
	var params filterParams
	var k = koanf.New(".")

	filterCmd := &cobra.Command{
		Use:   "logex [flags] file-name",
		Short: "logex is a tool for filtering and formatting structured log files",
		Run: func(cmd *cobra.Command, args []string) {
			params.fileNames = args
			err := loadConfiguration(&params, k, cmd)
			if err != nil {
				log.Fatal(err)
			}

			err = doFilter(&params, cmd)
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.MinimumNArgs(1),
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

	k.Unmarshal("properties", &params.propertiesConfig)

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
		"Specify a jq expression for filtering or transformation. Example: '.level==\"info\" or .level==\"warn\"'")

	params.include = reg.StringsP(
		"include",
		"i",
		nil,
		"Include only records containing any of the specified substrings")

	params.exclude = reg.StringsP(
		"exclude",
		"e",
		nil,
		"Exclude records containing any of the specified substrings")

	params.includeRegexp = reg.Strings(
		"include-regexp",
		nil,
		"Include only records that match any of the specified regular expressions")

	params.excludeRegexp = reg.Strings(
		"exclude-regexp",
		nil,
		"Exclude records that match any of the specified regular expressions")

	params.durationMs = reg.Strings(
		"duration-ms",
		nil,
		"Treat specified fields as duration strings and convert them to milliseconds (useful for filtering)")

	params.selectProps = reg.Strings(
		"select",
		nil,
		"Property names to output, other properties will be skipped")

	params.hideProps = reg.Strings(
		"hide",
		nil,
		"Property names to hide")

	params.expandProps = reg.Strings(
		"expand",
		nil,
		"Parse property names with string values as JSON objects for use in filters and other operations")

	params.showErrors = reg.Bool(
		"show-errors",
		false,
		"Show processing errors")

	params.headProps = reg.StringsP(
		"txt-head",
		"t",
		nil,
		"Specify property names whose values will be displayed at the beginning of the record without\n"+
			"printing property names. Other properties will follow. Applicable for text format.")

	params.orderProps = reg.Strings(
		"order",
		nil,
		"Specify property names to be displayed at the beginning of the record. Other properties will follow. "+
			"\nApplicable for text format.")

	params.textNoNewLine = reg.Bool(
		"txt-nonl",
		false,
		"Do not add new lines after each record.\nApplicable for text format.")

	params.textNoProp = reg.Bool(
		"txt-noprop",
		false,
		"Exclude printing properties except those explicitly selected in --txt-head or --order.\nApplicable for text format.")

	params.textDelim = reg.String(
		"txt-delim",
		"|",
		"Delimiter between text properties")

	params.outputFormat = reg.String(
		"format",
		"text",
		"Output format, can be \"text\" or \"json\"")

	params.distinctBy = reg.String(
		"distinct-by",
		"",
		"Return distinct records based on the specified property names")

	params.mergeBy = reg.Strings(
		"merge",
		[]string{"ts"},
		"Merge multiple files into single stream of records by specified fields (usually by timestamp)")

	params.highlights = reg.StringsP(
		"highlight",
		"l",
		nil,
		"Highlight substrings in the output")

	params.first = reg.Int(
		"first",
		0,
		"Print only the first N matched records",
	)

	params.last = reg.Int(
		"last",
		0,
		"Print only the last N matched records",
	)

	params.context = reg.Int(
		"context",
		0,
		"Print N additional records before and after matches",
	)

	params.metadata = reg.StringP(
		"metadata",
		"m",
		"rnum",
		"Add metadata fields. Format: name[:property-name]. \nExamples:\n"+
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

	input := lo.Map(params.fileNames, func(fileName string, _ int) fileDescr {
		var reader io.Reader
		var close func() error
		var err error
		if fileName == "-" {
			reader = cmd.InOrStdin()
			fileName = "stdin"
			close = func() error { return nil }
		} else {
			close, reader, err = steps.OpenFile(fileName)
			if err != nil {
				return fileDescr{err: err}
			}
		}

		return fileDescr{
			fileName: fileName,
			r:        reader,
			close:    close}
	})

	for _, fd := range input {
		if fd.err != nil {
			return fd.err
		}
	}

	defer func() {
		for _, fd := range input {
			fd.close()
		}
	}()

	return runPipeline(params, input, cmd.OutOrStdout())
}

func runPipeline(params *filterParams, input []fileDescr, w io.Writer) error {
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

	var formatJSONToText pipeline.Step[steps.JSON, string]

	if len(params.headProps()) > 0 || params.outputFormat() == "text" {
		formatJSONToText, err = steps.JsonToText(
			opts,
			params.headProps(),
			params.orderProps(),
			params.textNoNewLine(),
			params.textNoProp(),
			params.textDelim(),
			slices.Concat(params.include(), params.highlights()),
			params.propertiesConfig)
		if err != nil {
			return err
		}
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
		steps.Hide(opts, params.hideProps()),
		steps.Select(opts, params.selectProps()),
		steps.Context(opts, params.context(), params.context()),
	)

	postProcessJSON := pipeline.Combine(
		steps.DistinctBy(opts, params.distinctBy()),
		steps.First(opts, params.first()),
		steps.Last(opts, params.last()),
	)

	multiJsons := lo.Map(input, func(f fileDescr, _ int) pipeline.Seq[steps.JSON] {
		return processJSON(
			steps.StrToJson(opts, params.durationMs())(
				processStringInput(steps.ReadByLines(f.fileName, f.r))))
	})

	mergedJsons := steps.Merge(opts, params.mergeBy(), multiJsons)

	return steps.WriteLines(
		w,
		params.showErrors(),
		formatJSONToText(
			postProcessJSON(mergedJsons)))
}
