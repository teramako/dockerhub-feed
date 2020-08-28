package main

import (
	"fmt"
	"strings"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"log"
	"time"
	"github.com/gorilla/feeds"
)

type DockerHubRepository struct {
	User string `json:"user"`
	Name string `json:"name"`
	Namespace string `json:"namespace"`
	Description string `json:"description"`
	Last_updated time.Time `json:"last_updated"`
}
type DockerHubTagImage struct {
	Architecture string `json:"architecture"`
	Features     string `json:"features"`
	Variant      string `json:"variant"`
	Digest       string `json:"digest"`
	Os           string `json:"os"`
	Os_features  string `json:"os_features"`
	Os_version   string `json:"os_version"`
	Size         uint64 `json:"size"`
}

type DockerHubTag struct {
	Creator               uint      `json:"creator"`
	Id                    uint      `json:"id"`
	Image_id              uint      `json:"image_id"`
	Last_updated          time.Time `json:"last_updated"`
	Last_updater          uint      `json:"last_updater"`
	Last_updater_username string    `json:"last_updater_username"`
	Name                  string    `json:"name"`
	Repository            uint      `json:"repository"`
	Full_size             uint64    `json:"full_size"`
	V2                    bool      `json:"v2"`
	Images                []DockerHubTagImage `json:"images"`
}
func formatSize (b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
type DockerHubTagRoot struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results []DockerHubTag `json:"results"`
}

func GetDockerHubURL (namespace string, name string) string {
	baseURL := "https://hub.docker.com"
	if namespace == "library" {
		return fmt.Sprintf("%s/%s/%s", baseURL, "_", name)
	}
	return fmt.Sprintf("%s/r/%s/%s", baseURL, namespace, name)
}
func FetchRepository (namespace string, name string) DockerHubRepository {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/", namespace, name)
	res, _ := http.Get(url)
	defer res.Body.Close()

	buf, _ := ioutil.ReadAll(res.Body)
	var repoInfo DockerHubRepository
	if err := json.Unmarshal(buf, &repoInfo); err != nil {
		log.Fatal(err)
	}
	return repoInfo
}
func FetchTags (namespace string, name string) DockerHubTagRoot {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags", namespace, name)
	res, _ := http.Get(url)
	defer res.Body.Close()

	buf, _ := ioutil.ReadAll(res.Body)
	var tags DockerHubTagRoot
	if err := json.Unmarshal(buf, &tags); err != nil {
		log.Fatal(err)
	}
	return tags
}
func CreateDockerTagFeed (namespace string, name string) *feeds.Feed {
	repoInfo := FetchRepository(namespace, name)
	tags := FetchTags(namespace, name)
	url := GetDockerHubURL(namespace, name)
	feed := &feeds.Feed{
		Title: fmt.Sprintf("%s - Docker Hub", name),
		Link: &feeds.Link{Href: url},
		Description: repoInfo.Description,
		Author: &feeds.Author{Name: name},
		Updated: repoInfo.Last_updated,
	}
	for _, tag := range tags.Results {
		var str strings.Builder
		str.WriteString(`<section xmlns="http://www.w3.org/1999/xhtml">`)
		str.WriteString(fmt.Sprintf(`<h1>%s:%s</h1>`, name, tag.Name))
		str.WriteString(`<dl>`)
		for _, img := range tag.Images {
			str.WriteString(fmt.Sprintf(`<dt>Digest</dt><dd>%s</dd>`, img.Digest))
			str.WriteString(fmt.Sprintf(`<dt>OS/ARCH</dt><dd>%s/%s</dd>`, img.Os, img.Architecture))
			str.WriteString(fmt.Sprintf(`<dt>Size</dt><dd>%s</dd>`, formatSize(img.Size)))
		}
		str.WriteString(`</dl>`)
		str.WriteString(`</section>`)

		feed.Add(&feeds.Item{
			Title: fmt.Sprintf("%s:%s - Docker Hub Tags", name, tag.Name),
			Id: fmt.Sprintf("tag:hub.docker.com,%s,%s,%s,%s", namespace, name, tag.Name, tag.Last_updated.Format(time.RFC3339)),
			Link: &feeds.Link{Href: url},
			Description: str.String(),
			Author: &feeds.Author{Name: tag.Last_updater_username},
			Updated: tag.Last_updated,
		})
	}
	return feed
}
func atomHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	log.Printf("/atom: %v", q)
	name := q.Get("name")
	namespace := q.Get("user")
	if namespace == "" {
		namespace = "library"
	}
	if name == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	feed := CreateDockerTagFeed(namespace, name)
	atom, err := feed.ToAtom()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/atom+xml")
	w.Header().Set("Last-Modified", feed.Updated.Format(time.RFC1123))
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, atom)
}

func main() {
	http.HandleFunc("/atom", atomHandler)
	http.ListenAndServe(":8080", nil)
}

// vim: set sw=4 ts=4 sts=0 noet list:
