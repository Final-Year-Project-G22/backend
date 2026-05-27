package dto

import "github.com/danielgtaylor/huma/v2"

type UploadBusinessImageFormData struct {
	File huma.FormFile `form:"file"`
}

type UploadBusinessImageInput struct {
	RawBody huma.MultipartFormFiles[UploadBusinessImageFormData]
}

type UploadBusinessImageResponse struct {
	ImageURL string `json:"imageUrl"`
}

type UploadBusinessImageOutput struct {
	Body UploadBusinessImageResponse
}
