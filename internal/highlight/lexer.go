package highlight

import (
	"strings"
	"unicode"
)

// TokenType represents the type of a syntax token
type TokenType int

const (
	TokenCommand TokenType = iota
	TokenKeyword
	TokenFlag
	TokenString
	TokenVariable
	TokenOperator
	TokenComment
	TokenNumber
	TokenPath
	TokenDefault
)

// Token represents a syntax token with its type and value
type Token struct {
	Type  TokenType
	Value string
}

// Shell keywords and built-ins
var shellKeywords = map[string]bool{
	"if": true, "then": true, "else": true, "elif": true, "fi": true,
	"case": true, "esac": true, "for": true, "while": true, "until": true,
	"do": true, "done": true, "in": true, "function": true, "select": true,
	"time": true, "coproc": true, "return": true, "break": true, "continue": true,
	"export": true, "local": true, "readonly": true, "declare": true, "typeset": true,
	"alias": true, "unalias": true, "set": true, "unset": true, "shift": true,
	"source": true, "exec": true, "eval": true, "test": true, "cd": true,
	"echo": true, "printf": true, "read": true, "exit": true, "true": true, "false": true,
}

// Lexer performs syntax highlighting for shell commands
type Lexer struct {
	input string
	pos   int
	tokens []Token
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
		tokens: []Token{},
	}
}

// Tokenize performs lexical analysis on the input
func (l *Lexer) Tokenize() []Token {
	isFirstWord := true
	
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		
		// Skip whitespace
		if unicode.IsSpace(rune(ch)) {
			l.addToken(TokenDefault, string(ch))
			l.pos++
			continue
		}
		
		// Comments
		if ch == '#' {
			l.tokenizeComment()
			continue
		}
		
		// Strings
		if ch == '"' || ch == '\'' || ch == '`' {
			l.tokenizeString(ch)
			isFirstWord = false
			continue
		}
		
		// Variables
		if ch == '$' {
			l.tokenizeVariable()
			isFirstWord = false
			continue
		}
		
		// Operators and special characters
		if strings.ContainsRune("|&;<>()[]{}!*", rune(ch)) {
			l.tokenizeOperator()
			if ch == '|' || ch == '&' || ch == ';' {
				isFirstWord = true
			}
			continue
		}
		
		// Flags (start with -)
		if ch == '-' && l.pos+1 < len(l.input) && !unicode.IsSpace(rune(l.input[l.pos+1])) {
			l.tokenizeFlag()
			isFirstWord = false
			continue
		}
		
		// Words (commands, keywords, paths, etc.)
		if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '.' || ch == '/' || ch == '_' || ch == '~' {
			word := l.tokenizeWord()
			
			// Determine token type
			if isFirstWord {
				if shellKeywords[word] {
					l.tokens[len(l.tokens)-1].Type = TokenKeyword
				} else {
					l.tokens[len(l.tokens)-1].Type = TokenCommand
				}
				isFirstWord = false
			} else if shellKeywords[word] {
				l.tokens[len(l.tokens)-1].Type = TokenKeyword
			} else if l.isPath(word) {
				l.tokens[len(l.tokens)-1].Type = TokenPath
			} else if l.isNumber(word) {
				l.tokens[len(l.tokens)-1].Type = TokenNumber
			}
			
			continue
		}
		
		// Default
		l.addToken(TokenDefault, string(ch))
		l.pos++
	}
	
	return l.tokens
}

// tokenizeWord reads a word from the input
func (l *Lexer) tokenizeWord() string {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsSpace(rune(ch)) || strings.ContainsRune("|&;<>()[]{}\"'`$#", rune(ch)) {
			break
		}
		l.pos++
	}
	
	word := l.input[start:l.pos]
	l.addToken(TokenDefault, word)
	return word
}

// tokenizeString reads a quoted string
func (l *Lexer) tokenizeString(quote byte) {
	start := l.pos
	l.pos++ // Skip opening quote
	
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == quote {
			l.pos++ // Skip closing quote
			break
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.pos += 2 // Skip escaped character
			continue
		}
		l.pos++
	}
	
	l.addToken(TokenString, l.input[start:l.pos])
}

// tokenizeVariable reads a variable reference
func (l *Lexer) tokenizeVariable() {
	start := l.pos
	l.pos++ // Skip $
	
	if l.pos < len(l.input) && l.input[l.pos] == '{' {
		// ${VAR} syntax
		l.pos++
		for l.pos < len(l.input) && l.input[l.pos] != '}' {
			l.pos++
		}
		if l.pos < len(l.input) {
			l.pos++ // Skip }
		}
	} else {
		// $VAR syntax
		for l.pos < len(l.input) {
			ch := l.input[l.pos]
			if !unicode.IsLetter(rune(ch)) && !unicode.IsDigit(rune(ch)) && ch != '_' {
				break
			}
			l.pos++
		}
	}
	
	l.addToken(TokenVariable, l.input[start:l.pos])
}

// tokenizeFlag reads a command flag
func (l *Lexer) tokenizeFlag() {
	start := l.pos
	l.pos++ // Skip -
	
	// Handle -- or single -
	if l.pos < len(l.input) && l.input[l.pos] == '-' {
		l.pos++
	}
	
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsSpace(rune(ch)) || strings.ContainsRune("|&;<>()[]{}\"'`$", rune(ch)) {
			break
		}
		l.pos++
	}
	
	l.addToken(TokenFlag, l.input[start:l.pos])
}

// tokenizeOperator reads an operator or special character
func (l *Lexer) tokenizeOperator() {
	start := l.pos
	ch := l.input[l.pos]
	l.pos++
	
	// Handle multi-character operators
	if l.pos < len(l.input) {
		next := l.input[l.pos]
		if (ch == '|' && next == '|') || (ch == '&' && next == '&') ||
		   (ch == '>' && next == '>') || (ch == '<' && next == '<') ||
		   (ch == '>' && next == '&') || (ch == '<' && next == '&') {
			l.pos++
		}
	}
	
	l.addToken(TokenOperator, l.input[start:l.pos])
}

// tokenizeComment reads a comment
func (l *Lexer) tokenizeComment() {
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.pos++
	}
	l.addToken(TokenComment, l.input[start:l.pos])
}

// isPath checks if a word looks like a file path
func (l *Lexer) isPath(word string) bool {
	return strings.HasPrefix(word, "/") || 
	       strings.HasPrefix(word, "./") || 
	       strings.HasPrefix(word, "../") ||
	       strings.HasPrefix(word, "~/")
}

// isNumber checks if a word is a number
func (l *Lexer) isNumber(word string) bool {
	for _, ch := range word {
		if !unicode.IsDigit(ch) {
			return false
		}
	}
	return len(word) > 0
}

// addToken adds a token to the token list
func (l *Lexer) addToken(tokenType TokenType, value string) {
	l.tokens = append(l.tokens, Token{
		Type:  tokenType,
		Value: value,
	})
}
