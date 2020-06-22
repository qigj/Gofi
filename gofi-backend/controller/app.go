package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/sirupsen/logrus"
	"gofi/db"
	"gofi/env"
	"gofi/i18n"
	"gofi/tool"
	"path/filepath"
)

//UpdateConfiguration 更新设置
func UpdateConfiguration(ctx iris.Context) {

	configuration := db.ObtainConfiguration()
	// 初始化完成且处于Preview环境,不允许更改设置项
	if env.IsPreview() && configuration.Initialized {
		_, _ = ctx.JSON(NewResource().Fail().Message(i18n.Translate(i18n.OperationNotAllowedInPreviewMode)).Build())
		return
	}

	configuration.Initialized = true

	// 用客户端给定的Configuration覆盖数据库持久化的Configuration
	// 避免Body为空的时候ReadJson报错,导致后续不能默认初始化，这里用ContentLength做下判断
	if err := ctx.ReadJSON(configuration); ctx.GetContentLength() != 0 && err != nil {
		logrus.Error(err)
		_, _ = ctx.JSON(NewResource().Fail().Build())
	}

	path := configuration.CustomStoragePath
	defaultStorageDir := tool.GetDefaultStorageDir()

	// 是否使用默认地址
	useDefaultDir := path == "" || path == defaultStorageDir

	logrus.Printf("try to update configuration ,path is %v \n", path)
	logrus.Printf("useDefaultDir param is %v \n", useDefaultDir)

	if useDefaultDir {
		// 如果文件夹不存在，创建文件夹
		tool.MkdirIfNotExist(defaultStorageDir)

		configuration.CustomStoragePath = ""

		// 写入到配置文件,指定AllCols才会更新empty string
		db.UpdateConfiguration(configuration)

		logrus.Infof("use default path %s, setup success", defaultStorageDir)

		GetConfiguration(ctx)
	} else {
		// 判断给定的目录是否存在
		if !tool.FileExist(path) {
			_, _ = ctx.JSON(NewResource().Fail().Message(i18n.Translate(i18n.DirIsNotExist, path)))
			return
		}

		// 判断给定的路径是否是目录
		if !tool.IsDirectory(path) {
			_, _ = ctx.JSON(NewResource().Fail().Message(i18n.Translate(i18n.IsNotDir, path)))
			return
		}

		// 更新配置文件的仓库目录
		configuration.CustomStoragePath = filepath.Clean(path)

		// 写入到配置文件
		db.UpdateConfiguration(configuration)

		// 路径合法，初始化成功，持久化该路径。
		logrus.Infof("setup success,storage path is %s", path)

		GetConfiguration(ctx)
	}

}

//Setup 初始化
func Setup(ctx iris.Context) {
	// 已经初始化过
	if db.ObtainConfiguration().Initialized {
		_, _ = ctx.JSON(NewResource().Fail().Message(i18n.Translate(i18n.GofiIsAlreadyInitialized)).Build())
		return
	}

	UpdateConfiguration(ctx)
}

//GetConfiguration 获取设置项
func GetConfiguration(ctx iris.Context) {
	configuration := db.ObtainConfiguration()
	_, _ = ctx.JSON(NewResource().Payload(configuration).Build())
}
