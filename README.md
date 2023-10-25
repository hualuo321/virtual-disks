<a name="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/hualuo321/virtual-disks">
    <img src="images/error404.png" alt="Logo" width="180" height="180">
  </a>
</div>

# Virtual Disks
Go Library for Virtual Disk Development Kit (a.k.a. virtual-disks) 是一个 Golang 包装器，用于访问 VMware 虚拟磁盘开发套件 API (VDDK)，该 API 是一个 SDK，可帮助开发人员创建访问虚拟机上存储的应用程序。

虚拟磁盘为用户提供两级API：

* 低级 API，直接在 Golang 中公开所有 VDDK API。
* 高级 API，为用户提供一些常用的功能，例如 IO 读写。

用户可以选择通过高级 API 使用主要功能，或者使用低级 API 来实现自己的功能组合。

- 虚拟磁盘开发工具包VDDK：是VMWare开发的一组基于虚拟磁盘API的软件开发工具，用于与虚拟环境中的虚拟磁盘进行交互。它提供了一组API和工具，方便开发人员创建，管理以及操作虚拟磁盘。
- 虚拟磁盘API：是一组C函数调用，用于操作VMDK格式（虚拟机磁盘）的虚拟磁盘文件。这些库函数可以操作Workstation上的托管磁盘，或vCenter Server管理的ESXi主机的文件系统卷上的受管磁盘。
- 托管磁盘(hosted)：是指由虚拟机所运行的物理主机上的本地磁盘。这些磁盘是虚拟机使用的一部分主机存储资源。托管磁盘的特点包括性能较好，适用于需要高性能的工作负载，因为它们直接连接到主机硬件。"Hosted" 是一个术语，意味着虚拟化平台和磁盘是由客户操作系统（例如Windows或Linux）托管的。
- 受管磁盘(managed)：是云计算平台（如Azure）提供的一种云存储服务，用于虚拟机的持久性数据存储。受管磁盘由云服务提供商进行管理和维护，用户不需要关心底层硬件和存储设备。
- vsphere：是一个包含多个虚拟化和云计算技术的整体产品组合，由VMware开发。它包括虚拟化服务器（ESXi）、虚拟化存储、虚拟网络和其他关键组件，用于构建和管理虚拟化数据中心。
- vCenter Server：是vSphere产品组合的关键组件之一，用于管理虚拟机、虚拟化主机（ESXi主机）、存储、网络和其他虚拟化资源。
- VMware Workstation：是一个桌面虚拟化产品，用于在个人计算机上创建和运行虚拟机。
- ESXi主机：用于虚拟化服务器硬件，可以同时托管多个虚拟机，每个虚拟机都运行自己的操作系统和应用程序。这些虚拟机相互隔离，共享主机的计算、存储和网络资源。

# Dependency
Virtual-disks 需要 Virtual Disk Development Kit (VDDK) 才能与 vSphere 连接。

