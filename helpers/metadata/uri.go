package metadata

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/adrg/xdg"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/sdk"
)

var validURIRegex = regexp.MustCompile(`^metadata://([a-zA-Z\.]*)_([A-Fa-f0-9]{64})$`)

func GetURIForAgent(agent, hash string) string {
	return fmt.Sprintf("metadata://%s_%s", agent, hash)
}

func IsValidMetadataURI(uri string) bool {
	if uri == "" {
		return false
	}

	if !strings.HasPrefix(uri, "metadata://") {
		return false
	}

	if !validURIRegex.MatchString(uri) {
		return false
	}

	return true
}

// GetURIComponents returns the agent and hash components of a metadata URI.
func GetURIComponents(uri string) (string, string) {
	if !IsValidMetadataURI(uri) {
		return "", ""
	}

	parts := validURIRegex.FindStringSubmatch(uri)

	return parts[1], parts[2]
}

func GetCombinedFilepathForURI(uri string, item database.ItemMetadata, filetype string) string {
	if !IsValidMetadataURI(uri) {
		return ""
	}

	UUID := strings.ReplaceAll(item.UUID.String(), "-", "")
	UUIDPrefix := UUID[:2]

	cleanURI := strings.TrimPrefix(uri, "metadata://")

	metadataDir, err := xdg.DataFile(
		filepath.Join("meteorae", "metadata", item.Type.String(), UUIDPrefix, UUID, "combined", filetype, cleanURI))
	if err != nil {
		return ""
	}

	return metadataDir
}

func GetFilepathForURI(uri string, item database.ItemMetadata, filetype string) string {
	if !IsValidMetadataURI(uri) {
		return ""
	}

	agent, hash := GetURIComponents(uri)

	return GetFilepathForAgentAndHash(agent, hash, item.UUID.String(), item.Type, filetype)
}

func GetFilepathForAgentAndHash(agent, hash, uuid string, itemType sdk.ItemType, filetype string) string {
	// Remove dashes from the UUID.
	uuid = strings.ReplaceAll(uuid, "-", "")
	uuidPrefix := uuid[:2]

	metadataDir, err := xdg.DataFile(
		filepath.Join("meteorae", "metadata", itemType.String(), uuidPrefix, uuid, agent, filetype, hash))
	if err != nil {
		return ""
	}

	return metadataDir
}
