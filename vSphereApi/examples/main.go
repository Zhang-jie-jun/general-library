package main

import "C"
import (
	"fmt"
	"github.com/Zhang-jie-jun/general-library/vSphereApi"
)

var Client *vSphereApi.VSphereApi

func login() {
	var login vSphereApi.LoginInfo
	login.Ip = "192.168.212.108"
	login.UserName = "administrator@vsphere.local"
	login.PassWord = "Password@123"
	login.Port = 902
	var err error
	Client, err = vSphereApi.NewClient(&login)
	if err != nil {
		fmt.Printf("create client failed! case:%v", err)
		panic(err)
	}
}

func showDataSource(path string, level int) {
	formatStr := ""
	for i := 0; i < level; i++ {
		formatStr += "===="
	}
	datasources, err := Client.GetPathSubObjects(path)
	if err != nil {
		fmt.Printf("api GetPathSubObjects failed! case:%v\n", err)
		return
	}
	for _, datasource := range datasources {
		if datasource.Type == vSphereApi.INVENTORY_VIRTALMACHINE {
			fmt.Println(fmt.Sprintf("%s %s(%s)", formatStr, datasource.Name, datasource.Uuid))
		} else {
			fmt.Println(fmt.Sprintf("%s %s", formatStr, datasource.Name))
		}
		if datasource.Expanded {
			showDataSource(datasource.Path, level+1)
		}
	}
}

func showDataSourceComputeResource(path string, level int) {
	formatStr := ""
	for i := 0; i < level; i++ {
		formatStr += "===="
	}
	datasources, err := Client.GetDataSourcesByComputeResource(path, false)
	if err != nil {
		fmt.Printf("api GetDataSourcesByComputeResource failed! case:%v\n", err)
		return
	}
	for _, datasource := range datasources {
		if datasource.Type == vSphereApi.INVENTORY_VIRTALMACHINE {
			fmt.Println(fmt.Sprintf("%s %s(%s)", formatStr, datasource.Name, datasource.Uuid))
		} else {
			fmt.Println(fmt.Sprintf("%s %s", formatStr, datasource.Name))
		}
		if datasource.Expanded {
			showDataSourceComputeResource(datasource.Path, level+1)
		}
	}
}

func showDataSourceByVmTemplate(path string, level int) {
	formatStr := ""
	for i := 0; i < level; i++ {
		formatStr += "===="
	}
	datasources, err := Client.GetDataSourcesByVmTemplate(path, false)
	if err != nil {
		fmt.Printf("api GetDataSourcesByComputeResource failed! case:%v\n", err)
		return
	}
	for _, datasource := range datasources {
		if datasource.Type == vSphereApi.INVENTORY_VIRTALMACHINE {
			fmt.Println(fmt.Sprintf("%s %s(%s)", formatStr, datasource.Name, datasource.Uuid))
		} else {
			fmt.Println(fmt.Sprintf("%s %s", formatStr, datasource.Name))
		}
		if datasource.Expanded {
			showDataSourceByVmTemplate(datasource.Path, level+1)
		}
	}
}

func getVms() {
	vms, err := Client.GetVMs()
	if err != nil {
		fmt.Printf("api GetPathSubObjects failed! case:%v\n", err)
		return
	}
	for _, vm := range vms {
		fmt.Println(fmt.Sprintf("%s(%s)", vm.Name, vm.Uuid))
	}
	fmt.Println(len(vms))
}

func getNfsStore() {
	dataStoreVec, err := Client.GetNasDataStore()
	if err != nil {
		fmt.Printf("api GetNasDatastore failed! case:%v\n", err)
		return
	}
	for _, dataStore := range dataStoreVec {
		fmt.Println(fmt.Sprintf("%s", dataStore))
	}
}

func register() {
	var registerParam vSphereApi.RegisterVmParam
	registerParam.HostPath = "ha-datacenter/host/localhost6.7.3Ub"
	registerParam.ResourcePoolPath = "ha-datacenter/host/localhost6.7.3Ub/Resources/Jack(开发环境)"
	registerParam.FolderPath = "ha-datacenter/vm"
	registerParam.VmName = "tangula_test"
	registerParam.VmxPath = "[tangula_891edc27ad3a599ed9327cd8ea9c7b84] tangula_test.vmx"
	err := Client.RegisterVm(&registerParam)
	if err != nil {
		fmt.Printf("api register failed! case:%v\n", err)
	}
}

func unRegister() {
	err := Client.UnRegisterVm("50111d9b-aa6d-7734-f400-be4ba0eaab0a")
	if err != nil {
		fmt.Printf("api register failed! case:%v\n", err)
	}
}

func removeNasStore() {
	err := Client.RemoveNasDataStore("tangula_891edc27ad3a599ed9327cd8ea9c7b84")
	if err != nil {
		fmt.Printf("api register failed! case:%v\n", err)
	}
}

func customizeVM() {
	// 打开虚拟机电源
	//ex := Client.PowerOff("5004136a-e084-0329-4333-8490c98cd334")
	//if ex != nil {
	//	fmt.Printf("api PowerOff failed! case:%v\n", ex)
	//	return
	//}
	// 配置虚拟机
	var ipAddr vSphereApi.IpAddr
	ipAddr.Ip = "192.168.212.35"
	ipAddr.Netmask = "255.255.255.0"
	ipAddr.Gateway = "192.168.212.1"
	ipAddr.Hostname = "tangula"
	ex := Client.Customize("6db35e8f-57ed-4e8c-843a-5a81f8a84b55", &ipAddr)
	if ex != nil {
		fmt.Printf("api Customize failed! case:%v\n", ex)
		return
	}
	// 打开虚拟机电源
	ex = Client.PowerOn("6db35e8f-57ed-4e8c-843a-5a81f8a84b55")
	if ex != nil {
		fmt.Printf("api PowerOn failed! case:%v\n", ex)
		return
	}
}

func main() {
	login()
	defer Client.Logout()
	showDataSource("", 0)
	//showDataSourceComputeResource("", 0)
	//showDataSourceByVmTemplate("", 0)
	//getNfsStore()
	//getVms()
	vmObj, err := Client.GetVMObj("501127b6-cf5a-c334-ffc6-5d12b230e32c")
	if err != nil {
		fmt.Printf("api GetVMObj failed! case:%v\n", err)
		return
	}
	fmt.Printf("vmmore:%s\n", vmObj)
	//register()
	//unRegister()
	//removeNasStore()
	//customizeVM()
}
