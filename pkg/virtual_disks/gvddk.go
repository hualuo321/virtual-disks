package virtual_disks

import "C"
import (
	"fmt"
	"io"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware/virtual-disks/pkg/disklib"
)

// OpenFCD 用于打开一个 FCD 虚拟磁盘。（接受一系列参数来建立与虚拟磁盘的连接）
// 服务器名称、证书指纹、用户名、密码、FCD ID、数据存储、FCD session ID、标志、只读标志、传输模式、访问标识和日志记录器
func OpenFCD(serverName string, thumbPrint string, userName string, password string, fcdId string, fcdssid string, datastore string,
	flags uint32, readOnly bool, transportMode string, identity string, logger logrus.FieldLogger) (DiskReaderWriter, disklib.VddkError) {
	// 创建全局参数对象，包含了连接虚拟磁盘所需的信息
	globalParams := disklib.NewConnectParams("", serverName, thumbPrint, userName, password, fcdId, datastore, 
	fcdssid, "", identity, "", flags, readOnly, transportMode)
	// 调用 Open 函数以实际打开虚拟磁盘，传递全局参数和日志记录器
	return Open(globalParams, logger)
}

// Open 用于打开虚拟磁盘，并建立与虚拟磁盘的连接
func Open(globalParams disklib.ConnectParams, logger logrus.FieldLogger) (DiskReaderWriter, disklib.VddkError) {
	// 调用 PrepareForAccess 函数以准备虚拟磁盘以进行访问
	err := disklib.PrepareForAccess(globalParams)
	if err != nil {
		return DiskReaderWriter{}, err
	}
	// 调用 ConnectEx 函数以建立虚拟磁盘的连接
	conn, err := disklib.ConnectEx(globalParams)
	// 如果打开虚拟磁盘失败，断开连接并结束访问，然后返回错误
	if err != nil {
		disklib.EndAccess(globalParams)
		return DiskReaderWriter{}, err
	}
	// 调用 Open 函数以打开虚拟磁盘
	dli, err := disklib.Open(conn, globalParams)
	if err != nil {
		disklib.Disconnect(conn)
		disklib.EndAccess(globalParams)
		return DiskReaderWriter{}, err
	}
	// 获取虚拟磁盘信息
	info, err := disklib.GetInfo(dli)
	// 如果获取信息失败，断开连接并结束访问，然后返回错误
	if err != nil {
		disklib.Disconnect(conn)
		disklib.EndAccess(globalParams)
		return DiskReaderWriter{}, err
	}
	// 创建虚拟磁盘句柄，包括连接、全局参数、信息
	diskHandle := NewDiskHandle(dli, conn, globalParams, info)
	// 创建并返回一个包装了虚拟磁盘句柄的 DiskReaderWriter
	return NewDiskReaderWriter(diskHandle, logger), nil
}

// DiskReaderWriter 类型表示虚拟磁盘的读写操作对象。
type DiskReaderWriter struct {
	diskHandle DiskConnectHandle	// 虚拟磁盘连接句柄，用于执行读写操作
	offset     *int64				// 当前读写操作的偏移量，可以被多个线程共享
	mutex      *sync.Mutex 			// 用于保护 offset 和数据访问的互斥锁
	logger     logrus.FieldLogger	// 用于记录日志的记录器
}

// Read 方法用于从虚拟磁盘读取数据，将数据读入切片 p 中，并返回读取的字节数。
// 该方法在多线程环境中使用互斥锁来保护读取操作，以确保多个线程不会同时读取相同的数据。
func (this DiskReaderWriter) Read(p []byte) (n int, err error) {
	this.mutex.Lock()												// 添加互斥锁，只有一个携程可以访问资源
	defer this.mutex.Unlock()										// 延迟释放锁
	bytesRead, err := this.diskHandle.ReadAt(p, *this.offset)		// 读取偏移量数据到p中
	*this.offset += int64(bytesRead)								// 更新偏移量的值
	this.logger.Infof("Read returning %d, len(p) = %d, offset=%d\n", bytesRead, len(p), *this.offset)
	return bytesRead, err											// 记录日志并返回
}

// Write 方法用于向虚拟磁盘写入数据，将数据从切片 p 写入虚拟磁盘，并返回写入的字节数。
// 该方法在多线程环境中使用互斥锁来保护写入操作，以确保多个线程不会同时写入相同的数据。
func (this DiskReaderWriter) Write(p []byte) (n int, err error) {
	this.mutex.Lock()												// 添加互斥锁，只有一个携程可以访问资源
	defer this.mutex.Unlock()										// 延迟释放锁
	bytesWritten, err := this.diskHandle.WriteAt(p, *this.offset)	// 读取偏移量数据到p中
	*this.offset += int64(bytesWritten)								// 更新偏移量的值
	this.logger.Infof("Write returning %d, len(p) = %d, offset=%d\n", bytesWritten, len(p), *this.offset)
	return bytesWritten, err										// 记录日志并返回
}

