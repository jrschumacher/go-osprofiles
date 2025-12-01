package globalerrors

import "errors"

var ErrDeletingDefaultProfile = errors.New("error: cannot delete the default profile")
