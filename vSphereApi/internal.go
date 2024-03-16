package vSphereApi

import (
	"errors"
	"fmt"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

// 获取根对象
func (v *VSphereApi) getRootPath() types.ManagedObjectReference {
	return v.vmomi.ServiceContent.RootFolder
}

// 根据清单路径获取对象
func (v VSphereApi) findByInventoryPath(path string) (*types.ManagedObjectReference, error) {
	req := types.FindByInventoryPath{
		This:          *v.vmomi.Client.ServiceContent.SearchIndex,
		InventoryPath: path,
	}

	res, err := methods.FindByInventoryPath(v.ctx, v.vmomi.Client, &req)
	if err != nil {
		return nil, err
	}

	if res.Returnval == nil {
		return nil, err
	}
	obj := object.NewReference(v.vmomi.Client, *res.Returnval).Reference()
	return &obj, err
}

// 根据uuid获取虚拟机对象
func (v *VSphereApi) findByUUid(uuid string) (*types.ManagedObjectReference, error) {
	ins := true
	req := types.FindByUuid{
		This:         *v.vmomi.Client.ServiceContent.SearchIndex,
		Uuid:         uuid,
		VmSearch:     true,
		InstanceUuid: &ins,
	}
	res, err := methods.FindByUuid(v.ctx, v.vmomi.Client, &req)
	if err != nil {
		return nil, err
	}
	if res.Returnval == nil {
		return nil, errors.New("pointer is nil")
	}
	//obj := object.NewReference(v.vmomi.Client, *res.Returnval).Reference()
	return res.Returnval, nil
}

// 获取指定对象属性
func (v *VSphereApi) getObjectProperty(parentPath string, objRef *types.ManagedObjectReference) (*DataSource, error) {
	var dataSources DataSource
	if objRef == nil {
		return nil, NewError(ERROR_OBJECT_IS_NULL)
	}
	switch objRef.Type {
	case INVENTORY_FOLDER:
		var inventoryObj mo.Folder
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_FOLDER
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_DATACENTER:
		var inventoryObj mo.Datacenter
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_DATACENTER
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_CLUSTERCOMPUTERESOURCE:
		var inventoryObj mo.ClusterComputeResource
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_CLUSTERCOMPUTERESOURCE
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_COMPUTERESOURCE:
		var inventoryObj mo.ComputeResource
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_COMPUTERESOURCE
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_DATASTORE:
		var inventoryObj mo.Datastore
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_DATASTORE
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_DISTRIBUTEDVIRTUALSWITCH:
		var inventoryObj mo.DistributedVirtualSwitch
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Expanded = false
		dataSources.CheckAble = true
		dataSources.Type = INVENTORY_DISTRIBUTEDVIRTUALSWITCH
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_HOSTSYSTEM:
		var inventoryObj mo.HostSystem
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_HOSTSYSTEM
		dataSources.Expanded = false
		dataSources.CheckAble = false
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_NETWORK:
		var inventoryObj mo.Network
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_NETWORK
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_RESOURCEPOOL:
		var inventoryObj mo.ResourcePool
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_RESOURCEPOOL
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_VIRTUALAPP:
		var inventoryObj mo.VirtualApp
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_VIRTUALAPP
		dataSources.Expanded = true
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	case INVENTORY_VIRTALMACHINE:
		var inventoryObj mo.VirtualMachine
		err := v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*objRef}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
		dataSources.Name = inventoryObj.Name
		if inventoryObj.Config != nil {
			dataSources.Uuid = inventoryObj.Config.InstanceUuid
		}
		dataSources.Type = INVENTORY_VIRTALMACHINE
		dataSources.Expanded = false
		dataSources.CheckAble = true
		if parentPath == "" {
			dataSources.Path = inventoryObj.Name
		} else {
			dataSources.Path = fmt.Sprintf("%s/%s", parentPath, inventoryObj.Name)
		}
	default:
		dataSources.Name = ""
		dataSources.Uuid = ""
		dataSources.Type = INVENTORY_UNDEFINE
		dataSources.Expanded = false
		dataSources.CheckAble = false
		dataSources.Path = parentPath
	}
	return &dataSources, nil
}

// 获取所有主机对象
func (v *VSphereApi) getAllHostSystem(path string) ([]*types.ManagedObjectReference, error) {
	var hosts []*types.ManagedObjectReference
	datasources, err := v.GetPathSubObjects(path)
	if err != nil {
		return nil, err
	}
	for _, datasource := range datasources {
		if datasource.Type == INVENTORY_HOSTSYSTEM {
			host, err := v.findByInventoryPath(datasource.Path)
			if err != nil {
				return nil, err
			}
			hosts = append(hosts, host)
			continue
		}
		if datasource.Type == INVENTORY_RESOURCEPOOL || datasource.Type == INVENTORY_DATASTORE ||
			datasource.Type == INVENTORY_NETWORK || datasource.Type == INVENTORY_VIRTUALAPP ||
			datasource.Type == INVENTORY_DISTRIBUTEDVIRTUALSWITCH || datasource.Type == INVENTORY_VIRTALMACHINE {
			continue
		}
		if datasource.Expanded {
			tempHosts, err := v.getAllHostSystem(datasource.Path)
			if err != nil {
				return nil, err
			}
			hosts = append(hosts, tempHosts...)
		}
	}
	return hosts, nil
}

// 根据名称获取主机对象
func (v *VSphereApi) getHostSystemByName(hostName string) (*types.ManagedObjectReference, error) {
	hostVec, err := v.getAllHostSystem("")
	if err != nil {
		return nil, err
	}
	for _, iter := range hostVec {
		hostInfo, err := v.getObjectProperty("", iter)
		if err != nil {
			return nil, err
		}
		if hostInfo.Name == hostName {
			return iter, nil
		}
	}
	err = NewError(ERROR_NOT_FOUND)
	return nil, err
}

// 根据主机路径获取主机对象
func (v *VSphereApi) getHostObjectByPath(hostPath string) (*mo.HostSystem, error) {
	computeResourceRef, err := v.findByInventoryPath(hostPath)
	if err != nil {
		return nil, err
	}
	result, err := v.getObjectProperty("", computeResourceRef)
	if err != nil {

		return nil, err
	}
	var inventoryObj mo.HostSystem
	if result.Type == INVENTORY_CLUSTERCOMPUTERESOURCE {
		var ccrObj mo.ClusterComputeResource
		err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*computeResourceRef}, []string{}, &ccrObj)
		if err != nil {
			return nil, err
		}
		if len(ccrObj.Host) == 0 {
			return nil, err
		}
		err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{ccrObj.Host[0]}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
	} else if result.Type == INVENTORY_COMPUTERESOURCE {
		var crObj mo.ComputeResource
		err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{*computeResourceRef}, []string{}, &crObj)
		if err != nil {
			return nil, err
		}
		if len(crObj.Host) == 0 {
			return nil, err
		}
		err = v.vmomi.Retrieve(v.ctx, []types.ManagedObjectReference{crObj.Host[0]}, []string{}, &inventoryObj)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	return &inventoryObj, nil
}

// 根据主机路径获取主机存储管理对象
func (v *VSphereApi) getHostDataStoreSystem(hostPath string) (*object.HostDatastoreSystem, error) {
	hostObj, err := v.getHostObjectByPath(hostPath)
	if err != nil {
		return nil, err
	}
	dataStoreSystem := object.NewHostDatastoreSystem(v.vmomi.Client, hostObj.ConfigManager.DatastoreSystem.Reference())
	return dataStoreSystem, nil
}
