package template_dto

type CreateTemplatereq struct {
	Type        string `json:"type" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Environment string `json:"environment" binding:"required"`
}