可以从此处下载 VDDK：[https://code.vmware.com/web/sdk/7.0/vddk](https://code.vmware.com/web/sdk/7.0/vddk)。虚拟磁盘需要 7.0.0 VDDK 版本。

安装完成后，解压到 `/usr/local/` 目录：
```shell
> cd /usr/local

> sudo tar xzf <path to VMware-vix-disklib-*version*.x86_64.tar.gz>.
```

VDDK 可免费供个人和内部使用。重新分发需要免费许可证，请联系 VMware
获得许可证。

下面列出了所需的 Linux 库包：

Ubuntu:
* libc6
* libssl1.0.2
* libssl1.0.0
* libssl-dev
* libcurl3
* libexpat1-dev
* libffi6
* libgcc1
* libglib2.0-0
* libsqlite3-0
* libstdc++6
* libxml2
* zlib1g

Centos:
* openssl-libs
* libcurl
* expat-devel
* libffi 
* libgcc 
* glib2 
* sqlite
* libstdc++
* libxml2
* zlib

# Use cases
Virtual-disks 提供对 virtual disks 的访问，为应用程序供应商提供一系列用例，包括：
* 备份与虚拟机关联的特定卷或所有卷。
* 获取 vmbk 文件的 IO Reader 和 Writer。
* 将备份代理连接到 vSphere 并备份存储集群上的所有虚拟机。
* 操作 virtual disks 以进行碎片整理、扩展、转换、重命名或缩小文件系统映像。
* 将子磁盘链附加到父磁盘链。

在虚拟机备份中，获取 vmbk 文件的 IO Reader 和 Writer 主要用于读取虚拟机备份数据并将其写入备份存储目标。
1. 读取备份数据（IO Reader）：
	- 读取虚拟机的磁盘镜像数据：IO Reader 会读取虚拟机的磁盘数据，包括虚拟机的主磁盘和任何附加磁盘。
	- 读取虚拟机配置信息：IO Reader 还可能读取虚拟机的配置信息，例如虚拟机的硬件设置、网络配置等。
2. 写入备份数据（IO Writer）：
	- 将备份数据写入备份存储目标：IO Writer 用于将从虚拟机读取的数据写入备份存储设备，例如磁盘、云存储等。
	- 组织备份数据：IO Writer 负责将备份数据按照特定的备份格式或协议进行组织和写入，以确保备份数据的完整性和可还原性。

在虚拟机备份中，将子磁盘链附加到父磁盘链是一种备份策略，通常用于创建完整的虚拟机备份。
1. 创建完整备份：虚拟机通常包含一个主虚拟磁盘和多个附加虚拟磁盘。将子磁盘链附加到父磁盘链可以确保备份包含了虚拟机的所有数据。
2. 数据一致性：通过将子磁盘链附加到父磁盘链，可以确保备份捕获了虚拟机中所有磁盘的一致状态，避免了恢复虚拟机时数据不一致的问题。
3. 简化备份管理：只需要备份一个完整的虚拟机实例，而不是分别备份每个磁盘。
4. 快速还原：当需要还原虚拟机时，备份包含了所有磁盘的完整镜像，可以更快速地还原整个虚拟机。

# Low level API
## Set up
### Init
```$xslt
/**
 * 初始化库。
 * 必须在程序开始时调用，每个进程只应调用一次，应在程序结束时调用 Exit 进行清理。
 */
func Init(majorVersion uint32, minorVersion uint32, dir string) VddkError {}
```

### PrepareForAccess
```$xslt
/**
 * 通知主机不要重新定位虚拟机。
 * 每个 PrepareForAccess 调用都应该有一个匹配的 EndAccess 调用。
 */
func PrepareForAccess(appGlobal ConnectParams) VddkError {}
```

### Connect
```$xslt
/**
 * 将库连接到本地/远程服务器。
 * 始终在程序结束之前调用 Disconnect，这会使任何打开的文件句柄无效。
 * VixDiskLib_PrepareForAccess 应在每次连接之前调用。
 */
func Connect(appGlobal ConnectParams) (VixDiskLibConnection, VddkError) {} 
```

### ConnectEx
```$xslt
/**
 * 创建传输 context 以访问属于特定虚拟机的特定快照的磁盘。
 * 使用此传输 context，调用者可以使用可用于托管虚拟机的最高效的数据访问协议来打开虚拟磁盘，从而提高 I/O 性能。 
 * 如果您使用此调用而不是 Connect()，则应提供额外的输入参数 Transportmode 和 snapshotref。
 */
func ConnectEx(appGlobal ConnectParams) (VixDiskLibConnection, VddkError) {}
```
## Disk operation
### Create a local or remote disk
```$xslt
/**
 * 连接到主机后，在本地创建一个新的虚拟磁盘。
 * 在 createParams 中，您必须指定磁盘类型、适配器、硬件版本和容量（以扇区数表示）。
 * 对于托管磁盘，首先创建托管类型虚拟磁盘，然后使用 Clone() 将虚拟磁盘转换为托管磁盘。
 */
func Create(connection VixDiskLibConnection, path string, createParams VixDiskLibCreateParams, progressCallbackData string) VddkError {}
```
### Open a local or remote disk
库连接到工作站或服务器后，Open 将打开虚拟磁盘。 使用 SAN 或 HotAdd 传输，打开远程磁盘进行写入需要预先存在的快照。使用不同的打开标志来修改打开指令：
* VIXDISKLIB_FLAG_OPEN_UNBUFFERED – 禁用主机磁盘缓存。
* VIXDISKLIB_FLAG_OPEN_SINGLE_LINK – 打开当前链接，而不是整个链（仅限托管磁盘）。
* VIXDISKLIB_FLAG_OPEN_READ_ONLY – 以只读方式打开虚拟磁盘。
* VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB – 开放 NBDSSL 传输、zlib 压缩。
* VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ – 开放 NBDSSL 传输、fastlz 压缩。
* VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ – 开放 NBDSSL 传输、skipz 压缩。
应该有一个匹配的 VixDiskLib Close。
```$xslt
/**
 * 打开虚拟磁盘。
 */
func Open(conn VixDiskLibConnection, params ConnectParams) (VixDiskLibHandle, VddkError) {}
```

托管磁盘与非托管磁盘：
1. 管理：
	- 托管磁盘是由云服务提供商（如 Azure，Google Cloud 等）全面托管和管理的。这包括磁盘的创建，备份，快照及恢复等任务。用户只需要关注磁盘的使用，而无需担心低层基础设施。
	- 非托管磁盘需要用户自己来管理。这包括磁盘的创建，备份，快照及恢复等任务。用户需要更多的自定义和维护工作。
2. 部署：
	- 托管磁盘部署简单，只需要选择磁盘类型，容量和性能级别，然后将其附加到虚拟机或云实例。
	- 非托管磁盘需要用户自行创建和管理虚拟磁盘，包括存储账号设置，网络连接和磁盘配置。
3. 备份和快照：
	- 托管磁盘通常具有集成的备份和快照功能，使用户能够轻松创建、管理和还原备份以及生成快照。
	- 非托管磁盘用户需要自己设置备份策略和手动创建磁盘快照。
4. 灾难恢复：
	- 托管磁盘通常支持跨数据中心和区域的数据冗余和灾难恢复功能，提供更高的数据可用性。
	- 非托管磁盘用户需要自行设计和实施跨区域的灾难恢复策略。

### Read and Write disk IO
```$xslt
/**
 * 此函数从打开的虚拟磁盘读取一系列扇区。
 */
func Read(diskHandle VixDiskLibHandle, startSector uint64, numSectors uint64, buf []byte) VddkError {}
```
```$xslt
/**
 * 此函数写入打开的虚拟磁盘。
 */
func Write(diskHandle VixDiskLibHandle, startSector uint64, numSectors uint64, buf []byte) VddkError {}
```
### Metadata handling
```$xslt
/**
 * 从磁盘读取元数据键。
 */
func ReadMetadata(diskHandle VixDiskLibHandle, key string, buf []byte, bufLen uint, requiredLen uint) VddkError {}
```
```$xslt
/**
 * 从磁盘获取元数据表。
 */
func GetMetadataKeys(diskHandle VixDiskLibHandle, buf []byte, bufLen uint, requireLen uint) VddkError {}
```
```$xslt
/**
 * 将元数据表写入磁盘。
 */
func WriteMetadata(diskHandle VixDiskLibHandle, key string, val string) VddkError {}
```

在云计算环境中，当连接虚拟磁盘（无论是托管磁盘还是非托管磁盘）到虚拟机或云实例时，通常可以获取关于这个磁盘的一些元数据（metadata）。这些元数据包括有关磁盘的信息，以帮助使用者管理和使用磁盘。
- 磁盘ID：磁盘在云平台中的唯一标识符，用于标识特定磁盘。
- 磁盘大小：元数据通常包括磁盘的容量，以便您知道磁盘可以存储多少数据。
- 磁盘类型：这指的是磁盘的种类，例如标准硬盘、SSD等。
- 创建时间：您可以查看磁盘创建的日期和时间，以了解磁盘的生命周期。
- 挂载点：如果磁盘已经挂载到虚拟机或云实例上，元数据可能包括挂载点的信息。
- 操作系统信息：有关虚拟机操作系统的信息，帮助确定磁盘是否包含操作系统数据。
- 状态信息：获取有关磁盘的状态信息，如是否在线、是否附加到虚拟机等。
- 备份和快照信息：若磁盘的备份或快照数据可用，元数据可能包括相关信息，如备份策略和快照时间戳。

### Block allocation
```$xslt
/**
 * 确定分配的块。
 */
func QueryAllocatedBlocks(diskHandle VixDiskLibHandle, startSector VixDiskLibSectorType, numSectors VixDiskLibSectorType, chunkSize VixDiskLibSectorType) ([]VixDiskLibBlock, VddkError) {}
```
## Shut down
所有虚拟磁盘 API 应用程序都应在程序结束时调用这些函数。

### Disconnect
```$xslt
/**
 * 关闭连接。与 Connect 匹配。
 */
func Disconnect(connection VixDiskLibConnection) VddkError {}
```
### EndAccess
```$xslt
/**
 * 通知主机虚拟机磁盘已关闭，因此现在可以允许依赖于要关闭的虚拟磁盘的操作，例如 vMotion。
 * 该函数在内部重新启用 vSphere API 方法 RelocateVM_Task。
 */
func EndAccess(appGlobal ConnectParams) VddkError {}
```
### Exit
```$xslt
/** 
 * 释放 VixDiskLib 持有的所有资源。
 */
func Exit() {}
```
# High level API and data structure
## API
### Open
```$xslt
/**
 * 将处理磁盘的设置操作，包括准备访问、连接、打开。如果在设置阶段发生故障，将回滚到初始状态。
 * 返回一个允许对磁盘进行读或写操作的 DiskReaderWriter。
 */
func Open(globalParams disklib.ConnectParams, logger logrus.FieldLogger) (DiskReaderWriter, disklib.VddkError) {}
```
### Read
```$xslt
/**
 * Read 将最多 len(p) 个字节读取到 p 中。它返回读取的字节数 (0 <= n <= len(p)) 以及遇到的任何错误。
 */
func (this DiskReaderWriter) Read(p []byte) (n int, err error) {}
```
### ReadAt
```$xslt
/** 
 * 从给定的偏移量读取。
 */
func (this DiskReaderWriter) ReadAt(p []byte, off int64) (n int, err error) {}
```
### Write
```$xslt
/**
 * Write 将 p 中的 len(p) 个字节写入底层数据流。 它返回从 p (0 <= n <= len(p)) 写入的字节数。
 */
func (this DiskReaderWriter) Write(p []byte) (n int, err error) {}
```
### WriteAt
```$xslt
/**
 * 从给定的偏移量写入。
 */
func (this DiskConnectHandle) WriteAt(p []byte, off int64) (n int, err error) {}
```
### Block allocation
```$xslt
/**
 * 确定分配的块。
 */
func (this DiskReaderWriter) QueryAllocatedBlocks(startSector disklib.VixDiskLibSectorType, numSectors disklib.VixDiskLibSectorType, chunkSize disklib.VixDiskLibSectorType) ([]disklib.VixDiskLibBlock, disklib.VddkError) {}
```
### Close
```$xslt
/**
 * 清理所有持有的资源。最后应该调用该函数。
 */
func (this DiskReaderWriter) Close() error {} 
```

## Data structure
### DiskReaderWriter
```$xslt
type DiskReaderWriter struct {
	readerAt io.ReaderAt
	writerAt io.WriterAt
	closer   io.Closer
	offset  *int64
	mutex    sync.Mutex                                                     
	logger   logrus.FieldLogger
}
```