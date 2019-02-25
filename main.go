package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

//Wakatime is interface for wakatime API
type Wakatime interface {
	Current() string
}

//WClient is client implementation of Wakatime interface
type WClient struct {
	apiKey string
	host   string
	client *http.Client
}

//WakatimeResponse represents general response
type WakatimeResponse struct {
	Data     []DurationResponse `json:"data"`
	Branches []string           `json:"branches"`
	Start    string             `json:"start"`
	End      string             `json:"end"`
	Timezone string             `json:"timezone"`
}

//DurationResponse represents data on /durations endpoint
type DurationResponse struct {
	Project  string  `json:"project,omitempty"`
	Time     float64 `json:"time,omitempty"`
	Duration float64 `json:"duration,omitempty"`
}

//TotalDuration return sum of duration
func (w *WakatimeResponse) TotalDuration() float64 {
	totalDuration := float64(0)
	for _, d := range w.Data {
		totalDuration += d.Duration
	}
	return totalDuration
}

//NewWakatimeClient create instance of Wakatime
func NewWakatimeClient() Wakatime {
	apiKey := os.Getenv("WAKATIME_API_KEY")
	client := http.DefaultClient
	host := "https://wakatime.com/api/v1"
	return &WClient{
		apiKey: apiKey,
		client: client,
		host:   host,
	}

}

//Current returns user current duration
func (w *WClient) Current() string {
	endpoint := "/users/current/durations"
	date := time.Now().Format("2006-01-02")

	query := map[string]string{
		"api_key": w.apiKey,
		"date":    date,
	}

	result, err := w.doGet(endpoint, query, map[string]string{})
	if err != nil {
		log.Println(err)
		return "n/a"
	}

	return result

}

func (w *WClient) doGet(endpoint string, params map[string]string, headers map[string]string) (string, error) {
	urlString := w.host + endpoint
	u, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	for k, v := range params {
		q.Add(k, v)
	}

	u.RawQuery = q.Encode()
	parsedURL := u.String()

	req, err := http.NewRequest("GET", parsedURL, nil)
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("NOT OK " + resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	result, err := w.parseBody(b)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func (w *WClient) parseBody(b []byte) ([]byte, error) {
	var response WakatimeResponse

	err := json.Unmarshal(b, &response)
	if err != nil {
		return []byte{}, err
	}

	d := response.TotalDuration()

	duration, err := formatDuration(d)
	if err != nil {
		return []byte{}, err
	}

	return []byte(duration), nil
}

func formatDuration(d float64) (string, error) {
	strDuration := fmt.Sprintf("%ds", int64(d))
	duration, err := time.ParseDuration(strDuration)
	if err != nil {
		return "", err
	}

	duration = duration.Round(time.Minute)
	h := duration / time.Hour
	duration -= h * time.Hour
	m := duration / time.Minute

	return fmt.Sprintf("%d hour %d minute", h, m), nil
}

func main() {
	w := NewWakatimeClient()
	fmt.Println(w.Current())
}
