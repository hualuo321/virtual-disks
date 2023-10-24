package disklib
// #cgo 指令用于配置与C语言的互操作性
// LDFLAGS 用于指定链接器标志，CFLAGS 用于指定编译器标志

// #cgo LDFLAGS: -L/usr/local/vmware-vix-disklib-distrib/lib64 -lvixDiskLib
// #cgo CFLAGS: -I/usr/local/vmware-vix-disklib-distrib/include
// #include "gvddk_c.h"
import "C"
import (
	"fmt"
	"unsafe"
)

//export GoLogWarn 是一个导出的C函数，用于在Go中记录警告信息
//export GoLogWarn
func GoLogWarn(buf *C.char) {
	fmt.Println(C.GoString(buf))
}

// Init 函数用于初始化虚拟磁盘库（虚拟磁盘库主版本号，次版本号，库路径）
func Init(majorVersion uint32, minorVersion uint32, dir string) VddkError {
	// 将 Go 字符串转换为 C 字符串
	libDir := C.CString(dir)
	// 延迟，用于释放字符串内存
	defer C.free(unsafe.Pointer(libDir))
	// 调用 C 库中的初始化函数，返回错误码
	result := C.Init(C.uint32(majorVersion), C.uint32(minorVersion), libDir)
	if result != 0 {
		return NewVddkError(uint64(result), fmt.Sprintf("Initialize failed. The error code is %d.", result))
	}
	return nil
}

// InitEx 函数类似于 Init，但还接受配置文件作为参数（虚拟磁盘库主版本号，次版本号，库路径，配置文件路径）
func InitEx(majorVersion uint32, minorVersion uint32, dir string, configFile string) VddkError {
	var result C.VixError
	libDir := C.CString(dir)
	defer C.free(unsafe.Pointer(libDir))
	if configFile == "" {
		// 如果 configFile 为空，则执行与 Init 函数相同的初始化操作。
		result = C.Init(C.uint32(majorVersion), C.uint32(minorVersion), libDir)
	} else {
		// 如果提供了配置文件，调用 C.InitEx 函数执行初始化。
		config := C.CString(configFile)
		defer C.free(unsafe.Pointer(config))
		result = C.InitEx(C.uint32(majorVersion), C.uint32(minorVersion), libDir, config)
	}
	// 判断是否初始化成功
	if result != 0 {
		return NewVddkError(uint64(result), fmt.Sprintf("Initialize failed. The error code is %d.", result))
	}
	return nil
}

// prepareConnectParams 函数用于准备连接虚拟磁盘所需的参数（全局参数）（指向连接参数的指针，全局参数的切片）
func prepareConnectParams(appGlobal ConnectParams) (*C.VixDiskLibConnectParams, []*C.char) {
	// 将 Go 字符串转换为 C 字符串
	vmxSpec := C.CString(appGlobal.vmxSpec)
	serverName := C.CString(appGlobal.serverName)
	thumbPrint := C.CString(appGlobal.thumbPrint)
	userName := C.CString(appGlobal.userName)
	password := C.CString(appGlobal.password)
	fcdId := C.CString(appGlobal.fcdId)
	ds := C.CString(appGlobal.ds)
	fcdssId := C.CString(appGlobal.fcdssId)
	cookie := C.CString(appGlobal.cookie)
	// 将上述 C 字符串添加到切片中
	var cParams = []*C.char{vmxSpec, serverName, thumbPrint, userName, password, fcdId, ds, fcdssId, cookie}
	// 创建一个连接参数结构体的指针 cnxParams，并分配内存
	var cnxParams *C.VixDiskLibConnectParams = C.VixDiskLib_AllocateConnectParams()
	// 根据 appGlobal 中的参数，设置 cnxParams 结构体的各个字段，以构造连接参数。
	if appGlobal.fcdId != "" {
		cnxParams.specType = C.VIXDISKLIB_SPEC_VSTORAGE_OBJECT
		C.Params_helper(cnxParams, fcdId, ds, fcdssId, true, false)
	} else if appGlobal.vmxSpec != "" {
		cnxParams.specType = C.VIXDISKLIB_SPEC_VMX
		cnxParams.vmxSpec = vmxSpec
	}
	cnxParams.thumbPrint = thumbPrint
	cnxParams.serverName = serverName
	if appGlobal.cookie == "" {
		cnxParams.credType = C.VIXDISKLIB_CRED_UID
		C.Params_helper(cnxParams, cookie, userName, password, false, false)
	} else {
		cnxParams.credType = C.VIXDISKLIB_CRED_SESSIONID
		C.Params_helper(cnxParams, cookie, userName, password, false, true)
	}
	// 将 cnxParams 和 cParams 返回，以便在调用方使用它们进行虚拟磁盘连接。
	return cnxParams, cParams
}

