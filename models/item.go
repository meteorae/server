package models

import "github.com/meteorae/meteorae-server/database"

type Item interface {
	IsItem()
}

type MetadataModel struct {
	ID    uint
	Parts []database.MediaPart
}

func (MetadataModel) IsItem() {}

type MediaPart struct {
	FilePath string
}
