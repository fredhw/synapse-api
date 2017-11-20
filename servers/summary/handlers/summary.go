package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

//PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

//PreviewVideo represents a preview video for a page
type PreviewVideo struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

//PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
	Videos      []*PreviewVideo `json:"videos,omitempty"`
}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, err
	}

	// check response status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GET response status %v", resp.StatusCode)
	}

	// check response content type
	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("content-type not text/html but was %v", ctype)
	}

	return resp.Body, nil
}

//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	base, _ := url.Parse(pageURL)
	tokenizer := html.NewTokenizer(htmlStream)
	summ := PageSummary{}

	for {
		tokenType := tokenizer.Next()

		// if error token
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				return &summ, nil
			}
			return nil, tokenizer.Err()
		}

		// if start tag token
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			token := tokenizer.Token()
			switch token.Data {
			case "link":
				var rel string
				var href string
				var sizes string
				var itype string
				for _, attr := range token.Attr {
					switch attr.Key {
					case "rel":
						rel = attr.Val
					case "href":
						href = attr.Val
					case "sizes":
						sizes = attr.Val
					case "type":
						itype = attr.Val
					}
				}
				if rel == "icon" && len(href) > 0 {
					img := PreviewImage{}
					u, _ := url.Parse(href)
					abs := base.ResolveReference(u)
					img.URL = abs.String()
					if len(sizes) > 0 {
						s := strings.Split(sizes, "x")
						// avoid "any" case
						if len(s) > 1 {
							w, _ := strconv.Atoi(s[1])
							h, _ := strconv.Atoi(s[0])
							img.Width = w
							img.Height = h
						}
					}
					if len(itype) > 0 {
						img.Type = itype
					}
					summ.Icon = &img
				}
			case "title":
				tokenizer.Next()
				if len(summ.Title) == 0 {
					summ.Title = tokenizer.Token().Data
				}
			case "meta":
				var prop string
				var content string
				var name string

				for _, attr := range token.Attr {
					switch attr.Key {
					case "property":
						prop = attr.Val
					case "name":
						name = attr.Val
					case "content":
						content = attr.Val
					}
				}

				if len(name) > 0 {
					switch name {
					case "description":
						if len(summ.Description) == 0 {
							summ.Description = content
						}
					case "author":
						summ.Author = content
					case "keywords":
						words := strings.Split(content, ",")
						for _, word := range words {
							word := strings.TrimSpace(word)
							summ.Keywords = append(summ.Keywords, word)
						}
					}
				}

				if len(prop) > 0 && len(content) > 0 {
					switch prop {
					case "og:type":
						summ.Type = content
					case "twitter:card":
						if len(summ.Type) == 0 {
							summ.Type = content
						}
					case "og:url":
						summ.URL = content
					case "og:title":
						summ.Title = content
					case "twitter:title":
						if len(summ.Title) == 0 {
							summ.Title = content
						}
					case "og:site_name":
						summ.SiteName = content
					case "og:description":
						summ.Description = content
					case "twitter:description":
						if len(summ.Description) == 0 {
							summ.Description = content
						}
					case "twitter:image":
						found := 0
						for _, image := range summ.Images {
							if image.URL == content {
								found = 1
							}
						}
						if found == 0 {
							prev := PreviewImage{}
							u, _ := url.Parse(content)
							abs := base.ResolveReference(u)
							prev.URL = abs.String()
							summ.Images = append(summ.Images, &prev)
						}
					case "og:image":
						prev := PreviewImage{}
						u, _ := url.Parse(content)
						abs := base.ResolveReference(u)
						prev.URL = abs.String()
						summ.Images = append(summ.Images, &prev)
					case "og:image:url":
						found := 0
						for _, image := range summ.Images {
							if image.URL == content {
								found = 1
							}
						}
						if found == 0 {
							prev := PreviewImage{}
							u, _ := url.Parse(content)
							abs := base.ResolveReference(u)
							prev.URL = abs.String()
							summ.Images = append(summ.Images, &prev)
						}
					case "og:image:secure_url":
						prev := summ.Images[len(summ.Images)-1]
						prev.SecureURL = content
					case "og:image:type":
						prev := summ.Images[len(summ.Images)-1]
						prev.Type = content
					case "og:image:width":
						prev := summ.Images[len(summ.Images)-1]
						w, _ := strconv.Atoi(content)
						prev.Width = w
					case "og:image:height":
						prev := summ.Images[len(summ.Images)-1]
						h, _ := strconv.Atoi(content)
						prev.Height = h
					case "og:image:alt":
						prev := summ.Images[len(summ.Images)-1]
						prev.Alt = content
					case "og:video":
						prev := PreviewVideo{}
						u, _ := url.Parse(content)
						abs := base.ResolveReference(u)
						prev.URL = abs.String()
						summ.Videos = append(summ.Videos, &prev)
					case "og:video:url":
						found := 0
						for _, video := range summ.Videos {
							if video.URL == content {
								found = 1
							}
						}
						if found == 0 {
							prev := PreviewVideo{}
							u, _ := url.Parse(content)
							abs := base.ResolveReference(u)
							prev.URL = abs.String()
							summ.Videos = append(summ.Videos, &prev)
						}
					case "og:video:secure_url":
						prev := summ.Videos[len(summ.Videos)-1]
						prev.SecureURL = content
					case "og:video:type":
						prev := summ.Videos[len(summ.Videos)-1]
						prev.Type = content
					case "og:video:width":
						prev := summ.Videos[len(summ.Videos)-1]
						w, _ := strconv.Atoi(content)
						prev.Width = w
					case "og:video:height":
						prev := summ.Videos[len(summ.Videos)-1]
						h, _ := strconv.Atoi(content)
						prev.Height = h
					}
				}
			}
		}
	}
}
