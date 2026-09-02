package domain

import "time"

type MediaStatus string

const (
	Pending     MediaStatus = "pending"
	Downloading MediaStatus = "downloading"
	Downloaded  MediaStatus = "downloaded"
	Failed      MediaStatus = "failed"
)

type Media struct {
	externalID   string
	status       MediaStatus
	mimeType     string
	url          *string
	downloadedAt *time.Time
	err          *string
}

func (m Media) ExternalID() string       { return m.externalID }
func (m Media) Status() MediaStatus      { return m.status }
func (m Media) MimeType() string         { return m.mimeType }
func (m Media) Url() *string             { return m.url }
func (m Media) DownloadedAt() *time.Time { return m.downloadedAt }
func (m Media) Error() *string           { return m.err }

func NewMedia(
	externalID string,
	status MediaStatus,
	mimeType string,
	url *string,
	downloadedAt *time.Time,
	err *string,
) Media {
	return Media{
		externalID,
		status,
		mimeType,
		url,
		downloadedAt,
		err,
	}
}
