package moondeck

const streamDeckMiniProductID = 0x0080

func init() {
	streamDeckTypes[streamDeckMiniProductID] = streamDeckType{
		productID:            streamDeckMiniProductID,
		name:                 "Stream Deck Mini",
		cols:                 3,
		rows:                 2,
		keys:                 6,
		pixels:               80,
		dpi:                  138,
		padding:              16,
		featureReportSize:    17,
		firmwareOffset:       5,
		keyStateOffset:       1,
		translateKey:         translateRightToLeft,
		imagePageSize:        1024,
		imagePageHeaderSize:  16,
		imagePageHeader:      streamDeckRev1PageHeader,
		toImageFormat:        toBMP,
		commandFirmware:      streamDeckRev1Firmware,
		commandReset:         streamDeckRev1Reset,
		commandSetBrightness: streamDeckRev1SetBrightness,
	}
}
