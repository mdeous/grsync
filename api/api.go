package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

const (
	host           = "http://192.168.0.1"
	propertiesPath = "/v1/props"
	photosPath     = "/v1/photos"
)

func readURI(uri string) ([]byte, error) {
	resp, err := http.Get(host + uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to read URI %s: %v", uri, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func GetDeviceInfo() (props *Properties, err error) {
	data, err := readURI(propertiesPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &props)
	if err != nil {
		return nil, err
	}

	return props, nil
}

func GetPhotos() (Photos, error) {
	data, err := readURI(photosPath)
	if err != nil {
		return Photos{}, err
	}
	var photos Photos
	err = json.Unmarshal(data, &photos)
	if err != nil {
		return Photos{}, err
	}
	return photos, nil
}

func DownloadPhoto(photoPath string, destDir string) (destPath string, err error) {
	// Determine destination file and folder
	destPath = path.Join(destDir, photoPath)
	destPhotoDir := path.Dir(destPath)

	// Check if destination file already exists
	if _, err := os.Stat(destPath); os.IsExist((err)) {
		return "", fmt.Errorf("file already exists: %s", destPath)
	}

	// Check if destination folder exists
	if _, err := os.Stat(destPhotoDir); os.IsNotExist(err) {
		err = os.MkdirAll(destPhotoDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create destination folder: %v", err)
		}
	}

	// Download photo
	photoUri := photosPath + photoPath
	photoData, err := readURI(photoUri)
	if err != nil {
		return "", fmt.Errorf("failed to download photo %s: %v", photoPath, err)
	}
	err = os.WriteFile(destPath, photoData, 0750)
	if err != nil {
		return "", fmt.Errorf("failed to write photo file: %v", err)
	}
	return destPath, nil
}

// RESP SAMPLE: /v1/photos
// {
//     "errCode": 200,
//     "errMsg": "OK",
//     "dirs": [
//         {
//             "name": "106_0509",
//             "files": [
//                 "R0000086.JPG",
//                 "R0000086.DNG"
//             ]
//         }
//     ]
// }
// END SAMPLE

// RESP SAMPLE: /v1/props
// {
//     "errCode": 200,
//     "errMsg": "OK",
//     "manufacturer": "RICOH IMAGING COMPANY, LTD.",
//     "model": "RICOH GR IIIx",
//     "serialNo": "0090776",
//     "firmwareVersion": "1.42",
//     "macAddress": "FC:84:A7:45:18:FD",
//     "channelList": [
//         1,
//         2,
//         3,
//         4,
//         5,
//         6,
//         7,
//         8,
//         9,
//         10,
//         11
//     ],
//     "hot": false,
//     "battery": 100,
//     "operationModeList": [
//         "capture",
//         "playback"
//     ],
//     "stillFormatList": [
//         "jpeg",
//         "dng",
//         "rawdng"
//     ],
//     "storages": [
//         {
//             "name": "in",
//             "equipped": true,
//             "available": false,
//             "writable": false,
//             "format": "rawdng",
//             "remain": 0,
//             "recordableTime": 0,
//             "numOfPhotos": 0,
//             "numOfMovies": 0
//         },
//         {
//             "name": "sd1",
//             "equipped": true,
//             "available": true,
//             "writable": true,
//             "format": "rawdng",
//             "remain": 1852,
//             "recordableTime": 11646,
//             "numOfPhotos": 59,
//             "numOfMovies": 0
//         }
//     ],
//     "datetime": "2025-05-09T23:33:42",
//     "operationMode": "capture",
//     "stillFormat": "rawdng",
//     "geoTagging": "on",
//     "gpsInfo": "48.8924177,2.2849635,87.0999984741211,2025-05-09T12:57:55,WGS84",
//     "bdName": "GR_48AB93",
//     "bleEnableCondition": "enablePowerOn",
//     "channel": 6,
//     "ssid": "GR_4518FD",
//     "key": "QX_U$XO8",
//     "powerOffTransfer": "off",
//     "autoResize": "off",
//     "focusSettingList": [
//         "manual",
//         "snap",
//         "inf",
//         "spot",
//         "multiauto",
//         "pinpoint",
//         "tracking",
//         "continuous",
//         "multiauto_center",
//         "zone_select"
//     ],
//     "focused": false,
//     "focusEffectiveArea": [
//         68,
//         64
//     ],
//     "focusMode": "af",
//     "focusSetting": "tracking",
//     "stillSizeList": [
//         "L3",
//         "M3",
//         "S3",
//         "XS3"
//     ],
//     "movieSizeList": [
//         "FHD24p",
//         "FHD30p",
//         "FHD60p"
//     ],
//     "captureModeList": [
//         "still",
//         "movie"
//     ],
//     "shootModeList": [
//         "single",
//         "self2s",
//         "self10s",
//         "continuous",
//         "interval",
//         "intervalSelf2s",
//         "intervalSelf10s",
//         "intervalComp",
//         "intervalCompSelf2s",
//         "intervalCompSelf10s",
//         "multiExp",
//         "multiExpSelf2s",
//         "multiExpSelf10s",
//         "bracket",
//         "bracketSelf2s",
//         "bracketSelf10s"
//     ],
//     "exposureModeList": [
//         "U1"
//     ],
//     "meteringModeList": [
//         "multi",
//         "center",
//         "spot",
//         "highlight"
//     ],
//     "WBModeList": [
//         "auto",
//         "multiAuto",
//         "daylight",
//         "shade",
//         "cloud",
//         "daylightFluorescent",
//         "dayWhiteFluorescent",
//         "coolWhiteFluorescent",
//         "warmWhiteFluorescent",
//         "tungsten",
//         "cte",
//         "colorTemp1",
//         "manual1",
//         "custom1",
//         "custom2",
//         "custom3"
//     ],
//     "effectList": [
//         "off",
//         "col_vivid",
//         "efc_monochrome",
//         "efc_softMonochrome",
//         "efc_hardMonochrome",
//         "efc_highContrast",
//         "efc_posiFilm",
//         "efc_bleachBypass",
//         "efc_retro",
//         "efc_HDRTone",
//         "col_custom1",
//         "col_custom2"
//     ],
//     "resoList": [
//         "720x480",
//         "1080x720"
//     ],
//     "capturing": false,
//     "stateStill": "idle",
//     "stateMovie": "idle",
//     "countDown": "idle",
//     "shotsTotal": -1,
//     "shotsCurrent": 1,
//     "avList": [],
//     "tvList": [],
//     "svList": [
//         "100",
//         "125",
//         "160",
//         "200",
//         "250",
//         "320",
//         "400",
//         "500",
//         "640",
//         "800",
//         "1000",
//         "1250",
//         "1600",
//         "2000",
//         "2500",
//         "3200",
//         "4000",
//         "5000",
//         "6400",
//         "8000",
//         "10000",
//         "12800",
//         "16000",
//         "20000",
//         "25600",
//         "32000",
//         "40000",
//         "51200",
//         "64000",
//         "80000",
//         "102400",
//         "auto"
//     ],
//     "xvList": [
//         "-5.0",
//         "-4.7",
//         "-4.3",
//         "-4.0",
//         "-3.7",
//         "-3.3",
//         "-3.0",
//         "-2.7",
//         "-2.3",
//         "-2.0",
//         "-1.7",
//         "-1.3",
//         "-1.0",
//         "-0.7",
//         "-0.3",
//         "0.0",
//         "+0.3",
//         "+0.7",
//         "+1.0",
//         "+1.3",
//         "+1.7",
//         "+2.0",
//         "+2.3",
//         "+2.7",
//         "+3.0",
//         "+3.3",
//         "+3.7",
//         "+4.0",
//         "+4.3",
//         "+4.7",
//         "+5.0"
//     ],
//     "stillSize": "L3",
//     "movieSize": "FHD60p",
//     "captureMode": "still",
//     "shootMode": "multiExp",
//     "onePushBracket": false,
//     "WBMode": "multiAuto",
//     "exposureMode": "U1",
//     "meteringMode": "multi",
//     "av": "2.8",
//     "tv": "1.40",
//     "sv": "1600",
//     "xv": "0.0",
//     "effect": "off",
//     "cameraOrientation": "positive",
//     "liveState": "busy"
// }
