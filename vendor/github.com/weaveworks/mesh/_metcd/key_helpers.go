package metcd

// PrefixRangeEnd allows Get, Delete, and Watch requests to operate on all keys
// with a matching prefix. Pass the prefix to this function, and use the result
// as the RangeEnd value.
func PrefixRangeEnd(prefix []byte) []byte {
	// https://github.com/coreos/etcd/blob/17e32b6/clientv3/op.go#L187
	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		if end[i] < 0xff {
			end[i] = end[i] + 1
			end = end[:i+1]
			return end
		}
	}
	// next prefix does not exist (e.g., 0xffff);
	// default to WithFromKey policy
	return []byte{0}
}
