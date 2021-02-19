package main

import (
	"syscall"
	"time"
	"unsafe"
)

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
	ticker := time.NewTicker(time.Second * 1)
	ch := make(chan int)
	day := time.Now().Day()
	go func() {
		for true {
			select {
			case <-ticker.C:
				now := time.Now()
				hour := now.Hour()
				minute := now.Minute()
				if day != now.Day() {
					day = now.Day()
					closeFlag = false
				}
				//fmt.Println(time.Now().Hour())
				if hour == 9 {
					if minute == 55 && !closeFlag {
						ShowMessage2("打卡提示", "早上上班打卡")
					} else if minute != 55 && closeFlag {
						closeFlag = false
					}
				} else if hour == 11 {
					if minute == 49 && !closeFlag {
						ShowMessage2("午餐提示", "早上下班打卡, 身体革命本钱, 人是铁饭是钢! ")
					} else if minute != 49 && closeFlag {
						closeFlag = false
					}
				} else if hour == 13 {
					if minute == 28 && !closeFlag {
						ShowMessage2("打卡提示", "下午上班打卡")
					} else if minute != 28 && closeFlag {
						closeFlag = false
					}
				} else if hour == 17 {
					if minute == 39 && !closeFlag {
						ShowMessage2("打卡提示", "下午下班打卡")
					} else if minute != 39 && closeFlag {
						closeFlag = false
					}
				}
			}
		}
		ticker.Stop()
		ch <- 0
	}()
	<-ch // 通过通道阻塞，让任务可以执行完指定的次数。
}
