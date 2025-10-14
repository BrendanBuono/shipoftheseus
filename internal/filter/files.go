// Package filter provides utilities for filtering files and lines during analysis.
// It identifies files that should be skipped (binary, generated, vendor) and
// detects comment-only or blank lines that shouldn't be analyzed.
package filter

import (
	"path/filepath"
	"strings"
)

// skipDirectories lists directory names that should be excluded from analysis.
// These typically contain dependencies, generated code, or build artifacts.
var skipDirectories = map[string]bool{
	"vendor":       true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"__pycache__":  true,
	".git":         true,
	".svn":         true,
	".hg":          true,
	"coverage":     true,
	"tmp":          true,
	"temp":         true,
}

// binaryExtensions lists file extensions that indicate binary (non-text) files.
var binaryExtensions = map[string]bool{
	// Images
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	".ico": true, ".svg": true, ".webp": true, ".tiff": true,
	// Executables
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".a": true,
	".o": true, ".obj": true, ".lib": true,
	// Archives
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true,
	".7z": true, ".rar": true,
	// Media
	".mp3": true, ".mp4": true, ".avi": true, ".mov": true, ".wmv": true,
	".flv": true, ".wav": true, ".ogg": true,
	// Documents
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true,
	// Fonts
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	// Other binary formats
	".db": true, ".sqlite": true, ".dat": true, ".bin": true,
	".pyc": true, ".pyo": true, ".class": true,
}

// generatedFilePatterns lists substrings that indicate generated code.
var generatedFilePatterns = []string{
	".generated.",
	".gen.",
	".pb.go",      // Protocol buffers
	".pb.gw.go",   // gRPC gateway
	"_generated.",
	"_gen.",
	".g.go",
	"generated_",
	"gen_",
	"wire_gen.go", // Wire dependency injection
	"mock_",       // Mock files
	"_mock.go",
}

// ShouldSkipFile determines if a file should be excluded from analysis.
// Returns true for binary files, generated code, and files in skip directories.
func ShouldSkipFile(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	// Check if file is in a directory we should skip
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if skipDirectories[part] {
			return true
		}
	}

	// Check file extension for binary files
	ext := strings.ToLower(filepath.Ext(path))
	if binaryExtensions[ext] {
		return true
	}

	// Check for generated file patterns
	filename := filepath.Base(path)
	for _, pattern := range generatedFilePatterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}

	return false
}

// IsTextFile checks if a file extension indicates it's likely a text file.
// Returns true for known code/text extensions, false for binary or unknown.
func IsTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	// If it's a known binary extension, it's not text
	if binaryExtensions[ext] {
		return false
	}

	// Common text file extensions
	textExtensions := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
		".jsx": true, ".java": true, ".c": true, ".cpp": true, ".h": true,
		".hpp": true, ".cs": true, ".rb": true, ".php": true, ".swift": true,
		".kt": true, ".rs": true, ".scala": true, ".r": true, ".m": true,
		".sql": true, ".sh": true, ".bash": true, ".zsh": true, ".fish": true,
		".html": true, ".css": true, ".scss": true, ".sass": true, ".less": true,
		".xml": true, ".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".md": true, ".txt": true, ".rst": true, ".tex": true,
		".vim": true, ".el": true, ".lua": true, ".pl": true, ".pm": true,
		".makefile": true, ".mk": true, ".cmake": true,
		".proto": true, ".thrift": true, ".graphql": true, ".gql": true,
	}

	// If we recognize it as text, return true
	if textExtensions[ext] {
		return true
	}

	// Files without extension might be text (e.g., Makefile, Dockerfile)
	if ext == "" {
		filename := strings.ToLower(filepath.Base(path))
		if filename == "makefile" || filename == "dockerfile" ||
			filename == "jenkinsfile" || filename == "vagrantfile" ||
			filename == "gemfile" || filename == "rakefile" {
			return true
		}
	}

	// Unknown extension - be conservative and assume it might be text
	// The caller should verify with actual content if needed
	return true
}
