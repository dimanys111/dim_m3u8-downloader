package downloader

import (
	"bufio"
	"errors"
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

func Download(url, filename string, inCh, outCh chan string) (string, error) {
	// received file from server
	if strings.Contains(url, "@XXX@") {
		f, err := os.Create(filename)
		if err != nil {
			return "", err
		}
		defer f.Close()
		i := 1
		for true {
			l := strings.Replace(url, "@X@", strconv.Itoa(i), 1)

			if resp, err := http.Get(l); err != nil || resp.StatusCode != 200 {
				outCh <- "is_stop"
				return "", err
			} else {
				defer resp.Body.Close()
				if part, err := ioutil.ReadAll(resp.Body); err != nil {
					outCh <- "is_stop"
					return "", err
				} else if _, err = f.Write(part); err != nil {
					outCh <- "is_stop"
					return "", err
				}
			}
			i++
			select {
			case msg := <-inCh:
				if msg == "stop" {
					outCh <- "is_stop"
					return filename, err
				}
			default:
			}
		}
	}
	var writeList []string
	var s1 string
	var f *os.File
	for true {
		var br bool = false
		resp, err := http.Get(url)

		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		// create output file
		if f == nil {
			s := resp.Request.URL.Path
			ind1 := strings.LastIndex(s, "/") + 1
			s1 = s[0:ind1]
			if filename == "" {
				var t string = time.Now().Format("20060102150405")
				filename = t + ".mp4"
			}
			f, err = os.Create(filename)
			if err != nil {
				return "", err
			}
			defer f.Close()
		}

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
						return "", err
					} else {
						defer resp.Body.Close()
						if part, err := ioutil.ReadAll(resp.Body); err != nil {
							return "", err
						} else if _, err = f.Write(part); err != nil {
							return "", err
						}
					}
				}
			}

			select {
			case msg := <-inCh:
				if msg == "stop" {
					outCh <- "is_stop"
					return filename, err
				}
			default:
			}
		}

		if br {
			continue
		}

		if !b {
			return "", errors.New("XXX")
		}
		if err = scanner.Err(); err != nil {
			return filename, err
		}

		select {
		case msg := <-inCh:
			if msg == "stop" {
				outCh <- "is_stop"
				return filename, err
			}
		default:
		}
		fmt.Println("123")
		break
	}
	outCh <- "is_stop"
	return filename, errors.New("YYY")
}
