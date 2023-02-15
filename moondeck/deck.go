package moondeck

import (
	"errors"
	"image"
	"image/color"

	"github.com/karalabe/hid"
	"maze.io/moondeck/util"
)

var (
	ErrNoDevices = errors.New("moondeck: no supported devices could be found")
)

const (
	usbVendorIDElgato = 0x0fd9
)

type Deck interface {
	// Open the hardware device.
	Open() error

	// Close the hardware device.
	Close() error

	// Reset the device state.
	Reset() error

	// Name of the device.
	Name() string

	// Path of the device.
	Path() string

	// Version of the device.
	Version() string

	// Manufacturer of the device.
	Manufacturer() string

	// ID is the USB vendor and product ID.
	ID() string

	// SerialNumber of the device.
	SerialNumber() string

	// Size is the number of keys rows and columns.
	Size() util.Size

	// Button returns the requested button at index.
	Button(int) (Button, bool)

	// Buttons is the numebr of keys.
	Buttons() int

	// ButtonEvents returns all button press events.
	ButtonEvents() <-chan ButtonEvent

	// ButtonSize is the number of pixels per button.
	ButtonSize() util.Size

	// SetBrightness sets the device brightness (range 0-100).
	SetBrightness(uint8) error

	// SetColor sets all buttons the a uniform color.
	SetColor(color.Color) error

	// SetImage sets all buttons to an image.
	SetImage(image.Image) error
}

// Discover all connected and supported Deck devices.
func Discover() ([]Deck, error) {
	var (
		decks []Deck
		devs  = hid.Enumerate(usbVendorIDElgato, 0)
	)
	for _, info := range devs {
		if t, ok := streamDeckTypes[info.ProductID]; ok {
			decks = append(decks, newStreamDeck(info, t))
		}
	}
	if len(decks) == 0 {
		return nil, ErrNoDevices
	}
	return decks, nil
}

// Open the first supported Deck device.
func Open() (Deck, error) {
	d, err := Discover()
	if err != nil {
		return nil, err
	}
	if err = d[0].Open(); err != nil {
		return nil, err
	}
	return d[0], nil
}
