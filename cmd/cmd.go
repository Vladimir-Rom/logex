package commands

import (
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/spf13/cobra"
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
	fileName      string
	kqlFilter     string
	jq            string
	include       []string
	exclude       []string
	includeRegexp []string
	excludeRegexp []string
	selectProps   []string
	durationMs    []string
	showErrors    bool
	textFormat    []string
	textNoNewLine bool
	textDelim     string
	textNoProp    bool
	distinctBy    string
	highlights    []string
	first         int
	last          int
	context       int
	metadata      string
}

func createRootCmd() *cobra.Command {
	var params filterParams
	filterCmd := &cobra.Command{
		Use:   "logex [flags] file-name",
		Short: "logex is a tool for filtering and formatting structured log files",
		Run: func(cmd *cobra.Command, args []string) {
			params.fileName = args[0]
			err := doFilter(&params)
			if err != nil {
				log.Fatal(err)
			}
		},
		Args: cobra.ExactArgs(1),
	}

	filterCmd.Flags().StringVarP(
		&params.kqlFilter,
		"kql",
		"f",
		"",
		"filter in the Kibana Query Language format. Example: 'level:(error OR warn)'")

	filterCmd.Flags().StringVar(
		&params.jq,
		"jq",
		"",
		"jq expression for filtering or transformation. Example: '.level==\"info\" or .level==\"warn\"'")

	filterCmd.Flags().StringSliceVarP(
		&params.include,
		"include",
		"i",
		nil,
		"include only records with any of specified substrings")

	filterCmd.Flags().StringSliceVarP(
		&params.exclude,
		"exclude",
		"e",
		nil,
		"exclude records with any of specified substrings")

	filterCmd.Flags().StringSliceVar(
		&params.includeRegexp,
		"include-regexp",
		nil,
		"include only records which matched with any of specified regular expressions")

	filterCmd.Flags().StringSliceVar(
		&params.excludeRegexp,
		"exclude-regexp",
		nil,
		"exclude records which matched with any of specified regular expressions")

	filterCmd.Flags().StringSliceVar(
		&params.durationMs,
		"duration-ms",
		nil,
		"treat specified fields as duration string and convert it to milliseconds")

	filterCmd.Flags().StringSliceVar(
		&params.selectProps,
		"select",
		nil,
		"property names to output")

	filterCmd.Flags().BoolVar(
		&params.showErrors,
		"show-errors",
		false,
		"show processing errors")

	filterCmd.Flags().StringSliceVarP(
		&params.textFormat,
		"txt-format",
		"t",
		nil,
		"property names which will be printed first in the plain text format")

	filterCmd.Flags().BoolVar(
		&params.textNoNewLine,
		"txt-nonl",
		false,
		"do not add new lines after each record")

	filterCmd.Flags().BoolVar(
		&params.textNoProp,
		"txt-noprop",
		false,
		"do not print properties except these selected in the format string (txt-format)")

	filterCmd.Flags().StringVar(
		&params.textDelim,
		"txt-delim",
		"|",
		"delimiter between text properties")

	filterCmd.Flags().StringVar(
		&params.distinctBy,
		"distinct-by",
		"",
		"returns distinct records according to the specified property name")

	filterCmd.Flags().StringSliceVarP(
		&params.highlights,
		"highlight",
		"l",
		nil,
		"highlight substrings in output")

	filterCmd.Flags().IntVar(
		&params.first,
		"first",
		0,
		"print only first N matched records",
	)

	filterCmd.Flags().IntVar(
		&params.last,
		"last",
		0,
		"print only last N matched records",
	)

	filterCmd.Flags().IntVar(
		&params.context,
		"context",
		0,
		"print N additional records before and after matched",
	)

	filterCmd.Flags().StringVarP(
		&params.metadata,
		"metadata",
		"m",
		"rnum",
		"add metadata fields. Format: name[:property-name]. \nExamples:\n"+
			"'rnum' - adds rnum field with record number\n"+
			"'rnum:r1 file:f1' - adds field r1 record number and f1 with name of logfile. ",
	)

	return filterCmd
}

func doFilter(params *filterParams) error {
	var reader io.Reader
	var fileName string
	if params.fileName == "-" {
		reader = os.Stdin
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

	return runPipeline(params, fileName, reader, os.Stdout)
}

func runPipeline(params *filterParams, filename string, r io.Reader, w io.Writer) error {
	opts := pipeline.PipelineOptions{
		ContextEnabled: params.context > 0,
	}
	filterByKQL, err := steps.FilterByKQL(opts, params.kqlFilter)
	if err != nil {
		return err
	}

	filterByJq, err := steps.FilterByJq(opts, params.jq)
	if err != nil {
		return err
	}

	addMeta, err := steps.AddMeta(opts, params.metadata)
	if err != nil {
		return err
	}

	input := steps.ReadByLines(filename, r)

	var formatJSONToText pipeline.Step[steps.JSON, string]

	if len(params.textFormat) > 0 {
		formatJSONToText = steps.JsonToText(
			opts,
			params.textFormat,
			params.textNoNewLine,
			params.textNoProp,
			params.textDelim,
			slices.Concat(params.include, params.highlights))
	} else {
		formatJSONToText = steps.JsonToStr(opts)
	}

	includeRegexp, err := steps.IncludeRegexp(opts, params.includeRegexp)
	if err != nil {
		return err
	}
	excludeRegexp, err := steps.ExcludeRegexp(opts, params.excludeRegexp)
	if err != nil {
		return err
	}

	processStringInput := pipeline.Combine(
		steps.RemovePrefix(opts),
		steps.ExcludeSubstringsAny(opts, params.exclude),
		steps.IncludeSubstringsAny(opts, params.include),
		includeRegexp,
		excludeRegexp,
	)

	processJSON := pipeline.Combine(
		addMeta,
		filterByKQL,
		filterByJq,
		steps.DistinctBy(opts, params.distinctBy),
		steps.Select(opts, params.selectProps),
		steps.Context(opts, params.context, params.context),
		steps.First(opts, params.first),
		steps.Last(opts, params.last),
	)

	return steps.WriteLines(
		w,
		params.showErrors,
		formatJSONToText(
			processJSON(
				steps.StrToJson(opts, params.durationMs)(
					processStringInput(input)))))
}
