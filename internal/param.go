package internal

import (
	"fmt"
	"reflect"
	"strings"
)

type ParamType int

const (
	NoParam ParamType = iota
	RequiredParam
	OptionalParam
	FixedLengthParam
	ArbitraryLengthParam
)

func ParamTypes(f reflect.Type) []ParamType {
	var types []ParamType
	for i := 0; i < f.NumIn(); i++ {
		switch f.In(i).Kind() {
		case reflect.Interface:
			// Only context.Context for now, which counts as NoParam for CLI.
			types = append(types, NoParam)
		case reflect.String:
			types = append(types, RequiredParam)
		case reflect.Pointer:
			types = append(types, OptionalParam)
		case reflect.Array:
			types = append(types, FixedLengthParam)
		case reflect.Slice:
			types = append(types, ArbitraryLengthParam)
		}
	}
	return types
}

func ParamsUsage(names []string, types []ParamType) string {
	var usage strings.Builder
	for i, name := range names {
		name = NormalizeToKebabCase(name)
		switch types[i] {
		case RequiredParam:
			usage.WriteString(fmt.Sprintf(" <%v>", name))
		case OptionalParam:
			usage.WriteString(fmt.Sprintf(" [%v]", name))
		case FixedLengthParam:
			usage.WriteString(fmt.Sprintf(" <%v...>", name))
		case ArbitraryLengthParam:
			usage.WriteString(fmt.Sprintf(" [%v...]", name))
		}
	}
	return usage.String()
}
