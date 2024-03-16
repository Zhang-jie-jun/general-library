package VixDiskLibApi

/*
#cgo CFLAGS: -I./vixDiskLib/include
#cgo LDFLAGS: -L./vixDiskLib/lib64 -lvixDiskLib
#include <stdarg.h>
#include <vixDiskLib.h>

typedef void (VixDiskLibGenericLogFunc)(const char *fmt, va_list args);
void LogFunc (char* fmt, va_list args);
void WarnFunc (char* fmt, va_list args);
void PanicFunc (char* fmt, va_list args);
*/
import "C"

import (
    "bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
    "os"
	"unsafe"
)

var VixDiskLibLogPath = "/var/log/vixDiskLib"

//export GoLogFunc
func GoLogFunc(info *C.char){
    fileName := fmt.Sprintf("%s/VixDiskLibInfo.log", VixDiskLibLogPath)
    fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer func() {
		_ = fp.Close()
	}()
 	// 追加写入
	w := bufio.NewWriter(fp)
	_, err = w.WriteString(C.GoString(info))
	if err != nil {
		return
	}
	_ = w.Flush()
	_ = fp.Sync()
}
//export GoWarnFunc
func GoWarnFunc(info *C.char){
    fileName := fmt.Sprintf("%s/VixDiskLibWarn.log", VixDiskLibLogPath)
    fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer func() {
		_ = fp.Close()
	}()
 	// 追加写入
	w := bufio.NewWriter(fp)
	_, err = w.WriteString(C.GoString(info))
	if err != nil {
		return
	}
	_ = w.Flush()
	_ = fp.Sync()
}
//export GoPanicFunc
func GoPanicFunc(info *C.char){
    fileName := fmt.Sprintf("%s/VixDiskLibPanic.log", VixDiskLibLogPath)
    fp, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	defer func() {
		_ = fp.Close()
	}()
 	// 追加写入
	w := bufio.NewWriter(fp)
	_, err = w.WriteString(C.GoString(info))
	if err != nil {
		return
	}
	_ = w.Flush()
	_ = fp.Sync()
}

const (
	AUTO_METHOD    string = "auto"
	NBD_METHOD     string = "nbd"
	HOT_ADD_METHOD string = "hotadd"
	SAN_METHOD     string = "san"
)

type OperationFlags uint32

const (
	// disable host disk caching
	VIXDISKLIB_FLAG_OPEN_UNBUFFERED OperationFlags = C.VIXDISKLIB_FLAG_OPEN_UNBUFFERED
	// don't open parent disk(s)
	VIXDISKLIB_FLAG_OPEN_SINGLE_LINK OperationFlags = C.VIXDISKLIB_FLAG_OPEN_SINGLE_LINK
	// open read-only
	VIXDISKLIB_FLAG_OPEN_READ_ONLY OperationFlags = C.VIXDISKLIB_FLAG_OPEN_READ_ONLY
	// open for nbd with compression by zlib
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB OperationFlags = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_ZLIB
	// open for nbd with compression by fastlz
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ OperationFlags = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_FASTLZ
	// open for nbd with compression by skipz
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ OperationFlags = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ
	// mask for compression algorithms
	VIXDISKLIB_FLAG_OPEN_COMPRESSION_MASK OperationFlags = C.VIXDISKLIB_FLAG_OPEN_COMPRESSION_MASK
)

