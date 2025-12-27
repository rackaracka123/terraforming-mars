package dto

// mapSlice converts a slice of one type to a slice of another type using the provided mapper function.
func mapSlice[T, U any](items []T, mapper func(T) U) []U {
	if len(items) == 0 {
		return []U{}
	}
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = mapper(item)
	}
	return result
}

// ptrCast converts a pointer of one type to a pointer of another type using the provided converter function.
// Returns nil if the input pointer is nil.
func ptrCast[T, U any](val *T, conv func(T) U) *U {
	if val == nil {
		return nil
	}
	converted := conv(*val)
	return &converted
}
