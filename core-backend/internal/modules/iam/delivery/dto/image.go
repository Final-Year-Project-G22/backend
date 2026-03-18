package dto

import "github.com/danielgtaylor/huma/v2"

type UploadAvatarFormData struct {
	File huma.FormFile `form:"file"`
}

type UploadAvatarInput struct {
	RawBody huma.MultipartFormFiles[UploadAvatarFormData]
}

type UploadAvatarResponse struct {
	ImageURL string `json:"imageUrl"`
}

type UploadAvatarOutput struct {
	Body UploadAvatarResponse
}
