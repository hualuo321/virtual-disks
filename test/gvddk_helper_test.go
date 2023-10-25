package main

import (
	"github.com/vmware/virtual-disks/pkg/disklib"
	"os"
	"strings"
	"testing"
)

/*
	这段代码定义了一个测试函数，用于验证VMware服务器的指纹是否与预期的指纹匹配。
	这是为了确保VMware服务器的证书是有效且未被篡改的。
*/
func TestGetThumbPrintForServer(t *testing.T) {
	// 从环境变量中获取VMware服务器的主机名和端口
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	// 获取预期的VMware服务器指纹
	vmwareThumbprint := os.Getenv("THUMBPRINT")
	// 如果预期的指纹未在环境变量中设置，则跳过测试
	if vmwareThumbprint == "" {
		t.Skip("Skipping testing if environment variables are not set.")
	}
	// 使用disklib库获取VMware服务器的实际指纹
	thumbprint, err := disklib.GetThumbPrintForServer(host, port)
	// 如果获取过程中出现错误，记录错误
	if err != nil {
		t.Errorf("Thumbprint for %s:%s failed, err = %s\n", host, port, err)
	}
	// 记录实际获取到的指纹
	t.Logf("Thumbprint for %s:%s is %s\n", host, port, thumbprint)
	// 将实际获取到的指纹与预期的指纹进行比较
	// 如果不匹配，则可能证书已被更新或存在问题，因此记录一个错误
	if strings.Compare(vmwareThumbprint, thumbprint) != 0 {
		t.Errorf("Thumbprint %s does not match expected thumbprint %s for %s - check to see if cert has been updated at %s\n",
			thumbprint, vmwareThumbprint, host, host)
	}
}
