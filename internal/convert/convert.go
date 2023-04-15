package convert

import (
	"fmt"
	"reflect"
	"strings"
)

var convertUtils = map[reflect.Kind]func(reflect.Value, reflect.Value) error{
	reflect.String:  converNormal,
	reflect.Int:     converNormal,
	reflect.Int16:   converNormal,
	reflect.Int32:   converNormal,
	reflect.Int64:   converNormal,
	reflect.Uint:    converNormal,
	reflect.Uint16:  converNormal,
	reflect.Uint32:  converNormal,
	reflect.Uint64:  converNormal,
	reflect.Float32: converNormal,
	reflect.Float64: converNormal,
	reflect.Uint8:   converNormal,
	reflect.Int8:    converNormal,
}

//Convert
//示例
/*
	type Target struct {
		A int `viper:"aint"`
		B string `viper:"bstr"`
	}
	src :=map[string]interface{}{
		"aint":1224,
		"bstr":"124132"
	}

	var t Target
	Convert(src,&t)

*/
//fix循环引用的问题
var _ = func() struct{} {
	convertUtils[reflect.Map] = convertMap
	convertUtils[reflect.Array] = convertSlice
	convertUtils[reflect.Slice] = convertSlice
	return struct{}{}
}()

func Convert(src interface{}, dst interface{}) (err error) {

	dstRef := reflect.ValueOf(dst)
	if dstRef.Kind() != reflect.Ptr {
		return fmt.Errorf("dst is not ptr")
	}

	dstRef = reflect.Indirect(dstRef)

	srcRef := reflect.ValueOf(src)
	if srcRef.Kind() == reflect.Ptr || srcRef.Kind() == reflect.Interface {
		srcRef = srcRef.Elem()
	}
	if f, ok := convertUtils[srcRef.Kind()]; ok {
		return f(srcRef, dstRef)
	}

	return fmt.Errorf("no implemented:%s", srcRef.Type())
}

func converNormal(src reflect.Value, dst reflect.Value) error {
	if dst.CanSet() {
		if src.Type() == dst.Type() {
			dst.Set(src)
		} else if src.CanConvert(dst.Type()) {
			dst.Set(src.Convert(dst.Type()))
		} else {
			return fmt.Errorf("can not convert:%s:%s", src.Type().String(), dst.Type().String())
		}
	}
	return nil
}

func convertSlice(src reflect.Value, dst reflect.Value) error {
	if dst.Kind() != reflect.Array && dst.Kind() != reflect.Slice {
		return fmt.Errorf("error type:%s", dst.Type().String())
	}
	l := src.Len()
	target := reflect.MakeSlice(dst.Type(), l, l)
	if dst.CanSet() {
		dst.Set(target)
	}
	for i := 0; i < l; i++ {
		srcValue := src.Index(i)
		if srcValue.Kind() == reflect.Ptr || srcValue.Kind() == reflect.Interface {
			srcValue = srcValue.Elem()
		}
		if f, ok := convertUtils[srcValue.Kind()]; ok {
			err := f(srcValue, dst.Index(i))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func convertMap(src reflect.Value, dst reflect.Value) error {
	if src.Kind() != reflect.Map || dst.Kind() != reflect.Struct {
		if src.Kind() == reflect.Interface {
			return convertMap(src.Elem(), dst)
		} else {
			return fmt.Errorf("src or dst type error,%s,%s", src.Type().String(), dst.Type().String())
		}
	}
	dstType := dst.Type()
	num := dstType.NumField()
	exist := map[string]int{}
	for i := 0; i < num; i++ {
		k := dstType.Field(i).Tag.Get("viper")
		if k == "" {
			k = dstType.Field(i).Name
		}
		if strings.Contains(k, ",") {
			taglist := strings.Split(k, ",")
			if taglist[0] == "" {

				k = dstType.Field(i).Name
			} else {
				k = taglist[0]

			}

		}
		exist[k] = i
	}

	keys := src.MapKeys()
	for _, key := range keys {
		if index, ok := exist[key.String()]; ok {
			v := dst.Field(index)
			if v.Kind() == reflect.Struct {
				err := convertMap(src.MapIndex(key), v)
				if err != nil {
					return err
				}
			} else {
				if v.CanSet() {
					if v.Type() == src.MapIndex(key).Elem().Type() {
						v.Set(src.MapIndex(key).Elem())
					} else if src.MapIndex(key).Elem().CanConvert(v.Type()) {
						v.Set(src.MapIndex(key).Elem().Convert(v.Type()))
					} else if f, ok := convertUtils[src.MapIndex(key).Elem().Kind()]; ok && f != nil {
						err := f(src.MapIndex(key).Elem(), v)
						if err != nil {
							return err
						}
					} else {
						return fmt.Errorf("error type:d(%s)s(%s)", v.Type(), src.Type())
					}
				}
			}
		}
	}

	return nil
}
