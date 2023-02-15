package moondeck

import (
	"image"

	"github.com/disintegration/gift"
)

const (
	streamDeckOrigMK2Name       = "Stream Deck MK.2"
	streamDeckOrigMK2ProductID  = 0x0080
	streamDeckOrigMK2ButtonRows = 3
	streamDeckOrigMK2ButtonCols = 5
	streamDeckOrigMK2ImageSize  = 1024
)

func streamDeckOrigMK2ImageHeaderFunc(bytesLeft, buttonIndex, pageNumber uint) []byte {
	var l uint
	if streamDeckOrigMK2ImageSize < bytesLeft {
		l = streamDeckOrigMK2ImageSize
	} else {
		l = bytesLeft
	}

	return []byte{
		0x02,
		0x07,
		byte(buttonIndex),
		getHeaderElement(l, bytesLeft),
		byte(l & 0xff),
		byte(l >> 8),
		byte(pageNumber & 0xff),
		byte(pageNumber >> 8),
	}
}

func init() {
	g := gift.New(
		gift.Resize(72, 72, gift.LanczosResampling),
		gift.Rotate180(),
	)

	streamDeckTypes[streamDeckOrigMK2ProductID] = streamDeckType{
		Name:                streamDeckOrigMK2Name,
		ProductID:           streamDeckOrigMK2ProductID,
		ButtonRows:          streamDeckOrigMK2ButtonRows,
		ButtonCols:          streamDeckOrigMK2ButtonCols,
		ButtonReadOffset:    4,
		ResetPacket:         resetPacket32(),
		BrightnessPacket:    brightnessPacket32(),
		ImageSize:           [2]int{72, 72},
		ImageFormat:         imageJPEGType,
		ImagePayloadPerPage: streamDeckOrigMK2ImageSize,
		ImageHeaderFunc:     streamDeckOrigMK2ImageHeaderFunc,
		ImageProcessFunc: func(i image.Image) image.Image {
			o := image.NewRGBA(image.Rect(0, 0, 72, 72))
			g.Draw(o, i)
			return o
		},
	}
}
