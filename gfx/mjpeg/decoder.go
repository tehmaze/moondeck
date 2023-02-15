package mjpeg

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
)

// Client is our HTTP client.
var Client = http.DefaultClient

// Decoder for motion JPEG (MJPEG).
type Decoder struct {
	r *multipart.Reader
	c io.Closer
}

// NewDecoder return new instance of Decoder
func NewDecoder(r io.ReadCloser, b string) *Decoder {
	d := new(Decoder)
	d.c = r
	d.r = multipart.NewReader(r, b)
	return d
}

// NewDecoderFromURL returns a deco
func NewDecoderFromURL(u string) (*Decoder, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	return NewDecoderFromResponse(res)
}

// NewDecoderFromResponse return new instance of Decoder from http.Response.
func NewDecoderFromResponse(res *http.Response) (*Decoder, error) {
	_, param, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return NewDecoder(res.Body, strings.Trim(param["boundary"], "-")), nil
}

func (d *Decoder) Close() (err error) {
	if d.c != nil {
		err = d.c.Close()
		d.c = nil
		d.r = nil
	}
	return
}

// NextFrame do decoding raw bytes
func (d *Decoder) NextFrame() ([]byte, error) {
	p, err := d.r.NextPart()
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, p)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *Decoder) Next() (image.Image, error) {
	if d == nil || d.r == nil {
		return nil, io.ErrUnexpectedEOF
	}
	p, err := d.r.NextPart()
	if err != nil {
		return nil, err
	}
	return jpeg.Decode(p)
}
