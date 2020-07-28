package lobsters

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/drgarcia1986/gonews/providers"
	"github.com/drgarcia1986/gonews/story"
)

type Lobsters struct{}

func (h *Lobsters) GetStories(storyType, limit int) (<-chan chan *providers.StoryRequest, error) {
	stories, err := get()
	if err != nil {
		return nil, err
	}

	generator := make(chan chan *providers.StoryRequest, len(stories))
	go func() {
		defer close(generator)
		for _, s := range stories {
			f := make(chan *providers.StoryRequest, 1)
			f <- &providers.StoryRequest{s, nil}
			close(f)

			generator <- f
		}
	}()

	return generator, nil
}

func (h *Lobsters) Name() string {
	return "Lobsters"
}

func New() providers.Provider {
	return new(Lobsters)
}

func get() ([]*story.Story, error) {

	resp, err := http.DefaultClient.Get("https://lobste.rs/hottest.json")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	type item struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		Comments    int    `json:"comment_count"`
		CommentsURL string `json:"comments_url"`
	}
	content, _ := ioutil.ReadAll(resp.Body)
	items := []item{}
	err = json.Unmarshal(content, &items)
	if err != nil {
		return nil, err
	}
	stories := make([]*story.Story, len(items))
	for i, item := range items {
		stories[i] = &story.Story{
			Title:         item.Title,
			URL:           item.URL,
			CommentsCount: item.Comments,
			CommentsURL:   item.CommentsURL,
		}
	}
	return stories, nil
}
