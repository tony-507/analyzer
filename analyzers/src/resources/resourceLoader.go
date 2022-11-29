package resources

import (
	"strings"
)

type Resource struct {
	StreamType []string `json:"StreamType"`
}

type ResourceLoader struct {
	resource Resource
}

func (r *ResourceLoader) Query(path string, key interface{}) string {
	sPath := strings.Split(path, "/")
	switch sPath[0] {
	case "streamType":
		typeNum, ok := key.(int)
		if !ok {
			panic("Wrong query format for streamType")
		}
		return r.resource.StreamType[typeNum-1]
	}
	return ""
}

func CreateResourceLoader() ResourceLoader {

	streamType := []string{
		"MPEG 1 video",
		"MPEG 2 video",
		"MPEG 1 audio",
		"MPEG 2 audio",
		"MPEG 2 table data",
		"(AC-3/ DVB subtitle) packetized data for MPEG-2",
		"MHEG",
		"DSM CC",
		"H.222 and ISO/IEC 13818-1'11172-1 auxiliary data",
		"DSM CC multiprotocol encapsulation",
		"DSM CC U-N messages",
		"DSM CC stream descriptors",
		"DSM CC tabled data",
		"13818-1 auxiliary data",
		"ADTS AAC",
		"MPEG-4 H.263 based video",
		"MPEG-4 LOAD multi-format framed audio",
		"MPEG-4 FlexMux in a packetized stream",
		"MPEG-4 FlexMux in ISO/IEC 14496 tables",
		"DSM CC synchronized download protocol",
		"Packetized metadata",
		"Sectioned metadata",
		"DSM CC Data Carousel metadata",
		"DSM CC Object Carousel metadata",
		"Synchronized download protocol metadata",
		"IPMP",
		"H.264 video",
		"MPEG-4 raw audio",
		"MPEG-4 text",
		"MPEG-4 auxiliary video",
		"MPEG-4 AVC",
		"MPEG-4 AVC",
		"JPEG 2000 video",
		"Reserved",
		"Reserved",
		"H.265 UHD video",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"Reserved",
		"IPMP (DRM)",
		"H.262 DigiCipher II",
		"(ATSC and Bluray) AC-3 audio",
		"SCTE subtitle",
		"(Bluray) Dolby TrueHD loseless audio",
		"(Bluray) Dolby Digital Plus audio",
		"(Bluray) DTS 8 channel audio",
		"SCTE-35 DPI data"}

	resource := Resource{StreamType: streamType}

	return ResourceLoader{resource: resource}
}
