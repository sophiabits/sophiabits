package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

const FEED_URL = "https://sophiabits.com/feed.json"
const OUTPUT_PATH = "../README.md"
const TEMPLATE_PATH = "../.template.md"

type JSONFeedItem struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
	Url   string   `json:"url"`
}

type JSONFeed struct {
	Items []JSONFeedItem `json:"items"`
}

// Improves appearance of tag in a sentence
func formatTag(tag string) string {
	if tag == "APIs" || tag == "AWS" || tag == "DevOps" || tag == "EdTech" || tag == "macOS" || tag == "SEO" || tag == "UI" {
		return tag
	}
	return strings.ToLower(tag)
}

// Select a tag which makes sense in a sentence; otherwise fall back to "technology"
// TODO: Make chosen tag deterministic; maybe hash post title?
func pickTag(tags []string) string {
	badTags := []string{"essay", "mobile", "retrospective", "review", "tutorial"}

	validTags := []string{}
	for i := range tags {
		if !slices.Contains(badTags, strings.ToLower(tags[i])) {
			validTags = append(validTags, tags[i])
		}
	}

	if len(validTags) == 0 {
		return "technology"
	}

	return validTags[rand.Intn(len(validTags))]
}

func main() {
	rand.Seed(time.Now().Unix())

	// Grab feed
	res, err := http.Get(FEED_URL)
	if err != nil {
		log.Fatalf("error fetching feed: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("error reading feed: %v", err)
	}

	feed := JSONFeed{}
	err = json.Unmarshal(body, &feed)
	if err != nil {
		log.Fatalf("error parsing feed: %v", err)
	}

	// Assume first item is the most recent
	item := feed.Items[0]
	linkString := fmt.Sprintf("[%s](%s)", item.Title, item.Url)

	template, err := ioutil.ReadFile(TEMPLATE_PATH)
	if err != nil {
		log.Fatalf("failed to read in template: %v", err)
	}

	// Substitute values into template
	data := string(template)
	data = strings.Replace(data, "{{LINK}}", linkString, 1)
	data = strings.Replace(data, "{{TAG}}", formatTag(pickTag(item.Tags)), 1)
	data = strings.Replace(data, "{{TIMESTAMP}}", time.Now().Format("1 Jan 2006"), 1)

	// Write generated README
	file, err := os.Create(OUTPUT_PATH)
	if err != nil {
		log.Fatalf("error creating README.md: %v", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		log.Fatalf("error writing README.md: %v", err)
	}

	file.Sync()
}
