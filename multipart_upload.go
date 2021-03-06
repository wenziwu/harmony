package harmony

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

type multipartPayload interface {
	json() ([]byte, error)
}

// multipartFromFiles generate a multipart body given a payload and some files.
// It returns the raw generated body along the content type of this body.
func multipartFromFiles(payload multipartPayload, files ...File) ([]byte, string, error) {
	// Underlying buffer the multipart body will be written to.
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Send the endpoint parameters as JSON in a the "payload_json" part.
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", `form-data; name="payload_json"`)
	h.Set("Content-Type", "application/json")
	pw, err := w.CreatePart(h)
	if err != nil {
		return nil, "", err
	}

	b, err := payload.json()
	if err != nil {
		return nil, "", err
	}
	if _, err = pw.Write(b); err != nil {
		return nil, "", err
	}

	// Create a new part for each file.
	for i, f := range files {
		cd := fmt.Sprintf(`form-data; name="file%d"; filename="%s"`, i, f.name)

		h = textproto.MIMEHeader{}
		h.Set("Content-Disposition", cd)
		h.Set("Content-Type", "application/octet-stream")

		pw, err = w.CreatePart(h)
		if err != nil {
			return nil, "", err
		}

		if _, err = io.Copy(pw, f.reader); err != nil {
			return nil, "", err
		}

		if err = f.reader.Close(); err != nil {
			return nil, "", err
		}
	}

	if err = w.Close(); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), w.FormDataContentType(), nil
}
