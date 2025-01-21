//go:build darwin
// +build darwin

package pkg

/*
#include <os/log.h>
#include <stdlib.h>

void logInfo(const char *message) {
    os_log(OS_LOG_DEFAULT, "%{public}s", message);
}
*/
import "C"
import (
	"unsafe"
)

// logMessage logs an informational message to Unified Logging.
func LogMessage(message string) {
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage)) // Free C string after use
	C.logInfo(cMessage)                    // Call C function to log info
}
