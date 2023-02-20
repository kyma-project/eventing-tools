package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// Map maps the given object fields with the given ConfigMap data items.
func Map(cm *corev1.ConfigMap, obj interface{}) {
	objVal := reflect.ValueOf(obj).Elem()
	for i := 0; i < objVal.NumField(); i++ {
		field := objVal.Type().Field(i)
		for name, value := range cm.Data {
			if isMatching(field, name) {
				func() {
					defer func() {
						if r := recover(); r != nil {
							fmt.Printf("Failed to set value for field:[%v] error:[%v]\n", name, r)
						}
					}()

					fieldValue := objVal.FieldByName(field.Name)
					fieldType := fieldValue.Type()

					switch fieldType.Kind() {
					case reflect.String:
						{
							val := reflect.ValueOf(value)
							fieldValue.Set(val.Convert(fieldType))
						}
					case reflect.Int64:
						{
							if fieldType.PkgPath() == "time" && fieldType.Name() == "Duration" {
								if d, err := time.ParseDuration(value); err == nil {
									fieldValue.Set(reflect.ValueOf(int64(d)).Convert(fieldType))
								}
							}
						}
					case reflect.Int:
						{
							if val, err := strconv.ParseInt(value, 0, fieldType.Bits()); err == nil {
								fieldValue.SetInt(val)
							}
						}
					case reflect.Bool:
						{
							if val, err := strconv.ParseBool(value); err == nil {
								fieldValue.SetBool(val)
							}
						}
					}
				}()

				// break after first match
				break
			}
		}
	}
}

// isMatching returns true if the given struct field is matching the given name, otherwise returns false.
// Matching is done by struct field name or tag.
func isMatching(field reflect.StructField, name string) bool {
	if strings.EqualFold(field.Name, name) {
		return true
	}
	t := strings.ToLower(string(field.Tag))
	n := strings.ToLower(name)
	return strings.Contains(t, n)
}
