package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	if err := scrape(deeplClient, slackEndpoint); err != nil {
		log.Fatal(err)
	}
}

const baseURL = "https://openaccess.thecvf.com"
const message = `
[%d/1660]
*%s*
%s
[<%s|abstract>] [<%s|pdf>]
%s
%s
DeepL API quota: %d/500000
`

func scrape(deeplClient *deepl.Client, slackEndpoint string) error {

	res, err := http.Get("https://openaccess.thecvf.com/CVPR2021?day=all")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}

	doc.Find(".ptitle").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		relativePath, exists := s.Find("a").Attr("href")
		if !exists {
			log.Fatal("link doesn't exist")
		}
		title := s.Find("a").Text()

		fullLink := baseURL + relativePath

		enAbstract, jaAbstract, authors, pdfRelativePath, err := translateAbstract(fullLink, deeplClient)
		if err != nil {
			log.Fatal(err)
		}

		status, err := deeplClient.GetAccountStatus(context.TODO())
		if err != nil {
			log.Fatal(err)
		}

		text := fmt.Sprintf(message, i+1, title, authors, fullLink, baseURL+pdfRelativePath, enAbstract, jaAbstract, status.CharacterCount)

		sendToSlack(text, slackEndpoint)
	})

	return nil
}

func translateAbstract(link string, deeplClient *deepl.Client) (string, string, string, string, error) {

	res, err := http.Get(link)
	if err != nil {
		return "", "", "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", "", "", err
	}

	authors := doc.Find("#authors").Find("b").Find("i").Text()
	pdfRelativePath, exists := doc.Find("dd").Find("a").First().Attr("href")
	if !exists {
		log.Fatal("link doesn't exist")
	}
	enAbstract := doc.Find("#abstract").Text()
	translateResponse, err := deeplClient.TranslateSentence(context.Background(), enAbstract, "EN", "JA")
	if err != nil {
		return "", "", "", "", err
	}

	return enAbstract, translateResponse.Translations[0].Text, authors, pdfRelativePath, nil
}

func sendToSlack(text string, slackEndpoint string) {
	body := &RequestBody{
		Text: text,
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
