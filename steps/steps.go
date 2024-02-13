package steps

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charlievieth/strcase"
	"github.com/vladimir-rom/gokql"
	"github.com/vladimir-rom/logex/pipeline"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func Noop[V any]() pipeline.Step[V, V] {
	return func(in pipeline.Seq[V]) pipeline.Seq[V] {
		return in
	}
}

func RemovePrefix() pipeline.Step[string, string] {
	return pipeline.NewStep[string, string](func(line pipeline.Item[string], yield pipeline.Yield[string]) bool {
		if ind := strings.Index(line.Value, "{"); ind > 0 {
			return yield(line.WithValue(line.Value[ind:]), nil)
		} else {
			return yield(line, nil)
		}
	})
}

func ExcludeSubstrings(substrings []string) pipeline.Step[string, string] {
	if len(substrings) == 0 {
		return Noop[string]()
	}

	return pipeline.NewStep[string, string](func(line pipeline.Item[string], yield pipeline.Yield[string]) bool {
		for _, s := range substrings {
			if strcase.Contains(line.Value, s) {
				return true
			}
		}
		return yield(line, nil)
	})
}

func IncludeSubstrings(substrings []string) pipeline.Step[string, string] {
	if len(substrings) == 0 {
		return Noop[string]()
	}

	return pipeline.NewStep[string, string](func(line pipeline.Item[string], yield pipeline.Yield[string]) bool {
		for _, s := range substrings {
			if !strcase.Contains(line.Value, s) {
				return true
			}
		}
		return yield(line, nil)
	})
}

func StrToJson() pipeline.Step[string, JSON] {
	return pipeline.NewStep[string, JSON](func(line pipeline.Item[string], yield pipeline.Yield[JSON]) bool {
		var res JSON
		err := json.Unmarshal([]byte(line.Value), &res)
		if err != nil {
			res = make(JSON)
			res["raw"] = line
		}

		return yield(pipeline.ToItem[string, JSON](line, res), nil)
	})
}

func JsonToStr() pipeline.Step[JSON, string] {
	return pipeline.NewStep[JSON, string](func(obj pipeline.Item[JSON], yield pipeline.Yield[string]) bool {
		b, err := json.Marshal(obj.Value)
		return yield(pipeline.ToItem[JSON, string](obj, string(b)), err)
	})
}

func Select(properties []string) pipeline.Step[JSON, JSON] {
	if len(properties) == 0 {
		return Noop[JSON]()
	}

	return pipeline.NewStep[JSON, JSON](func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		result := make(JSON)
		for _, property := range properties {
			if value, ok := obj.Value[property]; ok {
				result[property] = value
			}
		}

		return yield(obj.WithValue(result), nil)
	})
}

func FilterByKQL(filter string) (pipeline.Step[JSON, JSON], error) {
	if len(filter) == 0 {
		return Noop[JSON](), nil
	}

	expression, err := gokql.Parse(filter)
	if err != nil {
		return nil, fmt.Errorf("filter parsing error: %w", err)
	}

	return pipeline.NewStep[JSON, JSON](func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		matched, err := expression.Match(gokql.MapEvaluator{Map: obj.Value})
		if err != nil {
			return yield(obj.WithValue(nil), err)
		} else if matched {
			return yield(obj, nil)
		}

		return true
	}), nil
}

func DistinctBy(property string) pipeline.Step[JSON, JSON] {
	if len(property) == 0 {
		return Noop[JSON]()
	}

	processed := make(map[string]struct{})

	return pipeline.NewStep[JSON, JSON](func(obj pipeline.Item[JSON], yield pipeline.Yield[JSON]) bool {
		if v, ok := obj.Value[property]; ok {
			key := fmt.Sprintf("%v", v)
			if _, ok := processed[key]; ok {
				return true
			}
			processed[key] = struct{}{}
			return yield(obj, nil)
		}

		return true
	})
}

func OpenFile(fileName string) (close func() error, reader io.Reader, err error) {
	raw, err := os.Open(fileName)
	if err != nil {
		return nil, nil, err
	}

	return raw.Close, transform.NewReader(raw, unicode.BOMOverride(unicode.UTF8.NewDecoder())), nil
}

func ReadByLines(r io.Reader) pipeline.Seq[string] {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	return func(yield pipeline.Yield[string]) {
		for scanner.Scan() {
			if !yield(pipeline.NewItem1(scanner.Text()), scanner.Err()) {
				return
			}
		}
	}
}

func WriteLines(w io.Writer, showErrors bool, lines pipeline.Seq[string]) error {
	for line, err := range lines {
		if err != nil {
			if showErrors {
				_, err := fmt.Fprintln(w, err)
				if err != nil {
					return err
				}
			}
		} else {
			_, err := fmt.Fprintln(w, line.Value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
