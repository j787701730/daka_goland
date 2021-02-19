package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type Configs map[string]json.RawMessage

var configPath string = "./config.json"

type Desc struct {
	Time    string `json:"time"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type MainConfig struct {
	Data []Desc `json:"data"`
}

var conf *MainConfig
var confs Configs

var instanceOnce sync.Once

//从配置文件中载入json字符串
func LoadConfig(path string) (Configs, *MainConfig) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Panicln("load config conf failed: ", err)
	}
	mainConfig := &MainConfig{}
	err = json.Unmarshal(buf, mainConfig)
	if err != nil {
		log.Panicln("decode config file failed:", string(buf), err)
	}
	allConfigs := make(Configs, 0)
	err = json.Unmarshal(buf, &allConfigs)
	if err != nil {
		log.Panicln("decode config file failed:", string(buf), err)
	}

	return allConfigs, mainConfig
}

func Init(path string) *MainConfig {
	if conf != nil && path != configPath {
		log.Printf("the config is already initialized, oldPath=%s, path=%s", configPath, path)
	}
	instanceOnce.Do(func() {
		allConfigs, mainConfig := LoadConfig(path)
		configPath = path
		conf = mainConfig
		confs = allConfigs
	})

	return conf
}

func IntPtr(n int) uintptr {
	return uintptr(n)
}

func StrPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}

var closeFlag = false

// windows下的另一种DLL方法调用
func ShowMessage2(title, text string) {
	user32dll, _ := syscall.LoadLibrary("user32.dll")
	user32 := syscall.NewLazyDLL("user32.dll")
	MessageBoxW := user32.NewProc("MessageBoxW")
	MessageBoxW.Call(IntPtr(0), StrPtr(text), StrPtr(title), IntPtr(0))
	defer syscall.FreeLibrary(user32dll)
	closeFlag = true
}

func main() {
	path := configPath
	//fmt.Println("path: ", path)
	Init(path)
	value := confs["data"]
	//fmt.Println(string(value))

	var data []Desc
	err := json.Unmarshal([]byte(value), &data)
	//fmt.Println(strings.Split(data[0].Time, ":")[0])
	//fmt.Println(len(data))
	if err != nil {
		log.Panicln("decode config file failed:", string("x"), err)
	}

	ticker := time.NewTicker(time.Second * 1)
	ch := make(chan int)
	day := time.Now().Day()
	go func() {
		for true {
			select {
			case <-ticker.C:
				now := time.Now()
				hour := strconv.Itoa(now.Hour())
				minute := strconv.Itoa(now.Minute())
				if day != now.Day() {
					day = now.Day()
					closeFlag = false
				}

				for j := 0; j < len(data); j++ {
					dataTime := strings.Split(data[j].Time, ":")
					//	strings.Split(data[0].Time, ":")[0]
					if hour == dataTime[0] {
						//fmt.Println(hour)
						if minute == dataTime[1] && !closeFlag {
							ShowMessage2(data[j].Title, data[j].Message)
						} else if minute != dataTime[1] && closeFlag {
							closeFlag = false
						}
					}
				}
			}
		}
		ticker.Stop()
		ch <- 0
	}()
	<-ch // 通过通道阻塞，让任务可以执行完指定的次数。
}
