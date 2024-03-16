package main

import (
	"fmt"
	"github.com/Zhang-jie-jun/general-library/VixDiskLibApi"
	"github.com/Zhang-jie-jun/general-library/vSphereApi"
	"io"
	"os"
	"time"
)

var IP = "192.168.212.52"
var Username = "administrator@vsphere.local"
var Password = "Password@123"
var VmUuid = "501127b6-cf5a-c334-ffc6-5d12b230e32c"
var ThumbPoint = "4F:04:58:8E:6A:29:99:7E:AD:A1:F2:C1:DE:20:7A:66:FB:DC:90:71"

var Client *vSphereApi.VSphereApi

func login() {
	var login vSphereApi.LoginInfo
	login.Ip = IP
	login.UserName = Username
	login.PassWord = Password
	login.Port = 902
	var err error
	Client, err = vSphereApi.NewClient(&login)
	if err != nil {
		fmt.Printf("create client failed! case:%v", err)
		panic(err)
	}
}

func main() {
	var totalSize, compleSize int64

	login()
	defer Client.Logout()
	vmObj, err := Client.GetVMObj("501127b6-cf5a-c334-ffc6-5d12b230e32c")
	if err != nil {
		fmt.Printf("api GetVMObj failed! case:%v\n", err)
		return
	}
	err = Client.CreateSnapShot(VmUuid, "test_snapshot", "this is golang test snapshot.")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		err = Client.RemoveSnapshot(VmUuid, "test_snapshot")
		if err != nil {
			fmt.Println(err)
		}
	}()
	snapObj, err := Client.FindSnapshot(VmUuid, "test_snapshot")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = VixDiskLibApi.Init("./log/", "../vixDiskLibxx", "./vixDiskLib.config", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer VixDiskLibApi.Exit()
	modes := VixDiskLibApi.GetTransportModesList()
	fmt.Printf("modes:%s\n", modes)
	param := &VixDiskLibApi.ConnectParam{ServerIp: IP, UserName: Username, PassWord: Password, ThumbPoint: ThumbPoint,
		VmRef: fmt.Sprintf("moref=%s", vmObj), SnapshotRef: snapObj, TransMode: VixDiskLibApi.NBD_METHOD,
		ReadOnly: true}
	client, err := VixDiskLibApi.NewVixDiskLibApi(param, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Disconnect()
	defer client.CleanParams()
	startTime := time.Now().UnixNano()
	lastTime := startTime
	{
		f, err := os.OpenFile("/vmware/go-vmware/vmdk.data", os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer f.Close()
		handle, err := VixDiskLibApi.NewDiskHandle(client, "[data50] vddk_test/vddk_test.vmdk", VixDiskLibApi.VIXDISKLIB_FLAG_OPEN_READ_ONLY)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer handle.CloseDisk()
		mode, err := handle.GetCurTransportMode()
		if err != nil {
			fmt.Printf("GetCurTransportMode error:%+v\n", err)
			return
		}
		fmt.Printf("vm mode:%s\n", mode)
		size, err := handle.GetDiskSize()
		if err != nil {
			fmt.Printf("GetDiskSize error:%+v\n", err)
			return
		}
		fmt.Printf("disk size:%d\n", size)
		blocks, err := handle.QueryAllocatedBlocks()
		if err != nil {
			fmt.Printf("QueryAllocatedBlocks error:%+v\n", err)
			return
		}
		fmt.Printf("disk blocks:%+v\n", blocks)
		for _, block := range blocks {
			totalSize += block.Length
			tempInfo := block
			//secSize := (tempInfo.Length + int64(4*1024*1024-1)) / int64(4*1024*1024)
			secSize := (tempInfo.Length + int64(64*1024-1)) / int64(64*1024)
			var temp VixDiskLibApi.BlockInfo
			for i := int64(0); i < secSize; i++ {
				temp.Start = tempInfo.Start
				//temp.Length = Min(tempInfo.Length, int64(4*1024*1024))
				temp.Length = Min(tempInfo.Length, int64(64*1024))
				startSector := temp.Start / 512
				sectorLen := (temp.Length + 511) / 512
				buf := make([]byte, sectorLen*512)
				ret, err := handle.ReadDisk(startSector, sectorLen, buf)
				if err != nil && err != io.EOF {
					fmt.Printf("ReadDisk error:%+v\n", err)
					return
				}
				_, err = f.Seek(temp.Start, 0)
				if err != nil {
					fmt.Printf("fp Seek error:%+v\n", err)
					return
				}
				_, err = f.WriteAt(buf, temp.Start)
				if err != nil {
					fmt.Printf("fp Write error:%+v\n", err)
					return
				}
				compleSize += ret
				tempInfo.Start += temp.Length
				tempInfo.Length -= temp.Length
				currentTime := time.Now().UnixNano()
				if currentTime-lastTime > 3000000000 {
					fmt.Printf("完成数据量为:%d\n", compleSize)
					lastTime = currentTime
				}
			}
		}
	}
	fmt.Printf("读取数据总量为:%d, 完成数据量为:%d\n", totalSize, compleSize)
	endTime := time.Now().UnixNano()
	t := (endTime - startTime) / 1000000000
	m := compleSize / (1024 * 1024)
	fmt.Printf("耗时:%d s, 平均速度为:%d Mb/s\n", t, m/int64(t))
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}
