package models

type MetadataModel struct {
	ID    uint
	Parts []MediaPart
}

func (MetadataModel) IsItem() {}

type MediaPart struct {
	FilePath string
}
