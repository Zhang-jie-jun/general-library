//go:build linux
// +build linux

package ceph

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/ceph/go-ceph/rados"
	"github.com/ceph/go-ceph/rbd"
	"io"
	"os/exec"
	"strings"
	"time"
)

var (
	Connect *rados.Conn //连接引擎
)

func init() {
	err := Client.loginCeph()
	if err != nil {
		panic(err)
	}
}

func CloseCeph() {
	if Connect != nil {
		Connect.Shutdown()
	}
}

type client struct{}

var Client = client{}

// 连接ceph集群
func (c *client) loginCeph() (err error) {
	Connect, err = rados.NewConn()
	if err != nil {
		return err
	}

	err = Connect.ReadDefaultConfigFile()
	if err != nil {
		return err
	}

	err = Connect.Connect()
	if err != nil {
		return err
	}
	return err
}

// 执行ceph命令
func (c *client) MonCommand(command []byte) (buf []byte, err error) {
	if Connect == nil {
		err = Client.loginCeph()
		if err != nil {
			return
		}
	}
	buf, _, err = Connect.MonCommand(command)
	if err != nil {
		return
	}
	return
}

// 获取集群信息
func (c *client) GetClusterInfo() (info *ClusterInfo, err error) {
	if Connect == nil {
		err = Client.loginCeph()
		if err != nil {
			return
		}
	}
	stats, err := Connect.GetClusterStats()
	if err != nil {
		return
	}
	info = &ClusterInfo{TotalSize: stats.Kb * 1024, UsedSize: stats.Kb_used * 1024,
		AvailSize: stats.Kb_avail * 1024, TotalObject: stats.Num_objects}
	return
}

// 创建存储池
func (c *client) CreatePool(poolName string) (err error) {
	if Connect == nil {
		err = Client.loginCeph()
		if err != nil {
			return err
		}
	}
	err = Connect.MakePool(poolName)
	if err != nil {
		return err
	}
	return err
}

// 删除存储池
func (c *client) DeletePool(poolName string) (err error) {
	if Connect == nil {
		err = Client.loginCeph()
		if err != nil {
			return err
		}
	}
	err = Connect.DeletePool(poolName)
	if err != nil {
		return err
	}
	return err
}

func (c *client) openPool(poolName string) (ctx *rados.IOContext, err error) {
	if Connect == nil {
		err = Client.loginCeph()
		if err != nil {
			return
		}
	}
	ctx, err = Connect.OpenIOContext(poolName)
	if err != nil {
		return ctx, err
	}
	return ctx, err
}

// 创建image
func (c *client) CreateImage(poolName, imageName string, size uint64) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	options := rbd.NewRbdImageOptions()
	err = rbd.CreateImage(ctx, imageName, size, options)
	if err != nil {
		return err
	}
	return err
}

// 删除image
func (c *client) DeleteImage(poolName, imageName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	err = rbd.RemoveImage(ctx, imageName)
	if err != nil {
		return err
	}
	return err
}

func (c *client) getImage(ctx *rados.IOContext, imageName string) (image *rbd.Image, err error) {
	image, err = rbd.OpenImage(ctx, imageName, rbd.NoSnapshot)
	if err != nil {
		return image, err
	}
	return image, err
}

// 复制image
func (c *client) CopyImage(poolName, imageName, destName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		w := image.Close()
		if w != nil {
		}
	}()
	err = image.Copy(ctx, destName)
	if err != nil {
		return err
	}
	return err
}

// 从快照克隆image[克隆前需要检查快照是否受保护]
func (c *client) CloneImageBySnapshot(poolName, imageName, snapName, destName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	features, err := image.GetFeatures()
	if err != nil {
		return err
	}
	_, err = image.Clone(snapName, ctx, destName, features, 22)
	if err != nil {
		return err
	}
	return err
}

// 分离image快照依赖
func (c *client) FlattenImage(poolName, imageName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	err = image.Flatten()
	if err != nil {
		return err
	}
	return err
}

// 重置image大小
func (c *client) ReSizeImage(poolName, imageName string, size uint64) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	err = image.Resize(size)
	if err != nil {
		return err
	}
	return err
}

