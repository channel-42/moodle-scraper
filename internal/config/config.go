package config

type config struct {
	BaseUrl     string
	Version     string
	DownloadAll bool
}

type user struct {
	Username string
	Password string
}

var Config *config = &config{
	BaseUrl:     "https://moodle.hs-hannover.de",
	Version:     "v0.1.0",
	DownloadAll: false,
}

var User *user = &user{}

var Banner string = `
                                       █████ ████                                                                               
                                      ░░███ ░░███                                                                               
 █████████████    ██████   ██████   ███████  ░███   ██████      █████   ██████  ████████   ██████   ████████   ██████  ████████ 
░░███░░███░░███  ███░░███ ███░░███ ███░░███  ░███  ███░░███    ███░░   ███░░███░░███░░███ ░░░░░███ ░░███░░███ ███░░███░░███░░███
 ░███ ░███ ░███ ░███ ░███░███ ░███░███ ░███  ░███ ░███████    ░░█████ ░███ ░░░  ░███ ░░░   ███████  ░███ ░███░███████  ░███ ░░░ 
 ░███ ░███ ░███ ░███ ░███░███ ░███░███ ░███  ░███ ░███░░░      ░░░░███░███  ███ ░███      ███░░███  ░███ ░███░███░░░   ░███     
 █████░███ █████░░██████ ░░██████ ░░████████ █████░░██████     ██████ ░░██████  █████    ░░████████ ░███████ ░░██████  █████    
░░░░░ ░░░ ░░░░░  ░░░░░░   ░░░░░░   ░░░░░░░░ ░░░░░  ░░░░░░     ░░░░░░   ░░░░░░  ░░░░░      ░░░░░░░░  ░███░░░   ░░░░░░  ░░░░░     
                                                                                                    ░███                        
                                                                                                    █████                       
                                                                                                   ░░░░░                        
																								   `
