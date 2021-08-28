package downloader

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func Download(url, filename string, inCh, outCh chan string) {
	var writeList []string
	var s1 string
	var file *os.File
	var err error
	// create output file
	if file == nil {
		if filename == "" {
			var t string = time.Now().Format("20060102150405")
			filename = t + ".mp4"
		}
		file, err = os.Create(filename)
		if err != nil {
			outCh <- "error"
			return
		}
		defer file.Close()
	}
	// received file from server
	if strings.Contains(url, "@XXX@") {
		i := 1
		for true {
			l := strings.Replace(url, "@X@", strconv.Itoa(i), 1)

			if resp, err := http.Get(l); err != nil || resp.StatusCode != 200 {
				outCh <- "error"
				return
			} else {
				defer resp.Body.Close()
				if part, err := ioutil.ReadAll(resp.Body); err != nil {
					outCh <- "error"
					return
				} else if _, err = file.Write(part); err != nil {
					outCh <- "error"
					return
				}
			}
			i++
			select {
			case msg := <-inCh:
				if msg == "stop" {
					outCh <- "is_stop"
					return
				}
			default:
			}
		}
	}
	for true {
		br := false
		resp, err := http.Get(url)
		if err != nil {
			outCh <- "error"
			return
		}
		defer resp.Body.Close()

		s := resp.Request.URL.Path
		ind1 := strings.LastIndex(s, "/") + 1
		s1 = s[0:ind1]

		// read server response line by line
		scanner := bufio.NewScanner(resp.Body)
		b := false
		for scanner.Scan() {
			l := scanner.Text()
			bb := false
			// if line contains url address
			if strings.HasPrefix(l, "http") {
				bb = true
				// download file part
			} else if strings.Contains(l, ".ts") {
				bb = true
				l = resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + s1 + l
			} else if strings.Contains(l, ".m3u8") {
				url = resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + s1 + l
				br = true
				break
			}

			if bb {
				b = true
				if !contains(writeList, l) {
					fmt.Println(l)
					writeList = append(writeList, l)
					if resp, err := http.Get(l); err != nil {
						outCh <- "error"
						return
					} else {
						defer resp.Body.Close()
						var part []byte
						if part, err = ioutil.ReadAll(resp.Body); err != nil {
							outCh <- "error"
							return
						}
						if _, err = file.Write(part); err != nil {
							outCh <- "error"
							return
						}
					}
				}
			}

			select {
			case msg := <-inCh:
				if msg == "stop" {
					outCh <- "is_stop"
					return
				}
			default:
			}
		}

		if br {
			continue
		}

		if !b {
			outCh <- "error"
			return
		}
		if err = scanner.Err(); err != nil {
			outCh <- "error"
			return
		}

		select {
		case msg := <-inCh:
			if msg == "stop" {
				outCh <- "is_stop"
				return
			}
		default:
		}
		fmt.Println("123")
		break
	}
	outCh <- "finish"
}
