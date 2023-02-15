package moondeck

import (
	"image"

	"github.com/disintegration/gift"
)

const (
	streamDeckOrigV2Name       = "Stream Deck V2"
	streamDeckOrigV2ProductID  = 0x006d
	streamDeckOrigV2ButtonRows = 3
	streamDeckOrigV2ButtonCols = 5
	streamDeckOrigV2ImageSize  = 1024
)

func streamDeckOrigV2ImageHeaderFunc(bytesLeft, buttonIndex, pageNumber uint) []byte {
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

	streamDeckTypes[streamDeckOrigV2ProductID] = streamDeckType{
		Name:                streamDeckOrigV2Name,
		ProductID:           streamDeckOrigV2ProductID,
		ButtonRows:          streamDeckOrigV2ButtonRows,
		ButtonCols:          streamDeckOrigV2ButtonCols,
		ButtonReadOffset:    4,
		ResetPacket:         resetPacket32(),
		BrightnessPacket:    brightnessPacket32(),
		ImageSize:           [2]int{72, 72},
		ImageFormat:         imageJPEGType,
		ImagePayloadPerPage: streamDeckOrigV2ImageSize,
		ImageHeaderFunc:     streamDeckOrigV2ImageHeaderFunc,
		ImageProcessFunc: func(i image.Image) image.Image {
			o := image.NewRGBA(image.Rect(0, 0, 72, 72))
			g.Draw(o, i)
			return o
		},
	}
}
