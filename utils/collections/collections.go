package collections

func Filter[T any](originalList []T, callback func(T) bool) []T {
	newList := make([]T, 0, len(originalList))

	for _, value := range originalList {
		if callback(value) {
			newList = append(newList, value)
		}
	}

	return newList
}
