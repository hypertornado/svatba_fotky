package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var templates = template.Must(template.ParseGlob("templates/*"))

func main() {
	var err error

	images := NewImages()

	for _, v := range []string{"br", "images"} {
		err = LoadPhotos(v, images)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	images.Print()

	Start(8543, images)
}

func Start(port int, images *Images) {

	handler := &ServerHandler{images}

	server := &http.Server{
		Addr:           ":" + strconv.Itoa(port),
		Handler:        handler,
		ReadTimeout:    2 * time.Minute,
		WriteTimeout:   2 * time.Minute,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Starting", port)
	server.ListenAndServe()
}

type ServerHandler struct {
	images *Images
}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.URL.Path)

	if r.Method == "POST" && r.URL.Path == "/upload" {
		info, err := h.ServeUpload(w, r)
		if err != nil {
			h.ServePage(w, r, "index", map[string]string{"title": "Error", "error": "error occured, sorry"})
		} else {
			h.images.Add(info.Path, info.Uploaded)
			v := url.Values{}
			v.Add("path", info.Path)
			Redirect(w, "/uploaded?"+v.Encode())
			return
		}
		return
	}

	if r.Method == "GET" && r.URL.Path == "/slideshow" {
		h.ServePage(w, r, "slideshow", map[string]interface{}{"title": "Slideshow"})
		return
	}

	if r.Method == "GET" && r.URL.Path == "/api" {
		var result []byte

		var img *ImgInfo
		typ := r.URL.Query().Get("type")

		switch typ {
		case "latest":
			img = h.images.GetLast()
		case "friends":
			img = h.images.GetFriends()
		case "children":
			img = h.images.GetChildren()
		case "smart":
			img = h.images.GetSmart()
		default:
			img = h.images.GetRandom()
		}

		result, err := json.Marshal(img)
		if err != nil {
			fmt.Println(err)
		}
		w.Write(result)
		return
	}

	if r.Method == "GET" && r.URL.Path == "/images" {
		h.ServePage(w, r, "images", map[string]interface{}{"title": "Všechny obrázky", "images": h.images.ImgInfos})
	}

	if r.Method == "GET" && r.URL.Path == "/uploaded" {
		url := r.URL.Query().Get("path")
		h.ServePage(w, r, "uploaded", map[string]string{"title": "Nahráno", "imgUrl": url})
		return
	}

	if r.Method == "GET" && r.URL.Path == "/" {
		h.ServePage(w, r, "index", map[string]string{"title": "Nahraj fotku"})
		return
	}

	if r.Method == "GET" && r.URL.Path == "/api/image" {
		h.ServeGetImage(w, r)
		return
	}

	h.ServeStatic(w, r)
	return
}

func (h *ServerHandler) ServeUpload(w http.ResponseWriter, r *http.Request) (info *ImgInfo, err error) {

	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return
	}
	boundary := params["boundary"]
	mr := multipart.NewReader(r.Body, boundary)

	for {
		p, err := mr.NextPart()
		if err != nil {
			return nil, err
		}

		name := getFileName(p.FileName())

		fmt.Println(name)

		f, err := os.Create("static/images/" + name)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(f, p)
		if err != nil {
			return nil, err
		}

		info = &ImgInfo{
			Path:     "/images/" + name,
			Uploaded: time.Now(),
		}

		return info, nil
	}

	return nil, errors.New("bad data")
}

func getFileName(str string) string {
	str = strings.Replace(str, "/", "", -1)
	path := randSeq(20) + "-" + str
	return path
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (h *ServerHandler) ServePage(w http.ResponseWriter, r *http.Request, pageName string, data interface{}) {

	w.Header().Set("Content-Type", "text/html")

	err := templates.ExecuteTemplate(w, pageName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Redirect(w http.ResponseWriter, path string) {
	w.Header().Add("Content-type", "text/html")
	w.Header().Add("Location", path)
	w.WriteHeader(301)
}

func (h *ServerHandler) ServeGetImage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GET IMAGE"))
}

func (h *ServerHandler) ServeStatic(w http.ResponseWriter, r *http.Request) {
	f := http.FileServer(http.Dir("static"))
	f.ServeHTTP(w, r)
	return
}
