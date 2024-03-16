package vSphereApi

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
)

func NewClient(login *LoginInfo) (client *VSphereApi, err error) {
	v := &VSphereApi{loginInfo: login, ctx: context.Background()}
	err = v.Login()
	if err == nil {
		client = v
	}
	return
}

func (v *VSphereApi) Login() (err error) {
	userInfo := &url.URL{
		Scheme: "https",
		Host:   v.loginInfo.Ip,
		Path:   "/sdk",
	}
	userInfo.User = url.UserPassword(v.loginInfo.UserName, v.loginInfo.PassWord)
	client, err := govmomi.NewClient(v.ctx, userInfo, true)
	if err != nil {
		return err
	}
	if client == nil {
		return NewError(ERROR_OBJECT_IS_NULL)
	}
	v.vmomi = client
	return
}

func (v *VSphereApi) ReLogin() {
	if v.vmomi == nil {
		err := v.Login()
		if err != nil {

		}
	}
}

func (v *VSphereApi) Logout() {
	if v.vmomi != nil {
		_ = v.vmomi.Logout(v.ctx)
	}
}

// 平台类型
func (v *VSphereApi) IsvCenter() bool {
	return v.vmomi.Client.IsVC()
}

// 获取平台版本
func (v *VSphereApi) GetVersion() string {
	version := v.vmomi.ServiceContent.About.ApiVersion
	if v.vmomi.Client.IsVC() {
		return fmt.Sprintf("vCenter %s", version)
	} else {
		return fmt.Sprintf("ESXi %s", version)
	}
}

func (v *VSphereApi) GetVMObj(uuid string) (string, error) {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return "", err
	}
	return vmRef.Value, nil
}

// 获取所有虚拟机
func (v *VSphereApi) GetVMs() ([]DataSource, error) {
	var rootObj []mo.Folder
	objs := []types.ManagedObjectReference{v.getRootPath()}
	err := v.vmomi.Retrieve(v.ctx, objs, []string{}, &rootObj)
	if err != nil {
		return nil, err
	}
	m := view.NewManager(v.vmomi.Client)

	cv, err := m.CreateContainerView(v.ctx, objs[0], []string{"VirtualMachine"}, true)
	if err != nil {
		return nil, err
	}
	var vms []mo.VirtualMachine
	err = cv.Retrieve(v.ctx, []string{"VirtualMachine"}, []string{}, &vms)
	if err != nil {
		return nil, err
	}
	var dataSources []DataSource
	for _, iter := range vms {
		var dataSource DataSource
		dataSource.Type = INVENTORY_VIRTALMACHINE
		dataSource.Name = iter.Name
		dataSource.Expanded = false
		if iter.Config != nil {
			dataSource.Uuid = iter.Config.InstanceUuid
		}
		dataSources = append(dataSources, dataSource)
	}
	return dataSources, nil
}

// 获取指定路径下的子对象
func (v *VSphereApi) GetPathSubObjects(path string) ([]*DataSource, error) {
	v.ReLogin()
	var dataSources []*DataSource
	if path == "" {
		var rootObj []mo.Folder
		objs := []types.ManagedObjectReference{v.getRootPath()}
		err := v.vmomi.Retrieve(v.ctx, objs, []string{}, &rootObj)
		if err != nil {
			return nil, err
		}
		for _, obj := range rootObj[0].ChildEntity {
			dataSource, err := v.getObjectProperty("", &obj)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, dataSource)
		}
	} else {
		obj, err := v.findByInventoryPath(path)
		if err != nil {

			return nil, err
		}
		switch obj.Type {
		case INVENTORY_FOLDER:
			var folder mo.Folder
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &folder)
			if err != nil {
				return nil, err
			}
			for _, temp := range folder.ChildEntity {
				dataSource, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, dataSource)
			}
		case INVENTORY_DATACENTER:
			var dc mo.Datacenter
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &dc)
			if err != nil {
				return nil, err
			}
			// 获取计算资源
			hostFolder, err := v.getObjectProperty(path, &dc.HostFolder)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, hostFolder)
			// 获取存储
			dataStoreFolder, err := v.getObjectProperty(path, &dc.DatastoreFolder)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, dataStoreFolder)
			// 获取网卡
			networkFolder, err := v.getObjectProperty(path, &dc.NetworkFolder)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, networkFolder)
			// 获取vmFolder
			vmFolder, err := v.getObjectProperty(path, &dc.VmFolder)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, vmFolder)
		case INVENTORY_CLUSTERCOMPUTERESOURCE:
			var ccr mo.ClusterComputeResource
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &ccr)
			if err != nil {
				return nil, err
			}
			// 获取主机
			for _, temp := range ccr.Host {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
			resourcesPool, err := v.getObjectProperty(path, ccr.ResourcePool)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, resourcesPool)
		case INVENTORY_COMPUTERESOURCE:
			var cr mo.ComputeResource
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &cr)
			if err != nil {
				return nil, err
			}
			// 获取主机
			for _, temp := range cr.Host {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
			data, err := v.getObjectProperty(path, cr.ResourcePool)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, data)
		case INVENTORY_HOSTSYSTEM:
		case INVENTORY_DATASTORE:
			var ds mo.Datastore
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &ds)
			if err != nil {
				return nil, err
			}
			for _, temp := range ds.Vm {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
		case INVENTORY_NETWORK:
			var net mo.Network
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &net)
			if err != nil {
				return nil, err
			}
			for _, temp := range net.Vm {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
		case INVENTORY_RESOURCEPOOL:
			var rs mo.ResourcePool
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &rs)
			if err != nil {
				return nil, err
			}
			for _, temp := range rs.ResourcePool {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
			for _, temp := range rs.Vm {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
		case INVENTORY_VIRTUALAPP:
			var vApp mo.VirtualApp
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &vApp)
			if err != nil {
				return nil, err
			}
			for _, temp := range vApp.ResourcePool.ResourcePool {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
			for _, temp := range vApp.Vm {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
		default:
			return nil, NewError(ERROR_NOT_FOUND)
		}
	}
	return dataSources, nil
}

