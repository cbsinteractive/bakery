package client

// MessageResponse will hold non channel responses from the requests
type MessageResponse struct {
	Message string                 `json:"message,omitempty"`
	Error   map[string]interface{} `json:"errors,omitempty"`
}

// ChannelOutput is the data structed of Channel.Outputs field
type ChannelOutput struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	PlaybackURL string `json:"playback_url,omitempty"`
	CaptionsURL string `json:"playback_url_auto_captions,omitempty"`
	AdsURL      string `json:"playback_url_ads,omitempty"`
	OriginURL   string `json:"origin_url,omitempty"`
	DRM         bool   `json:"drm,omitempty"`
	DVR         bool   `json:"dvr,omitempty"`
	AudioOnly   bool   `json:"audio_only,omitempty"`
}

// Channel is the data structure returned by the propeller endpoint.
// The defined fields may not represent the complete data structure, since
// we do not use them all. The URL fields point to the location of the master
// program manifests.
type Channel struct {
	// For error messaging
	MessageResponse
	// Org is the orgID
	Org string `json:"organization_id,omitempty"`

	// ID is the channelID
	ID string `json:"id,omitempty"`

	// PlaybackURL is the prefered, cache-friendly manifest endpoint
	PlaybackURL string `json:"playback_url,omitempty"`

	// CaptionsPlaybackURL is the captions enabled endpoint
	CaptionsURL string `json:"playback_url_auto_captions,omitempty"`

	// AdsPlaybackURL is the captions enabled endpoint
	AdsURL string `json:"playback_url_ads,omitempty"`

	// OriginURL is the uncached origin-sourced manifest endpoint
	OriginURL string `json:"origin_url,omitempty"`

	// DRM protected content must be decrypted or downloaded from origin
	DRM bool `json:"drm,omitempty"`

	DVR         bool     `json:"dvr,omitempty"`
	OriginOnly  bool     `json:"origin_only,omitempty"`
	Type        string   `json:"channel_type,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	InputType   string   `json:"input_type,omitempty"`
	Status      string   `json:"status,omitempty"`
	Region      string   `json:"region,omitempty"`
	Ads         bool     `json:"ads,omitempty"`
	Archive     bool     `json:"archive,omitempty"`
	AudioOnly   bool     `json:"audio_only,omitempty"`
	Captions    bool     `json:"auto_captions,omitempty"`
	Latency     bool     `json:"low_latency,omitempty"`
	Tags        []string `json:"tags,omitempty"`

	Outputs []ChannelOutput `json:"outputs,omitempty"`
}

// Fields returns useful fields for logging and tracing
func (c *Channel) Fields() []interface{} {
	return []interface{}{
		"orgID", c.Org,
		"channelID", c.ID,
		"type", c.Type,
		"name", c.Name,
		"description", c.Description,
		"dvr", c.DVR,
		"region", c.Region,
		"status", c.Status,
		"playback_url", c.PlaybackURL,
		"captions_url", c.CaptionsURL,
		"ads_url", c.AdsURL,
		"ads", c.Ads,
		"origin_url", c.OriginURL,
		"created_at", c.CreatedAt,
		"drm", c.DRM,
		"type", c.Type,
		"input_type", c.InputType,
		"archive", c.Archive,
		"origin_only", c.OriginOnly,
		"audio_only", c.AudioOnly,
		"captions", c.Captions,
		"latency", c.Latency,
		"tags", c.Tags,
	}
}

// FindOutput returns the ChannelOutput with the given id, or nil if none exist
func (c *Channel) FindOutput(id string) *ChannelOutput {
	for _, output := range c.Outputs {
		if output.ID == id {
			return &output
		}
	}
	return nil
}
