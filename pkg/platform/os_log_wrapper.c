#include <os/log.h>

void log_to_unified(const char *message) {
    os_log(OS_LOG_DEFAULT, "%{public}s", message);
}