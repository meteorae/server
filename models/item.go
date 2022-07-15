package models

import "github.com/meteorae/meteorae-server/database"

type MetadataModel struct {
	ID    uint
	Parts []database.MediaPart
}

func (MetadataModel) IsItem() {}

type MediaPart struct {
	FilePath string
}
