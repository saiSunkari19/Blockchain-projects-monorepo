package twitterapis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/dghubble/oauth1"
)

const mediaUploadURLV1 = "https://upload.twitter.com/1.1/media/upload.json"
const tweetPostURLV2 = "https://api.twitter.com/2/tweets"

func RunTwitterAPI() {

	// Post Tweet

	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	fmt.Println("consumerKey", consumerKey)

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessSecret)

	httpClient := config.Client(oauth1.NoContext, token)

	ctx := context.Background()

	rateMgr := NewRateLimitManager()

	imageURL := "https://ipfs.clutchplay.ai/ipfs/QmZiNLcJQcg7Jttm77trffmSVj2UAyL1Zc7T7fWXHVKv23"

	// Each image results in one tweet
	imageData, filename, err := DownloadImage(ctx, imageURL)
	if err != nil {
		fmt.Errorf("failed to get image data", err.Error())
	}

	fmt.Println("downloaded image")
	mediaID, err := UploadMedia(ctx, imageData, filename, httpClient)
	if err != nil {
		fmt.Errorf("failed to upload media %v", err)
	}

	mediaIDs := []string{mediaID}

	fmt.Println("uploaded image", mediaID)

	payload := map[string]interface{}{"text": ""}
	if len(mediaIDs) > 0 {
		payload["media"] = map[string]interface{}{"media_ids": mediaIDs}
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Errorf("failed to marshal tweet payload: %w", err)
	}
	fmt.Println("marshaled the data", string(jsonPayload))

	// Check rate limits

	check, duration := rateMgr.CheckOverallRateLimits()

	fmt.Println("check, duration", check, duration.Seconds())

	resp, err := makeRequest(ctx, "POST", tweetPostURLV2, bytes.NewBuffer(jsonPayload), "application/json", httpClient)
	rateMgr.UpdateFromHeader(resp.Header)

	if err != nil {
		defer resp.Body.Close()
		rateMgr.UpdateFromHeader(resp.Header)
	}

	// tweetText := "Posting from Go using Twitter API v2 and OAuth1.0a"
	// body, _ := json.Marshal(map[string]string{"text": tweetText})

	// tweetRequest, _ := http.NewRequest("POST", "https://api.twitter.com/2/tweets", bytes.NewBuffer(body))
	// tweetRequest.Header.Set("Content-Type", "application/json")

	// resp, err := httpClient.Do(tweetRequest)
	// if err != nil {
	// 	panic(err)
	// }
	// defer resp.Body.Close()

	// for k, v := range resp.Header() {
	// 	fmt.Printf("%s: %s\n", k, v)

	// }

	fmt.Println("x-user-limit-24hour-limit", resp.Header.Get("x-user-limit-24hour-limit"))
	fmt.Println("x-user-limit-24hour-reset", resp.Header.Get("x-user-limit-24hour-reset"))
	fmt.Println("x-user-limit-24hour-remaining", resp.Header.Get("x-user-limit-24hour-remaining"))
	fmt.Println("x-user-limit-24hour-limit", resp.Header.Get("x-user-limit-24hour-limit"))

	fmt.Println("x-app-limit-24hour-limit", resp.Header.Get("x-app-limit-24hour-limit"))
	fmt.Println("x-app-limit-24hour-reset", resp.Header.Get("x-app-limit-24hour-reset"))
	fmt.Println("x-app-limit-24hour-remaining", resp.Header.Get("x-app-limit-24hour-remaining"))
	fmt.Println("x-app-limit-24hour-limit", resp.Header.Get("x-app-limit-24hour-limit"))

	respBody, _ := io.ReadAll(resp.Body)
	fmt.Println("Status: ", resp.Status)
	fmt.Println("Body: ", string(respBody))

}

func DownloadImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request for image download %s: %w", imageURL, err)
	}

	dlClient := http.DefaultClient
	resp, err := dlClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download image %s: status %d", imageURL, resp.StatusCode)
	}

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	filename := "image.jpg"
	if resp.Header.Get("Content-Type") == "image/webp" {
		filename = "image.webp"
	}

	return imageData, filename, nil
}

func UploadMedia(ctx context.Context, imageData []byte, filename string, httpClient *http.Client) (string, error) {

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file for media: %w", err)
	}
	_, err = part.Write(imageData)
	if err != nil {
		return "", fmt.Errorf("failed to write image data to form: %w", err)
	}
	writer.Close()

	fmt.Println("Make request")

	resp, err := makeRequest(ctx, "POST", mediaUploadURLV1, &requestBody, writer.FormDataContentType(), httpClient)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
		}
		return "", fmt.Errorf("media upload request failed: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("Request done...")

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("media upload failed, status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var mediaResponse TweetMediaResponseV1
	if err := json.NewDecoder(resp.Body).Decode(&mediaResponse); err != nil {
		return "", fmt.Errorf("failed to decode media upload response: %w", err)
	}

	if mediaResponse.MediaIDString == "" {
		return "", errors.New("media upload response did not contain media_id_string")
	}

	fmt.Println("media id response", string(mediaUploadURLV1))
	return mediaResponse.MediaIDString, nil
}

type TweetMediaResponseV1 struct {
	MediaID          int64  `json:"media_id"`
	MediaIDString    string `json:"media_id_string"`
	Size             int    `json:"size"`
	ExpiresAfterSecs int    `json:"expires_after_secs"`
	Image            struct {
		ImageType string `json:"image_type"`
		Width     int    `json:"w"`
		Height    int    `json:"h"`
	} `json:"image"`
}

func makeRequest(ctx context.Context, method, url string, body io.Reader, contentType string, httpClient *http.Client) (*http.Response, error) {

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", url, err)

	}

	if resp.StatusCode == http.StatusTooManyRequests { // 429
		defer resp.Body.Close()
		return resp, fmt.Errorf("twitter API rate limit (HTTP 429) for %s", url)
	}

	return resp, nil

}
