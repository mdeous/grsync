package api

type Properties struct {
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmwareVersion"`
	Battery         int    `json:"battery"`
	SerialNumber    string `json:"serialNo"`
}

type PhotosDir struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

type Photos struct {
	Dirs []PhotosDir `json:"dirs"`
}
