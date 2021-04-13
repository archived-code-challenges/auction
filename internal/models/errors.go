package models

// These errors are returned by the services and can be used to provide error codes to the
// API results.
const (
	ErrNotFound ModelError = "models: not_found, resource not found"
	ErrLowValue ModelError = "models: low_value, bid amount should be higher than highest"
	ErrConflict ModelError = "models: conflict, resource is being used"
	ErrRequired ModelError = "models: required, value cannot be empty"
)

// PublicError is an error that returns a string code that can be presented to the API user.
type PublicError interface {
	error
	Public() string
	Detail() string
}

// ModelError defines errors exported by this package. This type implement a Public() method that
// extracts a unique error code defined for each error value exported.
type ModelError string

// Error returns the exact original message of the e value.
func (e ModelError) Error() string {
	return string(e)
}

// Public extracts the error code string present on the value of e.
//
// An error code is defined as the string after the package prefix and colon, and before the comma that follows this string. Example:
//		"models: error_code, this is a validation error"
func (e ModelError) Public() string {
	// remove the prefix
	s := string(e)[len("models: "):]

	// extract the error code
	for i := 1; i < len(s); i++ {
		if s[i] == ',' {
			s = s[:i]
			break
		}
	}

	return s
}

// Detail extracts the error detail string present on the value of e.
//
// An error detail is defined as the string after the package prefix and colon, and after the comma that follows this string. Example:
//		"models: error_code, this is the error detail string"
func (e ModelError) Detail() string {
	// remove the prefix
	s := string(e)[len("models: "):]

	// extract the error code
	for i := 1; i < len(s); i++ {
		if s[i] == ',' {
			s = s[i+2:] // +2 removes the comma and the space
			break
		}
	}

	return s
}

type ValidationError map[string]PublicError

// Error returns the list of fields with validation errors. The specific error for each field is not included.
func (v ValidationError) Error() string {
	ret := "models: validation error on fields "
	for k := range v {
		ret += k + ", "
	}

	return ret[:len(ret)-2]
}

// Is helps xerrors.Is check if a target error is a ValidationError. If err (the target) contains
// fields, the error values for each field are compared to the ones in v and their values must
// match. If err contains a subset of the fields in v, it is considered to match.
func (v ValidationError) Is(err error) bool {
	ve, ok := err.(ValidationError)
	if !ok {
		return false
	}

	for k := range ve {
		if v[k] != ve[k] {
			return false
		}
	}

	return true
}
