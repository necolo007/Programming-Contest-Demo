package file_dto

type UploadReq struct {
	Name string `json:"name" binding:"required"`
	Path string `json:"Path" binding:"required"`
	Type string
}
