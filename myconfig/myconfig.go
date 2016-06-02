package myconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lessgo/lessgo"
	confpkg "github.com/lessgo/lessgo/config"
)

// 快速创建自己简单的ini配置
// 参数类型：支持两级嵌套、仅包含基础数据类型的结构体指针
func Sync(structPtr interface{}, defaultSection ...string) (err error) {
	v := reflect.ValueOf(structPtr)
	if v.Kind() != reflect.Ptr {
		return errors.New("SyncConfig's param must be struct pointer type.")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errors.New("SyncConfig's param must be struct pointer type.")
	}
	num := v.NumField()
	var section string
	if len(defaultSection) > 0 {
		section = defaultSection[0]
	}
	t := v.Type()
	fname := strings.ToLower(filepath.Join(lessgo.CONFIG_DIR, t.Name()+".myconfig"))

	// 遍历二级结构体
	subStructPtrs := map[string]interface{}{}
	for i := 0; i < num; i++ {
		typField := t.Field(i)
		valField := v.Field(i)
		if !valField.CanSet() {
			continue
		}
		// 字段为指针时获取内存实例
		if valField.Kind() == reflect.Ptr {
			valField = valField.Elem()
		}
		// 字段为struct时
		if valField.Kind() == reflect.Struct {
			subStructPtrs[typField.Name] = valField.Addr().Interface()
		}
	}

	// 打开配置文件
	iniconf, err := confpkg.NewConfig("ini", fname)
	if err == nil {
		os.Remove(fname)
		// 读取配置信息
		readSingleConfig(section, structPtr, iniconf)
		// 读取下一级struct配置
		for k, v := range subStructPtrs {
			readSingleConfig(k, v, iniconf)
		}
	}
	os.MkdirAll(filepath.Dir(fname), 0777)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	f.Close()
	iniconf, err = confpkg.NewConfig("ini", fname)
	if err != nil {
		return err
	}
	writeSingleConfig(section, structPtr, iniconf)
	for k, v := range subStructPtrs {
		writeSingleConfig(k, v, iniconf)
	}
	return iniconf.SaveConfigFile(fname)
}

func readSingleConfig(section string, p interface{}, iniconf confpkg.Configer) {
	pt := reflect.TypeOf(p)
	if pt.Kind() != reflect.Ptr {
		return
	}
	pt = pt.Elem()
	if pt.Kind() != reflect.Struct {
		return
	}
	pv := reflect.ValueOf(p).Elem()

	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		if !pf.CanSet() {
			continue
		}
		name := pt.Field(i).Name
		fullname := getfullname(section, name)
		switch pf.Kind() {
		case reflect.String:
			str := iniconf.DefaultString(fullname, pf.String())
			pf.SetString(str)

		case reflect.Int, reflect.Int64:
			num := int64(iniconf.DefaultInt64(fullname, pf.Int()))
			pf.SetInt(num)

		case reflect.Bool:
			pf.SetBool(iniconf.DefaultBool(fullname, pf.Bool()))
		}
	}
}

func writeSingleConfig(section string, p interface{}, iniconf confpkg.Configer) {
	pt := reflect.TypeOf(p)
	if pt.Kind() != reflect.Ptr {
		return
	}
	pt = pt.Elem()
	if pt.Kind() != reflect.Struct {
		return
	}
	pv := reflect.ValueOf(p).Elem()

	for i := 0; i < pt.NumField(); i++ {
		pf := pv.Field(i)
		if !pf.CanSet() {
			continue
		}
		fullname := getfullname(section, pt.Field(i).Name)
		switch pf.Kind() {
		case reflect.String, reflect.Int, reflect.Int64, reflect.Bool:
			iniconf.Set(fullname, fmt.Sprint(pf.Interface()))
		}
	}
}

// section name and key name case insensitive
func getfullname(section, name string) string {
	if section == "" {
		return strings.ToLower(name)
	}
	return strings.ToLower(section + "::" + name)
}
