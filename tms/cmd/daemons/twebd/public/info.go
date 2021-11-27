package public

import "prisma/tms"

type InfoResponse struct {
	ID         string             `json:"id"`
	DatabaseID string             `json:"databaseId"`
	TrackID    string             `json:"trackId"`
	RegistryID string             `json:"registryId"`
	MarkerID   string             `json:"markerId"`
	LookupID   string             `json:"lookupId"`
	Target     *tms.Target        `json:"target"`
	Metadata   *tms.TrackMetadata `json:"metadata"`
	Registry   *Registry          `json:"registry"`
}

type Registry struct {
	Incidents []string `json:"incidents"`
}
