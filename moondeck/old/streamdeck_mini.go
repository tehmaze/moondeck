package moondeck

import (
	"image"

	"github.com/disintegration/gift"
)

const (
	streamDeckMiniName       = "Stream Deck Mini"
	streamDeckMiniProductID  = 0x0063
	streamDeckMiniButtonRows = 2
	streamDeckMiniButtonCols = 3
	streamDeckMiniImageSize  = 1024
)

func streamDeckMiniImageHeaderFunc(bytesLeft, buttonIndex, pageNumber uint) []byte {
	var l uint
	if streamDeckMiniImageSize < bytesLeft {
		l = streamDeckMiniImageSize
	} else {
		l = bytesLeft
	}

	return []byte{
		0x02,
		0x01,
		byte(pageNumber),
		0,
		getHeaderElement(l, bytesLeft),
		byte(buttonIndex + 1),
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func getHeaderElement(l, r uint) byte {
	if l == r {
		return 0x01
	}
	return 0x00
}

func init() {
	g := gift.New(
		gift.Resize(80, 80, gift.LanczosResampling),
		gift.Rotate90(),
		gift.FlipVertical(),
	)

	streamDeckTypes[streamDeckMiniProductID] = streamDeckType{
		Name:                streamDeckMiniName,
		ProductID:           streamDeckMiniProductID,
		ButtonRows:          streamDeckMiniButtonRows,
		ButtonCols:          streamDeckMiniButtonCols,
		ButtonReadOffset:    1,
		ResetPacket:         resetPacket17(),
		BrightnessPacket:    brightnessPacket17(),
		ImageFormat:         imageBMPType,
		ImageSize:           [2]int{80, 80},
		ImagePayloadPerPage: streamDeckMiniImageSize,
		ImageHeaderFunc:     streamDeckMiniImageHeaderFunc,
		ImageProcessFunc: func(i image.Image) image.Image {
			o := image.NewRGBA(image.Rect(0, 0, 80, 80))
			g.Draw(o, i)
			return o
		},
	}
}
