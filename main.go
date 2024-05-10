package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
)

const (
	apiEndpoint = "https://api.rss2json.com/v1/api.json?rss_url="
)

type VideoResponse struct {
	Items []struct {
		Link string `json:"link"`
	} `json:"items"`
}

type Playlist struct {
	Title string
	URL   string
}

type ViewData struct {
	LatestVideoLink string
	VideoID         string
}

func fetchLatestVideo(channelID string) (string, error) {
	channelURL := "https://www.youtube.com/feeds/videos.xml?channel_id=" + channelID
	resp, err := http.Get(apiEndpoint + url.QueryEscape(channelURL))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch RSS feed: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var videoResp VideoResponse
	err = json.Unmarshal(body, &videoResp)
	if err != nil {
		return "", err
	}

	if len(videoResp.Items) == 0 {
		return "", fmt.Errorf("no video items found in RSS feed")
	}

	latestVideoLink := videoResp.Items[0].Link
	return latestVideoLink, nil
}

func getLatestVideoID(latestVideoLink string) (string, error) {
	videoID := latestVideoLink
	fmt.Println("Latest video link:", latestVideoLink)
	// re := regexp.MustCompile(`embed/([^?]+)`)
	re := regexp.MustCompile(`[?&]v=([^&]+)`)

	// Find the submatch
	match := re.FindStringSubmatch(videoID)

	// Check if the match is found
	if len(match) > 1 {
		extractedValue := match[1]
		fmt.Println("Extracted value:", extractedValue)
		return extractedValue, nil
	} else {
		return "", fmt.Errorf("no match found")
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		channelID := r.FormValue("channelID")
		latestVideoLink, err := fetchLatestVideo(channelID)
		if err != nil {
			http.Error(w, "Failed to fetch latest video", http.StatusInternalServerError)
			return
		}
		videoID, err := getLatestVideoID(latestVideoLink)
		if err != nil {
			http.Error(w, "Failed to extract video ID", http.StatusInternalServerError)
			return
		}

		data := ViewData{
			LatestVideoLink: latestVideoLink,
			VideoID:         videoID,
		}

		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
	} else {
		tmpl, err := template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			return
		}
	}
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.ListenAndServe(":8080", nil)
}
