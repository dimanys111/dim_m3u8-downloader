package main

import (
	"dimanys111/m3u8-downloader/downloader"
	"dimanys111/m3u8-downloader/gui"
)

func main() {
	a := gui.New(downloader.Download)
	a.Run()
}
