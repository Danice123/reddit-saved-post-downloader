package main

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func init() {
	mime.AddExtensionType(".jpg", "image/jpg")
}

func getGenericChild(parent interface{}, key string) (interface{}, bool) {
	cast, ok := parent.(map[string]interface{})
	if !ok {
		panic("Bad parent " + key)
	}
	child, ok := cast[key]
	return child, ok
}

func HandleGallery(client *reddit.Client, postId string) (map[string]string, error) {
	req, err := client.NewRequest(http.MethodGet, fmt.Sprintf("comments/%s", postId), nil)
	if err != nil {
		return nil, err
	}

	var responseGeneric []interface{}
	_, err = client.Do(context.TODO(), req, &responseGeneric)
	if err != nil {
		return nil, err
	}

	var data interface{}
	for _, listing := range responseGeneric {
		listingData, _ := getGenericChild(listing, "data")
		children, _ := getGenericChild(listingData, "children")
		for _, child := range children.([]interface{}) {
			if t, _ := getGenericChild(child, "kind"); t.(string) == "t3" {
				data, _ = getGenericChild(child, "data")
				break
			}
		}
	}

	var images interface{}
	var gallery interface{}
	if _, ok := getGenericChild(data, "crosspost_parent"); ok {
		data, _ = getGenericChild(data, "crosspost_parent_list")
		images, _ = getGenericChild(data.([]interface{})[0], "media_metadata")
		gallery, _ = getGenericChild(data.([]interface{})[0], "gallery_data")
		gallery, _ = getGenericChild(gallery, "items")
	} else {
		if removed, ok := getGenericChild(data, "removed_by_category"); ok && removed != nil {
			return nil, errors.New("removed gallery post")
		}
		images, _ = getGenericChild(data, "media_metadata")
		gallery, _ = getGenericChild(data, "gallery_data")
		gallery, _ = getGenericChild(gallery, "items")
	}

	idMap := map[string]int{}
	for _, image := range gallery.([]interface{}) {
		media, _ := getGenericChild(image, "media_id")
		id, _ := getGenericChild(image, "id")
		idMap[media.(string)] = int(id.(float64))
	}

	imageToFetch := map[string]string{}
	for _, image := range images.(map[string]interface{}) {

		mediaId, _ := getGenericChild(image, "id")
		mediaMime, _ := getGenericChild(image, "m")
		ext, err := mime.ExtensionsByType(mediaMime.(string))
		if err != nil {
			panic(err)
		}
		if len(ext) == 0 {
			panic("No extensions for " + mediaMime.(string))
		}
		imageToFetch[fmt.Sprintf("%s_%d%s", postId, idMap[mediaId.(string)], ext[0])] = fmt.Sprintf("https://i.redd.it/%s%s", mediaId.(string), ext[0])
	}

	return imageToFetch, nil
}
