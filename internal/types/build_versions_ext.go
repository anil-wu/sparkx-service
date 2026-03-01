package types

type CreateBuildVersionDraftReq struct {
	ProjectId          int64  `json:"projectId"`
	SoftwareManifestId int64  `json:"softwareManifestId"`
	VersionNumber      int64  `json:"versionNumber,optional"`
	Description        string `json:"description,optional"`
	EntryPath          string `json:"entryPath,optional"`
}

type CreateBuildVersionDraftResp struct {
	BuildVersionId       int64  `json:"buildVersionId"`
	ProjectId            int64  `json:"projectId"`
	SoftwareManifestId   int64  `json:"softwareManifestId"`
	VersionNumber        int64  `json:"versionNumber"`
	Description          string `json:"description"`
	PreviewStoragePrefix string `json:"previewStoragePrefix"`
	EntryPath            string `json:"entryPath"`
	CreatedBy            int64  `json:"createdBy"`
	CreatedAt            string `json:"createdAt"`
}

type UpdateBuildVersionReq struct {
	Id                        int64  `path:"id"`
	BuildVersionFileId        int64  `json:"buildVersionFileId,optional"`
	BuildVersionFileVersionId int64  `json:"buildVersionFileVersionId,optional"`
	PreviewStoragePrefix      string `json:"previewStoragePrefix,optional"`
	EntryPath                 string `json:"entryPath,optional"`
}

type UpdateBuildVersionResp struct {
	BuildVersionId            int64  `json:"buildVersionId"`
	PreviewStoragePrefix      string `json:"previewStoragePrefix"`
	EntryPath                 string `json:"entryPath"`
	BuildVersionFileId        int64  `json:"buildVersionFileId"`
	BuildVersionFileVersionId int64  `json:"buildVersionFileVersionId"`
}
