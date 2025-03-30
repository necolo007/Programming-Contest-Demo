package file_dto

import "mime/multipart"

type UploadFileRequest struct {
	File     *multipart.FileHeader `form:"file" binding:"required"`
	Category string                `form:"category" binding:"required,oneof=pdf txt word"`
	Public   int                   `form:"public" binding:"required,oneof=0 1"`
}
