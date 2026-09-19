// Package clipboard isolates OS clipboard access from cast's application logic.
package clipboard

// Clipboard is the small interface needed by clipboard history services.
type Clipboard interface {
	Read() (string, error)
	Write(string) error
}
