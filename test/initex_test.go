package main

import (
	"github.com/vmware/virtual-disks/pkg/disklib"
	"os"
	"testing"
)

// TestInitEx 用于测试disklib库的初始化以及准备访问磁盘的功能。
// 为了运行此测试用例，以下环境变量都是必需的:
// LIBPATH: vddk库的路径。
// CONFIGFILE: 包含自定义日志级别设置的配置文件路径，例如：verbosevixDiskLib.transport.LogLevel=4。
// VC IP, THUMBPRINT, USERNAME, PASSWORD, FCDID, DATASTORE
// IDENTITY: 仅用于身份跟踪目的的自定义名称，限于50个字符。
func TestInitEx(t *testing.T) {
	// 从环境变量中获取库路径
	path := os.Getenv("LIBPATH")
	// 如果LIBPATH未设置，则跳过测试
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	// 从环境变量中获取配置文件路径
	config := os.Getenv("CONFIGFILE")
	// 如果CONFIGFILE未设置，则跳过测试
	if config == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	// 使用提供的库路径和配置文件路径初始化disklib库
	res := disklib.InitEx(7, 0, path, config)
	// 如果初始化失败，记录错误
	if res != nil {
		t.Errorf("Init failed, got error code: %d, error message: %s.", res.VixErrorCode(), res.Error())
	}
	// 获取其他必要的环境变量
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
	// 准备访问磁盘
	err1 := disklib.PrepareForAccess(params)
	// 如果准备访问过程中出现错误，记录错误
	if err1 != nil {
		t.Errorf("Prepare for access failed. Error code: %d. Error message: %s.", err1.VixErrorCode(), err1.Error())
	}
	// 结束磁盘访问
	disklib.EndAccess(params)
}