// Seek 方法用于在虚拟磁盘上设置当前的读写位置（偏移量）。
// 它接受一个偏移量和相对位置参数（whence），并返回新的偏移量和可能的错误。
func (this DiskReaderWriter) Seek(offset int64, whence int) (int64, error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	desiredOffset := *this.offset
	switch whence {
	case io.SeekStart:
		desiredOffset = offset
	case io.SeekCurrent:
		desiredOffset += offset
	case io.SeekEnd:
		// 后续完善
		return *this.offset, errors.New("Seek from SeekEnd not implemented")
	}

	if desiredOffset < 0 {
		return 0, errors.New("Cannot seek to negative offset")
	}
	*this.offset = desiredOffset
	return *this.offset, nil
}

// ReadAt 方法用于从虚拟磁盘的指定偏移量处读取数据，并将数据读入切片 p 中。
// 该方法直接调用底层虚拟磁盘连接句柄的 ReadAt 方法来执行读取操作。
func (this DiskReaderWriter) ReadAt(p []byte, off int64) (n int, err error) {
	return this.diskHandle.ReadAt(p, off)
}

// WriteAt 方法用于向虚拟磁盘的指定偏移量处写入数据，将数据从切片 p 写入虚拟磁盘。
// 该方法直接调用底层虚拟磁盘连接句柄的 WriteAt 方法来执行写入操作。
func (this DiskReaderWriter) WriteAt(p []byte, off int64) (n int, err error) {
	return this.diskHandle.WriteAt(p, off)
}

// Close 方法用于关闭虚拟磁盘连接。
// 它直接调用底层虚拟磁盘连接句柄的 Close 方法来执行关闭操作。
func (this DiskReaderWriter) Close() error {
	return this.diskHandle.Close()
}

// QueryAllocatedBlocks 方法用于查询虚拟磁盘上已分配的数据块。
// 它接受起始扇区、扇区数量和块大小作为参数，并返回已分配的数据块信息和可能的错误。
func (this DiskReaderWriter) QueryAllocatedBlocks(startSector disklib.VixDiskLibSectorType, numSectors disklib.VixDiskLibSectorType, chunkSize disklib.VixDiskLibSectorType) ([]disklib.VixDiskLibBlock, disklib.VddkError) {
	return this.diskHandle.QueryAllocatedBlocks(startSector, numSectors, chunkSize)
}

// NewDiskReaderWriter 函数用于创建一个新的虚拟磁盘读写操作对象。
// 它接受虚拟磁盘连接句柄（DiskConnectHandle）和日志记录器（logger）作为参数，
// 并返回一个初始化的 DiskReaderWriter 对象，用于执行虚拟磁盘的读写操作。
func NewDiskReaderWriter(diskHandle DiskConnectHandle, logger logrus.FieldLogger) DiskReaderWriter {
	var offset int64
	offset = 0
	var mutex sync.Mutex
	retVal := DiskReaderWriter{
		diskHandle: diskHandle,
		offset:     &offset,
		mutex:      &mutex,
		logger:     logger,
	}
	return retVal
}

// DiskConnectHandle 类型表示虚拟磁盘连接句柄，用于管理虚拟磁盘的访问和信息。
type DiskConnectHandle struct {
	mutex  *sync.Mutex
	dli    disklib.VixDiskLibHandle
	conn   disklib.VixDiskLibConnection
	params disklib.ConnectParams
	info   disklib.VixDiskLibInfo
}

// NewDiskHandle 函数用于创建一个新的虚拟磁盘连接句柄。
// 它接受虚拟磁盘句柄（dli）、连接句柄（conn）、连接参数（params）和虚拟磁盘信息（info）作为参数。
// 并返回一个初始化的 DiskConnectHandle 对象，用于管理虚拟磁盘的访问和信息。
func NewDiskHandle(dli disklib.VixDiskLibHandle, conn disklib.VixDiskLibConnection, params disklib.ConnectParams,
	info disklib.VixDiskLibInfo) DiskConnectHandle {
	var mutex sync.Mutex
	return DiskConnectHandle{
		mutex:  &mutex,			// 互斥锁，用于保护虚拟磁盘连接句柄的访问
		dli:    dli,			// 虚拟磁盘句柄，用于执行虚拟磁盘操作
		conn:   conn,			// 虚拟磁盘连接句柄，用于建立和维护虚拟磁盘连接
		params: params,			// 连接参数，包括连接信息和认证信息
		info:   info,			// 虚拟磁盘信息，包括大小和属性
	}
}

