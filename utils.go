package main

import (
	"path/filepath"
	"strings"
)

type Language int

const (
	C Language = iota + 1
	CPlusPlus
	Java
	Python
	CSharp
	Assembly
	Go
)

type Comment struct {
	Line                 string
	BlockStart, BlockEnd string
}

var extensionsLang = map[string]Language{
	".c":   C,
	".cpp": CPlusPlus, ".cxx": CPlusPlus, ".cc": CPlusPlus,
	".hpp": CPlusPlus, ".hxx": CPlusPlus, ".hh": CPlusPlus, ".h": CPlusPlus,
	".java": Java,
	".py":   Python,
	".cs":   CSharp,
	".asm":  Assembly, ".nasm": Assembly, ".s": Assembly,
	".go": Go,
}

var langComment = map[Language]Comment{
	C:         {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	CPlusPlus: {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	Java:      {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	Python:    {Line: "#"},
	CSharp:    {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
	Assembly:  {Line: ";"},
	Go:        {Line: "//", BlockStart: "/*", BlockEnd: "*/"},
}

func DetectLanguage(path string) (Language, bool) {
	lang, ok := extensionsLang[filepath.Ext(path)]
	return lang, ok
}

func ExtractFirstComment(src string, lang Language) (string, bool) {
	comment := langComment[lang]
	src = strings.TrimSpace(src)

	if after, ok := strings.CutPrefix(src, comment.Line); ok {
		content, _, _ := strings.Cut(after, "\n")
		return strings.TrimSpace(content), true
	}

	if comment.BlockStart != "" {
		if after, ok := strings.CutPrefix(src, comment.BlockStart); ok {
			src = strings.TrimSpace(after)
			content, _, ok := strings.Cut(src, comment.BlockEnd)
			if !ok {
				return "", false
			}

			return strings.TrimSpace(content), true
		}
	}

	return "", false
}
