package autofiber

// FieldErrorDetail represents a single field validation error
type FieldErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag,omitempty"`
}

// ValidationResponseError is used for response validation errors
type ValidationResponseError struct {
	Message string             `json:"error"`
	Details []FieldErrorDetail `json:"details,omitempty"`
}

// ValidationRequestError is used for request validation errors
type ValidationRequestError struct {
	Message string             `json:"error"`
	Details []FieldErrorDetail `json:"details,omitempty"`
}

// Error implements the error interface for ValidationResponseError
func (e *ValidationResponseError) Error() string {
	return e.Message
}

// Error implements the error interface for ValidationRequestError
func (e *ValidationRequestError) Error() string {
	return e.Message
}
