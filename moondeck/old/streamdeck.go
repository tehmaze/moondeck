package moondeck

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"

	"github.com/karalabe/hid"
	"golang.org/x/image/bmp"
)

const (
	elgatoVendorID = 0x0fd9
	imageBMPType   = "image/bmp"
	imageJPEGType  = "image/jpeg"
)

type StreamDeck struct {
	dev                  *hid.Device
	deviceType           streamDeckType
	button               map[int]Button
	buttonPressListeners []func(int, *StreamDeck, error)
}

func NewStreamDeck() (*StreamDeck, error) {
	devices := hid.Enumerate(elgatoVendorID, 0)
	if len(devices) == 0 {
		return nil, errors.New("moondeck: no compatible devices found")
	}

	var (
		sd  = new(StreamDeck)
		err error
	)
	for _, dev := range devices {
		if t, ok := streamDeckTypes[dev.ProductID]; ok {
			sd.deviceType = t
			if sd.dev, err = dev.Open(); err != nil {
				return nil, fmt.Errorf("moondeck: failed to open %s: %w", t.Name, err)
			}
			sd.button = make(map[int]Button, t.ButtonRows*t.ButtonCols)
			for i := 0; i < t.ButtonRows*t.ButtonCols; i++ {
				sd.button[i] = &streamDeckButton{deck: sd, index: i}
			}
			sd.Reset()
			go sd.listener()
			return sd, nil
		}
	}

	return nil, errors.New("moondeck: no compatible devices found")
}

func (sd *StreamDeck) Close() error {
	return sd.dev.Close()
}

func (sd *StreamDeck) Size() (int, int) {
	return sd.deviceType.ButtonRows, sd.deviceType.ButtonCols
}

func (sd *StreamDeck) Reset() error {
	_, err := sd.dev.SendFeatureReport(sd.deviceType.ResetPacket)
	return err
}

func (sd *StreamDeck) listener() {
	var (
		buttons = sd.deviceType.ButtonRows * sd.deviceType.ButtonCols
		mask    = make([]bool, buttons)
		data    = make([]byte, buttons+int(sd.deviceType.ButtonReadOffset))
	)
	for {
		if _, err := sd.dev.Read(data); err != nil {
			sd.sendButtonPressEvent(-1, err)
			break
		}
		for i := uint(0); i < uint(buttons); i++ {
			if data[sd.deviceType.ButtonReadOffset+i] == 1 {
				if !mask[i] {
					sd.sendButtonPressEvent(int(i), nil)
				}
				mask[i] = true
			} else {
				mask[i] = false
			}
		}
	}
}

func (sd *StreamDeck) sendButtonPressEvent(i int, err error) {
	log.Printf("moondeck: button %d pressed: %v", i, err)
	for _, f := range sd.buttonPressListeners {
		f(i, sd, err)
	}
}

func (sd *StreamDeck) SetBrightness(percent int) error {
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}
	_, err := sd.dev.SendFeatureReport(append(sd.deviceType.BrightnessPacket, byte(percent)))
	return err
}

func (sd *StreamDeck) Button(i int) (Button, bool) {
	b, ok := sd.button[i]
	return b, ok
}

type Button interface {
	Size() image.Point
	SetColor(color.Color) error
	SetImage(image.Image) error
}

type streamDeckButton struct {
	deck  *StreamDeck
	index int
}

func (b *streamDeckButton) Size() image.Point {
	return image.Pt(b.deck.deviceType.ImageSize[0], b.deck.deviceType.ImageSize[1])
}

func (b *streamDeckButton) SetColor(c color.Color) error {
	s := b.Size()
	i := image.NewRGBA(image.Rect(0, 0, s.X, s.Y))
	draw.Draw(i, i.Bounds(), image.NewUniform(c), image.Point{}, draw.Src)
	return b.SetImage(i)
}

