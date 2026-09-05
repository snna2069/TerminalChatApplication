package server

type room struct {
	name    string
	clients map[*client]struct{}
}

func newRoom(name string) *room {
	return &room{name: name, clients: make(map[*client]struct{})}
}
