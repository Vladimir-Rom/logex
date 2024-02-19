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
	include       []string
	exclude       []string
	selectProps   []string
	showErrors    bool
	textFormat    []string
	textNoNewLine bool
	textDelim     string
	textNoProp    bool
	distinctBy    string
	highlights    []string
	first         int
	last          int
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
		"filter-kql",
		"f",
		"",
		"filter in the Kibana Query Language format. Example: 'level:(error OR warn)'")

	filterCmd.Flags().StringArrayVarP(
		&params.include,
		"include",
		"i",
		nil,
		"include only records with specified substrings")

	filterCmd.Flags().StringArrayVarP(
		&params.exclude,
		"exclude",
		"e",
		nil,
		"exclude all records with specified substrings")

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

	return filterCmd
}

func doFilter(params *filterParams) error {
	var reader io.Reader
	if params.fileName == "-" {
		reader = os.Stdin
	} else {
		close, r, err := steps.OpenFile(params.fileName)
		if err != nil {
			return err
		}
		reader = r
		defer close()
	}

	return runPipeline(params, reader, os.Stdout)
}

func runPipeline(params *filterParams, r io.Reader, w io.Writer) error {
	filterByKQL, err := steps.FilterByKQL(params.kqlFilter)
	if err != nil {
		return err
	}

	input := steps.ReadByLines(r)

	var formatter pipeline.Step[steps.JSON, string]

	if len(params.textFormat) > 0 {
		formatter = steps.JsonToText(
			params.textFormat,
			params.textNoNewLine,
			params.textNoProp,
			params.textDelim,
			slices.Concat(params.include, params.highlights))
	} else {
		formatter = steps.JsonToStr()
	}

	return steps.WriteLines(
		w,
		params.showErrors,
		formatter(
			steps.Last(params.last)(
				steps.First(params.first)(
					steps.Select(params.selectProps)(
						steps.DistinctBy(params.distinctBy)(
							filterByKQL(
								steps.StrToJson()(
									steps.IncludeSubstrings(params.include)(
										steps.ExcludeSubstrings(params.exclude)(
											steps.RemovePrefix()(input)))))))))))
}
