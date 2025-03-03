package file_handler

import (
	"Programming-Demo/internal/app/File/file_dto"
	f "Programming-Demo/pkg/utils/file"
	"github.com/gin-gonic/gin"
	"net/http"
)

func FileUpload(c *gin.Context) {
	var req file_dto.UploadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "请求错误"})
		return
	}

	file, err := c.FormFile(req.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文件获取错误"})
		return
	}

	if f.IsValidateFileType(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文件格式错误"})
		return
	}

	if err = c.SaveUploadedFile(file, req.Path); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "文件保存错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "文件保存成功"})
}