// freeParams 函数用于释放 C 字符串数组。（参数切片）
func freeParams(params []*C.char) {
	for i, _ := range params {
		C.free(unsafe.Pointer(params[i]))
	}
	return
}

// Connect 函数用于连接虚拟磁盘。（连接参数）（虚拟磁盘连接信息对象，错误码）
func Connect(appGlobal ConnectParams) (VixDiskLibConnection, VddkError) {
	var connection VixDiskLibConnection
	// 准备连接虚拟磁盘所需的参数。这个函数会返回指向连接参数的指针 cnxParams 和全局参数的切片 toFree。
	cnxParams, toFree := prepareConnectParams(appGlobal)
	defer freeParams(toFree)
	// 利用连接参数 cnxParams ，连接对象的指针 &connection.conn 尝试连接
	err := C.Connect(cnxParams, &connection.conn)
	if err != 0 {
		return VixDiskLibConnection{}, NewVddkError(uint64(err), fmt.Sprintf("Connect failed. The error code is %d.", err))
	}
	// 如果连接成功，则返回一个 VixDiskLibConnection 对象，其中包含了连接的相关信息。
	return connection, nil
}

// ConnectEx 函数类似于 Connect，但还接受连接模式作为参数。（连接参数）（虚拟磁盘连接信息对象，错误码）
func ConnectEx(appGlobal ConnectParams) (VixDiskLibConnection, VddkError) {
	var connection VixDiskLibConnection
	cnxParams, toFree := prepareConnectParams(appGlobal)
	defer freeParams(toFree)
	modes := C.CString(appGlobal.mode)
	defer C.free(unsafe.Pointer(modes))
	// 传递连接参数 cnxParams、只读标志 readOnly、连接模式 modes，以及连接对象的指针 &connection.conn。
	err := C.ConnectEx(cnxParams, C._Bool(appGlobal.readOnly), modes, &connection.conn)
	if err != 0 {
		return VixDiskLibConnection{}, NewVddkError(uint64(err), fmt.Sprintf("ConnectEx failed. The error code is %d.", err))
	}
	// 如果连接成功，则返回一个 VixDiskLibConnection 对象，其中包含了连接的相关信息。
	return connection, nil
}

// PrepareForAccess 准备虚拟磁盘以进行访问。（全局参数）
func PrepareForAccess(appGlobal ConnectParams) VddkError {
	// 将 Go 字符串转换为 C 字符串
	name := C.CString(appGlobal.identity)
	defer C.free(unsafe.Pointer(name))
	// 获取连接参数
	cnxParams, toFree := prepareConnectParams(appGlobal)
	defer freeParams(toFree)
	// 调用 C 库中的 PrepareForAccess 函数
	result := C.PrepareForAccess(cnxParams, name)
	if result != 0 {
		return NewVddkError(uint64(result), fmt.Sprintf("Prepare for access failed. The error code is %d.", result))
	}
	// 准备访问操作成功
	return nil
}

