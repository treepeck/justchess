package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"justchess/internal/db"
	"strings"
)

// HTML frame layout used on every page.
const siteLayout = "_web/layouts/site.tmpl"

var (
	// Prefix of the page head part.
	headPrefix = []byte("<!--{")
	// Postfix of the page head part.
	headPostfix = []byte("}-->")

	noHeadPostfix = errors.New("web: the page file doesn't contain the head postfix")
)

// page is a parsed template with data used to fill it up.
type page struct {
	// Player who requested the page. **Always** request-scoped.
	Player db.Player
	// Page title. Either parsed from page head or request-scoped.
	Title string
	// Optional page script. Not included if empty.
	Script string
	// Arbitrary data to fill up the template. Either parsed from page head or
	// request-scoped.
	Data any
	// Parsed template. Changes to a file will not be displayed automatically and
	// require a server restart.
	tmpl *template.Template
}

// parsePage parses [page] from a given file.
func parsePage(path string, file []byte) (page, error) {
	// Parse head if the template file contains it.
	head, _, err := parsePageHead(file)
	if strings.HasSuffix(string(file), string(headPrefix)) && err != nil {
		return page{}, err
	}

	// Include the site layout which contains boilerplate HTML.
	files := []string{siteLayout, path}

	// If page contains some layouts, include them as well.
	lays, exists := head["layouts"]
	if exists {
		// Convert []any to []string.
		for _, lay := range lays.([]any) {
			files = append(files, lay.(string))
		}
	}

	// Parse the resulting template.
	t, err := template.ParseFiles(files...)
	if err != nil {
		return page{}, err
	}

	// Parse page title if specified in head.
	title := ""
	if head["title"] != nil {
		title = head["title"].(string)
	}

	// Parse script if specified in head.
	script := ""
	if head["script"] != nil {
		script = head["script"].(string)
	}

	return page{
		Title:  title,
		Script: script,
		Data:   head,
		tmpl:   t,
	}, nil
}

// parsePageHead parses the page head into a map[string]any. To be parsed,
// the head must begin with [headPrefix], contain JSON fields in a
// standardized format, and end with [headPostfix].
//
// Example:
//
//	 <!--{
//			"foo": "bar",
//			"test": 1,
//			"test1": false
//	 }-->
func parsePageHead(file []byte) (map[string]any, int, error) {
	// body is the remaining content after head.
	head := make(map[string]any)
	// Index of the head ending.
	end := 0

	// Parse head is present.
	end = bytes.Index(file, headPostfix)
	if end < 0 {
		return nil, end, noHeadPostfix
	}
	// Strip prefix and preserve postfix with "\n".
	file = file[len(headPrefix)-1 : end+1]
	// Parse head fields.
	if err := json.Unmarshal(file, &head); err != nil {
		return nil, end, err
	}
	// Strip postfix and trailing "\n".
	end += len(headPostfix) + 1
	return head, end, nil
}
