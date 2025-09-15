package bind

import (
	"reflect"
	"strconv"
	"strings"
)

const (
	TagOptions = "options"

	RequiredTag = "required"
	LevelTag    = "level"

	DefaultLevel = 1
)

// FieldBinding represents one destination struct field with zero or more tag candidates.
type FieldBinding struct {
	FieldIndex int
	Candidates []TagCandidate // evaluated in order, first match wins
	Options    TagCandidate   // options from `options` tag, if any
}

func (f *FieldBinding) Level() int {
	if f != nil && f.Options.MinimumLevel > 0 {
		return f.Options.MinimumLevel
	}
	return DefaultLevel
}

// TagCandidate is (kind, value) for a field tag, e.g. (json, "id") or (path, "org_id").
type TagCandidate struct {
	Kind    string
	Value   string
	Options []string

	// Struct tag options that are universal to all.
	MinimumLevel int
	Required     bool
}

// collectFieldBindings collects all exported fields from the struct type rt,
// along with their tag candidates.
func collectFieldBindings(rt reflect.Type) []FieldBinding {
	var out []FieldBinding

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)

		// skip unexported
		if sf.PkgPath != "" {
			continue
		}
		cands, optsCandidate := collectCandidatesForField(sf)

		out = append(out, FieldBinding{
			FieldIndex: i,
			Candidates: cands,
			Options:    optsCandidate,
		})
	}
	return out
}

// collectCandidatesForField collects all tag candidates for a given struct field.
func collectCandidatesForField(sf reflect.StructField) ([]TagCandidate, TagCandidate) {
	var cands []TagCandidate
	var optsCandidate TagCandidate

	for tag, value := range getTags(string(sf.Tag)) {
		if strings.HasPrefix(value, "-") {
			continue
		}
		value, options := parseTag(value)
		candidate := TagCandidate{
			Kind:         tag,
			Value:        value,
			MinimumLevel: 1,
		}
		candidate.parseOptions(options)

		if tag == TagOptions {
			optsCandidate = candidate
			continue
		}

		cands = append(cands, candidate)
	}

	return cands, optsCandidate
}

// parseTag breaks up a struct tag value into its name and options.
// E.g. "id,omitempty,hello" would return ("id", ["omitempty", "hello"]).
func parseTag(tag string) (string, []string) {
	split := strings.Split(tag, ",")
	if len(split) > 1 {
		return split[0], split[1:]
	}
	return split[0], nil
}

// parseOptions parses options for a tag candidate.
func (t *TagCandidate) parseOptions(options []string) {
	if t.Kind != TagOptions {
		t.Options = options
		return
	}
	// If this is an options tag, the 'value' is also an option.
	options = append(options, t.Value)
	t.Value = ""

	for _, opt := range options {
		if opt == RequiredTag {
			t.Required = true
			continue
		} else if strings.HasPrefix(opt, LevelTag+"=") {
			levelStr := strings.TrimPrefix(opt, LevelTag+"=")
			level, err := strconv.Atoi(levelStr)
			if err == nil && level > 0 {
				t.MinimumLevel = level
			}
			continue
		}
		t.Options = append(t.Options, opt)
	}
}

// getTags parses through a tag, collecting all key:"value" pairs into a map.
// This code is taken and modified from reflect/StructTag.
func getTags(tag string) map[string]string {
	tags := make(map[string]string)

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			break
		}

		tags[name] = value
	}

	return tags
}
