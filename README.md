# Logex - Make Logs Readable Again

Logex is a tool designed to enhance the readability of log files by filtering and formatting structured data.

Structured logs consist of JSON records primarily intended for automatic processing by systems like Elasticsearch and Kibana (ELK). However, there are instances where human has to read such files, for example when dealing with JSON logs from Docker containers or log files associated with an issue in a bug tracker. In such cases, manually parsing through these logs can be challenging. Logex serves as a solution to restore the readability of logs to a more human-friendly format like it was in the pre-ELK era. Make logs readable again!

### Main Features of Logex
1. **JSON Formatting:** Converts JSON data into plain text for easier comprehension.
2. **Filtering Options:** Supports filtering JSON records using the Kibana Query Language, JQ queries, plain text filters, or regular expressions.
3. **Output Colorization:** Enhances log visualization through color highlighting for better distinction.
4. **Multiple log files merging:** Merge multiple log files into single stream of records by specified fields (usually by timestamp)
5. and more

## Command line help
```
logex is a tool for filtering and formatting structured log files

Usage:
  logex [flags] file-name

Flags:
      --config string            configuration file name
      --context int              Print N additional records before and after matches
      --distinct-by string       Return distinct records based on the specified property names
      --duration-ms strings      Treat specified fields as duration strings and convert them to milliseconds (useful for filtering)
  -e, --exclude strings          Exclude records containing any of the specified substrings
      --exclude-regexp strings   Exclude records that match any of the specified regular expressions
      --expand strings           Parse property names with string values as JSON objects for use in filters and other operations
      --first int                Print only the first N matched records
      --format string            Output format, can be "text" or "json" (default "text")
  -h, --help                     help for logex
      --hide strings             Property names to hide
  -l, --highlight strings        Highlight substrings in the output
  -i, --include strings          Include only records containing any of the specified substrings
      --include-regexp strings   Include only records that match any of the specified regular expressions
      --jq string                Specify a jq expression for filtering or transformation. Example: '.level=="info" or .level=="warn"'
  -f, --kql string               Filter in the Kibana Query Language format. Example: 'level:(error OR warn)'
      --last int                 Print only the last N matched records
      --merge strings            Merge multiple files into single stream of records by specified fields (usually by timestamp) (default [ts])
  -m, --metadata string          Add metadata fields. Format: name[:property-name].
                                 Examples:
                                 'rnum' - adds an rnum field with the record number
                                 'rnum:r1 file:f1' - adds field r1 with the record number and f1 with the name of the logfile (default "rnum")
      --order strings            Specify property names to be displayed at the beginning of the record. Other properties will follow.
                                 Applicable for text format.
      --select strings           Property names to output, other properties will be skipped
      --show-errors              Show processing errors
      --txt-delim string         Delimiter between text properties (default "|")
  -t, --txt-head strings         Specify property names whose values will be displayed at the beginning of the record without
                                 printing property names. Other properties will follow. Applicable for text format.
      --txt-nonl                 Do not add new lines after each record.
                                 Applicable for text format.
      --txt-noprop               Exclude printing properties except those explicitly selected in --txt-head or --order.
                                 Applicable for text format.
```

### Configuration

Configuration example:
```yaml

```