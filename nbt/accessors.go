package nbt

func Get[T Tag](c Compound, key string) (T, bool) {
	value, ok := c[key].(T)

	return value, ok
}

func Items[T Tag](list List) ([]T, bool) {
	items := make([]T, 0, len(list.Items))

	for _, item := range list.Items {
		typed, ok := item.(T)
		if !ok {
			return nil, false
		}

		items = append(items, typed)
	}

	return items, true
}
