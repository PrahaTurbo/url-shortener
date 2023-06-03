package app

type application struct {
	urls map[string][]byte
	Addr string
}

func NewApp() application {
	return application{
		urls: make(map[string][]byte),
		Addr: "localhost:8080",
	}
}