// mapError 函数用于将 VddkError 转换为标准错误类型，以便处理特定错误情况。
// 它根据 VddkError 中的 VixErrorCode 来映射错误。
func mapError(vddkError disklib.VddkError) error {
	switch vddkError.VixErrorCode() {
	case disklib.VIX_E_DISK_OUTOFRANGE:
		return io.EOF
	default:
		return vddkError
	}
}

// aligned 函数用于检查给定长度和偏移量是否对齐到虚拟磁盘的扇区大小。
// 它用于确保读写操作的对齐性，以提高性能和避免不必要的内部处理。
func aligned(len int, off int64) bool {
	return len%disklib.VIXDISKLIB_SECTOR_SIZE == 0 && off%disklib.VIXDISKLIB_SECTOR_SIZE == 0
}

// ReadAt 方法用于从虚拟磁盘中指定偏移量处读取数据，并将其写入给定的字节切片 p。
// 它接受偏移量（off）和目标字节切片（p）作为参数，并返回读取的字节数以及可能的错误。
func (this DiskConnectHandle) ReadAt(p []byte, off int64) (n int, err error) {
	capacity := this.Capacity()
	// 如果偏移量超出容量，则返回EOF（文件末尾）
	if off >= capacity {
		return 0, io.EOF
	}
	// 如果读取的数据跨越文件末尾，需要将 p 切片截断
	if off+int64(len(p)) > capacity {
		readLen := int32(capacity - off)
		p = p[0:readLen]
	}
	// 计算起始扇区
	startSector := off / disklib.VIXDISKLIB_SECTOR_SIZE
	var total int = 0
	// 如果读取的数据不对齐，需要加锁，以确保读/修改/写操作是原子的
	if !aligned(len(p), off) {
		this.mutex.Lock()
		defer this.mutex.Unlock()
	}
	// 处理偏移量不对齐的部分
	/*
	这段代码处理了读取偏移量不对齐的情况。它首先从虚拟磁盘中读取一个扇区的数据到临时缓冲区 tmpBuf，
	然后根据偏移量和目标切片的长度确定需要复制的数据部分。最后，将复制的数据部分拷贝到目标切片 p 中，
	并更新起始扇区和已读取的总字节数。这确保了不对齐的数据部分也能正确地被处理。
	*/
	if off%disklib.VIXDISKLIB_SECTOR_SIZE != 0 {
		// 创建一个临时缓冲区 tmpBuf，用于读取一个虚拟磁盘扇区的数据
		tmpBuf := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		// 从虚拟磁盘的起始扇区（startSector）读取一个扇区的数据，存储在 tmpBuf 中
		err := disklib.Read(this.dli, (uint64)(startSector), 1, tmpBuf)
		if err != nil {
			return 0, mapError(err)
		}
		// 计算相对于虚拟磁盘扇区的偏移量，以确定要从 tmpBuf 中复制的数据部分
		srcOff := int(off % disklib.VIXDISKLIB_SECTOR_SIZE)
		// 计算要复制的字节数 count，不超过目标字节切片 p 的长度
		count := disklib.VIXDISKLIB_SECTOR_SIZE - srcOff
		if count > len(p) {
			count = len(p)
		}
		// 计算 tmpBuf 中要复制的数据的结束位置 srcEnd
		srcEnd := srcOff + count
		// 从 tmpBuf 中提取 srcOff 到 srcEnd 范围内的数据，然后复制到目标字节切片 p 的开头
		tmpSlice := tmpBuf[srcOff:srcEnd]
		copy(p[:count], tmpSlice)
		// 更新起始扇区，准备处理下一个扇区
		startSector = startSector + 1
		// 更新总字节数（已读取的字节数）
		total = total + count
	}
	// 处理对齐的部分
	/*
	这段代码处理了对齐的数据部分，首先计算需要处理的对齐扇区的数量。然后，它计算目标字节切片 p 中对齐数据的起始和结束偏移量，
	使用 disklib.Read 从虚拟磁盘中读取这些对齐扇区的数据，并将其存储在目标字节切片 p 的指定范围内。最后，它更新起始扇区和
	已读取的总字节数，以确保对齐数据部分被正确处理。
	*/
	numAlignedSectors := (len(p) - total) / disklib.VIXDISKLIB_SECTOR_SIZE
	// 计算需要处理的对齐扇区数量，即目标字节切片 p 中剩余的字节数除以扇区大小
	if numAlignedSectors > 0 {
		// 如果有对齐的扇区需要处理
		// 计算目标字节切片 p 中对齐数据的起始偏移量 desOff 和结束偏移量 desEnd
		desOff := total
		desEnd := total + numAlignedSectors*disklib.VIXDISKLIB_SECTOR_SIZE
		// 从虚拟磁盘的起始扇区（startSector）读取多个对齐扇区的数据，存储在目标字节切片 p 的指定范围中
		err := disklib.Read(this.dli, (uint64)(startSector), (uint64)(numAlignedSectors), p[desOff:desEnd])
		if err != nil {
			return total, mapError(err)
		}
		// 更新起始扇区，准备处理下一个对齐扇区
		startSector = startSector + int64(numAlignedSectors)
		// 更新已读取的总字节数 total，将其增加对齐扇区的字节数
		total = total + numAlignedSectors*disklib.VIXDISKLIB_SECTOR_SIZE
	}
	// 处理剩余的不对齐部分
	if (len(p) - total) > 0 {
		// 如果仍有剩余的字节需要处理
		// 创建一个临时缓冲区 tmpBuf，用于读取一个虚拟磁盘扇区的数据
		tmpBuf := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		// 从虚拟磁盘的起始扇区（startSector）读取一个扇区的数据，存储在 tmpBuf 中
		err := disklib.Read(this.dli, (uint64)(startSector), 1, tmpBuf)
		if err != nil {
			return total, mapError(err)
		}
		// 计算需要复制的字节数 count，即剩余未处理的字节数
		count := len(p) - total
		// 计算 tmpBuf 中要复制的数据的结束位置 srcEnd
		srcEnd := count
		// 从 tmpBuf 中提取 tmpBuf 的前 srcEnd 字节数据，然后复制到目标字节切片 p 的剩余部分
		tmpSlice := tmpBuf[0:srcEnd]
		copy(p[total:], tmpSlice)
	}
	// 返回已读取的总字节数 total
	return total, nil
}

