package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yeka/zip"
)

var (
	version  = "1.0.0"
	result   string
	resultMu sync.Mutex
	found    int32
	stopped  int32
	attempts int64
)

func main() {
	showVersion := flag.Bool("v", false, "显示版本信息")
	showVersionLong := flag.Bool("version", false, "显示版本信息")
	zipPath := flag.String("f", "", "ZIP文件路径 (必需)")
	wordlist := flag.String("w", "", "密码词典文件路径")
	minLen := flag.Int("min", 1, "密码最小长度 (用于暴力模式)")
	maxLen := flag.Int("max", 4, "密码最大长度 (用于暴力模式)")
	charset := flag.String("charset", "0123456789", "密码字符集 (用于暴力模式)")
	workers := flag.Int("workers", 1, "并发工作线程数")
	maxTime := flag.Duration("timeout", 0, "最大运行时间(0表示无限制)")
	showProgress := flag.Bool("progress", true, "显示进度")

	flag.Usage = func() {
		fmt.Println("╔══════════════════════════════════════════════════════════════╗")
		fmt.Printf("║                  ZIP 密码破解工具 v%s                    ║\n", version)
		fmt.Println("╠══════════════════════════════════════════════════════════════╣")
		fmt.Println("║  用法:                                                       ║")
		fmt.Println("║    tryunzip -f <zip文件> [选项]                               ║")
		fmt.Println("║                                                              ║")
		fmt.Println("║  暴力模式 (默认):                                             ║")
		fmt.Println("║    tryunzip -f test.zip                                       ║")
		fmt.Println("║    tryunzip -f test.zip -min 4 -max 6 -charset 0123456789   ║")
		fmt.Println("║                                                              ║")
		fmt.Println("║  词典模式:                                                   ║")
		fmt.Println("║    tryunzip -f test.zip -w passwords.txt                    ║")
		fmt.Println("║                                                              ║")
		fmt.Println("║  选项说明:                                                   ║")
		fmt.Println("║    -v, -version    显示版本信息                              ║")
		fmt.Println("║    -f <file>      ZIP文件路径 (必需)                          ║")
		fmt.Println("║    -w <file>      密码词典文件路径                           ║")
		fmt.Println("║    -min <n>       密码最小长度 (默认: 1)                    ║")
		fmt.Println("║    -max <n>       密码最大长度 (默认: 4)                    ║")
		fmt.Println("║    -charset <str> 字符集 (默认: 0123456789)                  ║")
		fmt.Println("║                    常用: abcdefghijklmnopqrstuvwxyz          ║")
		fmt.Println("║                          ABCDEFGHIJKLMNOPQRSTUVWXYZ          ║")
		fmt.Println("║                          0123456789                          ║")
		fmt.Println("║                          a-z A-Z 0-9                         ║")
		fmt.Println("║    -workers <n>    工作线程数 (默认: 1)                      ║")
		fmt.Println("║    -timeout <dur>  最大运行时间 (0表示无限制)               ║")
		fmt.Println("║    -progress       显示进度 (默认: true)                    ║")
		fmt.Println("║                                                              ║")
		fmt.Println("║  示例:                                                       ║")
		fmt.Println("║    tryunzip -f test.zip -charset a-z -min 3 -max 5          ║")
		fmt.Println("║    tryunzip -f test.zip -charset a-zA-Z0-9 -min 6 -max 8    ║")
		fmt.Println("║    tryunzip -f test.zip -w common.txt                        ║")
		fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	}

	flag.Parse()

	if *showVersion || *showVersionLong {
		fmt.Printf("tryunzip version %s\n", version)
		os.Exit(0)
	}

	if *zipPath == "" {
		fmt.Println("\n错误: 必须指定ZIP文件路径 (-f)")
		fmt.Println("使用 -help 查看帮助")
		os.Exit(1)
	}

	if _, err := os.Stat(*zipPath); os.IsNotExist(err) {
		fmt.Printf("\n错误: 文件不存在: %s\n", *zipPath)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Printf("║  ZIP文件: %-47s║\n", *zipPath)
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	passwordCh := make(chan string, 100)
	stopChan := make(chan struct{})
	var wg sync.WaitGroup
	startTime := time.Now()

	if *wordlist != "" {
		if _, err := os.Stat(*wordlist); os.IsNotExist(err) {
			fmt.Printf("║  错误: 词典文件不存在: %-36s║\n", *wordlist)
			fmt.Println("╚══════════════════════════════════════════════════════════════╝")
			os.Exit(1)
		}
		fmt.Printf("║  模式: %-53s║\n", "词典模式")
		fmt.Printf("║  词典: %-52s║\n", *wordlist)

		wg.Add(1)
		go func() {
			defer wg.Done()
			loadWordlist(*wordlist, passwordCh, stopChan)
			close(passwordCh)
		}()
	} else {
		expandedCharset := expandCharset(*charset)
		fmt.Printf("║  模式: %-53s║\n", "暴力模式")
		fmt.Printf("║  密码长度: %d-%d                                             ║\n", *minLen, *maxLen)
		fmt.Printf("║  字符集: %-51s║\n", expandedCharset)
		fmt.Printf("║  组合数: %-51s║\n", formatNumber(int64(calculateCombinations(len(expandedCharset), *minLen, *maxLen))))

		wg.Add(1)
		go func() {
			defer wg.Done()
			generatePasswords(expandedCharset, *minLen, *maxLen, passwordCh, stopChan)
			close(passwordCh)
		}()
	}

	fmt.Printf("║  工作线程: %-48d║\n", *workers)
	if *maxTime > 0 {
		fmt.Printf("║  超时时间: %-48s║\n", *maxTime)
	}
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if *showProgress {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if atomic.LoadInt32(&found) > 0 || atomic.LoadInt32(&stopped) > 0 {
						return
					}
					elapsed := time.Since(startTime)
					attemptsVal := atomic.LoadInt64(&attempts)
					if elapsed.Seconds() > 0 {
						rate := float64(attemptsVal) / elapsed.Seconds()
						fmt.Printf("\r  [进度] 已尝试: %s | 速率: %s/s | 耗时: %s",
							formatNumber(attemptsVal),
							formatNumber(int64(rate)),
							elapsed.Round(time.Second))
						if attemptsVal > 0 {
							fmt.Printf(" | 平均: %.2fms/次", elapsed.Seconds()*1000/float64(attemptsVal))
						}
					}
				case <-stopChan:
					return
				}
			}
		}()
	}

	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pwd := range passwordCh {
				if atomic.LoadInt32(&found) > 0 || atomic.LoadInt32(&stopped) > 0 {
					return
				}
				if tryPassword(*zipPath, pwd) {
					if atomic.CompareAndSwapInt32(&found, 0, 1) {
						resultMu.Lock()
						result = pwd
						resultMu.Unlock()
						close(stopChan)
					}
					return
				}
				atomic.AddInt64(&attempts, 1)
			}
		}()
	}

	if *maxTime > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(*maxTime)
			if atomic.CompareAndSwapInt32(&stopped, 0, 1) {
				close(stopChan)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(startTime)
	attemptsVal := atomic.LoadInt64(&attempts)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	if result != "" {
		fmt.Printf("║  ✓ 密码已找到: %-46s║\n", result)
		fmt.Printf("║    尝试次数: %-46s║\n", formatNumber(attemptsVal))
		fmt.Printf("║    耗时: %-50s║\n", elapsed.Round(time.Second))
		if elapsed.Seconds() > 0 {
			fmt.Printf("║    平均速率: %-45s║\n", formatNumber(int64(float64(attemptsVal)/elapsed.Seconds()))+"/s")
		}
		fmt.Println("╠══════════════════════════════════════════════════════════════╣")
		fmt.Println("║                                                              ║")
		fmt.Printf("║  解压命令: unzip -P %s %s               ║\n", result, *zipPath)
		fmt.Println("║                                                              ║")
	} else if atomic.LoadInt32(&stopped) > 0 {
		fmt.Println("║  ✗ 超时，未找到密码                                           ║")
		fmt.Printf("║    尝试次数: %-46s║\n", formatNumber(attemptsVal))
		fmt.Printf("║    耗时: %-50s║\n", elapsed.Round(time.Second))
	} else {
		fmt.Println("║  ✗ 未找到密码                                                 ║")
		fmt.Printf("║    尝试次数: %-46s║\n", formatNumber(attemptsVal))
		fmt.Printf("║    耗时: %-50s║\n", elapsed.Round(time.Second))
	}
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}