// open 打开虚拟磁盘。（虚拟磁盘连接信息，连接参数）（虚拟磁盘句柄）
func Open(conn VixDiskLibConnection, params ConnectParams) (VixDiskLibHandle, VddkError) {
	// 虚拟磁盘句柄，用于表示对虚拟磁盘文件的操作，如打开、读取、写入、关闭等。
	var dli VixDiskLibHandle
	filePath := C.CString(params.path)
	defer C.free(unsafe.Pointer(filePath))
	// 调用 C 库中的 Open 函数
	res := C.Open(conn.conn, filePath, C.uint32(params.flag))
	dli.dli = res.dli
	if res.err != 0 {
		return dli, NewVddkError(uint64(res.err), fmt.Sprintf("Open virtual disk file failed. The error code is %d.", res.err))
	}
	return dli, nil
}

// 结束虚拟磁盘的访问。
func EndAccess(appGlobal ConnectParams) VddkError {
	name := C.CString(appGlobal.identity)
	defer C.free(unsafe.Pointer(name))
	cnxParams, toFree := prepareConnectParams(appGlobal)
	// 调用 C 库中的 VixDiskLib_EndAccess 函数
	result := C.VixDiskLib_EndAccess(cnxParams, name)
	freeParams(toFree)
	if result != 0 {
		return NewVddkError(uint64(result), fmt.Sprintf("End access failed. The error code is %d.", result))
	}
	return nil
}

// 断开虚拟磁盘连接。
func Disconnect(connection VixDiskLibConnection) VddkError {
	res := C.VixDiskLib_Disconnect(connection.conn)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Disconnect failed. The error code is %d.", res))
	}
	return nil
}

// 退出虚拟磁盘库。
func Exit() {
	C.VixDiskLib_Exit()
}

// 将子磁盘链附加到父磁盘链。
func Attach(childHandle VixDiskLibHandle, parentHandle VixDiskLibHandle) VddkError {
	res := C.VixDiskLib_Attach(childHandle.dli, parentHandle.dli)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Attach child disk chain to the parent disk chain failed. The error code is %d.", res))
	}
	return nil
}

// 检查或修复虚拟磁盘文件。
func CheckRepair(connection VixDiskLibConnection, filename string, repair bool) VddkError {
	file := C.CString(filename)
	defer C.free(unsafe.Pointer(file))
	res := C.CheckRepair(connection.conn, file, C._Bool(repair))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Check repair failed. The error code is %d.", res))
	}
	return nil
}

// 清理虚拟磁盘连接。
func Cleanup(appGlobal ConnectParams, numCleanUp uint32, numRemaining uint32) VddkError {
	cnxParams, toFree := prepareConnectParams(appGlobal)
	defer freeParams(toFree)
	res := C.Cleanup(cnxParams, C.uint32(numCleanUp), C.uint32(numRemaining))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Clean up failed. The error code is %d.", res))
	}
	return nil
}

// 克隆虚拟磁盘。（目标虚拟磁盘的连接，目标虚拟磁盘的路径，源虚拟磁盘的连接，源虚拟磁盘的路径，虚拟磁盘的创建参数，进度回调数据，是否覆盖目标虚拟磁盘）
func Clone(dstConnection VixDiskLibConnection, dstPath string, srcConnection VixDiskLibConnection, srcPath string,
	params VixDiskLibCreateParams, progressCallbackData string, overWrite bool) VddkError {
	dst := C.CString(dstPath)
	defer C.free(unsafe.Pointer(dst))
	src := C.CString(srcPath)
	defer C.free(unsafe.Pointer(src))
	createParams := prepareCreateParams(params)
	cstr := C.CString(progressCallbackData)
	defer C.free(unsafe.Pointer(cstr))
	res := C.Clone(dstConnection.conn, dst, srcConnection.conn, src, createParams, unsafe.Pointer(&cstr), C._Bool(overWrite))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Clone a virtual disk failed. The error code is %d.", res))
	}
	return nil
}

