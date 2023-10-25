package main

import (
	"github.com/vmware/virtual-disks/pkg/disklib"
	"os"
	"testing"
)

func TestCreate(t *testing.T) {
	// 设置
	path := os.Getenv("LIBPATH")
	// 如果LIBPATH环境变量未设置，则跳过测试
	if path == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	// 初始化disklib库
	res := disklib.Init(7, 0, path)
	// 如果初始化失败，记录错误
	if res != nil {
		t.Errorf("Init failed, got error code: %d, error message: %s.", res.VixErrorCode(), res.Error())
	}
	// 获取环境变量的值
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
	if err1 != nil {
		t.Errorf("Prepare for access failed. Error code: %d. Error message: %s.", err1.VixErrorCode(), err1.Error())
	}
	// 连接到磁盘库
	conn, err2 := disklib.ConnectEx(params)
	// 如果连接失败，结束访问并记录错误
	if err2 != nil {
	 disklib.EndAccess(params)
		t.Errorf("Connect to vixdisk lib failed. Error code: %d. Error message: %s.", err2.VixErrorCode(), err2.Error())
	}
	// 断开连接
 	disklib.Disconnect(conn)
	// 结束访问
 	disklib.EndAccess(params)
}
