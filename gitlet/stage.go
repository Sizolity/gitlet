package gitlet

import (
	"encoding/json"
	"gitlet/config"
	"gitlet/utils"
	"log"
)

type Index struct {
	Entries map[string]string `json:"entries"`
}

func NewIndex() *Index {
	return &Index{
		Entries: make(map[string]string),
	}
}

func LoadIndex() *Index {
	idx := &Index{}
	data := utils.ReadFile(config.INDEX)
	if err := json.Unmarshal(data, idx); err != nil {
		log.Fatal(err)
	}
	if idx.Entries == nil {
		idx.Entries = make(map[string]string)
	}
	return idx
}

func (idx *Index) Save() {
	data, err := json.Marshal(idx)
	if err != nil {
		log.Fatal(err)
	}
	utils.WriteFileBytes(config.INDEX, data)
}

func (idx *Index) Update(filepath string, blobId string) {
	idx.Entries[filepath] = blobId
}

func (idx *Index) Remove(filepath string) {
	delete(idx.Entries, filepath)
}

func (idx *Index) Has(filepath string) bool {
	_, ok := idx.Entries[filepath]
	return ok
}

func (idx *Index) GetBlobId(filepath string) string {
	return idx.Entries[filepath]
}
