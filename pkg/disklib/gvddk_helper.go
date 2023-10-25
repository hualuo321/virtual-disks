package disklib

// #include "gvddk_c.h"
import "C"
import (
	"crypto/tls"		// 导入 TLS 库用于加密通信
	"fmt"
	"net/url"			// 导入 URL 处理库
)
import "crypto/sha1"	// 导入 SHA-1 哈希库用于计算指纹

// 声明用于打开磁盘的标志常量
const (
	VIXDISKLIB_FLAG_OPEN_UNBUFFERED         = C.VIXDISKLIB_FLAG_OPEN_UNBUFFERED
	VIXDISKLIB_FLAG_OPEN_SINGLE_LINK        = C.VIXDISKLIB_FLAG_OPEN_SINGLE_LINK
	VIXDISKLIB_FLAG_OPEN_READ_ONLY          = C.VIXDISKLIB_FLAG_OPEN_READ_ONLY
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB   = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ  = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_MASK   = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_MASK
)

// 声明传输模式的常量
const (
	NBD    = "nbd"
	NBDSSL = "nbdssl"
	HOTADD = "hotadd"
)

// 声明扇区大小的常量
const VIXDISKLIB_SECTOR_SIZE = C.VIXDISKLIB_SECTOR_SIZE

// 定义块的常量
const VIXDISKLIB_MIN_CHUNK_SIZE = C.VIXDISKLIB_MIN_CHUNK_SIZE
const VIXDISKLIB_MAX_CHUNK_SIZE = C.VIXDISKLIB_MAX_CHUNK_SIZE
const VIXDISKLIB_MAX_CHUNK_NUMBER = C.VIXDISKLIB_MAX_CHUNK_NUMBER

// 声明错误代码的常量
const VIX_E_DISK_OUTOFRANGE = C.VIX_E_DISK_OUTOFRANGE

// 定义磁盘类型的枚举类型
type VixDiskLibDiskType int
const (
	VIXDISKLIB_DISK_MONOLITHIC_SPARSE VixDiskLibDiskType = C.VIXDISKLIB_DISK_MONOLITHIC_SPARSE // monolithic file, sparse,
	VIXDISKLIB_DISK_MONOLITHIC_FLAT   VixDiskLibDiskType = C.VIXDISKLIB_DISK_MONOLITHIC_FLAT   // monolithic file, all space pre-allocated
	VIXDISKLIB_DISK_SPLIT_SPARSE      VixDiskLibDiskType = C.VIXDISKLIB_DISK_SPLIT_SPARSE      // disk split into 2GB extents, sparse
	VIXDISKLIB_DISK_SPLIT_FLAT        VixDiskLibDiskType = C.VIXDISKLIB_DISK_SPLIT_FLAT        // disk split into 2GB extents, pre-allocated
	VIXDISKLIB_DISK_VMFS_FLAT         VixDiskLibDiskType = C.VIXDISKLIB_DISK_VMFS_FLAT         // ESX 3.0 and above flat disks
	VIXDISKLIB_DISK_STREAM_OPTIMIZED  VixDiskLibDiskType = C.VIXDISKLIB_DISK_STREAM_OPTIMIZED  // compressed monolithic sparse
	VIXDISKLIB_DISK_VMFS_THIN         VixDiskLibDiskType = C.VIXDISKLIB_DISK_VMFS_THIN         // ESX 3.0 and above thin provisioned
	VIXDISKLIB_DISK_VMFS_SPARSE       VixDiskLibDiskType = C.VIXDISKLIB_DISK_VMFS_SPARSE       // ESX 3.0 and above sparse disks
	VIXDISKLIB_DISK_UNKNOWN           VixDiskLibDiskType = C.VIXDISKLIB_DISK_UNKNOWN           // unknown type
)

// 定义适配器类型的枚举类型
type VixDiskLibAdapterType int

const (
	VIXDISKLIB_ADAPTER_IDE           VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_IDE
	VIXDISKLIB_ADAPTER_SCSI_BUSLOGIC VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_SCSI_BUSLOGIC
	VIXDISKLIB_ADAPTER_SCSI_LSILOGIC VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_SCSI_LSILOGIC
	VIXDISKLIB_ADAPTER_UNKNOWN       VixDiskLibAdapterType = C.VIXDISKLIB_ADAPTER_UNKNOWN
)

// 定义扇区类型的无符号整数类型
type VixDiskLibSectorType uint64

// 定义用于连接虚拟机的参数结构
type ConnectParams struct {
	vmxSpec    string	// 虚拟机配置文件的路径
	serverName string	// 服务器名称或IP地址
	thumbPrint string	// 服务器证书的指纹
	userName   string	// 用户名
	password   string	// 密码
	fcdId      string	// FCD ID
	ds         string	// 数据存储的名称
	fcdssId    string	// FCD SS ID
	cookie     string	// Cookie
	identity   string	// 身份标识
	path       string	// 路径
	flag       uint32	// 标志
	readOnly   bool		// 是否只读
	mode       string	// 模式
}

// 定义 VixDiskLibHandle 结构，表示磁盘句柄
type VixDiskLibHandle struct {
	dli C.VixDiskLibHandle
}

// 定义 VixDiskLibConnection 结构，表示磁盘连接
type VixDiskLibConnection struct {
	conn C.VixDiskLibConnection
}

// 定义 vddkErrorImpl 结构，实现 VddkError 接口
type VddkError interface {
	Error() string
	VixErrorCode() uint64
}

