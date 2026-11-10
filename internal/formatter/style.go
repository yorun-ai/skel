package formatter

import "strings"

// styleBoundaries adds layout rules which require lossless comments and tokens.
// Negative boundaries are tight line breaks (no blank line).
func styleBoundaries(tokens []_Token, boundaries map[int]int) {
	if boundaries == nil {
		return
	}
	// Decorators are indivisible lines, attached to the following declaration.
	for i, token := range tokens {
		if token.value != "@" || token.kind != "Punctuation" {
			continue
		}
		if boundaries[token.offset] == 0 {
			boundaries[token.offset] = 1
		}
		end := i + 1
		if end >= len(tokens) {
			continue
		}
		if end+1 < len(tokens) && tokens[end+1].value == "(" {
			depth := 0
			for j := end + 1; j < len(tokens); j++ {
				if tokens[j].value == "(" {
					depth++
				}
				if tokens[j].value == ")" {
					depth--
					if depth == 0 {
						end = j
						break
					}
				}
			}
		}
		for j := end + 1; j < len(tokens); j++ {
			if tokens[j].kind == "Newline" || isComment(tokens[j]) {
				continue
			}
			if boundaries[tokens[j].offset] < 2 {
				boundaries[tokens[j].offset] = -1
			}
			break
		}
	}
	// Keep each item in a multiline parenthesized list on its own line,
	// leaving trailing comments beside the preceding item.
	multiline := multilineParens(tokens)
	stack := []int{}
	for i, token := range tokens {
		if token.kind != "Punctuation" {
			continue
		}
		switch token.value {
		case "(":
			stack = append(stack, i)
		case ")":
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case ",":
			if len(stack) == 0 || !multiline[stack[len(stack)-1]] {
				continue
			}
			for j := i + 1; j < len(tokens); j++ {
				if tokens[j].kind == "Newline" || isComment(tokens[j]) {
					continue
				}
				boundaries[tokens[j].offset] = -1
				break
			}
		}
	}
	// Group leading standalone comments with their declaration, never steal a
	// trailing comment from the preceding member.
	for i, token := range tokens {
		if boundaries[token.offset] != 2 {
			continue
		}
		start := i
		for j := i - 1; j >= 0; j-- {
			if tokens[j].kind == "Newline" {
				continue
			}
			if !isComment(tokens[j]) {
				break
			}
			if j > 0 && tokens[j-1].kind != "Newline" {
				break
			}
			start = j
		}
		if start != i {
			boundaries[tokens[start].offset] = 2
			boundaries[token.offset] = -1
		}
		// The first member of a block does not need a separating blank line.
		for j := start - 1; j >= 0; j-- {
			if tokens[j].kind == "Newline" {
				continue
			}
			if tokens[j].value == "{" {
				boundaries[tokens[start].offset] = -1
			}
			break
		}
	}
	fileHeaderBoundaries(tokens, boundaries)
}

func isComment(token _Token) bool {
	return token.kind == "LineComment" || token.kind == "BlockComment"
}

// multilineParens propagates multiline layout to enclosing parentheses.
func multilineParens(tokens []_Token) map[int]bool {
	result := make(map[int]bool)
	stack := []int{}
	braceDepth := 0
	openDepth := make(map[int]int)
	for i, t := range tokens {
		switch t.kind {
		case "Newline":
			for _, open := range stack {
				// An object argument already supplies its own multiline block layout.
				if openDepth[open] == braceDepth {
					result[open] = true
				}
			}
		case "Punctuation":
			switch t.value {
			case "{":
				braceDepth++
			case "}":
				braceDepth--
			case "(":
				stack = append(stack, i)
				openDepth[i] = braceDepth
			case ")":
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			}
		}
	}
	for _, open := range stack {
		delete(result, open)
	}
	return result
}

type _Audience struct {
	blockStart int
	start      int
	end        int
	actors     []int
}

// Short, comment-free audience lists fit on one line. Existing long or
// commented lists retain a multiline layout with one actor per continuation.
func layoutAudiences(tokens []_Token, audiences []_Audience, boundaries map[int]int) ([]_Token, map[int]int) {
	continuations := make(map[int]int)
	remove := make(map[int]bool)
	for _, audience := range audiences {
		if audience.blockStart > 0 {
			for _, token := range tokens {
				if token.offset >= audience.blockStart && token.kind == "Punctuation" && token.value == "}" {
					audience.end = token.offset + 1
					break
				}
			}
		}
		width := 0
		hasComment := false
		multiline := false
		first, end := -1, -1
		for i, t := range tokens {
			if t.offset < audience.start {
				continue
			}
			if t.offset >= audience.end {
				end = i
				break
			}
			if first < 0 {
				first = i
			}
			if t.kind == "Newline" {
				multiline = true
				continue
			}
			width += len([]rune(t.value)) + 1
			hasComment = hasComment || isComment(t)
		}
		if first < 0 {
			continue
		}
		if end < 0 {
			end = len(tokens)
		}
		if !hasComment && width <= 100 {
			for i := first + 1; i < end; i++ {
				if tokens[i].kind == "Newline" {
					remove[i] = true
				}
			}
		} else if multiline || hasComment {
			if first+1 < end {
				continuations[tokens[first+1].offset] = 1
			}
			if end < len(tokens) {
				continuations[tokens[end].offset] = -1
			}
			for _, offset := range audience.actors[1:] {
				boundaries[offset] = -1
			}
		}
	}
	result := make([]_Token, 0, len(tokens))
	for i, t := range tokens {
		if !remove[i] {
			result = append(result, t)
		}
	}
	return result, continuations
}

// fileHeaderBoundaries separates an explicitly marked file header from the
// following content. Ordinary leading comments keep the author's grouping.
func fileHeaderBoundaries(tokens []_Token, boundaries map[int]int) {
	marked := false
	comments := false
	newlines := 0
	for _, token := range tokens {
		if token.kind == "Newline" {
			if comments {
				newlines += strings.Count(token.value, "\n")
			}
			continue
		}
		if comments && (newlines > 1 || !isComment(token)) {
			if marked {
				boundaries[token.offset] = 2
			}
			return
		}
		if !isComment(token) {
			return
		}
		comments = true
		newlines = 0
		marked = marked || isFileHeaderComment(token.value)
	}
}

func isFileHeaderComment(value string) bool {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimPrefix(line, "/*")
		line = strings.TrimPrefix(strings.TrimSpace(line), "*")
		line = strings.TrimSpace(strings.TrimSuffix(line, "*/"))
		if strings.HasPrefix(strings.ToLower(line), "copyright ") ||
			strings.HasPrefix(line, "SPDX-License-Identifier:") ||
			strings.HasPrefix(line, "Code generated ") && strings.HasSuffix(line, "DO NOT EDIT.") {
			return true
		}
	}
	return false
}
