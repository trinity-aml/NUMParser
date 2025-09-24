package ml

type Collection struct {
	Title    string
	Overview string
	Prompt   string
	Image    string `json:",omitempty"`
}
