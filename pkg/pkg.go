//go:build darwin
// +build darwin

package pkg

/*
#include <os/log.h>
#include <stdlib.h>

void logInfo(const char *message) {
    os_log(OS_LOG_DEFAULT, "%{public}s", message);
}

void logError(const char *message) {
    os_log(OS_LOG_DEFAULT, "Error: %{public}s", message);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// logMessage logs an informational message to Unified Logging.
func LogMessage(message string) {
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage)) // Free C string after use
	C.logInfo(cMessage)                    // Call C function to log info
}

// logError logs an error message to Unified Logging.
func LogError(message string) {
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage)) // Free C string after use
	C.logError(cMessage)                   // Call C function to log error
}

func main() {
	// Log an informational message
	LogMessage("Hello, Unified Logging from Golang! VIRTRU")

	// Log an error message
	LogError("This is a simulated error message. VIRTRU")

	// Inform the user
	fmt.Println("Log messages have been written to Unified Logging.")
}