func (this DiskConnectHandle) WriteAt(p []byte, off int64) (n int, err error) {
	// 获取虚拟磁盘的容量，即虚拟磁盘的总扇区数
	capacity := this.Capacity()
	// 如果写操作的起始偏移量（off）或结束偏移量超出虚拟磁盘的容量，返回一个错误（io.ErrShortWrite）
	if off > capacity || off+int64(len(p)) > capacity {
		return 0, io.ErrShortWrite
	}
	// 如果写操作的数据不对齐（不是以扇区大小的倍数开始），需要加锁来确保对不对齐数据的读取、修改和写入的一致性。
	if !aligned(len(p), off) {
		// 加锁，防止多个携程同时访问不对齐数据，确保原子性的读取、修改和写入
		this.mutex.Lock()
		defer this.mutex.Unlock()
	}
	var total int64 = 0		// 总共已写入的字节数
	var srcOff int64 = 0 	// p 中要复制的数据的起始索引
	var srcEnd int64 = 0	// p 中要复制的数据的结束索引
	startSector := off / disklib.VIXDISKLIB_SECTOR_SIZE		// 起始扇区的索引
	// 如果写操作的偏移量（off）不是扇区大小的倍数，说明写操作不对齐，需要特殊处理。
	if off%disklib.VIXDISKLIB_SECTOR_SIZE != 0 {
		// 创建一个临时缓冲区 tmpBuf，用于存储一个虚拟磁盘扇区的数据
		tmpBuf := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		// 从虚拟磁盘读取一个扇区的数据，这是为了获取已存储在虚拟磁盘上的数据，以便后续修改。
		err := disklib.Read(this.dli, uint64(startSector), 1, tmpBuf)
		if err != nil {
			return 0, mapError(err)
		}
		// 计算写操作在扇区中的偏移量
		desOff := off % disklib.VIXDISKLIB_SECTOR_SIZE
		// 计算当前扇区中剩余可写入的字节数
		count := disklib.VIXDISKLIB_SECTOR_SIZE - desOff
		// 如果 p 中的数据不足以填满当前扇区，调整写入的字节数
		if int64(len(p)) < count {
			count = int64(len(p))
		}
		// 计算 p 中要复制的数据的结束索引
		desEnd := desOff + count
		srcEnd = srcOff + count
		// 将 p 中的数据复制到 tmpBuf 的适当位置，实现部分数据的写入
		copy(tmpBuf[desOff:desEnd], p[srcOff:srcEnd])
		// 将修改后的 tmpBuf 数据写回虚拟磁盘的当前扇区
		err = disklib.Write(this.dli, uint64(startSector), 1, tmpBuf)
		if err != nil {
			return 0, mapError(err)
		}
		// 更新下一扇区的索引、已写入的总字节数以及 p 中的数据的起始索引
		startSector = startSector + 1
		total = total + count
		srcOff = srcOff + count
	}
	// Middle aligned part, override directly
	// 如果待写入数据的剩余长度除以扇区大小大于零，说明还有完整的扇区需要直接覆盖写入。
	if (int64(len(p))-total)/disklib.VIXDISKLIB_SECTOR_SIZE > 0 {
		// 计算需要写入的完整扇区数
		numSector := (int64(len(p)) - total) / disklib.VIXDISKLIB_SECTOR_SIZE
		// 计算 p 中待写入数据的结束索引
		srcEnd = srcOff + numSector*disklib.VIXDISKLIB_SECTOR_SIZE
		// 直接将待写入数据 p 中的完整扇区数据写入虚拟磁盘
		err := disklib.Write(this.dli, uint64(startSector), uint64(numSector), p[srcOff:srcEnd])
		if err != nil {
			return int(total), mapError(err)
		}
		// 更新起始扇区索引、已写入的总字节数以及待写入数据 p 中的数据索引
		startSector = startSector + numSector
		total = total + numSector*disklib.VIXDISKLIB_SECTOR_SIZE
		srcOff = srcEnd
	}
	// End missing aligned part
	/*
	这段代码用于处理待写入数据的末尾不对齐部分。首先，它计算了尚未处理的字节数，并计算
	了 p 中待写入数据的结束索引。然后，它创建了一个临时缓冲区 tmpBuf 用于存储一个虚拟
	磁盘扇区的数据。接下来，它使用 disklib.Read 函数从虚拟磁盘读取一个扇区的数据，将待
	写入数据 p 中的数据复制到 tmpBuf 中的适当位置。最后，它使用 disklib.Write 函数将
	修改后的数据写回虚拟磁盘，以完成对不对齐部分的写入。
	*/
	if int64(len(p))-total > 0 {
		// 计算尚未处理的字节数
		count := int64(len(p)) - total
		// 计算待写入数据 p 中的结束索引
		srcEnd = srcOff + count
		// 创建一个临时缓冲区 tmpBuf 用于存储一个虚拟磁盘扇区的数据
		tmpBuf := make([]byte, disklib.VIXDISKLIB_SECTOR_SIZE)
		// 从虚拟磁盘读取一个扇区的数据
		err := disklib.Read(this.dli, uint64(startSector), 1, tmpBuf)
		if err != nil {
			return int(total), mapError(err)
		}
		// 将 p 中的数据复制到 tmpBuf 中的适当位置
		copy(tmpBuf[:count], p[srcOff:srcEnd])
		// 将修改后的数据写回虚拓展磁盘
		err = disklib.Write(this.dli, uint64(startSector), 1, tmpBuf)
		if err != nil {
			return int(total), errors.Wrap(err, "Write into disk in part 3 failed part3.")
		}
	}
	return len(p), nil
}

