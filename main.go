package main

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/vartanbeno/go-reddit/v2/reddit"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Scrape map[string]string
}

func client() *reddit.Client {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	client, err := reddit.NewClient(reddit.Credentials{
		ID:       os.Getenv("REDDIT_SCRIPT_ID"),
		Secret:   os.Getenv("REDDIT_SCRIPT_SECRET"),
		Username: os.Getenv("REDDIT_USERNAME"),
		Password: os.Getenv("REDDIT_PASSWORD"),
	}, reddit.WithUserAgent("golang:saved-post-download:v1.0.0 (by XxJewishRevengexX)"))
	if err != nil {
		panic(err)
	}
	return client
}

func downloader(u string) func(u string, path string) error {
	pu, err := url.Parse(u)
	if err != nil {
		println(err.Error())
		return nil
	}

	switch strings.TrimPrefix(pu.Host, "www.") {
	case "i.redd.it":
		fallthrough
	case "cdn-images.imagevenue.com":
		fallthrough
	case "i.imgur.com":
		return DownloadFile

	case "redgifs.com":
		return DownloadRG
	}

	return nil
}

func main() {
	client := client()

	d, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.Unmarshal(d, &config)
	if err != nil {
		panic(err)
	}

	posts, _, _, err := client.User.Saved(context.TODO(), &reddit.ListUserOverviewOptions{
		ListOptions: reddit.ListOptions{Limit: 100},
	})
	if err != nil {
		panic(err)
	}

	for _, post := range posts {
		if path, ok := config.Scrape[post.SubredditName]; ok {
			if strings.HasPrefix(post.URL, "https://www.reddit.com/gallery/") {
				imageIds, err := HandleGallery(client, post.ID)
				if err != nil {
					println(err.Error())
					continue
				}

				for name, u := range imageIds {
					err = DownloadNamedFile(u, filepath.Join(path, post.Title), name)
					if err != nil {
						panic(err)
					}
				}

				_, err = client.Post.Unsave(context.TODO(), post.FullID)
				if err != nil {
					panic(err)
				}
				continue
			}

			df := downloader(post.URL)
			if df != nil {
				err = df(post.URL, path)
				if err != nil {
					println(err.Error())
					continue
				}

				_, err = client.Post.Unsave(context.TODO(), post.FullID)
				if err != nil {
					panic(err)
				}
			} else {
				println(post.URL)
			}
		}
	}
}
