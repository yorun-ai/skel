package binding

import "strings"

func ChooseString(condition bool, trueChosen string, falseChosen string) string {
	if condition {
		return trueChosen
	}
	return falseChosen
}

func EnsureSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if strings.HasSuffix(text, ".") ||
		strings.HasSuffix(text, "。") ||
		strings.HasSuffix(text, "!") ||
		strings.HasSuffix(text, "！") ||
		strings.HasSuffix(text, "?") ||
		strings.HasSuffix(text, "？") {
		return text
	}
	return text + "."
}

func SplitDocLines(description string) []string {
	description = strings.TrimSpace(description)
	if description == "" {
		return nil
	}

	rawLines := strings.Split(description, "\n")
	lines := make([]string, 0, len(rawLines))
	for _, rawLine := range rawLines {
		lines = append(lines, strings.TrimSpace(rawLine))
	}
	return lines
}

func MergeDescriptionAndExample(description string, example string) string {
	description = strings.TrimSpace(description)
	example = strings.TrimSpace(example)
	example = CompactDocValue(example)

	if example == "" {
		return description
	}
	if description == "" {
		return "e.g. " + example
	}
	return description + " (e.g. " + example + ")"
}

func CompactDocValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	var builder strings.Builder
	inString := false
	escaped := false

	writeSpace := func() {
		if builder.Len() == 0 {
			return
		}
		current := builder.String()
		if strings.HasSuffix(current, " ") {
			return
		}
		builder.WriteByte(' ')
	}

	for _, char := range value {
		if inString {
			builder.WriteRune(char)
			if escaped {
				escaped = false
				continue
			}
			if char == '\\' {
				escaped = true
				continue
			}
			if char == '"' {
				inString = false
			}
			continue
		}

		switch char {
		case '"':
			inString = true
			builder.WriteRune(char)
		case '{', '[':
			builder.WriteRune(char)
			writeSpace()
		case '}', ']':
			current := builder.String()
			builder.Reset()
			builder.WriteString(strings.TrimRight(current, " "))
			writeSpace()
			builder.WriteRune(char)
		case ',':
			current := builder.String()
			builder.Reset()
			builder.WriteString(strings.TrimRight(current, " "))
			builder.WriteRune(char)
			builder.WriteByte(' ')
		default:
			if char == ' ' || char == '\n' || char == '\r' || char == '\t' {
				writeSpace()
				continue
			}
			builder.WriteRune(char)
		}
	}

	return strings.TrimSpace(builder.String())
}
