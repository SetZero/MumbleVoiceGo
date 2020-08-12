package main

import (
	"MumbleSound/src/mumble"
	"MumbleSound/src/rest"
)

func main() {
	go mumble.StartConnection()
	rest.RequestRouter()
}