// 准备虚拟磁盘的创建参数。（包含虚拟磁盘信息的参数结构体）
func prepareCreateParams(createSpec VixDiskLibCreateParams) *C.VixDiskLibCreateParams {
	var createParams *C.VixDiskLibCreateParams
	createParams.diskType = C.VixDiskLibDiskType(createSpec.diskType)
	createParams.adapterType = C.VixDiskLibAdapterType(createSpec.adapterType)
	createParams.hwVersion = C.uint16(createSpec.hwVersion)
	createParams.capacity = C.VixDiskLibSectorType(createSpec.capacity)
	return createParams
}

// 创建虚拟磁盘。
func Create(connection VixDiskLibConnection, path string, createParams VixDiskLibCreateParams, progressCallbackData string) VddkError {
	pathName := C.CString(path)
	defer C.free(unsafe.Pointer(pathName))
	// 准备虚拟磁盘的创建参数
	createSpec := prepareCreateParams(createParams)
	cstr := C.CString(progressCallbackData)
	defer C.free(unsafe.Pointer(cstr))
	// 创建虚拟磁盘
	res := C.Create(connection.conn, pathName, createSpec, unsafe.Pointer(&cstr))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Create a virtual disk failed. The error code is %d.", res))
	}
	return nil
}

// 创建虚拟磁盘的子磁盘。（虚拟磁盘的句柄，子磁盘的路径，子磁盘的类型，进度回调数据）
func CreateChild(diskHandle VixDiskLibHandle, childPath string, diskType VixDiskLibDiskType, progressCallbackData string) VddkError {
	child := C.CString(childPath)
	defer C.free(unsafe.Pointer(child))
	cstr := C.CString(progressCallbackData)
	defer C.free(unsafe.Pointer(cstr))
	res := C.CreateChild(diskHandle.dli, child, C.VixDiskLibDiskType(diskType), unsafe.Pointer(&cstr))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Create child virtual disk failed. The error code is %d.", res))
	}
	return nil
}

// 创建虚拟磁盘信息的 C 语言结构体。
func createDiskInfo(diskInfo *VixDiskLibInfo) (*C.VixDiskLibInfo, []*C.char) {
	var dliInfo *C.VixDiskLibInfo
	var bios C.VixDiskLibGeometry
	var phys C.VixDiskLibGeometry
	bios.cylinders = C.uint32(diskInfo.BiosGeo.Cylinders)
	bios.heads = C.uint32(diskInfo.BiosGeo.Heads)
	bios.sectors = C.uint32(diskInfo.BiosGeo.Sectors)
	phys.cylinders = C.uint32(diskInfo.PhysGeo.Cylinders)
	phys.heads = C.uint32(diskInfo.PhysGeo.Heads)
	phys.sectors = C.uint32(diskInfo.PhysGeo.Sectors)
	dliInfo.biosGeo = bios
	dliInfo.physGeo = phys
	dliInfo.capacity = C.VixDiskLibSectorType(diskInfo.Capacity)
	dliInfo.adapterType = C.VixDiskLibAdapterType(diskInfo.AdapterType)
	dliInfo.numLinks = C.int(diskInfo.NumLinks)
	dliInfo.parentFileNameHint = C.CString(diskInfo.ParentFileNameHint)
	dliInfo.uuid = C.CString(diskInfo.Uuid)
	var cParams = []*C.char{dliInfo.parentFileNameHint, dliInfo.uuid}
	return dliInfo, cParams
}

// 扩展虚拟磁盘的容量。
func Grow(connection VixDiskLibConnection, path string, capacity VixDiskLibSectorType, updateGeometry bool, callbackData string) VddkError {
	filePath := C.CString(path)
	defer C.free(unsafe.Pointer(filePath))
	cstr := C.CString(callbackData)
	defer C.free(unsafe.Pointer(cstr))
	res := C.Grow(connection.conn, filePath, C.VixDiskLibSectorType(capacity), C._Bool(updateGeometry), unsafe.Pointer(&cstr))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Grow failed. The error code is %d.", res))
	}
	return nil
}

// 列出支持的传输模式。
func ListTransportModes() string {
	res := C.VixDiskLib_ListTransportModes()
	modes := C.GoString(res)
	return modes
}

