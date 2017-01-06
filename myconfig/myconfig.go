package myconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/henrylee2cn/lessgo"
	confpkg "github.com/henrylee2cn/lessgo/config"
	"github.com/henrylee2cn/lessgo/utils"
)

/* 从结构体快速创建自己简单的ini配置。
 * 结构类型限制：
 * 传入Sync()的结构体必须为指针类型；
 * 内部支持最多嵌套一层子结构体；
 * 字段类型仅支持基础数据类型如：string、int、int64、bool、float 和 []string，共6种类型；
 * []string类型数据在配置项以英文分号 ";" 间隔，注意末尾不可有 ";"，否则认为该 ";" 后还有一个空字符串存在；
 * 根据实际应用场景，除[]string类型数据外，其他类型的数据在缺省相应配置项时均自动写入默认值。
 */
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
	fname := strings.TrimSuffix(t.Name(), "Config")
	fname = utils.SnakeString(fname) + ".myconfig"
	fname = filepath.Join(lessgo.CONFIG_DIR, fname)

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

	// 写入配置（缺省项将用传入的结构体相应字段自动补全）
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

		case reflect.Slice:
			pf.Set(reflect.ValueOf(iniconf.DefaultStrings(fullname, nil)))
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
		case reflect.Slice:
			var v string
			for i, count := 0, pf.Len(); i < count; i++ {
				v += ";" + fmt.Sprint(pf.Index(i).Interface())
			}
			if len(v) > 0 {
				iniconf.Set(fullname, v[1:])
			} else {
				iniconf.Set(fullname, v)
			}
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
