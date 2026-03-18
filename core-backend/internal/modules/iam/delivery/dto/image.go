package dto

type UploadAvatarInput struct {
	File []byte `multipart:"file"`
}

type UploadAvatarResponse struct {
	ImageURL string `json:"imageUrl"`
}

type UploadAvatarOutput struct {
	Body UploadAvatarResponse
}
