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

// ShellType represents the type of shell for syntax highlighting
type ShellType int

const (
	ShellBash ShellType = iota
	ShellPowerShell
)

// Shell keywords and built-ins for bash/zsh
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

// PowerShell keywords and cmdlets
var powershellKeywords = map[string]bool{
	"if": true, "else": true, "elseif": true, "switch": true, "for": true, "foreach": true,
	"while": true, "do": true, "until": true, "break": true, "continue": true, "return": true,
	"function": true, "filter": true, "param": true, "begin": true, "process": true, "end": true,
	"try": true, "catch": true, "finally": true, "throw": true, "trap": true,
	"class": true, "enum": true, "using": true, "namespace": true,
	"in": true, "exit": true, "where": true, "select": true,
}

var powershellCmdlets = map[string]bool{
	"Get-ChildItem": true, "Get-Content": true, "Set-Content": true, "Get-Item": true,
	"Get-Location": true, "Set-Location": true, "Get-Process": true, "Stop-Process": true,
	"Start-Process": true, "Get-Service": true, "Start-Service": true, "Stop-Service": true,
	"Write-Host": true, "Write-Output": true, "Read-Host": true, "Clear-Host": true,
	"Get-Command": true, "Get-Help": true, "Get-Member": true, "Measure-Object": true,
	"Select-Object": true, "Where-Object": true, "ForEach-Object": true, "Sort-Object": true,
	"Group-Object": true, "Export-Csv": true, "Import-Csv": true, "ConvertTo-Json": true,
	"ConvertFrom-Json": true, "Invoke-RestMethod": true, "Invoke-WebRequest": true,
	"New-Item": true, "Remove-Item": true, "Copy-Item": true, "Move-Item": true,
	"Test-Path": true, "Join-Path": true, "Split-Path": true,
}

// Lexer performs syntax highlighting for shell commands
type Lexer struct {
	input     string
	pos       int
	tokens    []Token
	shellType ShellType
}

// NewLexer creates a new lexer for the given input with default bash shell
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:     input,
		pos:       0,
		tokens:    []Token{},
		shellType: ShellBash,
	}
}

// NewLexerWithShell creates a new lexer for the given input and shell type
func NewLexerWithShell(input string, shellType ShellType) *Lexer {
	return &Lexer{
		input:     input,
		pos:       0,
		tokens:    []Token{},
		shellType: shellType,
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
			
			// Determine token type based on shell
			if isFirstWord {
				if l.isKeyword(word) {
					l.tokens[len(l.tokens)-1].Type = TokenKeyword
				} else if l.shellType == ShellPowerShell && l.isCmdlet(word) {
					l.tokens[len(l.tokens)-1].Type = TokenCommand
				} else {
					l.tokens[len(l.tokens)-1].Type = TokenCommand
				}
				isFirstWord = false
			} else if l.isKeyword(word) {
				l.tokens[len(l.tokens)-1].Type = TokenKeyword
			} else if l.shellType == ShellPowerShell && l.isCmdlet(word) {
				l.tokens[len(l.tokens)-1].Type = TokenCommand
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

// isKeyword checks if a word is a keyword based on shell type
func (l *Lexer) isKeyword(word string) bool {
	wordLower := strings.ToLower(word)
	if l.shellType == ShellPowerShell {
		return powershellKeywords[wordLower]
	}
	return shellKeywords[word]
}

// isCmdlet checks if a word is a PowerShell cmdlet
func (l *Lexer) isCmdlet(word string) bool {
	if l.shellType != ShellPowerShell {
		return false
	}
	return powershellCmdlets[word]
}

// isPath checks if a word looks like a file path
func (l *Lexer) isPath(word string) bool {
	// Unix paths
	if strings.HasPrefix(word, "/") || 
	   strings.HasPrefix(word, "./") || 
	   strings.HasPrefix(word, "../") ||
	   strings.HasPrefix(word, "~/") {
		return true
	}
	// Windows paths (PowerShell)
	if l.shellType == ShellPowerShell && len(word) >= 3 {
		if (word[1] == ':' && (word[2] == '\\' || word[2] == '/')) {
			return true
		}
	}
	return false
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
