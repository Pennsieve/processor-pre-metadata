package instance

import (
	"encoding/json"
	"time"
)

type Proxy struct {
	ProxyID
	ProxyPackage
}
type ProxyID struct {
	ID string `json:"id"`
}

type ProxyPackageContent struct {
	CreatedAt     time.Time `json:"createdAt"`
	DatasetId     string    `json:"datasetId"`
	DatasetNodeId string    `json:"datasetNodeId"`
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	NodeID        string    `json:"nodeId"`
	OwnerID       int       `json:"ownerId"`
	PackageType   string    `json:"packageType"`
	State         string    `json:"state"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ProxyPackage should also have "children" and "properties" slices, but we don't
// need them for now
type ProxyPackage struct {
	Content ProxyPackageContent `json:"content"`
}

// RawFromFile represents the structure of proxy instances as they appear in the downloaded files.
// The first json.RawMessage is a ProxyID and the second is a ProxyPackage
// A downloaded file will contain a []RawFromFile
type RawFromFile [2]json.RawMessage
