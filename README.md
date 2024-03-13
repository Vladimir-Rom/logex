# logex
```
logex is a tool for filtering and formatting structured log files

Usage:
  logex [flags] file-name

Flags:
      --context int              print N additional records before and after matched
      --distinct-by string       returns distinct records according to the specified property name
      --duration-ms strings      treat specified fields as duration string and convert it to milliseconds
  -e, --exclude strings          exclude records with any of specified substrings
      --exclude-regexp strings   exclude records which matched with any of specified regular expressions
      --expand strings           property names with string values to parse them as json objects. Can be used then in filters and other operations
      --first int                print only first N matched records
  -h, --help                     help for logex
  -l, --highlight strings        highlight substrings in output
  -i, --include strings          include only records with any of specified substrings
      --include-regexp strings   include only records which matched with any of specified regular expressions
      --jq string                jq expression for filtering or transformation. Example: '.level=="info" or .level=="warn"'
  -f, --kql string               filter in the Kibana Query Language format. Example: 'level:(error OR warn)'
      --last int                 print only last N matched records
  -m, --metadata string          add metadata fields. Format: name[:property-name].
                                 Examples:
                                 'rnum' - adds rnum field with record number
                                 'rnum:r1 file:f1' - adds field r1 record number and f1 with name of logfile.  (default "rnum")
      --select strings           property names to output
      --show-errors              show processing errors
      --txt-delim string         delimiter between text properties (default "|")
  -t, --txt-format strings       property names which will be printed first in the plain text format
      --txt-nonl                 do not add new lines after each record
      --txt-noprop               do not print properties except these selected in the format string (txt-format)
```