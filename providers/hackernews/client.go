package hackernews

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/drgarcia1986/gonews/providers"
	"github.com/drgarcia1986/gonews/story"
)

var (
	urlBase       = "https://hacker-news.firebaseio.com/v0"
	urlTopStories = fmt.Sprintf("%s/topstories.json", urlBase)
	urlNewStories = fmt.Sprintf("%s/newstories.json", urlBase)
	urlStoryBase  = fmt.Sprintf("%s/item", urlBase)
)

type HackerNews struct{}

func getStoryURL(id int) string {
	return fmt.Sprintf("%s/%d.json", urlStoryBase, id)
}

func getCommentsURL(id int) string {
	return fmt.Sprintf("https://news.ycombinator.com/item?id=%d", id)
}

func getStoryIds(url string) ([]int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ids := []int{}
	if err = json.Unmarshal(body, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func getStory(id int) (*story.Story, error) {
	url := getStoryURL(id)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type r struct {
		Title    string `json:"Title"`
		URL      string `json:"url"`
		Comments int    `json:"descendants"`
	}
	var temp r

	if err = json.Unmarshal(body, &temp); err != nil {
		return nil, err
	}

	return &story.Story{
		Title:         temp.Title,
		URL:           temp.URL,
		CommentsCount: temp.Comments,
		CommentsURL:   getCommentsURL(id),
	}, nil
}

func getURL(storyType int) string {
	switch storyType {
	case story.TopStories:
		return urlTopStories
	default:
		return urlNewStories
	}
}

func storiesGenerator(targetIds []int) <-chan chan *providers.StoryRequest {
	generator := make(chan chan *providers.StoryRequest, len(targetIds))

	go func() {
		defer close(generator)
		for _, id := range targetIds {
			generator <- func(id int) chan *providers.StoryRequest {
				future := make(chan *providers.StoryRequest, 1)
				go func() {
					defer close(future)
					story, err := getStory(id)
					future <- &providers.StoryRequest{story, err}
				}()
				return future
			}(id)
		}
	}()

	return generator
}

func (h *HackerNews) GetStories(storyType, limit int) (<-chan chan *providers.StoryRequest, error) {
	url := getURL(storyType)
	ids, err := getStoryIds(url)
	if err != nil {
		return nil, err
	}
	targetIds := ids[:limit]

	return storiesGenerator(targetIds), nil
}

func (h *HackerNews) Name() string {
	return "HackerNews"
}

func New() providers.Provider {
	return new(HackerNews)
}
