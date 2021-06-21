package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/DaikiYamakawa/deepl-go"
	"github.com/PuerkitoBio/goquery"
)

type RequestBody struct {
	Text string `json:"text,omitempty"`
}

func main() {

	deeplClient, err := deepl.New("https://api-free.deepl.com", nil)
	if err != nil {
		panic(err)
	}

	slackEndpoint := os.Getenv("SLACK_ENDPOINT")

	scrape(deeplClient, slackEndpoint)
}

func scrape(deeplClient *deepl.Client, slackEndpoint string) {

	res, err := http.Get("https://openaccess.thecvf.com/CVPR2021?day=all")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".ptitle").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		link, exists := s.Find("a").Attr("href")
		if !exists {
			log.Fatal("link doesn't exist")
		}
		translateAbstract(link, deeplClient, slackEndpoint)
	})
}

func translateAbstract(link string, deeplClient *deepl.Client, slackEndpoint string) {

	baseURL := "https://openaccess.thecvf.com"

	res, err := http.Get(baseURL + link)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	enAbstract := doc.Find("#abstract").Text()

	translateResponse, err := deeplClient.TranslateSentence(context.Background(), enAbstract, "EN", "JA")
	if err != nil {
		log.Fatal(err)
	}

	sendToSlack(translateResponse.Translations[0].Text, slackEndpoint)
}

func sendToSlack(abstract string, slackEndpoint string) {
	body := &RequestBody{
		Text: abstract,
	}

	jsonString, err := json.Marshal(body)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", slackEndpoint, bytes.NewBuffer(jsonString))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}
