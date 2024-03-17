# Logex - Make Logs Readable Again

Logex is a tool designed to enhance the readability of log files by filtering and formatting structured data.

Structured logs consist of JSON records primarily intended for processing by systems like Elasticsearch and Kibana (ELK). However, there are instances where human intervention is necessary, such as when dealing with JSON logs from Docker containers or multiple log files associated with an issue in a bug tracker. In such cases, manually parsing through these logs can be challenging. Logex serves as a solution to restore the readability of logs to a more human-friendly format reminiscent of pre-ELK era practices.

### Main Features of Logex
1. **JSON Formatting:** Converts JSON data into plain text for easier comprehension.
2. **Filtering Options:** Supports filtering JSON records using the Kibana Query Language, JQ queries, plain text filters, or regular expressions.
3. **Output Colorization:** Enhances log visualization through color highlighting for better distinction.
4. etc.

## Command line help
```
logex is a tool for filtering and formatting structured log files

Usage:
  logex [flags] file-name

Flags:
      --config string            configuration file name
      --context int              print N additional records before and after matches
      --distinct-by string       return distinct records based on the specified property names
      --duration-ms strings      treat specified fields as duration strings and convert them to milliseconds (useful for filtering)
  -e, --exclude strings          exclude records containing any of the specified substrings
      --exclude-regexp strings   exclude records that match any of the specified regular expressions
      --expand strings           parse property names with string values as JSON objects for use in filters and other operations
      --first int                print only the first N matched records
  -h, --help                     help for logex
      --hide strings             property names to hide
  -l, --highlight strings        highlight substrings in the output
  -i, --include strings          include only records containing any of the specified substrings
      --include-regexp strings   include only records that match any of the specified regular expressions
      --jq string                specify a jq expression for filtering or transformation. Example: '.level=="info" or .level=="warn"'
  -f, --kql string               Filter in the Kibana Query Language format. Example: 'level:(error OR warn)'
      --last int                 print only the last N matched records
  -m, --metadata string          add metadata fields. Format: name[:property-name].
                                 Examples:
                                 'rnum' - adds an rnum field with the record number
                                 'rnum:r1 file:f1' - adds field r1 with the record number and f1 with the name of the logfile (default "rnum")
      --select strings           property names to output
      --show-errors              show processing errors
      --txt-delim string         delimiter between text properties (default "|")
  -t, --txt-format strings       property names to be printed first in plain text format
      --txt-nonl                 do not add new lines after each record
      --txt-noprop               do not print properties except those selected in the format string (txt-format)
```