// 重命名镜像
func (c *client) ReNameImage(poolName, imageName, destImageName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	err = image.Rename(destImageName)
	if err != nil {
		return err
	}
	return err
}

// 刷新缓存数据到镜像
func (c *client) FlushImage(poolName, imageName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	err = image.Flush()
	if err != nil {
		return err
	}
	return err
}

// 创建image快照
func (c *client) CreateSnapshot(poolName, imageName, snapName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		w := image.Close()
		if w != nil {
		}
	}()
	_, err = image.CreateSnapshot(snapName)
	if err != nil {
		return err
	}
	return err
}

// 删除image快照
func (c *client) DeleteSnapshot(poolName, imageName, snapName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		w := image.Close()
		if w != nil {
		}
	}()
	snapshot := image.GetSnapshot(snapName)
	err = snapshot.Remove()
	if err != nil {
		return err
	}
	return err
}

// 回滚image快照
func (c *client) RollbackSnapshot(poolName, imageName, snapName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	snapshot := image.GetSnapshot(snapName)
	err = snapshot.Rollback()
	if err != nil {
		return err
	}
	return err
}

// 保护快照
func (c *client) ProtectSnapShot(poolName, imageName, snapName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	snapshot := image.GetSnapshot(snapName)
	err = snapshot.Protect()
	if err != nil {
		return err
	}
	return err
}

// 解除快照保护
func (c *client) UnProtectSnapShot(poolName, imageName, snapName string) (err error) {
	ctx, err := c.openPool(poolName)
	if err != nil {
		return err
	}
	defer ctx.Destroy()
	image, err := c.getImage(ctx, imageName)
	if err != nil {
		return err
	}
	defer func() {
		_ = image.Close()
	}()
	snapshot := image.GetSnapshot(snapName)
	err = snapshot.Unprotect()
	if err != nil {
		return err
	}
	return err
}

func (c *client) MapRBDImage(poolName, imageName string) (devPath string, err error) {
	cmd := fmt.Sprintf("rbd map %s --pool %s", imageName, poolName)
	result, err := RunCommand(cmd)
	if err != nil {
		return
	}
	if len(result) != 1 {
		info := "params error."
		err = errors.New(info)
		return
	}
	devPath = result[0]
	return
}

func (c *client) UnMapRBDImage(poolName, imageName string) (err error) {
	cmd := fmt.Sprintf("rbd unmap /dev/rbd/%s/%s", poolName, imageName)
	_, err = RunCommand(cmd)
	if err != nil {
		return
	}
	return
}

func (c *client) ShowMapRBDImage() (mapInfos []MapInfo, err error) {
	cmd := "rbd showmapped"
	result, err := RunCommand(cmd)
	if err != nil {
		return
	}
	for _, line := range result {
		temp := strings.Fields(line)
		if len(temp) != 5 {
			info := "params error."
			err = errors.New(info)
			return
		}
		// 过滤标签行
		if temp[0] == "id" {
			continue
		}
		var info MapInfo
		info.PoolName = temp[1]
		info.ImageName = temp[2]
		info.DevPath = temp[4]
		mapInfos = append(mapInfos, info)
	}
	return
}

func RunCommand(command string) (lines []string, err error) {
	// 检查命令是否可执行
	_, err = exec.LookPath(strings.Split(command, " ")[0])
	if err != nil {
		return
	}
	// 设置5分钟超时
	timeout := time.Duration(5) * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	name := "/bin/bash"
	args := []string{"-c", command}
	cmd := exec.CommandContext(ctx, name, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// 创建命令输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	// 执行命令
	err = cmd.Start()
	if err != nil {
		return
	}
	// 按行获取执行结果
	reader := bufio.NewReader(stdout)
	for {
		lineByte, _, err1 := reader.ReadLine()
		if err1 != nil {
			if err1 == io.EOF {
				break
			}
			break
		}
		// 去掉多余的换行符
		line := strings.Replace(string(lineByte), "\n", "", -1)
		lines = append(lines, line)
	}
	// 等待执行结束
	err1 := cmd.Wait()
	if err1 != nil {
		if stderr.Len() == 0 && lines == nil {
			// 文件描述符链接到os.DevNul上，忽略错误
			return
		}
		err = errors.New(err1.Error() + ":" + stderr.String())
		return
	}
	return
}
