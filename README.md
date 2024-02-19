# logex
```
logex is a tool for filtering and formatting structured log files

Usage:
  logex [flags] file-name

Flags:
      --distinct-by string    returns distinct records according to a specified property name
  -e, --exclude stringArray   exclude all records with specified substrings
  -f, --filter-kql string     filter in the Kibana Query Language format. Example: 'level:(error OR warn)'
      --first int             print only first N matched records
  -h, --help                  help for logex
  -l, --highlight strings     highlight substrings in output
  -i, --include stringArray   include only records with specified substrings
      --last int              print only last N matched records
      --select strings        property names to output
      --show-errors           show processing errors
      --txt-delim string      delimiter between text properties (default "|")
  -t, --txt-format strings    property names which will be printed first in plain text format
      --txt-nonl              do not add new lines after each record
      --txt-noprop            do not print properties except these selected in format string (txt-format)
```