// 重命名虚拟磁盘文件。
func Rename(srcFileName string, dstFileName string) VddkError {
	src := C.CString(srcFileName)
	defer C.free(unsafe.Pointer(src))
	dst := C.CString(dstFileName)
	defer C.free(unsafe.Pointer(dst))
	res := C.VixDiskLib_Rename(src, dst)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Rename failed. The error code is %d.", res))
	}
	return nil
}

// 获取克隆操作所需的空间大小。
func SpaceNeededForClone(srcHandle VixDiskLibHandle, diskType VixDiskLibDiskType, spaceNeeded uint64) VddkError {
	space := C.uint64(spaceNeeded)
	res := C.VixDiskLib_SpaceNeededForClone(srcHandle.dli, C.VixDiskLibDiskType(diskType), &space)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Get space needed for clone failed. The error code is %d.", res))
	}
	return nil
}

// 删除虚拟磁盘文件，包括所有的扩展。
func Unlink(connection VixDiskLibConnection, path string) VddkError {
	delete := C.CString(path)
	defer C.free(unsafe.Pointer(delete))
	res := C.VixDiskLib_Unlink(connection.conn, delete)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Delete the virtual disk including all the extents failed. The error code is %d.", res))
	}
	return nil
}

// 收缩虚拟磁盘的容量。
func Shrink(diskHandle VixDiskLibHandle, progressCallbackData string) VddkError {
	cstr := C.CString(progressCallbackData)
	defer C.free(unsafe.Pointer(cstr))
	res := C.Shrink(diskHandle.dli, unsafe.Pointer(&cstr))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Shrink failed. The error code is %d.", res))
	}
	return nil
}

// 对虚拟磁盘执行碎片整理。
func Defragment(diskHandle VixDiskLibHandle, progressCallbackData string) VddkError {
	cstr := C.CString(progressCallbackData)
	defer C.free(unsafe.Pointer(cstr))
	res := C.Defragment(diskHandle.dli, unsafe.Pointer(&cstr))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Defragment failed. The error code is %d.", res))
	}
	return nil
}

// 获取虚拟磁盘的传输模式。
func GetTransportMode(diskHandle VixDiskLibHandle) string {
	res := C.VixDiskLib_GetTransportMode(diskHandle.dli)
	mode := C.GoString(res)
	return mode
}

// 获取虚拟磁盘的元数据键。
func GetMetadataKeys(diskHandle VixDiskLibHandle, buf []byte, bufLen uint, requireLen uint) VddkError {
	cbuf := ((*C.char)(unsafe.Pointer(&buf[0])))
	res := C.GetMetadataKeys(diskHandle.dli, cbuf, C.size_t(bufLen), C.size_t(requireLen))
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("GetMetadataKeys failed. The error code is %d.", res))
	}
	return nil
}

// 关闭虚拟磁盘句柄，释放相关资源。
func Close(diskHandle VixDiskLibHandle) VddkError {
	res := C.VixDiskLib_Close(diskHandle.dli)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Close virtual disk failed. The error code is %d.", res))
	}
	return nil
}

// 写入虚拟磁盘的元数据。
func WriteMetadata(diskHandle VixDiskLibHandle, key string, val string) VddkError {
	w_key := C.CString(key)
	defer C.free(unsafe.Pointer(w_key))
	w_val := C.CString(val)
	defer C.free(unsafe.Pointer(w_val))
	res := C.VixDiskLib_WriteMetadata(diskHandle.dli, w_key, w_val)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Write meta data failed. The error code is %d.", res))
	}
	return nil
}

// 从虚拟磁盘中读取元数据。
func ReadMetadata(diskHandle VixDiskLibHandle, key string, buf []byte, bufLen uint, requiredLen uint) VddkError {
	readKey := C.CString(key)
	defer C.free(unsafe.Pointer(readKey))
	cbuf := ((*C.char)(unsafe.Pointer(&buf[0])))
	required := C.size_t(requiredLen)
	res := C.VixDiskLib_ReadMetadata(diskHandle.dli, readKey, cbuf, C.size_t(bufLen), &required)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Read meta data from virtual disk file failed. The error code is %d.", res))
	}
	return nil
}

