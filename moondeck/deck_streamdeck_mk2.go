package moondeck

const streamDeckMK2ProductID = 0x0080

func init() {
	streamDeckTypes[streamDeckMK2ProductID] = streamDeckType{
		productID:            streamDeckMK2ProductID,
		name:                 "Stream Deck MK.2",
		cols:                 5,
		rows:                 3,
		keys:                 15,
		pixels:               72,
		dpi:                  124,
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
