package types

type CreateReleaseReq struct {
	Id                           int64  `json:"id,optional"`
	ProjectId                    int64  `json:"projectId"`
	BuildVersionId               int64  `json:"buildVersionId"`
	ReleaseManifestFileId        int64  `json:"releaseManifestFileId"`
	ReleaseManifestFileVersionId int64  `json:"releaseManifestFileVersionId"`
	Name                         string `json:"name"`
	Channel                      string `json:"channel"`
	Platform                     string `json:"platform"`
	Status                       string `json:"status,optional"`
	VersionTag                   string `json:"versionTag,optional"`
	ReleaseNotes                 string `json:"releaseNotes,optional"`
	PublishedAt                  string `json:"publishedAt,optional"`
}

type CreateReleaseResp struct {
	ReleaseId                    int64  `json:"releaseId"`
	ProjectId                    int64  `json:"projectId"`
	BuildVersionId               int64  `json:"buildVersionId"`
	ReleaseManifestFileId        int64  `json:"releaseManifestFileId"`
	ReleaseManifestFileVersionId int64  `json:"releaseManifestFileVersionId"`
	Name                         string `json:"name"`
	Channel                      string `json:"channel"`
	Platform                     string `json:"platform"`
	Status                       string `json:"status"`
	VersionTag                   string `json:"versionTag"`
	ReleaseNotes                 string `json:"releaseNotes"`
	CreatedBy                    int64  `json:"createdBy"`
	CreatedAt                    string `json:"createdAt"`
	PublishedAt                  string `json:"publishedAt"`
}
