package codegen

// File is one generated UTF-8 source or metadata file. Path is a slash-separated
// relative path. Target names an output directory (empty is the primary output).
// CommentPrefix optionally selects a line comment marker for a new language:
// //, #, --, ; or %. Leave it empty for built-in Go/TS/Skel/JSON formats.
type File struct {
	Target        string
	Path          string
	Content       string
	CommentPrefix string
}