// 按计算资源的方式获取数据源
func (v *VSphereApi) GetDataSourcesByComputeResource(path string, isGetVm bool) ([]*DataSource, error) {
	v.ReLogin()
	var dataSources []*DataSource
	if path == "" {
		var rootObj []mo.Folder
		objs := []types.ManagedObjectReference{v.getRootPath()}
		err := v.vmomi.Retrieve(v.ctx, objs, []string{}, &rootObj)
		if err != nil {
			return nil, err
		}
		for _, obj := range rootObj[0].ChildEntity {
			dataSource, err := v.getObjectProperty("", &obj)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, dataSource)
		}
	} else {
		obj, err := v.findByInventoryPath(path)
		if err != nil {

			return nil, err
		}
		switch obj.Type {
		case INVENTORY_FOLDER:
			var folder mo.Folder
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &folder)
			if err != nil {
				return nil, err
			}
			for _, temp := range folder.ChildEntity {
				dataSource, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, dataSource)
			}
		case INVENTORY_DATACENTER:
			var dc mo.Datacenter
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &dc)
			if err != nil {
				return nil, err
			}
			// 获取计算资源
			hostFolder, err := v.getObjectProperty(path, &dc.HostFolder)
			if err != nil {

				return nil, err
			}
			// 再次获取，过滤掉隐藏路径[host]
			hostFolderSub, err := v.GetPathSubObjects(hostFolder.Path)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, hostFolderSub...)
		case INVENTORY_CLUSTERCOMPUTERESOURCE:
			var ccr mo.ClusterComputeResource
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &ccr)
			if err != nil {
				return nil, err
			}
			temp, err := v.getObjectProperty(path, ccr.ResourcePool)
			if err != nil {

				return nil, err
			}
			// 再次获取，过滤掉根资源池[Resources]
			subObjects, err := v.GetPathSubObjects(temp.Path)
			if err != nil {

				return nil, err
			}
			for _, subObject := range subObjects {
				// 是否获取虚拟机
				if !isGetVm && (subObject.Type == INVENTORY_VIRTALMACHINE) {
					continue
				}
				dataSources = append(dataSources, subObject)
			}
		case INVENTORY_COMPUTERESOURCE:
			var cr mo.ComputeResource
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &cr)
			if err != nil {
				return nil, err
			}
			temp, err := v.getObjectProperty(path, cr.ResourcePool)
			if err != nil {

				return nil, err
			}
			// 再次获取，过滤掉根资源池[Resources]
			subObjects, err := v.GetPathSubObjects(temp.Path)
			if err != nil {

				return nil, err
			}
			for _, subObject := range subObjects {
				// 是否获取虚拟机
				if !isGetVm && (subObject.Type == INVENTORY_VIRTALMACHINE) {
					continue
				}
				dataSources = append(dataSources, subObject)
			}
		case INVENTORY_RESOURCEPOOL:
			var rs mo.ResourcePool
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &rs)
			if err != nil {
				return nil, err
			}
			for _, temp := range rs.ResourcePool {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
			// 是否获取虚拟机
			if isGetVm {
				for _, temp := range rs.Vm {
					data, err := v.getObjectProperty(path, &temp)
					if err != nil {

						return nil, err
					}
					dataSources = append(dataSources, data)
				}
			}
		case INVENTORY_VIRTUALAPP:
			var vApp mo.VirtualApp
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &vApp)
			if err != nil {
				return nil, err
			}
			for _, temp := range vApp.ResourcePool.ResourcePool {
				data, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				dataSources = append(dataSources, data)
			}
			// 是否获取虚拟机
			if isGetVm {
				for _, temp := range vApp.Vm {
					data, err := v.getObjectProperty(path, &temp)
					if err != nil {

						return nil, err
					}
					dataSources = append(dataSources, data)
				}
			}
		default:
			return nil, NewError(ERROR_NOT_FOUND)
		}
	}
	return dataSources, nil
}

