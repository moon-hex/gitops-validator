package types

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Type     string `json:"type"`
	Severity string `json:"severity"` // error, warning, info
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Resource string `json:"resource,omitempty"`
	// Category is set by the orphaned-resource validator when path-based
	// categories are configured. Used for grouped output.
	Category string `json:"category,omitempty"`
}