// Close 关闭虚拟磁盘连接及相关资源。
func (this DiskConnectHandle) Close() error {
	// 尝试关闭虚拟磁盘句柄
	vErr := disklib.Close(this.dli)
	if vErr != nil {
		return errors.New(fmt.Sprintf(vErr.Error()+" with error code: %d", vErr.VixErrorCode()))
	}
	// 尝试断开虚拟磁盘连接
	vErr = disklib.Disconnect(this.conn)
	if vErr != nil {
		return errors.New(fmt.Sprintf(vErr.Error()+" with error code: %d", vErr.VixErrorCode()))
	}
	// 结束虚拟磁盘的访问
	vErr = disklib.EndAccess(this.params)
	if vErr != nil {
		return errors.New(fmt.Sprintf(vErr.Error()+" with error code: %d", vErr.VixErrorCode()))
	}

	return nil
}

// Capacity 返回虚拟磁盘的总容量（以字节为单位）。
func (this DiskConnectHandle) Capacity() int64 {
	return int64(this.info.Capacity) * disklib.VIXDISKLIB_SECTOR_SIZE
}

// QueryAllocatedBlocks 调用 VDDK 中的 QueryAllocatedBlocks 函数以查询虚拟磁盘上的已分配块信息。
func (this DiskConnectHandle) QueryAllocatedBlocks(startSector disklib.VixDiskLibSectorType, numSectors disklib.VixDiskLibSectorType, chunkSize disklib.VixDiskLibSectorType) ([]disklib.VixDiskLibBlock, disklib.VddkError) {
	return disklib.QueryAllocatedBlocks(this.dli, startSector, numSectors, chunkSize)
}
