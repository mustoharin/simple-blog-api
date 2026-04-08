package domain

import "time"

type Image struct {
	ID         string
	Filename   string
	S3Key      string
	URL        string
	UploadedBy string
	CreatedAt  time.Time
}
