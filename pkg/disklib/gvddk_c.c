#include "gvddk_c.h"
#include <string.h>

// 日志记录函数
void LogFunc(const char *fmt, va_list args)
{
    char *buf = malloc(1024);           // 这里创建了一个缓冲区并使用 sprintf 格式化日志消息。
    sprintf(buf, fmt, args);
    GoLogWarn(buf);                     // GoLogWarn 应该是 Go 中的记录日志的函数。
}

// ProgressFunc 函数用于处理进度回调。在此只是简单地返回true。
bool ProgressFunc(void *progressData, int percentCompleted)
{
    return true;
}

// Init函数用于初始化VixDiskLib库。
VixError Init(uint32 major, uint32 minor, char* libDir)
{
    VixError result = VixDiskLib_Init(major, minor, NULL, NULL, NULL, libDir);
    return result;
}

// 带额外信息的初始化
VixError InitEx(uint32 major, uint32 minor, char* libDir, char* configFile)
{
    VixError result = VixDiskLib_InitEx(major, minor, NULL, NULL, NULL, libDir, configFile);
    return result;
}

// 连接虚拟磁盘
VixError Connect(VixDiskLibConnectParams *cnxParams, VixDiskLibConnection *connection) {
    VixError vixError;
    vixError = VixDiskLib_Connect(cnxParams, connection);
    return vixError;
}

// 带额外信息的连接虚拟磁盘
VixError ConnectEx(VixDiskLibConnectParams *cnxParams, bool readOnly, char* transportModes, VixDiskLibConnection *connection) {
    VixError vixError;
    vixError = VixDiskLib_ConnectEx(cnxParams, readOnly, "", transportModes, connection);
    return vixError;
}

// 打开磁盘
DiskHandle Open(VixDiskLibConnection conn, char* path, uint32 flags)
{
    VixDiskLibHandle diskHandle;
    VixError vixError;
    vixError = VixDiskLib_Open(conn, path, flags, &diskHandle);
    DiskHandle myDli;
    myDli.dli = diskHandle;
    myDli.err = vixError;
    return myDli;
}

// 为连接磁盘准备数据
VixError PrepareForAccess(VixDiskLibConnectParams *cnxParams, char* identity)
{
    return VixDiskLib_PrepareForAccess(cnxParams, identity);
}

// 参数帮助
void Params_helper(VixDiskLibConnectParams *cnxParams, char* arg1, char* arg2, char* arg3, bool isFcd, bool isSession) {
    if (isFcd)
    {
        cnxParams->spec.vStorageObjSpec.id = arg1;
        cnxParams->spec.vStorageObjSpec.datastoreMoRef = arg2;
        if (strlen(arg3) != 0) {
            cnxParams->spec.vStorageObjSpec.ssId = arg3;
        }
    } else {
        if (isSession)
        {
            cnxParams->creds.sessionId.cookie = arg1;
            cnxParams->creds.sessionId.userName = arg2;
            cnxParams->creds.sessionId.key = arg3;
        } else {
            cnxParams->creds.uid.userName = arg2;
            cnxParams->creds.uid.password = arg3;
        }
    }
    return;
}

VixError Create(VixDiskLibConnection connection, char *path, VixDiskLibCreateParams *createParams, void *progressCallbackData)
{
    VixError vixError;
    vixError = VixDiskLib_Create(connection, path, createParams, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return vixError;
}

VixError CreateChild(VixDiskLibHandle diskHandle, char *childPath, VixDiskLibDiskType diskType, void *progressCallbackData)
{
    VixError vixError;
    vixError = VixDiskLib_CreateChild(diskHandle, childPath, diskType, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return vixError;
}

VixError Defragment(VixDiskLibHandle diskHandle, void *progressCallbackData)
{
    VixError vixError;
    vixError = VixDiskLib_Defragment(diskHandle, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return vixError;
}

VixError GetInfo(VixDiskLibHandle diskHandle, VixDiskLibInfo *info)
{
    VixError error;
    error = VixDiskLib_GetInfo(diskHandle, &info);
    return error;
}

VixError Grow(VixDiskLibConnection connection, char* path, VixDiskLibSectorType capacity, bool updateGeometry, void *progressCallbackData)
{
    VixError error;
    error = VixDiskLib_Grow(connection, path, capacity, updateGeometry, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return error;
}

VixError Shrink(VixDiskLibHandle diskHandle, void *progressCallbackData)
{
    VixError error;
    error = VixDiskLib_Shrink(diskHandle, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData);
    return error;
}

VixError CheckRepair(VixDiskLibConnection connection, char *file, bool repair)
{
    VixError error;
    error = VixDiskLib_CheckRepair(connection, file, repair);
    return error;
}

VixError Cleanup(VixDiskLibConnectParams *connectParams, uint32 numCleanedUp, uint32 numRemaining)
{
    VixError error;
    error = VixDiskLib_Cleanup(connectParams, &numCleanedUp, &numRemaining);
    return error;
}

VixError GetMetadataKeys(VixDiskLibHandle diskHandle, char *buf, size_t bufLen, size_t required)
{
    VixError error;
    error = VixDiskLib_GetMetadataKeys(diskHandle, buf, bufLen, &required);
    return error;
}

VixError Clone(VixDiskLibConnection dstConn, char *dstPath, VixDiskLibConnection srcConn, char *srcPath, VixDiskLibCreateParams *createParams,
               void *progressCallbackData, bool overWrite)
{
    VixError error;
    error = VixDiskLib_Clone(dstConn, dstPath, srcConn, srcPath, createParams, (VixDiskLibProgressFunc)&ProgressFunc, progressCallbackData, overWrite);
    return error;
}

/*
 * QueryAllocatedBlocks wraps the underlying VixDiskLib method of the same name.
 * It accepts an output descriptor data structure allocated in Go in which to return
 * details on the VixDiskLibBlockList on success.  The caller should use
 * BlockListCopyAndFree to copy the data to a Go slice and free the C memory
 * associated with the VixDiskLibBlockList.
 */
VixError QueryAllocatedBlocks(VixDiskLibHandle diskHandle, VixDiskLibSectorType startSector, VixDiskLibSectorType numSectors,
                              VixDiskLibSectorType chunkSize, BlockListDescriptor *bld)
{
    VixError vErr;
    VixDiskLibBlockList *bl;

    vErr = VixDiskLib_QueryAllocatedBlocks(diskHandle, startSector, numSectors, chunkSize, &bl);
    if VIX_FAILED(vErr) {
        return vErr;
    }

    bld->blockList = bl;
    bld->numBlocks = bl->numBlocks;

    return VIX_OK;
}

/*
 * BlockListCopyAndFree copies the block list to the specified array of blocks.
 * It frees the block list on completion and returns the error on free.
 */
VixError BlockListCopyAndFree(BlockListDescriptor *bld,  VixDiskLibBlock *ba)
{
    VixDiskLibBlockList *bl = bld->blockList;

    for (int i = 0; i < bl->numBlocks; i++) {
        *ba = *(&bl->blocks[i]);
        ba++;
    }

    return VixDiskLib_FreeBlockList(bl);
}
