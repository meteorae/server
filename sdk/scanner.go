package sdk

type FileScanner interface {
	Scan(path string, files, dirs *[]string, mediaList *[]Item, extensions []string, root string)
	Process(path string, files, dirs *[]string, mediaList *[]Item, extensions []string, root string)
}
