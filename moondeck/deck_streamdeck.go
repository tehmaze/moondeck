package moondeck

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"time"

	"github.com/disintegration/gift"
	"github.com/karalabe/hid"
	"golang.org/x/image/draw"
	"maze.io/moondeck/util"
)

const streamDeckProductID = 0x0060

var (
	streamDeckRev1Firmware      = []byte{0x04}
	streamDeckRev1Reset         = []byte{0x0b, 0x63}
	streamDeckRev1SetBrightness = []byte{0x05, 0x55, 0xd1, 0x01}
	streamDeckRev2Firmware      = []byte{0x05}
	streamDeckRev2Reset         = []byte{0x03, 0x02}
	streamDeckRev2SetBrightness = []byte{0x03, 0x08}
)

func streamDeckRev1PageHeader(pageIndex, keyIndex, payloadLength int, lastPage bool) []byte {
	var lastPageByte byte
	if lastPage {
		lastPageByte = 0x01
	}
	return []byte{
		0x02, 0x01,
		byte(pageIndex + 1), 0x00,
		lastPageByte,
		byte(keyIndex + 1),
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func streamDeckRev2PageHeader(pageIndex, keyIndex, payloadLength int, lastPage bool) []byte {
	var lastPageByte byte
	if lastPage {
		lastPageByte = 0x01
	}
	return []byte{
		0x02, 0x07,
		byte(keyIndex),
		lastPageByte,
		byte(payloadLength),
		byte(payloadLength >> 8),
		byte(pageIndex),
		byte(pageIndex >> 8),
	}
}

func toRGBA(i image.Image) *image.RGBA {
	switch i := i.(type) {
	case *image.RGBA:
		return i
	}
	o := image.NewRGBA(i.Bounds())
	draw.Copy(o, image.Point{}, i, i.Bounds(), draw.Src, nil)
	return o
}

var bmpHeader = []byte{
	0x42, 0x4d, 0xf6, 0x3c, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x36, 0x00, 0x00, 0x00, 0x28, 0x00,
	0x00, 0x00, 0x48, 0x00, 0x00, 0x00, 0x48, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x18, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xc0, 0x3c, 0x00, 0x00, 0xc4, 0x0e,
	0x00, 0x00, 0xc4, 0x0e, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func toBMP(i image.Image) ([]byte, error) {
	var (
		r = i.Bounds()
		b = make([]byte, len(bmpHeader)+r.Dx()*r.Dy())
		s = toRGBA(i)
	)
	copy(b, bmpHeader)

	o := len(bmpHeader)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		// flip image horizontally
		for x := r.Max.X - 1; x >= r.Min.X; x-- {
			c := s.RGBAAt(x, y)
			b[o+0] = c.B
			b[o+1] = c.G
			b[o+2] = c.R
			o += 3
		}
	}

	return b, nil
}

func toJPEG(i image.Image) ([]byte, error) {
	// flip image horizontally and vertically
	var (
		f  = image.NewRGBA(i.Bounds())
		r  = i.Bounds()
		dx = r.Dx()
		dy = r.Dy()
	)
	draw.Copy(f, image.Point{}, i, r, draw.Src, nil)
	for y := 0; y < dy/2; y++ {
		yy := r.Max.Y - y - 1
		for x := 0; x < dx; x++ {
			xx := r.Max.X - x - 1
			c := f.RGBAAt(x, y)
			f.SetRGBA(x, y, f.RGBAAt(xx, yy))
			f.SetRGBA(xx, yy, c)
		}
	}

	var b bytes.Buffer
	if err := jpeg.Encode(&b, f, &jpeg.Options{Quality: 100}); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

type streamDeckType struct {
	productID            uint16
	name                 string
	cols, rows           int
	keys                 int
	pixels               int
	dpi                  int
	padding              int
	featureReportSize    int
	firmwareOffset       int
	keyStateOffset       int
	translateKey         func(index, cols uint8) uint8
	imagePageSize        int
	imagePageHeaderSize  int
	toImageFormat        func(image.Image) ([]byte, error)
	imagePageHeader      func(pageIndex, keyIndex, payloadLength int, lastPage bool) []byte
	commandFirmware      []byte
	commandReset         []byte
	commandSetBrightness []byte
}

var streamDeckTypes = make(map[uint16]streamDeckType)

type streamDeck struct {
	streamDeckType
	dev           *hid.Device
	info          hid.DeviceInfo
	button        []*streamDeckButton
	buttonState   []byte
	buttonPress   []time.Time
	buttonTrigger []time.Time
}

func newStreamDeck(info hid.DeviceInfo, t streamDeckType) *streamDeck {
	d := &streamDeck{
		streamDeckType: t,
		info:           info,
		button:         make([]*streamDeckButton, t.keys),
		buttonState:    make([]byte, t.keys),
		buttonPress:    make([]time.Time, t.keys),
		buttonTrigger:  make([]time.Time, t.keys),
	}
	for i := range d.button {
		d.button[i] = newStreamDeckButton(i, d)
	}
	return d
}

func (d *streamDeck) Name() string {
	return d.name
}

func (d *streamDeck) Path() string {
	return d.info.Path
}

func (d *streamDeck) Size() util.Size {
	return util.Sz(d.cols, d.rows)
}

func (d *streamDeck) Open() (err error) {
	d.dev, err = d.info.Open()
	return
}

func (d *streamDeck) Close() (err error) {
	if d.dev != nil {
		err = d.dev.Close()
	}
	return
}

func (d *streamDeck) Reset() error {
	return d.sendFeatureReport(d.commandReset)
}

func (d *streamDeck) Version() string {
	r, err := d.getFeatureReport(d.commandFirmware)
	if err != nil {
		return ""
	}
	return string(r[d.firmwareOffset:])
}

func (d *streamDeck) SerialNumber() string {
	return d.info.Serial
}

func (d *streamDeck) Button(index int) (Button, bool) {
	if index < 0 || index >= len(d.button) {
		return nil, false
	}
	return d.button[index], true
}

func (d *streamDeck) Buttons() int {
	return d.keys
}

func (d *streamDeck) ButtonEvents() <-chan ButtonEvent {
	var (
		c = make(chan ButtonEvent)
		b = make([]byte, d.keyStateOffset+d.keys)
	)
	go func(c chan<- ButtonEvent) {
		defer close(c)

		// Trigger button presses
		go func(c chan<- ButtonEvent) {
			for now := range time.Tick(time.Second / 10) {
				for i, state := range d.buttonState {
					if state == 0 {
						continue
					}
					var (
						first    = d.buttonPress[i]
						firstAgo = now.Sub(first)
						last     = d.buttonTrigger[i]
						lastAgo  = now.Sub(last)
					)
					for _, s := range ButtonTriggerSchedule {
						if firstAgo < s.After {
							continue
						}
						if lastAgo < s.Trigger {
							continue
						}
						d.buttonTrigger[i] = now
						c <- ButtonEvent{
							Button:   d.button[i],
							Pressed:  true,
							Duration: firstAgo,
						}
					}
				}
			}
		}(c)

		// Read events from device
		for {
			copy(d.buttonState, b[d.keyStateOffset:])
			if _, err := d.dev.Read(b); err != nil {
				close(c)
				return
			}

			for i := d.keyStateOffset; i < len(b); i++ {
				j := uint8(i - d.keyStateOffset)
				if d.translateKey != nil {
					j = d.translateKey(j, uint8(d.cols))
				}
				if b[i] != d.buttonState[j] {
					log.Printf("moondeck: stream deck button %d: %x", j, b[i])

					var (
						duration time.Duration
						pressed  = b[i] == 1
					)
					if pressed {
						// Press action immediately triggers a press
						d.buttonPress[j] = time.Now()
						d.buttonTrigger[j] = d.buttonPress[j]
					} else {
						duration = time.Since(d.buttonPress[j])
					}

					c <- ButtonEvent{
						Button:   d.button[j],
						Pressed:  pressed,
						Duration: duration,
					}
				}
			}
		}
	}(c)
	return c
}

func (d *streamDeck) ButtonSize() util.Size {
	return util.Sz(d.pixels, d.pixels)
}

func (d *streamDeck) SetBrightness(v uint8) error {
	if v > 100 {
		v = 100
	}
	return d.sendFeatureReport(append(d.commandSetBrightness, v))
}

func (d *streamDeck) SetColor(c color.Color) error {
	i := image.NewUniform(c)
	for _, b := range d.button {
		if err := b.SetImage(i); err != nil {
			return err
		}
	}
	return nil
}

func (d *streamDeck) SetImage(i image.Image) error {
	var (
		r = i.Bounds()
		w = r.Dx() / d.cols
		h = r.Dy() / d.rows
		o = image.NewRGBA(image.Rect(0, 0, w, h))
		b int
	)
	for y := 0; y < d.rows; y++ {
		for x := 0; x < d.cols; x++ {
			draw.Draw(o, o.Bounds(), i, image.Pt(x*w, y*h), draw.Src)
			if err := d.button[b].SetImage(o); err != nil {
				return err
			}
			b++
		}
	}
	return nil
}

func (d *streamDeck) getFeatureReport(p []byte) ([]byte, error) {
	b := make([]byte, d.featureReportSize)
	copy(b, p)
	if _, err := d.dev.GetFeatureReport(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (d *streamDeck) sendFeatureReport(p []byte) error {
	b := make([]byte, d.featureReportSize)
	copy(b, p)
	_, err := d.dev.SendFeatureReport(b)
	return err
}

type streamDeckButton struct {
	index int
	deck  *streamDeck
}

func newStreamDeckButton(index int, deck *streamDeck) *streamDeckButton {
	return &streamDeckButton{
		index: index,
		deck:  deck,
	}
}

func (b streamDeckButton) Deck() Deck {
	return b.deck
}

func (b streamDeckButton) Index() int {
	return b.index
}

func (b streamDeckButton) Size() util.Size {
	return util.Sz(b.deck.pixels, b.deck.pixels)
}

func (b streamDeckButton) Pos() util.Point {
	x, y := b.index%b.deck.cols, b.index/b.deck.cols
	return util.Pt(x, y)
}

func (b streamDeckButton) SetColor(c color.Color) error {
	return b.SetImage(image.NewUniform(c))
}

func (b streamDeckButton) SetImage(i image.Image) error {
	r := i.Bounds()
	if r.Dx() != b.deck.pixels || r.Dy() != b.deck.pixels {
		// Resize with Lanczos resampling.
		o := image.NewRGBA(image.Rect(0, 0, b.deck.pixels, b.deck.pixels))
		gift.New(gift.Resize(b.deck.pixels, b.deck.pixels, gift.LanczosResampling)).Draw(o, i)
		i = o
	}

	p, err := b.deck.toImageFormat(i)
	if err != nil {
		return err
	}

	var (
		imageData = streamDeckImageData{
			data:     p,
			pageSize: b.deck.imagePageSize - b.deck.imagePageHeaderSize,
		}
		data   = make([]byte, b.deck.imagePageSize)
		last   bool
		header []byte
	)
	for page := 0; !last; page++ {
		var p []byte
		p, last = imageData.Page(page)
		header = b.deck.imagePageHeader(page, b.index, len(p), last)

		copy(data, header)
		copy(data[len(header):], p)

		if _, err = b.deck.dev.Write(data); err != nil {
			return fmt.Errorf("moondeck: image transfer to button %d failed: %w", b.index, err)
		}
	}
	return nil
}

type streamDeckImageData struct {
	data     []byte
	pageSize int
}

func (d streamDeckImageData) Page(index int) ([]byte, bool) {
	o := index * d.pageSize
	if o >= len(d.data) {
		return nil, true
	}

	l := d.pageLength(index)
	if o+l > len(d.data) {
		l = len(d.data) - o
	}

	return d.data[o : o+l], index == d.pageCount()-1
}

func (d streamDeckImageData) pageLength(index int) int {
	r := len(d.data) - index*d.pageSize
	if r > d.pageSize {
		return d.pageSize
	}
	if r > 0 {
		return r
	}
	return 0
}

func (d streamDeckImageData) pageCount() int {
	c := len(d.data) / d.pageSize
	if len(d.data)%d.pageSize > 0 {
		return c + 1
	}
	return c
}

func translateRightToLeft(index, cols uint8) uint8 {
	keyCol := index % cols
	return (index - keyCol) + (cols + 1) - keyCol
}

func init() {
	streamDeckTypes[streamDeckProductID] = streamDeckType{
		productID:            streamDeckProductID,
		name:                 "Stream Deck",
		cols:                 5,
		rows:                 3,
		keys:                 15,
		pixels:               72,
		dpi:                  124,
		padding:              16,
		featureReportSize:    17,
		firmwareOffset:       5,
		keyStateOffset:       1,
		translateKey:         translateRightToLeft,
		imagePageSize:        7819,
		imagePageHeaderSize:  16,
		imagePageHeader:      streamDeckRev1PageHeader,
		toImageFormat:        toBMP,
		commandFirmware:      streamDeckRev1Firmware,
		commandReset:         streamDeckRev1Reset,
		commandSetBrightness: streamDeckRev1SetBrightness,
	}
}
