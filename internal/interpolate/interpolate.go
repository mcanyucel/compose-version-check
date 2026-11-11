// Package interpolate expands Compose-style variables such as
// ${TAG:-latest} in strings.
package interpolate

import (
	"fmt"
	"strings"
)

// LookupFunc returns the value of a variable and whether it is set.
type LookupFunc func(name string) (string, bool)

// UnsetError is returned when a variable has no value and no default.
type UnsetError struct {
	Name    string
	Message string // from ${VAR:?message}; empty otherwise
}

func (e *UnsetError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("variable %s is not set: %s", e.Name, e.Message)
	}
	return fmt.Sprintf("variable %s is not set", e.Name)
}

// Expand replaces $VAR, ${VAR} and ${VAR<op>word} in s, using lookup to find
// values. Operators: :- - :? ? :+ +. "$$" is a literal "$".
func Expand(s string, lookup LookupFunc) (string, error) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '$' || i+1 == len(s) {
			b.WriteByte(s[i])
			continue
		}
		next := s[i+1]
		switch {
		case next == '$':
			b.WriteByte('$')
			i++
		case next == '{':
			end := matchingBrace(s, i+1)
			if end < 0 {
				return "", fmt.Errorf("unclosed ${ in %q", s)
			}
			value, err := expandBraced(s[i+2:end], lookup)
			if err != nil {
				return "", err
			}
			b.WriteString(value)
			i = end
		case isNameStart(next):
			j := i + 1
			for j < len(s) && isNameChar(s[j]) {
				j++
			}
			name := s[i+1 : j]
			value, ok := lookup(name)
			if !ok {
				return "", &UnsetError{Name: name}
			}
			b.WriteString(value)
			i = j - 1
		default:
			b.WriteByte('$')
		}
	}
	return b.String(), nil
}

// expandBraced handles the inside of ${...}, for example "TAG:-latest".
func expandBraced(expr string, lookup LookupFunc) (string, error) {
	n := 0
	for n < len(expr) && isNameChar(expr[n]) {
		n++
	}
	name, rest := expr[:n], expr[n:]
	if name == "" || !isNameStart(name[0]) {
		return "", fmt.Errorf("invalid variable name in ${%s}", expr)
	}

	op, word := "", ""
	for _, candidate := range []string{":-", ":?", ":+", "-", "?", "+"} {
		if strings.HasPrefix(rest, candidate) {
			op, word = candidate, rest[len(candidate):]
			break
		}
	}
	if op == "" && rest != "" {
		return "", fmt.Errorf("invalid expression ${%s}", expr)
	}

	value, set := lookup(name)
	// With a colon (":-", ":?", ":+"), an empty value counts as unset.
	if strings.HasPrefix(op, ":") && value == "" {
		set = false
	}

	switch strings.TrimPrefix(op, ":") {
	case "-":
		if set {
			return value, nil
		}
		return Expand(word, lookup)
	case "?":
		if set {
			return value, nil
		}
		return "", &UnsetError{Name: name, Message: word}
	case "+":
		if set {
			return Expand(word, lookup)
		}
		return "", nil
	default:
		if !set {
			return "", &UnsetError{Name: name}
		}
		return value, nil
	}
}

// matchingBrace returns the index of the "}" that closes the "{" at open,
// or -1 if there is none. Nested braces are skipped.
func matchingBrace(s string, open int) int {
	depth := 0
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func isNameStart(c byte) bool {
	return c == '_' || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isNameChar(c byte) bool {
	return isNameStart(c) || ('0' <= c && c <= '9')
}
