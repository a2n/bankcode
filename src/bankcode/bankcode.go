package bankcode

import (
	"net/http"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"encoding/json"
	"errors"
	"time"
	"os"
	"log"
	"sort"

	"github.com/qiniu/iconv"
)

// 更新，檢查本地和網頁版本，比較新舊。若有更新則產生 banks.json 和 banks.js。
func Update() error {
	newer, err := HasNewVersion()
	if err != nil {
		return err
	}

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	if !newer {
		log.Println("No need to update.")
	} else {
		log.Println("Getting the web page...")
		banks, err := Get()
		if err != nil {
			return err
		}

		log.Println("Writing the result.")
		err = Write(banks)
		if err != nil {
			return err
		}

		log.Println("Write files successfully.")
	}

	return nil
}

// 網頁地址
const URL = "http://www.esunbank.com.tw/event/announce/BankCode.htm"

// 檢查本地和網頁版本新舊
func HasNewVersion() (bool, error) {
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

type Bank struct {
	No string
	Name string
	Type string
}

// 下載網頁資料，並返回 map 結果。
func Get() ([]Bank, error) {
	// Bank
	// Type > Name > Number.
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

	banks := make([]Bank, 0)
	reNo := regexp.MustCompile(`\d+.?`)
	no := ""
	const SPACE =  "&nbsp;"
	for i := 17; i < len(ms); i++ {
		if ms[i][2] == SPACE {
			continue
		}

		if reNo.MatchString(ms[i][2]) {
			no = strings.Trim(ms[i][2], SPACE)
		} else {
			theType := ""
			switch ms[i][1] {
				case "2", "5": theType = "商業銀行"
				case "3": theType = "漁會信用部"
				case "4", "9": theType = "農會信用部"
				case "6": theType = "信用合作社"
			}

			banks = append(banks, Bank {
				No: no,
				Name: ms[i][2],
				Type: theType,
			})
		}
	}

	sort.Sort(ByNo(banks))

	return banks, nil
}

// 將整理的結果寫至檔案，分別是 bankcode.json 和 bankcode.js。
func Write(banks []Bank) error {
	// bankcode.json
	b, err := json.Marshal(banks)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("bankcode.json", b, 0600)
	if err != nil {
		return err
	}

	// bankcode.min.js
	raw, err := ioutil.ReadFile("bankcode.min.js.tpl")
	if err != nil {
		return err
	}
	tpl := strings.Replace(string(raw), "%CODE%", string(b), 1)
	err = ioutil.WriteFile("bankcode.min.js", []byte(tpl), 0600)
	if err != nil {
		return err
	}

	return nil
}

type ByNo []Bank

func (b ByNo) Len() int {
	return len(b)
}

func (b ByNo) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByNo) Less(i, j int) bool {
	return b[i].No < b[j].No
}
