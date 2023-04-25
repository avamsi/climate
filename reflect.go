package climate

import "reflect"

var errorType = reflect.TypeOf((*error)(nil)).Elem()

func typeIsError(t reflect.Type) bool {
	return t.Kind() == reflect.Interface && t.Implements(errorType)
}

func typeIsStringSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.String
}

func typeIsStructPointer(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

type reflection struct {
	ptr *reflection
	ot  reflect.Type
	ov  *reflect.Value
}

func (r *reflection) t() reflect.Type {
	if r == nil {
		return nil
	}
	if r.ot == nil {
		if r.ptr.t() != nil {
			r.ot = r.ptr.ot.Elem()
		} else {
			r.ot = r.ov.Type()
		}
	}
	return r.ot
}

func (r *reflection) v() *reflect.Value {
	if r == nil {
		return nil
	}
	if r.ov == nil {
		if r.ptr == nil {
			r.ptr = &reflection{}
		}
		if r.ptr.ov == nil {
			v := reflect.New(r.t())
			r.ptr.ov = &v
		}
		v := r.ptr.ov.Elem()
		r.ov = &v
	}
	return r.ov
}
