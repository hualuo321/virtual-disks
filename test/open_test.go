/*
这段代码基于 vmware/virtual-disks 包测试了虚拟机磁盘的打开、读取、写入和查询功能。
*/
package main

import (
	"fmt"
	"github.com/sirupsen/logrus"							// 用于日志记录
	"github.com/vmware/virtual-disks/pkg/disklib"			// 自定义的虚拟磁盘低级 API 库
	"github.com/vmware/virtual-disks/pkg/virtual_disks"		// 自定义的虚拟磁盘高级 API 库
	"os"
	"testing"
)

// TestOpen 是一个测试函数，用于测试虚拟磁盘的一些功能。
func TestOpen(t *testing.T) {
	fmt.Println("Test Open")
	// 定义版本信息，主版本号，次版本号
	var majorVersion uint32 = 7
	var minorVersion uint32 = 0
	// 从环境变量中获取 LIBPATH，如果未设置 LIBPATH，则跳过测试
	path := os.Getenv("LIBPATH")
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	// 初始化磁盘库：需要版本号和库路径
	disklib.Init(majorVersion, minorVersion, path)
	// 从环境变量中获取连接参数	
	serverName := os.Getenv("IP")					// IP
	thumPrint := os.Getenv("THUMBPRINT")			// 指纹
	userName := os.Getenv("USERNAME")				// 用户名
	password := os.Getenv("PASSWORD")				// 密码
	fcdId := os.Getenv("FCDID")						// FCD 是一个独立于虚拟机的磁盘，允许管理员直接管理操作没有与 VMware 关联的 VMDK。
	ds := os.Getenv("DATASTORE")					// 用于存储 VMware 虚拟机的资源的存储位置。
	identity := os.Getenv("IDENTITY")				// 标识
	params := disklib.NewConnectParams("", serverName,thumPrint, userName,
		password, fcdId, ds, "", "", identity, "", disklib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ,
		false, disklib.NBD)
	// 使用上述参数尝试打开虚拟磁盘。
	diskReaderWriter, err := virtual_disks.Open(params, logrus.New())
	// 如果出错，结束访问并记录错误
	if err != nil {
		disklib.EndAccess(params)
		t.Errorf("Open failed, got error code: %d, error message: %s.", err.VixErrorCode(), err.Error())
	}
	// 查询虚拟磁盘的已分配块。（假设至少有 1GiB 卷和 1MiB 块大小）
	abInitial, err := diskReaderWriter.QueryAllocatedBlocks(0, 2048*1024, 2048)
	if err != nil {
		t.Errorf("QueryAllocatedBlocks failed: %d, error message: %s", err.VixErrorCode(), err.Error())
	} else {
		// 打印查询到的块的信息
		fmt.Printf("Number of blocks: %d\n", len(abInitial))
		fmt.Printf("Offset      Length\n")
		for _, ab := range abInitial {
			fmt.Printf("0x%012x  0x%012x\n", ab.Offset(), ab.Length())
		}
	}
	// 读取虚拟磁盘的内容
	fmt.Printf("ReadAt test\n")
	buffer := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
	n, err4 := diskReaderWriter.Read(buffer)
	fmt.Printf("Read byte n = %d\n", n)
	fmt.Println(buffer)
	fmt.Println(err4)

	// 向虚拟磁盘写入内容
	fmt.Println("WriteAt start")
	buf1 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
	for i,_ := range(buf1) {
		buf1[i] = 'E'
	}
	n2, err2 := diskReaderWriter.WriteAt(buf1, 0)
	fmt.Printf("Write byte n = %d\n", n2)
	fmt.Println(err2)

	// 读取刚才写入的内容
	buffer2 := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
	n2, err5 := diskReaderWriter.ReadAt(buffer2, 0)
	fmt.Printf("Read byte n = %d\n", n2)
	fmt.Println(buffer2)
	fmt.Println(err5)

	// 再次查询虚拟磁盘的已分配块
	abFinal, err := diskReaderWriter.QueryAllocatedBlocks(0, 2048*1024, 2048)
	if err != nil {
		t.Errorf("QueryAllocatedBlocks failed: %d, error message: %s", err.VixErrorCode(), err.Error())
	} else {
		fmt.Printf("Number of blocks: %d\n", len(abInitial))
		fmt.Printf("Offset      Length\n")
		for _, ab := range abFinal {
			fmt.Printf("0x%012x  0x%012x\n", ab.Offset(), ab.Length())
		}
	}
	
	// 关闭虚拟磁盘连接
	diskReaderWriter.Close()
}