// 定义 VixDiskLibCreateParams 结构，表示创建磁盘的参数
type vddkErrorImpl struct {
	err_code uint64
	err_msg  string
}

// 该结构用于定义创建磁盘时的参数。
type VixDiskLibCreateParams struct {
	diskType    VixDiskLibDiskType
	adapterType VixDiskLibAdapterType
	hwVersion   uint16
	capacity    VixDiskLibSectorType
}

// VixDiskLib Block 是底层 C 类型的 Go 类型。
// 该结构用于表示虚拟磁盘上的数据块。
type VixDiskLibBlock C.VixDiskLibBlock

// 返回块的偏移量（以扇区为单位）。
func (b VixDiskLibBlock) Offset() VixDiskLibSectorType {
	return VixDiskLibSectorType(b.offset)
}

// 用于设置块的偏移量。
func (b *VixDiskLibBlock) SetOffset(offset VixDiskLibSectorType) {
	b.offset = C.VixDiskLibSectorType(offset)
}

// 返回块的长度（以扇区为单位）。
func (b VixDiskLibBlock) Length() VixDiskLibSectorType {
	return VixDiskLibSectorType(b.length)
}

// 方法用于设置块的长度。
func (b *VixDiskLibBlock) SetLength(length VixDiskLibSectorType) {
	b.length = C.VixDiskLibSectorType(length)
}

// 该结构用于表示虚拟磁盘的几何信息，包括圆柱数、磁头数和扇区数。
type VixDiskLibGeometry struct {
	Cylinders uint32
	Heads     uint32
	Sectors   uint32
}

// 该结构用于表示虚拟磁盘的信息，包括 BIOS 几何信息、物理几何信息、容量、适配器类型、链接数、父文件名提示和 UUID。
type VixDiskLibInfo struct {
	BiosGeo            VixDiskLibGeometry
	PhysGeo            VixDiskLibGeometry
	Capacity           VixDiskLibSectorType
	AdapterType        VixDiskLibAdapterType
	NumLinks           int
	ParentFileNameHint string
	Uuid               string
}

// 错误信息
func (this vddkErrorImpl) Error() string {
	return this.err_msg
}

// 错误码
func (this vddkErrorImpl) VixErrorCode() uint64 {
	return this.err_code
}

// 该函数用于创建 ConnectParams 结构，包括连接虚拟机所需的参数。
func NewConnectParams(vmxSpec string, serverName string, thumbPrint string, userName string, password string,
	fcdId string, ds string, fcdssId string, cookie string, identity string, path string, flag uint32, readOnly bool, mode string) ConnectParams {
	params := ConnectParams{
		vmxSpec:    vmxSpec,
		serverName: serverName,
		thumbPrint: thumbPrint,
		userName:   userName,
		password:   password,
		fcdId:      fcdId,
		ds:         ds,
		fcdssId:    fcdssId,
		cookie:     cookie,
		identity:   identity,
		path:       path,
		flag:       flag,
		readOnly:   readOnly,
		mode:       mode,
	}
	return params
}

// 该函数用于创建 VddkError 接口的实现，表示VDDK错误。
func NewVddkError(err_code uint64, err_msg string) VddkError {
	vddkError := vddkErrorImpl{
		err_code: err_code,
		err_msg:  err_msg,
	}
	return vddkError
}

// 该函数用于创建 VixDiskLibCreateParams 结构，包括创建磁盘所需的参数。
func NewCreateParams(diskType VixDiskLibDiskType, adapterType VixDiskLibAdapterType, hwVersion uint16, capacity VixDiskLibSectorType) VixDiskLibCreateParams {
	params := VixDiskLibCreateParams{
		diskType:    diskType,
		adapterType: adapterType,
		hwVersion:   hwVersion,
		capacity:    capacity,
	}
	return params
}

// 该函数用于从URL中获取服务器的证书指纹。
func GetThumbPrintForURL(url url.URL) (string, error) {
	return GetThumbPrintForServer(url.Hostname(), url.Port())
}

/*
 * 检索 TLS 服务器的“指纹”或“指纹”。 打开与指定的禁用安全性的服务器/端口的 TLS 连接，
 * 检索证书链并将指纹计算为服务器证书的 SHA-1 哈希值。 对于更高的安全性用途，允许用户
 * 指定指纹而不是自动检索它。
 */
 // 该函数用于检索TLS服务器的证书指纹。
func GetThumbPrintForServer(host string, port string) (string, error) {
	var address string
	if port != "" {
		address = host + ":" + port
	} else {
		address = host
	}
	config := tls.Config{
		InsecureSkipVerify: true, // Skip verify so we can get the thumbprint from any server
	}
	conn, err := tls.Dial("tcp", address, &config)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	peerCerts := conn.ConnectionState().PeerCertificates
	if len(peerCerts) > 0 {
		sha1 := sha1.New()
		sha1.Write(peerCerts[0].Raw)
		sha1Bytes := sha1.Sum(nil)
		var thumbPrint string = ""
		for _, curByte := range sha1Bytes {
			if thumbPrint != "" {
				thumbPrint = thumbPrint + ":"
			}

			thumbPrint = thumbPrint + fmt.Sprintf("%02X", curByte)
		}
		return thumbPrint, nil
	} else {
		return "", fmt.Errorf("no certs returned for " + host + ":" + port)
	}
}
