// Package response declares text responses reused in different HTTP handlers across packages.
package response

const (
	// Declaration of error messages.
	Unauthorized    string = "Invalid credentials"
	Conflict        string = "Not unique username or email"
	TokenMissing    string = "Token not found"
	TokenConflict   string = "You already have a pending token"
	CannotHash      string = "Cannot generate password hash"
	CannotSendEmail string = "Cannot send email. Please, ensure that email is valid"
	DatabaseError   string = "Database cannot be accessed. Please, try again later"
	CookieError     string = "Cannot generate secure cookie"
	BadRequest      string = "Malformed request body"
	InternalError   string = "Internal server error. Please try again later"
	NotFound        string = "Resource with the named ID does not exist"
)
