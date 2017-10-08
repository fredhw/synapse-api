package handlers

import (
	"encoding/json"
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

//SummaryHandler handles requests for the page summary API.
//This API expects one query string parameter named `url`,
//which should contain a URL to a web page. It responds with
//a JSON-encoded PageSummary struct containing the page summary
//meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	/*TODO: add code and additional functions to do the following:
	- Add an HTTP header to the response with the name
	 `Access-Control-Allow-Origin` and a value of `*`. This will
	  allow cross-origin AJAX requests to your server.
	- Get the `url` query string parameter value from the request.
	  If not supplied, respond with an http.StatusBadRequest error.
	- Call fetchHTML() to fetch the requested URL. See comments in that
	  function for more details.
	- Call extractSummary() to extract the page summary meta-data,
	  as directed in the assignment. See comments in that function
	  for more details
	- Close the response HTML stream so that you don't leak resources.
	- Finally, respond with a JSON-encoded version of the PageSummary
	  struct. That way the client can easily parse the JSON back into
	  an object

	Helpful Links:
	https://golang.org/pkg/net/http/#Request.FormValue
	https://golang.org/pkg/net/http/#Error
	https://golang.org/pkg/encoding/json/#NewEncoder
	*/

	w.Header().Add("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	pageURL := r.FormValue("url")
	if len(pageURL) == 0 {
		http.Error(w, "please provide a url", http.StatusBadRequest)
		return
	}

	respBody, err := fetchHTML(pageURL)
	if err != nil {
		http.Error(w, "error fetching URL %v\n", http.StatusBadRequest)
	}
	summ, err := extractSummary(pageURL, respBody)
	if err != nil {
		http.Error(w, "error extracting summary %v\n", http.StatusBadRequest)
	}
	defer respBody.Close()
	json.NewEncoder(w).Encode(summ)
}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	/*TODO: Do an HTTP GET for the page URL. If the response status
	code is >= 400, return a nil stream and an error. If the response
	content type does not indicate that the content is a web page, return
	a nil stream and an error. Otherwise return the response body and
	no (nil) error.

	To test your implementation of this function, run the TestFetchHTML
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestFetchHTML

	Helpful Links:
	https://golang.org/pkg/net/http/#Get
	*/
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, err
	}

	// check response status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("bad request: %v", resp.StatusCode)
	}

	// check response content type
	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("bad request: %v", http.StatusBadRequest)
	}

	return resp.Body, nil
}

//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	/*TODO: tokenize the `htmlStream` and extract the page summary meta-data
	according to the assignment description.

	To test your implementation of this function, run the TestExtractSummary
	test in summary_test.go. You can do that directly in Visual Studio Code,
	or at the command line by running:
		go test -run TestExtractSummary

	Helpful Links:
	https://drstearns.github.io/tutorials/tokenizing/
	http://ogp.me/
	https://developers.facebook.com/docs/reference/opengraph/
	https://golang.org/pkg/net/url/#URL.ResolveReference
	*/

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

			if "link" == token.Data {
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
			}

			if "title" == token.Data {
				tokenizer.Next()
				if len(summ.Title) == 0 {
					summ.Title = tokenizer.Token().Data
				}
			}

			if "meta" == token.Data {
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
								found = 1;
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
