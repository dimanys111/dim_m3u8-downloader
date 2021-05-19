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
				ind2 := strings.LastIndex(s, ".")
				filename = s[ind1:ind2] + ".mp4"
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
			} else if strings.HasSuffix(l, ".ts") {
				bb = true
				l = resp.Request.URL.Scheme + "://" + resp.Request.URL.Host + s1 + l
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
		time.Sleep(2 * time.Second)
	}

	return filename, errors.New("YYY")
}
