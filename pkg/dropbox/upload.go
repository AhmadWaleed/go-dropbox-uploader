package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ChunkedUploadFileSize of single chunk to upload to dropbox
// const ChunkedUploadFileSize = 100 * 1024 * 1024
const ChunkedUploadFileSize = 1

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
	Options *UploadOptions
}

// Upload file to dropbox size limit 150 mb
func (c *Dropbox) Upload() error {
	file, err := os.Open(c.Options.Source)
	if err != nil {
		return fmt.Errorf("could not open file %v", err)
	}

	c.Options.Reader = file
	c.Options.AutoRename = true

	_, err = c.Dropbox.content("/files/upload", c.Options, file)
	if err != nil {
		return fmt.Errorf("could not upload file %v", err)
	}

	return nil
}

// ChunkedUpload uploads file more then 150 mb of size
func (c *Dropbox) ChunkedUpload() error {
	file, err := os.Open(c.Options.Source)
	if err != nil {
		return err
	}

	defer file.Close()
	c.Options.AutoRename = true

	fileinfo, err := file.Stat()
	if err != nil {
		return err
	}

	filesize := int(fileinfo.Size())

	concurrency := filesize / ChunkedUploadFileSize

	chunksizes := make([]Chunk, concurrency)

	for i := 0; i < concurrency; i++ {
		chunksizes[i].BufferSize = ChunkedUploadFileSize
		chunksizes[i].Offset = int64(ChunkedUploadFileSize * i)
	}

	if remainder := filesize % ChunkedUploadFileSize; remainder != 0 {
		c := Chunk{BufferSize: remainder, Offset: int64(concurrency * ChunkedUploadFileSize)}
		concurrency++
		chunksizes = append(chunksizes, c)
	}

	s, err := c.StartSession()
	if err != nil {
		return err
	}

	sessionArgs := &UploadSessionAppendArg{
		Cursor: *s,
		Close:  false,
	}

	for itr, chunk := range chunksizes {

		if itr == len(chunksizes)-1 {
			sessionArgs.Close = true
		}

		sessionArgs.Cursor.Offset = chunk.Offset

		if err := c.AppendChunk(chunk, *sessionArgs, file); err != nil {
			if err = c.FinishSession(*s); err != nil {
				return err
			}
			break
		}
	}

	err = c.FinishSession(*s)
	if err != nil {
		return err
	}

	return nil
}

// UploadSessionCursor  Contains the upload session ID and the offset.
type UploadSessionCursor struct {
	SessionID string `json:"session_id"`
	Offset    int64  `json:"offset"`
}

// StartSession starts session for uploading file in chunks
func (c *Dropbox) StartSession() (*UploadSessionCursor, error) {
	opt := struct {
		Close bool `json:"close"`
	}{Close: false}

	res, err := c.content("/files/upload_session/start", opt, nil)
	if err != nil {
		return nil, err
	}

	var data UploadSessionCursor
	err = json.NewDecoder(res).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("could not decode session response %v", err)
	}

	return &data, nil
}

// Chunk contains buffer size and
type Chunk struct {
	BufferSize int
	Offset     int64
}

// UploadSessionAppendArg contains the upload session ID and the offset and close status.
type UploadSessionAppendArg struct {
	Cursor UploadSessionCursor `json:"cursor"`
	Close  bool                `json:"close"`
}

// AppendChunk appends more data to an upload session.
func (c *Dropbox) AppendChunk(chunk Chunk, s UploadSessionAppendArg, r io.ReaderAt) error {
	buffer := make([]byte, chunk.BufferSize)
	_, err := r.ReadAt(buffer, chunk.Offset)
	if err != nil {
		return err
	}

	_, err = c.content("/files/upload_session/append_v2", s, bytes.NewReader(buffer))
	if err != nil {
		return err
	}

	return nil
}

// FinishSession finishes an upload session
// and save the uploaded data to the given file path.
func (c *Dropbox) FinishSession(s UploadSessionCursor) error {
	options := struct {
		Cursor UploadSessionCursor `json:"cursor"`
		Commit UploadOptions       `json:"commit"`
	}{
		Cursor: s,
		Commit: *c.Options,
	}

	_, err := c.content("/files/upload_session/finish", options, nil)
	if err != nil {
		return err
	}

	return nil
}
