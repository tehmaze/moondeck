package moondeck

import (
	"image"

	"github.com/disintegration/gift"
)

const (
	streamDeckXLName       = "Stream Deck XL"
	streamDeckXLProductID  = 0x006c
	streamDeckXLButtonRows = 4
	streamDeckXLButtonCols = 8
	streamDeckXLImageSize  = 1024
)

func streamDeckXLImageHeaderFunc(bytesLeft, buttonIndex, pageNumber uint) []byte {
	var l uint
	if streamDeckXLImageSize < bytesLeft {
		l = streamDeckXLImageSize
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
		gift.Resize(96, 96, gift.LanczosResampling),
		gift.Rotate180(),
	)

	streamDeckTypes[streamDeckXLProductID] = streamDeckType{
		Name:                streamDeckXLName,
		ProductID:           streamDeckXLProductID,
		ButtonRows:          streamDeckXLButtonRows,
		ButtonCols:          streamDeckXLButtonCols,
		ButtonReadOffset:    4,
		ResetPacket:         resetPacket32(),
		BrightnessPacket:    brightnessPacket32(),
		ImageSize:           [2]int{96, 96},
		ImageFormat:         imageJPEGType,
		ImagePayloadPerPage: streamDeckXLImageSize,
		ImageHeaderFunc:     streamDeckXLImageHeaderFunc,
		ImageProcessFunc: func(i image.Image) image.Image {
			o := image.NewRGBA(image.Rect(0, 0, 96, 96))
			g.Draw(o, i)
			return o
		},
	}
}
