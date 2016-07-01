/*
   功能：监控配置目录下的文件变化自动载入(修改，删除--OK)
   作者：畅雨
   已知问题：
       1) 文件改名检测到但无法获知改名后文件故未更新----->可以用ReloadAll重新载入即可
*/
package directsql

import (
	"github.com/fsnotify/fsnotify"
	"github.com/lessgo/lessgo"
	"strings"
)

//start filesytem watcher
func (mss *ModelSqls) StartWatcher() error {
	var err error
	mss.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	lessgo.Log.Info("StartWatcher.....................")
	go func() {
		for {
			select {
			case event := <-mss.watcher.Events:
				//如果变更的文件是 .msql文件
				if strings.HasSuffix(event.Name, mss.extension) {
					if event.Op&fsnotify.Write == fsnotify.Write {
						lessgo.Log.Debug("Modified file:" + event.Name)
						err = mss.refreshModelFile(event.Name)
						if err != nil {
							lessgo.Log.Error(err.Error())
						}

					} else if event.Op&fsnotify.Remove == fsnotify.Remove {

						lessgo.Log.Debug("Delete file:" + event.Name)
						err = mss.removeModelFile(event.Name)
						if err != nil {
							lessgo.Log.Error(err.Error())
						}

					} else if event.Op == fsnotify.Rename {

						lessgo.Log.Debug("Rename file:" + event.Name)
						err = mss.renameModelFile(event.Name, event.Name)
						if err != nil {
							lessgo.Log.Error(err.Error())
						}
					}
				}
			case err := <-mss.watcher.Errors:
				if err != nil {
					lessgo.Log.Error(err.Error())
				}
			}
		}
	}()
	//增加监控路径
	for _, value := range mss.roots {
		err = mss.watcher.Add(value)
		if err != nil {
			lessgo.Log.Error(err.Error())
			//return
		}
	}
	return nil
}

//stop filesytem watcher
func (mss *ModelSqls) StopWatcher() error {
	if mss.watcher != nil {
		lessgo.Log.Info("StopWatcher.....................")
		return mss.watcher.Close()
	}
	return nil
}
