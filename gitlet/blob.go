package gitlet

import (
	"encoding/json"
	"gitlet/config"
	"gitlet/utils"
	"log"
	"path/filepath"
)

type Blob struct {
	Filename string `json:"filename"`
	FilePath string `json:"filepath"`
	Contents []byte `json:"contents"`
	HashId   string `json:"hashId"`
}

func NewBlob(filePath string, contents []byte) *Blob {
	return &Blob{
		Filename: filepath.Base(filePath),
		FilePath: filePath,
		Contents: contents,
		HashId:   utils.GenerateID(contents),
	}
}

func (b *Blob) Persist() {
	data, err := json.Marshal(b)
	if err != nil {
		log.Fatal(err)
	}
	utils.WriteFileBytes(config.BLOB+"/"+b.HashId, data)
}

func GetBlobById(id string) *Blob {
	path := config.BLOB + "/" + id
	if !utils.FileExists(path) {
		return nil
	}
	data := utils.ReadFile(path)
	b := &Blob{}
	err := json.Unmarshal(data, b)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
