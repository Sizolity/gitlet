package gitlet

import (
	"encoding/json"
	"gitlet/config"
	"gitlet/utils"
	"log"
	"sort"
	"strings"
)

type TreeEntry struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	HashId string `json:"hashId"`
}

type Tree struct {
	Entries []TreeEntry `json:"entries"`
	HashId  string      `json:"hashId"`
}

func NewTree(entries []TreeEntry) *Tree {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})
	data, _ := json.Marshal(entries)
	return &Tree{
		Entries: entries,
		HashId:  utils.GenerateID(data),
	}
}

func (t *Tree) Persist() {
	data, err := json.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	utils.WriteFileBytes(config.TREE+"/"+t.HashId, data)
}

func GetTreeById(id string) *Tree {
	path := config.TREE + "/" + id
	if !utils.FileExists(path) {
		return nil
	}
	data := utils.ReadFile(path)
	t := &Tree{}
	if err := json.Unmarshal(data, t); err != nil {
		log.Fatal(err)
	}
	return t
}

// BuildTree converts a flat filepath->blobId map into a tree hierarchy,
// persists all intermediate trees, and returns the root tree's HashId.
func BuildTree(flatIndex map[string]string) string {
	return buildTreeRecursive(flatIndex)
}

func buildTreeRecursive(entries map[string]string) string {
	blobs := make(map[string]string)
	subdirs := make(map[string]map[string]string)

	for path, blobId := range entries {
		parts := strings.SplitN(path, "/", 2)
		if len(parts) == 1 {
			blobs[parts[0]] = blobId
		} else {
			dir := parts[0]
			rest := parts[1]
			if subdirs[dir] == nil {
				subdirs[dir] = make(map[string]string)
			}
			subdirs[dir][rest] = blobId
		}
	}

	var treeEntries []TreeEntry

	for name, blobId := range blobs {
		treeEntries = append(treeEntries, TreeEntry{
			Name:   name,
			Type:   "blob",
			HashId: blobId,
		})
	}

	for dir, subEntries := range subdirs {
		subTreeId := buildTreeRecursive(subEntries)
		treeEntries = append(treeEntries, TreeEntry{
			Name:   dir,
			Type:   "tree",
			HashId: subTreeId,
		})
	}

	tree := NewTree(treeEntries)
	tree.Persist()
	return tree.HashId
}

// FlattenTree recursively resolves a tree into a flat filepath->blobId map.
func FlattenTree(treeId string) map[string]string {
	result := make(map[string]string)
	if treeId == "" {
		return result
	}
	flattenRecursive(treeId, "", result)
	return result
}

func flattenRecursive(treeId string, prefix string, result map[string]string) {
	tree := GetTreeById(treeId)
	if tree == nil {
		return
	}
	for _, entry := range tree.Entries {
		path := entry.Name
		if prefix != "" {
			path = prefix + "/" + entry.Name
		}
		if entry.Type == "blob" {
			result[path] = entry.HashId
		} else {
			flattenRecursive(entry.HashId, path, result)
		}
	}
}
