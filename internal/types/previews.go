package types

type PreviewBuildEntryReq struct {
	BuildVersionId int64 `path:"buildVersionId"`
}

type PreviewBuildAssetReq struct {
	BuildVersionId int64  `path:"buildVersionId"`
	AssetPath      string `path:"assetPath"`
}

type PreviewBuildAssetQueryReq struct {
	BuildVersionId int64  `path:"buildVersionId"`
	Path           string `form:"path"`
}

type PreviewBuildPreuploadReq struct {
	BuildVersionId int64  `path:"buildVersionId"`
	Name           string `json:"name"`
	FileFormat     string `json:"fileFormat"`
	SizeBytes      int64  `json:"sizeBytes"`
	Hash           string `json:"hash"`
	ContentType    string `json:"contentType,optional"`
}

type PreviewBuildPreuploadResp struct {
	UploadUrl   string `json:"uploadUrl"`
	ContentType string `json:"contentType"`
	StorageKey  string `json:"storageKey"`
}
