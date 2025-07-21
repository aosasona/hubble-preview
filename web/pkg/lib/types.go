package lib

import "reflect"

/*
Check if the value is a pointer
*/
func IsPointer(value any) bool {
	return reflect.TypeOf(value).Kind() == reflect.Ptr
}