// 从虚拟磁盘中读取数据。
func Read(diskHandle VixDiskLibHandle, startSector uint64, numSectors uint64, buf []byte) VddkError {
	cbuf := ((*C.uint8)(unsafe.Pointer(&buf[0])))
	res := C.VixDiskLib_Read(diskHandle.dli, C.VixDiskLibSectorType(startSector), C.VixDiskLibSectorType(numSectors), cbuf)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Read from virtual disk file failed. The error code is %d.", res))
	}
	return nil
}

// 向虚拟磁盘中写入数据。
func Write(diskHandle VixDiskLibHandle, startSector uint64, numSectors uint64, buf []byte) VddkError {
	cbuf := ((*C.uint8)(unsafe.Pointer(&buf[0])))
	res := C.VixDiskLib_Write(diskHandle.dli, C.VixDiskLibSectorType(startSector), C.VixDiskLibSectorType(numSectors), cbuf)
	if res != 0 {
		return NewVddkError(uint64(res), fmt.Sprintf("Write to virtual disk file failed. The error code is %d.", res))
	}
	return nil
}

// 获取虚拟磁盘的信息，如容量、几何信息等。
func GetInfo(diskHandle VixDiskLibHandle) (VixDiskLibInfo, VddkError) {
	var dliInfoPtr *C.VixDiskLibInfo
	res := C.VixDiskLib_GetInfo(diskHandle.dli, &dliInfoPtr)
	if res != 0 {
		return VixDiskLibInfo{}, NewVddkError(uint64(res), fmt.Sprintf("GetInfo failed. The error code is %d.", res))
	}
	dliInfo := *dliInfoPtr
	retInfo := VixDiskLibInfo{
		BiosGeo: VixDiskLibGeometry{
			Cylinders: uint32(dliInfo.biosGeo.cylinders),
			Heads:     uint32(dliInfo.biosGeo.heads),
			Sectors:   uint32(dliInfo.biosGeo.sectors),
		},
		PhysGeo: VixDiskLibGeometry{
			Cylinders: uint32(dliInfo.physGeo.cylinders),
			Heads:     uint32(dliInfo.physGeo.heads),
			Sectors:   uint32(dliInfo.physGeo.sectors),
		},
		Capacity:           VixDiskLibSectorType(dliInfo.capacity),
		AdapterType:        VixDiskLibAdapterType(dliInfo.adapterType),
		NumLinks:           int(dliInfo.numLinks),
		ParentFileNameHint: C.GoString(dliInfo.parentFileNameHint),
		Uuid:               C.GoString(dliInfo.uuid),
	}
	C.VixDiskLib_FreeInfo(dliInfoPtr)
	return retInfo, nil
}

// 查询虚拟磁盘中已分配的块。
func QueryAllocatedBlocks(diskHandle VixDiskLibHandle, startSector VixDiskLibSectorType, numSectors VixDiskLibSectorType, chunkSize VixDiskLibSectorType) ([]VixDiskLibBlock, VddkError) {
	ss := C.VixDiskLibSectorType(startSector)
	ns := C.VixDiskLibSectorType(numSectors)
	cs := C.VixDiskLibSectorType(chunkSize)
	var bld C.BlockListDescriptor

	res := C.QueryAllocatedBlocks(diskHandle.dli, ss, ns, cs, &bld)
	if res != 0 {
		return nil, NewVddkError(uint64(res), fmt.Sprintf("QueryAllocatedBlocks(%d, %d, %d) error: %d.", startSector, numSectors, chunkSize, res))
	}

	retList := make([]VixDiskLibBlock, bld.numBlocks)
	C.BlockListCopyAndFree(&bld, (*C.VixDiskLibBlock)(unsafe.Pointer(&retList[0])))

	return retList, nil
}