var isInit = false
var lowVersion = false

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: Init
//@description: Initializes VixDiskLib.
//@param libDir [in] Directory location where dependent libs are located.
//@param configPath [in] Configuration file path in local encoding.
//          configuration files are of the format
//                name = "value"
//          each name/value pair on a separate line. For a detailed
//          description of allowed values, refer to the VixDiskLib
//          documentation.
//@param isLowVersion [in] If the vsphere Version is lower than 6.0, it is true; else, it is false.
//@return: err error
func Init(logPath, libDir, configPath string, isLowVersion bool) error {
    if logPath != ""{
        if string([]byte(logPath)[len(logPath)-1:]) == "/" {
            logPath = logPath[0:len(logPath)-1]
        }
        VixDiskLibLogPath = logPath
    }
	if !isInit {
		dir := C.CString(libDir)
		defer C.free(unsafe.Pointer(dir))
		configFile := C.CString(configPath)
		defer C.free(unsafe.Pointer(configFile))
		ret := C.VixDiskLib_InitEx(C.uint(6), C.uint(7),
			(*C.VixDiskLibGenericLogFunc)(unsafe.Pointer(C.LogFunc)),
			(*C.VixDiskLibGenericLogFunc)(unsafe.Pointer(C.WarnFunc)),
			(*C.VixDiskLibGenericLogFunc)(unsafe.Pointer(C.PanicFunc)),
			dir,
			configFile)
		err := NewError(ret)
		if err != nil {
			return err
		}
		isInit = true
		lowVersion = isLowVersion
	}
	return nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: Exit
//@description: Cleans up VixDiskLib.
func Exit() {
	if isInit {
		C.VixDiskLib_Exit()
		isInit = false
	}
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: GetTransportModesList
//@description: Get a list of transport modes known to VixDiskLib. This list is also the
// 				default used if VixDiskLib_ConnectEx is called with transportModes set to NULL.
//@return: modes string
//@example: "file:san:hotadd:nbd".
func GetTransportModesList() string {
	modes := C.VixDiskLib_ListTransportModes()
	return C.GoString(modes)
}

type ConnectParam struct {
	ServerIp    string
    UserName    string
    PassWord    string 
    ThumbPoint  string
    VmRef       string 
    SnapshotRef string
	TransMode   string
	ReadOnly    bool
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: NewVixDiskLibApi
//@description:
//@param param [in] connectParams [in] NULL if manipulating local disks.
//@param connLocal [in] Whether to manipulating local disks.
//@return: api *VixDiskLibApi
func NewVixDiskLibApi(param *ConnectParam, connLocal bool) (*VixDiskLibApi, error) {
	vixApi := VixDiskLibApi{connLocal: connLocal}
	if !connLocal {
		vixApi.SetParams(param.ServerIp, param.UserName, param.PassWord, param.ThumbPoint, param.VmRef)
		err := vixApi.ConnectEx(param.SnapshotRef, param.TransMode, param.ReadOnly)
		if err != nil {
			return nil, err
		}
		return &vixApi, nil
	} else {
		err := vixApi.Connect()
		if err != nil {
			return nil, err
		}
		return &VixDiskLibApi{}, nil
	}
}

type VixDiskLibApi struct {
	connLocal bool
	cnxParams C.VixDiskLibConnectParams
	connect   C.VixDiskLibConnection
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: SetParams
//@description: Set connect param
func (api *VixDiskLibApi) SetParams(SerIp, UserName, PassWord, ThumbPoint, VmSpace string) {
	api.CleanParams()

	api.cnxParams.credType = C.VIXDISKLIB_CRED_UID

	ip := C.CString(SerIp)
	api.cnxParams.serverName = ip

	api.cnxParams.port = 443

	userName := C.CString(UserName)
	passWord := C.CString(PassWord)
	binary.LittleEndian.PutUint64(api.cnxParams.creds[0:8], uint64(uintptr(unsafe.Pointer(userName))))
	binary.LittleEndian.PutUint64(api.cnxParams.creds[8:16], uint64(uintptr(unsafe.Pointer(passWord))))

	vmSpace := C.CString(VmSpace)
	api.cnxParams.vmxSpec = vmSpace

	thumbPoint := C.CString(ThumbPoint)
	api.cnxParams.thumbPrint = thumbPoint
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: CleanParams
//@description: Clean connect param
func (api *VixDiskLibApi) CleanParams() {
	C.VixDiskLib_Connect(nil, &api.connect)

	if api.cnxParams.vmxSpec != nil {
		C.free(unsafe.Pointer(api.cnxParams.vmxSpec))
		api.cnxParams.vmxSpec = nil
	}
	if api.cnxParams.serverName != nil {
		C.free(unsafe.Pointer(api.cnxParams.serverName))
		api.cnxParams.serverName = nil
	}
	var ptr1 uint64
	buf1 := bytes.NewBuffer(api.cnxParams.creds[0:8])
    err := binary.Read(buf1, binary.LittleEndian, &ptr1)
	if err != nil {
		fmt.Println(err)
	}
	userName := unsafe.Pointer(uintptr(ptr1))
	if userName != nil {
		C.free(userName)
	}
	var ptr2 uint64
	buf2 := bytes.NewBuffer(api.cnxParams.creds[8:16])
	err = binary.Read(buf2, binary.LittleEndian, &ptr2)
	if err != nil {
		fmt.Println(err)
	}
	passWord := unsafe.Pointer(uintptr(ptr2))
	if passWord != nil {
		C.free(passWord)
	}

	if api.cnxParams.thumbPrint != nil {
		C.free(unsafe.Pointer(api.cnxParams.thumbPrint))
		api.cnxParams.thumbPrint = nil
	}
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: Connect
//@description: Connects to a local / remote server.
//@return: err error
func (api *VixDiskLibApi) Connect() error {
	api.Disconnect()
	ret := C.VixDiskLib_Connect(nil, &api.connect)
	err := NewError(ret)
	if err != nil {
		return err
	}
	return nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: ConnectEx
//@description: Create a transport context to access disks belonging to a particular snapshot of a
// particular virtual machine. Using this transport context will enable callers to open virtual
// disks using the most efficient data acces protocol available for managed virtual machines,
// hence getting better I/O performance.
//@return: err error
func (api *VixDiskLibApi) ConnectEx(snapshotRef string, transMode string, readOnly bool) error {
	api.Disconnect()
	ref := C.CString(snapshotRef)
	defer C.free(unsafe.Pointer(ref))
    mode := C.CString(transMode)
	if transMode == AUTO_METHOD {
		C.free(unsafe.Pointer(mode))
		mode = nil
	} else {
		defer C.free(unsafe.Pointer(mode))
	}
	var CReadOnly int
	if readOnly {
		CReadOnly = 1
	} else {
		CReadOnly = 0
	}
	ret := C.VixDiskLib_ConnectEx(&api.cnxParams,
		C.Bool(CReadOnly),
		ref,
		mode,
		&api.connect)
	err := NewError(ret)
	if err != nil {
		return err
	}
	return nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: Disconnect
//@description: Breaks an existing connection.
func (api *VixDiskLibApi) Disconnect() {
	if api.connect != nil {
		C.VixDiskLib_Disconnect(api.connect)
		api.connect = nil
	}
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: PrepareForAccess
//@description: This function is used to notify the host of the virtual machine that the disks of the virtual
// machine will be opened. The host disables operations on the virtual machine that may be adversely affected if
// they are performed while the disks are open by a third party application.
//@return: err error
func (api *VixDiskLibApi) PrepareForAccess() error {
	identity := C.CString("vixDiskLibApi-golang-identity")
	defer C.free(unsafe.Pointer(identity))
	ret := C.VixDiskLib_PrepareForAccess(&api.cnxParams, identity)
	err := NewError(ret)
	if err != nil {
		return err
	}
	return nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: EndAccess
//@description: This function is used to notify the host of a virtual machine that the virtual machine disks are
// closed and that the operations which rely on the virtual machine disks to be closed can now be allowed.
//@return: err error
func (api *VixDiskLibApi) EndAccess() error {
	identity := C.CString("vixDiskLibApi-golang-identity")
	defer C.free(unsafe.Pointer(identity))
	ret := C.VixDiskLib_EndAccess(&api.cnxParams, identity)
	err := NewError(ret)
	if err != nil {
		return err
	}
	return nil
}

type VixDiskHandle struct {
	*VixDiskLibApi
	handle C.VixDiskLibHandle
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: NewDiskHandle
//@description:
//@return: handle *VixDiskHandle
func NewDiskHandle(vixApi *VixDiskLibApi, diskPath string, flags OperationFlags) (*VixDiskHandle, error) {
	path := C.CString(diskPath)
	defer C.free(unsafe.Pointer(path))
	var handle C.VixDiskLibHandle
	ret := C.VixDiskLib_Open(vixApi.connect, path, C.uint(flags), &handle)
	err := NewError(ret)
	if err != nil {
		return nil, err
	}
	diskHandle := VixDiskHandle{VixDiskLibApi: vixApi, handle: handle}
	return &diskHandle, nil
}

func (handle *VixDiskHandle) checkParams() error {
	if handle.connect == nil {
		return NewError(VIX_CONNECT_IS_NIL)
	}
	if handle.handle == nil {
		return NewError(VIX_DISK_HANDLE_IS_NIL)
	}
	return nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: GetCurTransportMode
//@description: Returns a pointer to a static string identifying the transport mode that is used to access the virtual disk's data.
//@return: mode string, err error
func (handle *VixDiskHandle) GetCurTransportMode() (string, error) {
	if err := handle.checkParams(); err != nil {
		return "", err
	}
	mode := C.VixDiskLib_GetTransportMode(handle.handle)
	return C.GoString(mode), nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: GetDiskSize
//@description: Returns the current virtual disk size.
//@return: size int64, err error
func (handle *VixDiskHandle) GetDiskSize() (int64, error) {
	if err := handle.checkParams(); err != nil {
		return 0, err
	}
	var info *C.VixDiskLibInfo
	ret := C.VixDiskLib_GetInfo(handle.handle, &info)
	err := NewError(ret)
	if err != nil {
		return 0, err
	}
	size := int64(info.capacity * 512)
	C.VixDiskLib_FreeInfo(info)
	return size, nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: GetDiskMetadata
//@description: Retrieves the value of a metadata entry corresponding to the supplied key.
//@return: size int64, err error
func (handle *VixDiskHandle) GetDiskMetadata() (map[string]string, error) {
	return nil, nil
}

type BlockInfo struct {
	Start  int64
	Length int64
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: QueryAllocatedBlocks
//@description: Get the blocks allocated.
//@return: blocks []BlockInfo, err error
func (handle *VixDiskHandle) QueryAllocatedBlocks() ([]BlockInfo, error) {
	blockInfos := make([]BlockInfo, 0)
	if !lowVersion {
		capacity, err := handle.GetDiskSize()
		if err != nil {
			return nil, err
		}
        totalSector := capacity / int64(C.VIXDISKLIB_SECTOR_SIZE)
		chunkSize := (C.VixDiskLibSectorType)(C.VIXDISKLIB_MIN_CHUNK_SIZE)
		startSector := (C.VixDiskLibSectorType)(0)
		numChunk := (C.VixDiskLibSectorType)(totalSector / int64(chunkSize))
		for numChunk > 0 {
			numChunkToQuery := (C.VixDiskLibSectorType)(0)
			if numChunk > C.VIXDISKLIB_MAX_CHUNK_NUMBER {
				numChunkToQuery = (C.VixDiskLibSectorType)(C.VIXDISKLIB_MAX_CHUNK_NUMBER)
			} else {
				numChunkToQuery = numChunk
			}

			var blockList *C.VixDiskLibBlockList
			ret := C.VixDiskLib_QueryAllocatedBlocks(handle.handle, startSector, numChunkToQuery*chunkSize, chunkSize, &blockList)
			err := NewError(ret)
			if err != nil {
				return nil, err
			}
			if blockList != nil {
				for i := 0; i < int(blockList.numBlocks); i++ {
					var blockInfo BlockInfo
					blockInfo.Start = int64((blockList.blocks[i].offset) * (C.VIXDISKLIB_SECTOR_SIZE))
					blockInfo.Length = int64((blockList.blocks[i].length) * (C.VIXDISKLIB_SECTOR_SIZE))
					blockInfos = append(blockInfos, blockInfo)
				}
			}

			numChunk -= numChunkToQuery
			startSector += numChunkToQuery * chunkSize

			if blockList != nil {
				C.VixDiskLib_FreeBlockList(blockList)
			}
		}
	}
	return blockInfos, nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: ReadDisk
//@description: Reads a sector range.
//@return: length int64, err error
func (handle *VixDiskHandle) ReadDisk(startSector, numSectors int64, readBuffer []byte) (int64, error) {
	if err := handle.checkParams(); err != nil {
		return 0, err
	}
	if len(readBuffer) == 0 {
		return 0, nil
	}
	ret := C.VixDiskLib_Read(handle.handle,
                             (C.VixDiskLibSectorType)(startSector), 
                             (C.VixDiskLibSectorType)(numSectors),
                             (*C.uint8)(unsafe.Pointer(&readBuffer[0])))
	err := NewError(ret)
	if err != nil {
		return 0, err
	}
    length := numSectors * 512
	if length < int64(len(readBuffer)) {
		return length, io.EOF
	}
	return length, nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: WriteDisk
//@description: Writes a sector range.
//@return: length int64, err error
func (handle *VixDiskHandle) WriteDisk(startSector, numSectors int64, writeBuffer []byte) (int64, error) {
	if err := handle.checkParams(); err != nil {
		return 0, err
	}
	if len(writeBuffer) == 0 {
		return 0, nil
	}
	ret := C.VixDiskLib_Write(handle.handle,
                              (C.VixDiskLibSectorType)(startSector),
                              (C.VixDiskLibSectorType)(numSectors),
                              (*C.uint8)(unsafe.Pointer(&writeBuffer[0])))
	err := NewError(ret)
	if err != nil {
		return 0, err
	}
    length := numSectors * 512
	if length < int64(len(writeBuffer)) {
		return length, io.EOF
	}
	return length, nil
}

//@author: [jick.zhang](zhang.jiejun@outlook.com)
//@function: CloseDisk
//@description: Closes the disk.
func (handle *VixDiskHandle) CloseDisk() {
	if handle.handle != nil {
		C.VixDiskLib_Close(handle.handle)
		handle.handle = nil
	}
}
