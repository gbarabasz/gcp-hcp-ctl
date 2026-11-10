package rolebinding

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// conditionPreview escapes non-printing code points before counting the display
// length, so escape sequences are never cut in half or interpreted by a terminal.
func conditionPreview(condition *string) string {
	if condition == nil || *condition == "" {
		return "<none>"
	}
	var tokens []string
	total := 0
	for _, r := range *condition {
		var token string
		switch r {
		case '\n':
			token = `\n`
		case '\r':
			token = `\r`
		case '\t':
			token = `\t`
		default:
			if unicode.IsPrint(r) {
				token = string(r)
			} else if r <= 0xffff {
				token = fmt.Sprintf(`\u%04x`, r)
			} else {
				token = fmt.Sprintf(`\U%08x`, r)
			}
		}
		tokens = append(tokens, token)
		total += utf8.RuneCountInString(token)
	}
	var b strings.Builder
	count := 0
	for _, token := range tokens {
		n := utf8.RuneCountInString(token)
		if count+n > 40 {
			break
		}
		b.WriteString(token)
		count += n
	}
	if total > 40 {
		b.WriteString("...")
	}
	return b.String()
}

// conditionArgument returns a shell literal. For terminal controls, Bash and
// Zsh ANSI-C quoting encodes their UTF-8 bytes as fixed-width octal escapes.
func conditionArgument(condition *string) string {
	value := ""
	if condition != nil {
		value = *condition
	}
	unsafe := false
	for _, r := range value {
		if r != '\n' && !unicode.IsPrint(r) {
			unsafe = true
			break
		}
	}
	if !unsafe {
		return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
	}
	var b strings.Builder
	b.WriteString("$'")
	for _, r := range value {
		switch r {
		case '\n':
			b.WriteString(`\n`)
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		default:
			if unicode.IsPrint(r) {
				b.WriteRune(r)
			} else {
				for _, c := range []byte(string(r)) {
					fmt.Fprintf(&b, `\%03o`, c)
				}
			}
		}
	}
	b.WriteByte('\'')
	return b.String()
}

const conditionHelp = `The condition is a Cedar expression body, not a complete policy. The API
validates it (maximum 4096 characters; Namespace:: references and trailing
policy statements are rejected). Missing attributes fail closed; use "has"
guards. An omitted or empty condition is unrestricted; whitespace-only is
also unrestricted. Expressions are sent unchanged. A condition limits only
this binding, not independent grants. Changes may take time to propagate.

Text details print a copyable --condition argument; control characters use
Bash/Zsh ANSI-C quoting. JSON/YAML retain the complete returned value. List
previews are escaped and truncated. Explicit empty may remain visible as
"condition": "" in JSON rather than an absent field. See the access management
section of README for context fields, regional examples, and lockout warnings.`