// 按虚拟机模板的方式获取数据源
func (v *VSphereApi) GetDataSourcesByVmTemplate(path string, isGetVm bool) ([]*DataSource, error) {
	v.ReLogin()
	var dataSources []*DataSource
	if path == "" {
		var rootObj []mo.Folder
		objs := []types.ManagedObjectReference{v.getRootPath()}
		err := v.vmomi.Retrieve(v.ctx, objs, []string{}, &rootObj)
		if err != nil {
			return nil, err
		}
		for _, obj := range rootObj[0].ChildEntity {
			dataSource, err := v.getObjectProperty("", &obj)
			if err != nil {

				return nil, err
			}
			dataSources = append(dataSources, dataSource)
		}
	} else {
		obj, err := v.findByInventoryPath(path)
		if err != nil {

			return nil, err
		}
		switch obj.Type {
		case INVENTORY_FOLDER:
			var folder mo.Folder
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &folder)
			if err != nil {
				return nil, err
			}
			for _, temp := range folder.ChildEntity {
				dataSource, err := v.getObjectProperty(path, &temp)
				if err != nil {

					return nil, err
				}
				// 是否获取虚拟机
				if !isGetVm && (dataSource.Type == INVENTORY_VIRTALMACHINE) {
					continue
				}
				dataSources = append(dataSources, dataSource)
			}
		case INVENTORY_DATACENTER:
			var dc mo.Datacenter
			err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*obj}, []string{}, &dc)
			if err != nil {
				return nil, err
			}
			// 获取虚拟机文件夹
			vmFolder, err := v.getObjectProperty(path, &dc.VmFolder)
			if err != nil {

				return nil, err
			}
			// 再次获取，过滤掉隐藏路径[vm]
			vmFolderSub, err := v.GetPathSubObjects(vmFolder.Path)
			if err != nil {

				return nil, err
			}
			for _, subObject := range vmFolderSub {
				// 是否获取虚拟机
				if !isGetVm && (subObject.Type == INVENTORY_VIRTALMACHINE) {
					continue
				}
				dataSources = append(dataSources, subObject)
			}
		case INVENTORY_VIRTUALAPP:
			// todo: 过滤不展示
		default:
			return nil, NewError(ERROR_NOT_FOUND)
		}
	}
	return dataSources, nil
}

// 配置虚拟机
func (v *VSphereApi) Customize(uuid string, addr *IpAddr) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	if addr.Ip == "" || addr.Netmask == "" || addr.Gateway == "" {
		return nil
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	cam := types.CustomizationAdapterMapping{
		Adapter: types.CustomizationIPSettings{
			Ip:         &types.CustomizationFixedIp{IpAddress: addr.Ip},
			SubnetMask: addr.Netmask,
			Gateway:    []string{addr.Gateway},
		},
	}
	customSpec := types.CustomizationSpec{
		NicSettingMap: []types.CustomizationAdapterMapping{cam},
		Identity:      &types.CustomizationLinuxPrep{HostName: &types.CustomizationFixedName{Name: addr.Hostname}},
	}
	t, err := vmObj.Customize(v.ctx, customSpec)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

// 开机
func (v *VSphereApi) PowerOn(uuid string) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	t, err := vmObj.PowerOn(v.ctx)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

// 关机
func (v *VSphereApi) PowerOff(uuid string) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	t, err := vmObj.PowerOff(v.ctx)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

// 创建快照
func (v *VSphereApi) CreateSnapShot(uuid, name, desc string) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	t, err := vmObj.CreateSnapshot(v.ctx, name, desc, false, false)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *VSphereApi) FindSnapshot(uuid, name string) (string, error) {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return "", err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	snapRef, err := vmObj.FindSnapshot(v.ctx, name)
	if err != nil {
		return "", err
	}
	return snapRef.Value, err
}

func (v *VSphereApi) RemoveSnapshot(uuid, name string) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	consolidate := true
	t, err := vmObj.RemoveSnapshot(v.ctx, name, false, &consolidate)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *VSphereApi) RemoveAllSnapshot(uuid string) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	consolidate := true
	t, err := vmObj.RemoveAllSnapshot(v.ctx, &consolidate)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

