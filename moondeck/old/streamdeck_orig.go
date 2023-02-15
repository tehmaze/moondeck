package moondeck

import (
	"image"

	"github.com/disintegration/gift"
)

const (
	streamDeckOrigName       = "Stream Deck (original)"
	streamDeckOrigProductID  = 0x0060
	streamDeckOrigButtonRows = 3
	streamDeckOrigButtonCols = 5
	streamDeckOrigImageSize  = 8191
)

func streamDeckOrigImageHeaderFunc(bytesLeft, buttonIndex, pageNumber uint) []byte {
	var l uint
	if streamDeckOrigImageSize < bytesLeft {
		l = streamDeckOrigImageSize
	} else {
		l = bytesLeft
	}

	return []byte{
		0x02,
		0x01,
		byte(pageNumber + 1),
		0x00,
		getHeaderElement(l, bytesLeft),
		byte(buttonIndex + 1),
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func init() {
	g := gift.New(
		gift.Resize(72, 72, gift.LanczosResampling),
		gift.Rotate180(),
	)

	streamDeckTypes[streamDeckOrigProductID] = streamDeckType{
		Name:                streamDeckOrigName,
		ProductID:           streamDeckOrigProductID,
		ButtonRows:          streamDeckOrigButtonRows,
		ButtonCols:          streamDeckOrigButtonCols,
		ButtonReadOffset:    4,
		ResetPacket:         resetPacket17(),
		BrightnessPacket:    brightnessPacket17(),
		ImageSize:           [2]int{72, 72},
		ImageFormat:         imageBMPType,
		ImagePayloadPerPage: streamDeckOrigImageSize,
		ImageHeaderFunc:     streamDeckOrigImageHeaderFunc,
		ImageProcessFunc: func(i image.Image) image.Image {
			o := image.NewRGBA(image.Rect(0, 0, 72, 72))
			g.Draw(o, i)
			return o
		},
	}
}
