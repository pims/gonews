package story

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	TopStories = iota
	NewStories
)

type Story struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	CommentsURL   string `json:"comments_url"`
	CommentsCount int    `json:"comments"`
}

func Title(title string, comments int) string {
	return fmt.Sprintf("%s [%d]", title, comments)
}

func (s *Story) Domain(providerName string) string {
	u, err := url.Parse(s.URL)
	if err != nil {
		return "error"
	}

	h := strings.TrimPrefix(u.Host, "www.")
	if strings.HasPrefix(h, strings.Split(providerName, "-")[0]) {
		return ""
	}
	return h

}
