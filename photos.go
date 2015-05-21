package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type ImgInfo struct {
	Path     string    `json:"path"`
	Uploaded time.Time `json:"uploaded"`
}

func NewImages() *Images {
	ret := &Images{}
	ret.ImgInfos = []*ImgInfo{}
	return ret
}

type Images struct {
	ImgInfos []*ImgInfo
	mut      sync.Mutex
}

func (i *Images) Filter(term string) *Images {
	ret := NewImages()
	for _, v := range i.ImgInfos {
		if strings.Contains(v.Path, term) {
			ret.Add(v.Path, v.Uploaded)
		}
	}
	return ret
}

func (i *Images) Print() {
	fmt.Println("PRINTING")
	for _, v := range i.ImgInfos {
		fmt.Println(v.Path)
	}
}

func (i *Images) Add(path string, t time.Time) {
	i.mut.Lock()
	i.ImgInfos = append([]*ImgInfo{&ImgInfo{path, t}}, i.ImgInfos...)
	i.mut.Unlock()
}

func (i *Images) GetLast() *ImgInfo {
	return i.ImgInfos[0]
}

func (i *Images) GetSmart() *ImgInfo {
	ret := i.GetLast()

	if ret.Uploaded.Add(2*time.Minute).Unix() < time.Now().Unix() {
		ret = i.GetRandom()
	}
	return ret
}

func (i *Images) GetFriends() *ImgInfo {
	newI := i.Filter("images/")
	return newI.GetRandom()
}

func (i *Images) GetChildren() *ImgInfo {
	newI := i.Filter("br/")
	return newI.GetRandom()
}

func (i *Images) GetRandom() *ImgInfo {
	rand.Seed(time.Now().UTC().UnixNano())
	len := len(i.ImgInfos)

	index := rand.Intn(len)
	item := i.ImgInfos[index]
	return item
}

func LoadPhotos(folderName string, images *Images) error {
	files, err := ioutil.ReadDir("static/" + folderName)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.Name()[0:1] != "." {
			path := "/" + folderName + "/" + f.Name()
			images.Add(path, f.ModTime())
		}
	}
	return nil
}
