package generic

// EqualsAny checks if a base value of type T is equal to
// any of the other values of type T provided as arguments.
func EqualsAny[T comparable](base T, all ...T) bool {
	for _, v := range all {
		if v == base {
			return true
		}
	}
	return false
}

// EqualsAll checks if a base value of type T is equal to all of the
// other values of type T provided as arguments.
func EqualsAll[T comparable](base T, all ...T) bool {
	for _, v := range all {
		if v != base {
			return false
		}
	}
	return true
}
