package domain

import "time"

type Image struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	S3Key      string    `json:"s3_key"`
	URL        string    `json:"url"`
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}
