package dropbox

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// UploadOptions for file
type UploadOptions struct {
	Source         string    `json:"-"`
	Destination    string    `json:"path"`
	Mode           string    `json:"mode"` // add, overwrite, update
	AutoRename     bool      `json:"autorename"`
	Mute           bool      `json:"mute"`
	ClientModified string    `json:"client_modified,omitempty"`
	Reader         io.Reader `json:"-"`
}

// Dropbox uploader
type Dropbox struct {
	*Client
}

// Upload file to dropbox size limit 150 mb
func (c *Dropbox) Upload(opt UploadOptions) error {
	file, err := os.Open(opt.Source)
	if err != nil {
		return fmt.Errorf("could not open file %v", err)
	}

	opt.Reader = file
	opt.AutoRename = true

	res, err := c.Dropbox.upload("/files/upload", opt, file)
	if err != nil {
		return fmt.Errorf("could not upload file %v", err)
	}

	meta := struct {
		Tag            string    `json:".tag"`
		Name           string    `json:"name"`
		PathLower      string    `json:"path_lower"`
		PathDisplay    string    `json:"path_display"`
		ClientModified time.Time `json:"client_modified"`
		ServerModified time.Time `json:"server_modified"`
		Rev            string    `json:"rev"`
		Size           uint64    `json:"size"`
		ID             string    `json:"id"`
		ContentHash    string    `json:"content_hash,omitempty"`
	}{}

	if err := json.NewDecoder(res).Decode(&meta); err != nil {
		return fmt.Errorf("could not parse response %v", err)
	}

	return nil
}

// ChunkedUpload uploads file more then 150 mb of size
func (c *Dropbox) ChunkedUpload(opt UploadOptions) {
	c.StartSession()
}

// StartSession starts session for uploading file in chunks
func (c *Dropbox) StartSession() error {

	opt := struct {
		Close bool `json:"close"`
	}{Close: true}

	res, err := c.call("/files/upload_session/start", opt)
	if err != nil {
		return fmt.Errorf("could not start upload session %v", err)
	}

	var data struct {
		SessionID string `json:"session_id"`
	}

	err = json.NewDecoder(res).Decode(&data)
	if err != nil {
		return fmt.Errorf("could not decode session response %v", err)
	}

	return nil
}

// AppendChunk appends more data to an upload session.
func (c *Dropbox) AppendChunk() error {

	return nil
}

// FinishSession finishes an upload session
// and save the uploaded data to the given file path.
func (c *Dropbox) FinishSession() error {

	return nil
}
