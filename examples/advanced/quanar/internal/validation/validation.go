// Package validation checks a request against the rules an api declared, as data. Text from
// the path and query becomes numbers and booleans where a rule asks for them, so the quan
// receives them typed.
package validation

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"shop-quanar/internal/wire"
)

// Issue is a failed rule, as the quanar answers it with 422.
type Issue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

var email = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func number(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func coerce(rule wire.Rule, value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}
	switch rule.Type {
	case "integer", "number":
		if n, err := strconv.ParseFloat(strings.TrimSpace(text), 64); err == nil {
			return n
		}
	case "boolean":
		switch text {
		case "true":
			return true
		case "false":
			return false
		}
	}
	return value
}

func check(rule wire.Rule, value any, path string, issues *[]Issue) {
	fail := func(format string, args ...any) {
		*issues = append(*issues, Issue{path, fmt.Sprintf(format, args...)})
	}
	if value == nil || value == "" {
		if rule.Required {
			fail("is required")
		}
		return
	}
	switch rule.Type {
	case "string", "email", "enum":
		text, ok := value.(string)
		if !ok {
			fail("must be text")
			return
		}
		if rule.Type == "email" && !email.MatchString(text) {
			fail("must be an email address")
			return
		}
		if rule.Type == "enum" && !slices.Contains(rule.Values, text) {
			fail("must be one of %s", strings.Join(rule.Values, ", "))
			return
		}
		length := float64(len([]rune(text)))
		if rule.Min != nil && length < *rule.Min {
			fail("must be at least %s characters", number(*rule.Min))
		}
		if rule.Max != nil && length > *rule.Max {
			fail("must be at most %s characters", number(*rule.Max))
		}
		if rule.Pattern != "" {
			if re, err := regexp.Compile(rule.Pattern); err == nil && !re.MatchString(text) {
				fail("has the wrong format")
			}
		}
	case "integer", "number":
		n, ok := value.(float64)
		if !ok {
			fail("must be a number")
			return
		}
		if rule.Type == "integer" && n != math.Trunc(n) {
			fail("must be a whole number")
			return
		}
		if rule.Min != nil && n < *rule.Min {
			fail("must be at least %s", number(*rule.Min))
		}
		if rule.Max != nil && n > *rule.Max {
			fail("must be at most %s", number(*rule.Max))
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			fail("must be true or false")
		}
	case "array":
		list, ok := value.([]any)
		if !ok {
			fail("must be a list")
			return
		}
		plural := func(n float64) string {
			if n == 1 {
				return ""
			}
			return "s"
		}
		if rule.Min != nil && float64(len(list)) < *rule.Min {
			fail("must have at least %s item%s", number(*rule.Min), plural(*rule.Min))
		}
		if rule.Max != nil && float64(len(list)) > *rule.Max {
			fail("must have at most %s item%s", number(*rule.Max), plural(*rule.Max))
		}
		if rule.Items != nil {
			for i, item := range list {
				check(*rule.Items, item, fmt.Sprintf("%s[%d]", path, i), issues)
			}
		}
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			fail("must be an object")
			return
		}
		if rule.Fields != nil {
			checkFields(rule.Fields, object, path, issues)
		}
	}
}

func checkFields(fields map[string]wire.Rule, source map[string]any, prefix string, issues *[]Issue) {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		rule := fields[name]
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		if v, ok := source[name]; ok {
			source[name] = coerce(rule, v)
		}
		check(rule, source[name], path, issues)
	}
}

// Validate checks and coerces in place; it returns every issue found.
func Validate(rules *wire.Validation, params map[string]any, query map[string]any, body any) []Issue {
	issues := []Issue{}
	if rules.Params != nil {
		checkFields(rules.Params, params, "params", &issues)
	}
	if rules.Query != nil {
		checkFields(rules.Query, query, "query", &issues)
	}
	if rules.Body != nil {
		object, ok := body.(map[string]any)
		if !ok {
			issues = append(issues, Issue{"body", "must be a JSON object"})
		} else {
			checkFields(rules.Body, object, "body", &issues)
		}
	}
	return issues
}