func loadWordlist(wordlistPath string, passwordCh chan<- string, stopChan <-chan struct{}) {
	file, err := os.Open(wordlistPath)
	if err != nil {
		fmt.Printf("\n错误: 无法打开词典文件: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		select {
		case <-stopChan:
			return
		default:
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				passwordCh <- line
				lineNum++
				if lineNum%10000 == 0 {
					fmt.Printf("\r  已加载: %s 个密码", formatNumber(int64(lineNum)))
				}
			}
		}
	}
	if lineNum > 10000 {
		fmt.Println()
	}
}

func generatePasswords(charset string, minLen, maxLen int, passwordCh chan<- string, stopChan <-chan struct{}) {
	charsetLen := len(charset)
	if charsetLen == 0 {
		return
	}

	for length := minLen; length <= maxLen; length++ {
		if length == 1 {
			for i := 0; i < charsetLen; i++ {
				select {
				case <-stopChan:
					return
				case passwordCh <- string(charset[i]):
				}
			}
		} else {
			generateRecursive(make([]byte, length), 0, charset, charsetLen, passwordCh, stopChan)
		}
	}
}

func generateRecursive(current []byte, pos int, charset string, charsetLen int, passwordCh chan<- string, stopChan <-chan struct{}) {
	if pos == len(current) {
		select {
		case <-stopChan:
			return
		case passwordCh <- string(current):
		}
		return
	}
	for i := 0; i < charsetLen; i++ {
		select {
		case <-stopChan:
			return
		default:
			current[pos] = charset[i]
			generateRecursive(current, pos+1, charset, charsetLen, passwordCh, stopChan)
		}
	}
}

func tryPassword(zipPath, password string) bool {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return false
	}
	defer r.Close()

	for _, file := range r.File {
		file.SetPassword(password)

		rc, err := file.Open()
		if err != nil {
			if password == "" {
				continue
			}
			return false
		}

		_, readErr := io.ReadAll(rc)
		rc.Close()

		if readErr != nil && password != "" {
			return false
		}
	}

	return true
}

func expandCharset(input string) string {
	var result strings.Builder
	for i := 0; i < len(input); i++ {
		if i+1 < len(input) && input[i+1] == '-' {
			start := input[i]
			end := input[i+2]
			if i+2 < len(input) && start < end {
				for c := start; c <= end; c++ {
					result.WriteByte(c)
				}
				i += 2
			} else {
				result.WriteByte(start)
			}
		} else {
			result.WriteByte(input[i])
		}
	}
	return result.String()
}

func calculateCombinations(charsetLen, minLen, maxLen int) int {
	total := 0
	for l := minLen; l <= maxLen; l++ {
		power := 1
		for i := 0; i < l; i++ {
			power *= charsetLen
		}
		total += power
	}
	return total
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	if n < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(n)/1000000000)
}
