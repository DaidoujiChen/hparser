package hparser

// GMetaDatas 把 api 回傳的內容轉為 struct
type GMetaDatas struct {
	Items []GMetaData `json:"gmetadata"`
}

// GMetaData 每個 gallery 分別的內容
type GMetaData struct {
	GID          int      `json:"gid"`
	Token        string   `json:"token"`
	ArchiverKey  string   `json:"archiver_key"`
	Title        string   `json:"title"`
	TitleJpn     string   `json:"title_jpn"`
	Category     string   `json:"category"`
	Thumb        string   `json:"thumb"`
	Uploader     string   `json:"uploader"`
	Posted       string   `json:"posted"`
	Filecount    string   `json:"filecount"`
	Filesize     int      `json:"filesize"`
	Expunged     bool     `json:"expunged"`
	Rating       string   `json:"rating"`
	Torrentcount string   `json:"torrentcount"`
	Tags         []string `json:"tags"`
}