func (b *streamDeckButton) SetImage(i image.Image) error {
	// Convert image to the correct format/orientation
	if b.deck.deviceType.ImageProcessFunc != nil {
		i = b.deck.deviceType.ImageProcessFunc(i)
	}

	// Convert image to the correct encoding
	var o bytes.Buffer
	switch b.deck.deviceType.ImageFormat {
	case imageBMPType:
		bmp.Encode(&o, i)
	case imageJPEGType:
		jpeg.Encode(&o, i, &jpeg.Options{Quality: 100})
	default:
		return fmt.Errorf("moondeck: unsupported image format: %q", b.deck.deviceType.ImageFormat)
	}

	return b.setRawImage(o.Bytes())
}

func (b *streamDeckButton) setRawImage(rawImage []byte) error {
	// Based on set_key_image from https://github.com/abcminiuser/python-elgato-streamdeck/blob/master/src/StreamDeck/Devices/StreamDeckXL.py#L151

	pageNumber := 0
	bytesRemaining := len(rawImage)
	halfImage := len(rawImage) / 2
	bytesSent := 0

	for bytesRemaining > 0 {

		header := b.deck.deviceType.ImageHeaderFunc(uint(bytesRemaining), uint(b.index), uint(pageNumber))
		imageReportLength := int(b.deck.deviceType.ImagePayloadPerPage)
		imageReportHeaderLength := len(header)
		imageReportPayloadLength := imageReportLength - imageReportHeaderLength

		/*
			if halfImage > imageReportPayloadLength {
				log.Fatalf("image too large: %d", halfImage*2)
			}
		*/

		thisLength := 0
		if imageReportPayloadLength < bytesRemaining {
			if b.deck.deviceType.Name == "Stream Deck Original" {
				thisLength = halfImage
			} else {
				thisLength = imageReportPayloadLength
			}
		} else {
			thisLength = bytesRemaining
		}

		payload := append(header, rawImage[bytesSent:(bytesSent+thisLength)]...)
		padding := make([]byte, imageReportLength-len(payload))

		thingToSend := append(payload, padding...)
		_, err := b.deck.dev.Write(thingToSend)
		if err != nil {
			return err
		}

		bytesRemaining = bytesRemaining - thisLength
		pageNumber = pageNumber + 1
		bytesSent = bytesSent + thisLength
	}
	return nil
}

type streamDeckType struct {
	Name                string
	ButtonRows          int
	ButtonCols          int
	ButtonReadOffset    uint
	ProductID           uint16
	ResetPacket         []byte
	BrightnessPacket    []byte
	ImageFormat         string
	ImageSize           [2]int
	ImagePayloadPerPage uint
	ImageHeaderFunc     func(bytesLeft, buttonIndex, pageNumber uint) []byte
	ImageProcessFunc    func(image.Image) image.Image
}

var streamDeckTypes = make(map[uint16]streamDeckType)

// resetPacket17 gives the reset packet for devices which need it to be 17 bytes long
func resetPacket17() []byte {
	pkt := make([]byte, 17)
	pkt[0] = 0x0b
	pkt[1] = 0x63
	return pkt
}

// resetPacket32 gives the reset packet for devices which need it to be 32 bytes long
func resetPacket32() []byte {
	pkt := make([]byte, 32)
	pkt[0] = 0x03
	pkt[1] = 0x02
	return pkt
}

// brightnessPacket17 gives the brightness packet for devices which need it to be 17 bytes long
func brightnessPacket17() []byte {
	pkt := make([]byte, 5)
	pkt[0] = 0x05
	pkt[1] = 0x55
	pkt[2] = 0xaa
	pkt[3] = 0xd1
	pkt[4] = 0x01
	return pkt
}

// brightnessPacket32 gives the brightness packet for devices which need it to be 32 bytes long
func brightnessPacket32() []byte {
	pkt := make([]byte, 2)
	pkt[0] = 0x03
	pkt[1] = 0x08
	return pkt
}
