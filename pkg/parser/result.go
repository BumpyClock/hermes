package parser

// SetError sets error state for JS compatibility
func (r *Result) SetError(message string) {
	r.Error = true
	r.Message = message
}

// IsError checks if result contains an error
func (r *Result) IsError() bool {
	return r.Error
}