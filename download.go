package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func DownloadFile(url string, path string) error {
	mkdir := exec.Command("mkdir", "-p", path)
	mkdir.Stderr = os.Stderr
	err := mkdir.Run()
	if err != nil {
		return err
	}

	curl := exec.Command("curl", "-O", url, "-#")
	curl.Stdout = os.Stdout
	curl.Stderr = os.Stderr
	curl.Dir = path
	return curl.Run()
}

func DownloadNamedFile(url string, path string, name string) error {
	mkdir := exec.Command("mkdir", "-p", path)
	mkdir.Stderr = os.Stderr
	err := mkdir.Run()
	if err != nil {
		return err
	}

	curl := exec.Command("curl", "-o", name, url, "-#")
	curl.Stdout = os.Stdout
	curl.Stderr = os.Stderr
	curl.Dir = path
	return curl.Run()
}

type RGResp struct {
	Gif          *RGGif
	ErrorMessage *RGError
}

type RGGif struct {
	Id   string
	Urls map[string]string
}

type RGError struct {
	Code        string
	Description string
}

func DownloadRG(gifUrl string, path string) error {
	u, err := url.Parse(gifUrl)
	if err != nil {
		return err
	}
	tag := strings.TrimPrefix(u.Path, "/watch/")

	resp, err := http.Get("https://api.redgifs.com/v2/gifs/" + tag)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r RGResp
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}

	if r.ErrorMessage != nil {
		return fmt.Errorf("%s: %s", r.ErrorMessage.Description, gifUrl)
	}

	hdurl := r.Gif.Urls["hd"]
	return DownloadFile(hdurl, path)
}
