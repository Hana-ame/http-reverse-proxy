package main

import (
	"bufio"
	"fmt"
	"os"
)

func openfile() (ips []string) {
	// 打开文件
	file, err := os.Open("ips.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建一个 Scanner 对象
	scanner := bufio.NewScanner(file)

	// 按行读取文件内容
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		ips = append(ips, line)
	}

	// 检查 Scanner 过程中的错误
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}

	return
}
