package main

import (
	"github.com/vmware/virtual-disks/pkg/disklib"
	"github.com/vmware/virtual-disks/pkg/virtual_disks"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"testing"
)

// II vs II
// TestAligned 测试多线程写操作在对齐的情况下的行为。
// 它同时写入两种不同的字节（'A' 和 'B'）到磁盘的同一个位置，然后尝试读取该位置来验证写入的数据。
func TestAligned(t *testing.T) {
	// 打印测试信息
	fmt.Println("Test Multithread write for aligned case which skip lock: II vs II")
	// 设置磁盘库的版本信息
	var majorVersion uint32 = 7
	var minorVersion uint32 = 0
	// 从环境变量中获取库路径
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	// 初始化磁盘库
	disklib.Init(majorVersion, minorVersion, path)
	// 从环境变量中获取连接所需的参数
	serverName := os.Getenv("IP")
	thumPrint := os.Getenv("THUMBPRINT")
	userName := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fcdId := os.Getenv("FCDID")
	ds := os.Getenv("DATASTORE")
	identity := os.Getenv("IDENTITY")
	// 创建连接参数
	params := disklib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", disklib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, disklib.NBD)
	// 打开磁盘进行读写
	diskReaderWriter, err := virtual_disks.Open(params, logrus.New())
	if err != nil {
		disklib.EndAccess(params)
		t.Errorf("Open failed, got error code: %d, error message: %s.", err.VixErrorCode(), err.Error())
	}
	// 开始并发写操作
	// 使用两个goroutine在相同的偏移位置写入相同长度的数据，会产生竞争
	done := make(chan bool)
	fmt.Println("---------------------WriteAt start----------------------")
	// 第一个goroutine写入 'A' 字节
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		for i, _ := range (buf1) {
			buf1[i] = 'A'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 0)
		fmt.Printf("--------Write A byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()
	// 第二个goroutine写入 'B' 字节
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		for i, _ := range (buf1) {
			buf1[i] = 'B'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 0)
		fmt.Printf("--------Write B byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()
	// 等待两个goroutine完成
	for i := 0; i < 2; i++ {
		<-done
	}
	// 验证写入的数据，通过读取之前写入的位置
	fmt.Println("----------Read start to verify----------")
	buffer2 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
	n2, err5 := diskReaderWriter.ReadAt(buffer2, 0)
	fmt.Printf("Read byte n = %d\n", n2)
	fmt.Println(buffer2)
	fmt.Println(err5)
	// 关闭磁盘读写器
	diskReaderWriter.Close()
}

// I II III vs II III
// 这段代码的功能是测试多线程写操作，尤其是对齐不一致的情况，并行地将两种不同的字节值（'C' 和 'D'）写入磁盘，然后读取这些位置以验证写入的数据。
func TestMiss1(t *testing.T) {
	// 打印测试信息，表明这是一个多线程写入不对齐测试
	fmt.Println("Test Multithread write for miss aligned case which lock: I II III vs II III")
	var majorVersion uint32 = 7
	var minorVersion uint32 = 0
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	disklib.Init(majorVersion, minorVersion, path)
	serverName := os.Getenv("IP")
	thumPrint := os.Getenv("THUMBPRINT")
	userName := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fcdId := os.Getenv("FCDID")
	ds := os.Getenv("DATASTORE")
	identity := os.Getenv("IDENTITY")
	params := disklib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", disklib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, disklib.NBD)
	diskReaderWriter, err := virtual_disks.Open(params, logrus.New())
	if err != nil {
		disklib.EndAccess(params)
		t.Errorf("Open failed, got error code: %d, error message: %s.", err.VixErrorCode(), err.Error())
	}
	// 开始多线程写操作
	// 两个 goroutine 在不同的偏移位置写入不同长度的数据。
	// 由于它们写入的偏移位置和长度不同，所以不会有数据重叠或争用。
	done := make(chan bool)
	fmt.Println("---------------------WriteAt start----------------------")
	// 第一个goroutine写'C'字节到磁盘的指定位置
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
		for i, _ := range (buf1) {
			buf1[i] = 'C'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 500)
		fmt.Printf("--------Write C byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()
	// 第二个goroutine写'D'字节到磁盘的指定位置
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 2)
		for i, _ := range (buf1) {
			buf1[i] = 'D'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, disklib.VIXDISKLIB_SECTOR_SIZE)
		fmt.Printf("--------Write D byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()
	// 等待所有goroutines完成
	for i := 0; i < 2; i++ {
		<-done
	}
	// 验证写入的数据通过读取
	fmt.Println("----------Read start to verify----------")
	buffer2 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
	n2, err5 := diskReaderWriter.ReadAt(buffer2, 500)
	fmt.Printf("Read byte n = %d\n", n2)
	// 打印读取的数据
	fmt.Println(buffer2)
	fmt.Println(err5)
	// 关闭磁盘读写器
	diskReaderWriter.Close()
}

// I II vs I II III
func TestMiss2(t *testing.T) {
	fmt.Println("Test Multithread write for miss aligned case which lock: I II vs I II III")
	var majorVersion uint32 = 7
	var minorVersion uint32 = 0
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	disklib.Init(majorVersion, minorVersion, path)
	serverName := os.Getenv("IP")
	thumPrint := os.Getenv("THUMBPRINT")
	userName := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fcdId := os.Getenv("FCDID")
	ds := os.Getenv("DATASTORE")
	identity := os.Getenv("IDENTITY")
	params := disklib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", disklib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, disklib.NBD)
	diskReaderWriter, err := virtual_disks.Open(params, logrus.New())
	if err != nil {
		disklib.EndAccess(params)
		t.Errorf("Open failed, got error code: %d, error message: %s.", err.VixErrorCode(), err.Error())
	}
	// WriteAt
	// 两个 goroutine 都尝试在相同的偏移位置（500）写入数据，但数据的长度略有不同。
	// 这会导致数据竞争，因为两个 goroutine 可能会尝试在几乎相同的时间内修改相同的数据块。
	done := make(chan bool)
	fmt.Println("---------------------WriteAt start----------------------")
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 12)
		for i, _ := range (buf1) {
			buf1[i] = 'E'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 500)
		fmt.Printf("--------Write E byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()

	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
		for i, _ := range (buf1) {
			buf1[i] = 'F'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 500)
		fmt.Printf("--------Write F byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()

	for i := 0; i < 2; i++ {
		<-done
	}
	// Verify written data by read
	fmt.Println("----------Read start to verify----------")
	buffer2 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
	n2, err5 := diskReaderWriter.ReadAt(buffer2, 500)
	fmt.Printf("Read byte n = %d\n", n2)
	fmt.Println(buffer2)
	fmt.Println(err5)

	diskReaderWriter.Close()
}

// I II vs II III
func TestMiss3(t *testing.T) {
	fmt.Println("Test Multithread write for miss aligned case which lock: I II vs II III")
	var majorVersion uint32 = 7
	var minorVersion uint32 = 0
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	disklib.Init(majorVersion, minorVersion, path)
	serverName := os.Getenv("IP")
	thumPrint := os.Getenv("THUMBPRINT")
	userName := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fcdId := os.Getenv("FCDID")
	ds := os.Getenv("DATASTORE")
	identity := os.Getenv("IDENTITY")
	params := disklib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", disklib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, disklib.NBD)
	diskReaderWriter, err := virtual_disks.Open(params, logrus.New())
	if err != nil {
		disklib.EndAccess(params)
		t.Errorf("Open failed, got error code: %d, error message: %s.", err.VixErrorCode(), err.Error())
	}
	// WriteAt
	// 由于两个 goroutine 写入的位置可能存在部分重叠
	// 偏移量 500 和 disklib.VIXDISKLIB_SECTOR_SIZE 之间的差异小于 disklib.VIXDISKLIB_SECTOR_SIZE + 12
	// 所以可能会有数据竞争，取决于两个 goroutine 的执行顺序。
	done := make(chan bool)
	fmt.Println("---------------------WriteAt start----------------------")
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 12)
		for i, _ := range (buf1) {
			buf1[i] = 'G'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 500)
		fmt.Printf("--------Write G byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()

	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 2)
		for i, _ := range (buf1) {
			buf1[i] = 'H'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, disklib.VIXDISKLIB_SECTOR_SIZE)
		fmt.Printf("--------Write H byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()

	for i := 0; i < 2; i++ {
		<-done
	}
	// Verify written data by read
	fmt.Println("----------Read start to verify----------")
	buffer2 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
	n2, err5 := diskReaderWriter.ReadAt(buffer2, 500)
	fmt.Printf("Read byte n = %d\n", n2)
	fmt.Println(buffer2)
	fmt.Println(err5)

	diskReaderWriter.Close()
}

// I II III vs II
func TestMissAlign(t *testing.T) {
	fmt.Println("Test Multithread write for case which lock: I II III vs II")
	var majorVersion uint32 = 7
	var minorVersion uint32 = 0
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	disklib.Init(majorVersion, minorVersion, path)
	serverName := os.Getenv("IP")
	thumPrint := os.Getenv("THUMBPRINT")
	userName := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	fcdId := os.Getenv("FCDID")
	ds := os.Getenv("DATASTORE")
	identity := os.Getenv("IDENTITY")
	params := disklib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", disklib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, disklib.NBD)
	diskReaderWriter, err := virtual_disks.Open(params, logrus.New())
	if err != nil {
		disklib.EndAccess(params)
		t.Errorf("Open failed, got error code: %d, error message: %s.", err.VixErrorCode(), err.Error())
	}
	// WriteAt
	done := make(chan bool)
	fmt.Println("---------------------WriteAt start----------------------")
	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
		for i, _ := range (buf1) {
			buf1[i] = 'A'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, 500)
		fmt.Printf("--------Write A byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()

	go func() {
		buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		for i, _ := range (buf1) {
			buf1[i] = 'B'
		}
		n2, err2 := diskReaderWriter.WriteAt(buf1, disklib.VIXDISKLIB_SECTOR_SIZE)
		fmt.Printf("--------Write B byte n = %d\n", n2)
		fmt.Println(err2)
		done <- true
	}()

	for i := 0; i < 2; i++ {
		<-done
	}
	// Verify written data by read
	fmt.Println("----------Read start to verify----------")
	buffer2 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE + 14)
	n2, err5 := diskReaderWriter.ReadAt(buffer2, 500)
	fmt.Printf("Read byte n = %d\n", n2)
	fmt.Println(buffer2)
	fmt.Println(err5)

	diskReaderWriter.Close()
}
