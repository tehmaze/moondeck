package moondeck

const streamDeckXLProductID = 0x006c

func init() {
	streamDeckTypes[streamDeckXLProductID] = streamDeckType{
		productID:            streamDeckXLProductID,
		name:                 "Stream Deck XL",
		cols:                 8,
		rows:                 4,
		keys:                 32,
		pixels:               96,
		dpi:                  166,
		padding:              16,
		featureReportSize:    32,
		firmwareOffset:       6,
		keyStateOffset:       4,
		imagePageSize:        1024,
		imagePageHeaderSize:  8,
		imagePageHeader:      streamDeckRev2PageHeader,
		toImageFormat:        toJPEG,
		commandFirmware:      streamDeckRev2Firmware,
		commandReset:         streamDeckRev2Reset,
		commandSetBrightness: streamDeckRev2SetBrightness,
	}
}
