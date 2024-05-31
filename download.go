package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

type RGAuth struct {
	Token string
}

type RGResp struct {
	Gif   *RGGif
	Error *RGError
}

type RGGif struct {
	Id   string
	Urls map[string]string
}

type RGError struct {
	Code        string
	Description string
}

var RGtoken string

func DownloadRG(gifUrl string, path string) error {
	if RGtoken == "" {
		resp, err := http.Get("https://api.redgifs.com/v2/auth/temporary")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var r RGAuth
		err = json.Unmarshal(body, &r)
		if err != nil {
			return err
		}
		RGtoken = r.Token
	}

	u, err := url.Parse(gifUrl)
	if err != nil {
		return err
	}
	tag := strings.TrimPrefix(u.Path, "/watch/")

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, "https://api.redgifs.com/v2/gifs/"+tag, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+RGtoken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r RGResp
	err = json.Unmarshal(body, &r)
	if err != nil {
		return err
	}

	if r.Error != nil {
		return fmt.Errorf("%s: %s", r.Error.Description, gifUrl)
	}

	hdurl := r.Gif.Urls["hd"]
	return DownloadFile2(hdurl, filepath.Join(path, r.Gif.Id+".mp4"))
}

func DownloadFile2(url string, path string) error {
	outfile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outfile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(outfile, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
