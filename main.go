package main

import (
	"dimanys111/downloader"
	"dimanys111/gui"
)

func main() {
	a := gui.New(downloader.Download)
	a.Run()
}
