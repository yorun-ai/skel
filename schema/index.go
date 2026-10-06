package schema

// Entries returns declaration summaries in document order.
func Entries(document *Document) []*Entry {
	entries := make([]*Entry, 0, len(document.Declarations))
	for _, declaration := range document.Declarations {
		entries = append(entries, &Entry{
			Pub: declaration.Pub, Name: declaration.Name, Kind: declaration.Kind, SkelName: declaration.SkelName,
		})
	}
	return entries
}

// Find returns a declaration or nil when the name and kind are absent.
func Find(document *Document, kind DeclarationType, skelName string) *Declaration {
	for _, declaration := range document.Declarations {
		if declaration.Kind == kind && declaration.SkelName == skelName {
			return declaration
		}
	}
	return nil
}