// 恢复快照
func (v *VSphereApi) RecoverToSnapshot(uuid string, name string, suppressPowerOn bool) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	t, err := vmObj.RevertToSnapshot(v.ctx, name, suppressPowerOn)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (v *VSphereApi) GetHostNameByPath(hostPath string) (string, error) {
	hostObj, err := v.getHostObjectByPath(hostPath)
	if err != nil {

		return "", err
	}
	return hostObj.Name, nil
}

// 获取NFS存储列表
func (v *VSphereApi) GetNasDataStore() ([]string, error) {
	var dataStoreNames []string
	hostVec, err := v.getAllHostSystem("")
	if err != nil {
		return nil, err
	}
	for _, host := range hostVec {
		var inventoryObj mo.HostSystem
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*host}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		for _, storeObj := range inventoryObj.Datastore {
			var dataStore mo.Datastore
			err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{storeObj}, []string{}, &dataStore)
			if err != nil {
				return nil, err
			}
			if dataStore.Summary.Type == "NFS" || dataStore.Summary.Type == "NFS41" {
				dataStoreNames = append(dataStoreNames, dataStore.Name)
			}
		}
	}
	return dataStoreNames, nil
}

// 创建nas存储
func (v *VSphereApi) CreateNasDatastore(storeInfo *NasDataStoreParam) error {
	v.ReLogin()
	sysObj, err := v.getHostDataStoreSystem(storeInfo.HostPath)
	if err != nil {
		return err
	}
	hostNasSpec := types.HostNasVolumeSpec{
		RemoteHost:      storeInfo.RemoteHost,
		RemotePath:      storeInfo.RemotePath,
		LocalPath:       storeInfo.LocalPath,
		AccessMode:      storeInfo.AccessMode,
		Type:            storeInfo.StoreType,
		UserName:        "",
		Password:        "",
		RemoteHostNames: []string{storeInfo.RemoteHost},
		SecurityType:    "AUTH_SYS"}
	_, err = sysObj.CreateNasDatastore(v.ctx, hostNasSpec)
	if err != nil {
		return err
	}
	return nil
}

// 卸载NFS存储
func (v *VSphereApi) RemoveNasDataStore(storeName string) error {
	hostVec, err := v.getAllHostSystem("")
	if err != nil {
		return err
	}
	for _, host := range hostVec {
		var inventoryObj mo.HostSystem
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*host}, []string{}, &inventoryObj)
		if err != nil {
			return err
		}
		for _, storeObj := range inventoryObj.Datastore {
			var dataStore mo.Datastore
			err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{storeObj}, []string{}, &dataStore)
			if err != nil {
				return err
			}
			if dataStore.Name == storeName {
				ds := object.NewDatastore(v.vmomi.Client, dataStore.Reference())
				datastoreSystem := object.NewHostDatastoreSystem(v.vmomi.Client, inventoryObj.ConfigManager.DatastoreSystem.Reference())
				err := datastoreSystem.Remove(v.ctx, ds)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// 注册虚拟
func (v *VSphereApi) RegisterVm(param *RegisterVmParam) error {
	v.ReLogin()
	// 获取folder对象
	var folder *object.Folder
	if param.FolderPath != "" {
		folderObj, err := v.findByInventoryPath(param.FolderPath)
		if err != nil {
			return err
		}
		folder = object.NewFolder(v.vmomi.Client, *folderObj)
	} else {
		return NewError(ERROR_NOT_FOUND)
	}
	// 获取主机对象
	var hostSystem *object.HostSystem
	if param.HostPath != "" {
		hostObj, err := v.getHostObjectByPath(param.HostPath)
		if err != nil {

		} else {
			hostSystem = object.NewHostSystem(v.vmomi.Client, hostObj.Reference())
		}
	}
	// 获取资源池对象
	var resourcePool *object.ResourcePool
	if param.ResourcePoolPath != "" {
		resourcePoolObj, err := v.findByInventoryPath(param.ResourcePoolPath)
		if err != nil {

		} else {
			resourcePool = object.NewResourcePool(v.vmomi.Client, *resourcePoolObj)
		}
	}
	t, err := folder.RegisterVM(v.ctx, param.VmxPath, param.VmName, false, resourcePool, hostSystem)
	if err != nil {
		return err
	}
	err = t.Wait(v.ctx)
	if err != nil {
		return err
	}
	return nil
}

// 取消注册
func (v *VSphereApi) UnRegisterVm(uuid string) error {
	v.ReLogin()
	vmRef, err := v.findByUUid(uuid)
	if err != nil {
		return err
	}
	vmObj := object.NewVirtualMachine(v.vmomi.Client, vmRef.Reference())
	err = vmObj.Unregister(v.ctx)
	if err != nil {
		return err
	}
	return nil
}
