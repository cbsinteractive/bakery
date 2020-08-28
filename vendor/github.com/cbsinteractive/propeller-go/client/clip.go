package client

import (
	"fmt"
	"net/url"
)

// Clip is the data structure that represents
// a clip in propeller
type Clip struct {
	MessageResponse
	Channel           string   `json:"channel_id,omitempty"`
	ID                string   `json:"id,omitempty"`
	Name              string   `json:"name,omitempty"`
	Description       string   `json:"description,omitempty"`
	Start             string   `json:"start,omitempty"`
	End               string   `json:"end,omitempty"`
	Format            string   `json:"format,omitempty"`
	Status            string   `json:"status,omitempty"`
	StatusDescription string   `json:"status_description,omitempty"`
	PlaybackURL       string   `json:"url,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

const (
	//CreatedStatus reflect a successfully created clip
	CreatedStatus = "created"
	//PendingStatus reflects a pending clip creation
	PendingStatus = "pending"
	//ErrorStatus reflects an error when clip was created
	ErrorStatus = "error"
)

// URL returns the url available for playback.
func (c *Clip) URL() (*url.URL, error) {
	if c.Status == ErrorStatus {
		return nil, fmt.Errorf("%v", c.StatusDescription)
	}

	if c.Status == PendingStatus {
		return nil, fmt.Errorf("%v", "Clip not ready")
	}

	return url.Parse(c.PlaybackURL)
}
