package bot

// Login information for tmi
type Identity struct {
	Name  string `json:"name"`
	Oauth string `json:"oauth"`
}

// Youtube module options
type Youtube struct {
	APIKey string `json:"api_key"`
}

// Full bot options
type BotOptions struct {
	Identity    Identity `json:"identity"`
	Channel     string   `json:"channel"`
	Youtube     Youtube  `json:"youtube"`
	UIPort      int      `json:"ui_port"`
	OverlayPort int      `json:"overlay_port"`
	Announces   []string `json:""`
}
