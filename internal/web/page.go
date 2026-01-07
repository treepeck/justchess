package web

type Page map[string]any

type Form struct {
	IsSignUp bool
}

type Tooltip struct {
	Header  string
	Content []string
}
