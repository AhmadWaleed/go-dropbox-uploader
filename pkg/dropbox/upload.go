package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

const (
	// UploadFileSizeLimit of uploading file
	UploadFileSizeLimit = 150 * 1024 * 1024 // 150 MB
	// ChunkedUploadFileSize of single chunk
	ChunkedUploadFileSize = 50 * 1024 * 1024 // 50 MB
)

// UploadOptions for file
type UploadOptions struct {
	Source         string          `json:"-"`
	Destination    string          `json:"path"`
	Mode           string          `json:"mode"` // add, overwrite, update
	AutoRename     bool            `json:"autorename"`
	Mute           bool            `json:"mute"`
	ClientModified string          `json:"client_modified,omitempty"`
	File           *os.File        `json:"-"`
	Ctx            context.Context `json:"-"`
}

// Dropbox uploader
type Dropbox struct {
	*Client
	Options *UploadOptions
}

// Upload file to dropbox size limit 150 mb
func (c *Dropbox) Upload() error {
	_, err := c.Dropbox.content("/files/upload", c.Options, c.Options.File)
	if err != nil {
		return err
	}

	return nil
}

// ChunkedUpload uploads file more then 150 mb of size
func (c *Dropbox) ChunkedUpload() error {
	size, err := c.size()
	if err != nil {
		return err
	}

	chunks, _ := splitChunks(size)

	s, err := c.startSession()
	if err != nil {
		return err
	}

	args := &UploadSessionAppendArg{
		Cursor: s,
		Close:  false,
	}

	for i, chunk := range chunks {
		fmt.Printf("uploading... chunk %d\n", i+1)

		args.Cursor.Offset = chunk.Offset
		if err := c.appendChunk(chunk, *args); err != nil {
			if err = c.finishSession(UploadSessionCursor{
				SessionID: s.SessionID,
				Offset:    uint64(size)}); err != nil {
				return err
			}
			return err
		}
	}

	err = c.finishSession(UploadSessionCursor{SessionID: s.SessionID, Offset: uint64(size)})
	if err != nil {
		return err
	}

	return nil
}

// Chunk contains buffer size and
type Chunk struct {
	BufferSize int
	Offset     uint64
}

func splitChunks(size int) ([]Chunk, int) {
	splits := size / ChunkedUploadFileSize

	chunks := make([]Chunk, splits)

	for i := 0; i < splits; i++ {
		chunks[i].BufferSize = ChunkedUploadFileSize
		chunks[i].Offset = uint64(ChunkedUploadFileSize * i)
	}

	if remainder := size % ChunkedUploadFileSize; remainder != 0 {
		chunks = append(chunks, Chunk{
			BufferSize: remainder,
			Offset:     uint64(splits * ChunkedUploadFileSize),
		})
		splits++
	}

	return chunks, splits
}

func (c *Dropbox) size() (int, error) {
	fileinfo, err := c.Options.File.Stat()
	if err != nil {
		return 0, err
	}

	return int(fileinfo.Size()), nil
}

// UploadSessionCursor  Contains the upload session ID and the offset.
type UploadSessionCursor struct {
	SessionID string `json:"session_id"`
	Offset    uint64 `json:"offset"`
}

// starts session for uploading file in chunks
func (c *Dropbox) startSession() (*UploadSessionCursor, error) {
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

// UploadSessionAppendArg contains the upload session ID and the offset and close status.
type UploadSessionAppendArg struct {
	Cursor *UploadSessionCursor `json:"cursor"`
	Close  bool                 `json:"close"`
}

// appends more data to an upload session.
func (c *Dropbox) appendChunk(chunk Chunk, s UploadSessionAppendArg) error {
	buf := make([]byte, chunk.BufferSize)
	_, err := c.Options.File.ReadAt(buf, int64(chunk.Offset))
	if err != nil {
		return err
	}

	_, err = c.content("/files/upload_session/append_v2", s, bytes.NewReader(buf))
	if err != nil {
		return err
	}

	fmt.Println("bytestream to string: ", string(buf[:chunk.BufferSize]))

	return nil
}

// finishes an upload session and save the uploaded data to the given file path.
func (c *Dropbox) finishSession(s UploadSessionCursor) error {
	opt := struct {
		Cursor UploadSessionCursor `json:"cursor"`
		Commit UploadOptions       `json:"commit"`
	}{
		Cursor: s,
		Commit: *c.Options,
	}

	_, err := c.content("/files/upload_session/finish", opt, nil)
	if err != nil {
		return err
	}

	return nil
}
