package features

import "strings"

func keywordOffsetBefore(source string, offset int, keyword string) int {
	end := min(len(source), offset+len(keyword))
	for end > 0 {
		index := strings.LastIndex(source[:end], keyword)
		if index < 0 {
			return -1
		}
		beforeOK := index == 0 || !isIdentifierByte(source[index-1])
		after := index + len(keyword)
		afterOK := after == len(source) || !isIdentifierByte(source[after])
		if beforeOK && afterOK {
			return optionalPubOffsetBefore(source, index)
		}
		end = index
	}
	return -1
}

func optionalPubOffsetBefore(source string, offset int) int {
	end := offset
	for end > 0 && (source[end-1] == ' ' || source[end-1] == '\t') {
		end--
	}
	pub := "pub"
	if end >= 3 && source[end-3:end] == "ext" {
		pub = "ext"
	}
	if end >= 3 && source[end-3:end] == "api" {
		pub = "api"
	}
	start := end - len(pub)
	if start >= 0 && source[start:end] == pub &&
		(start == 0 || !isIdentifierByte(source[start-1])) {
		return start
	}
	return offset
}

func keywordOffsetAfterDecorators(source string, offset int, keyword string) int {
	for offset < len(source) {
		switch {
		case source[offset] == ' ' || source[offset] == '\t' || source[offset] == '\r' || source[offset] == '\n':
			offset++
		case strings.HasPrefix(source[offset:], "//"):
			end := strings.IndexByte(source[offset:], '\n')
			if end < 0 {
				return -1
			}
			offset += end
		case strings.HasPrefix(source[offset:], "/*"):
			end := strings.Index(source[offset+2:], "*/")
			if end < 0 {
				return -1
			}
			offset += end + 4
		case source[offset] == '@':
			next, ok := skipDecorator(source, offset)
			if !ok {
				return -1
			}
			offset = next
		default:
			start := offset
			for offset < len(source) && isIdentifierByte(source[offset]) {
				offset++
			}
			if offset == start {
				return -1
			}
			if source[start:offset] == keyword {
				return start
			}
			return -1
		}
	}
	return -1
}

func isIdentifierByte(value byte) bool {
	return value == '_' ||
		value >= 'a' && value <= 'z' ||
		value >= 'A' && value <= 'Z' ||
		value >= '0' && value <= '9'
}

func decoratorGapOnly(gap string) bool {
	for offset := 0; offset < len(gap); {
		switch {
		case gap[offset] == ' ' || gap[offset] == '\t' || gap[offset] == '\r' || gap[offset] == '\n':
			offset++
		case strings.HasPrefix(gap[offset:], "//"):
			end := strings.IndexByte(gap[offset:], '\n')
			if end < 0 {
				return true
			}
			offset += end
		case strings.HasPrefix(gap[offset:], "/*"):
			end := strings.Index(gap[offset+2:], "*/")
			if end < 0 {
				return false
			}
			offset += end + 4
		case gap[offset] == '@':
			next, ok := skipDecorator(gap, offset)
			if !ok {
				return false
			}
			offset = next
		default:
			return false
		}
	}
	return true
}

func skipDecorator(source string, offset int) (int, bool) {
	offset++
	start := offset
	for offset < len(source) {
		r := source[offset]
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
			break
		}
		offset++
	}
	if offset == start {
		return offset, false
	}
	for offset < len(source) && (source[offset] == ' ' || source[offset] == '\t') {
		offset++
	}
	if offset >= len(source) || source[offset] != '(' {
		return offset, true
	}

	depth := 0
	for offset < len(source) {
		switch {
		case strings.HasPrefix(source[offset:], `"""`):
			end := strings.Index(source[offset+3:], `"""`)
			if end < 0 {
				return offset, false
			}
			offset += end + 6
		case source[offset] == '"':
			offset++
			for offset < len(source) {
				if source[offset] == '\\' {
					offset += min(2, len(source)-offset)
					continue
				}
				offset++
				if source[offset-1] == '"' {
					break
				}
			}
		case source[offset] == '(':
			depth++
			offset++
		case source[offset] == ')':
			depth--
			offset++
			if depth == 0 {
				return offset, true
			}
		default:
			offset++
		}
	}
	return offset, false
}
