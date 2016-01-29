package main

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"strconv"
	"encoding/json"
	"errors"
	"time"
	"os"
	"log"

	"github.com/qiniu/iconv"
)

func main() {
	newer, err := hasNewVersion()
	if err != nil {
		panic(err)
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if !newer {
		log.Println("No need to update.")
	} else {
		log.Println("Getting the web page...")
		banks, err := get()
		if err != nil {
			panic(err)
		}

		log.Println("Writing the result.")
		err = write(banks)
		if err != nil {
			panic(err)
		}

		log.Println("Write banks.json successfully.")
	}
}

const URL = "http://www.esunbank.com.tw/event/announce/BankCode.htm"
func hasNewVersion() (bool, error) {
	// Remote
	resp, err := http.DefaultClient.Head(URL)
	if err != nil {
		return false, err
	}

	str, ok := resp.Header["Last-Modified"]
	if !ok {
		return false, errors.New("The Last-Modified header value doesn't exist.")
	}
	resp.Body.Close()

	last, err := time.Parse(time.RFC1123, str[0])
	if err != nil {
		return false, err
	}

	// Local
	const CONFIG = ".config"
	b, err := ioutil.ReadFile(CONFIG)
	firstTime := false
	if err != nil {
		if err != os.ErrNotExist {
			firstTime = true
			f, err := os.OpenFile(CONFIG, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0600)
			if err != nil {
				return false, err
			}
			f.Close()
		} else {
			return false, err
		}
	}

	local := time.Date(0, time.January, 1, 0, 0, 0, 0, time.Local)
	if !firstTime {
		local, err = time.Parse(time.RFC1123, string(b))
		if err != nil {
			return false, err
		}
	}

	newer := last.After(local)
	if newer {
		err = ioutil.WriteFile(CONFIG, []byte(str[0]), 0600)
		if err != nil {
			return false, err
		}
	}

	return newer, nil
}

// Bank
// Type > Name > Number.
func get() (map[string]map[string]uint16, error) {
	resp, err := http.DefaultClient.Get(URL)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	cd, err := iconv.Open("utf-8//ignore", "big5")
	if err != nil {
		return nil, fmt.Errorf("Open iconv failed, %s.", err.Error())
	}
	utf8 := cd.ConvString(string(b))
	cd.Close()

	re := regexp.MustCompile(`style(\d)">(.+?)</td>`)
	if !re.MatchString(utf8) {
		return nil, fmt.Errorf("The regexp doesn't matched.")
	}
	ms := re.FindAllStringSubmatch(utf8, -1)

	banks := map[string]map[string]uint16 {
		"商業銀行": make(map[string]uint16),
		"漁會信用部": make(map[string]uint16),
		"農會信用部": make(map[string]uint16),
		"信用合作社": make(map[string]uint16),
	}

	reNo := regexp.MustCompile(`\d+.?`)
	no := 0
	const SPACE =  "&nbsp;"
	for i := 17; i < len(ms); i++ {
		if ms[i][2] == SPACE {
			continue
		}

		if reNo.MatchString(ms[i][2]) {
			no, err = strconv.Atoi(strings.Trim(ms[i][2], SPACE))
			if err != nil {
				return nil, err
			}
		} else {
			theType := ""
			switch ms[i][1] {
				case "2", "5": theType = "商業銀行"
				case "3": theType = "漁會信用部"
				case "4", "9": theType = "農會信用部"
				case "6": theType = "信用合作社"
			}

			banks[theType][ms[i][2]] = uint16(no)
		}
	}

	return banks, nil
}

func write(banks interface{}) error {
	b, err := json.Marshal(banks)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("banks.json", b, 0600)
	if err != nil {
		return err
	}
	return nil
}
