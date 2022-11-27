package nginx

import "os"

// IsExecAny returns true if the file mode is executable by any user
func IsExecAny(mode os.FileMode) bool {
	return mode&0111 != 0
}
