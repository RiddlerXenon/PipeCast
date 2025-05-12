package monitor

/*
#cgo CFLAGS: -I../../internal/c_modules/source
#cgo LDFLAGS: -L../../internal/c_modules/lib -ldm -lpipewire-0.3 -ljson-c
#include "device_monitor.h"

extern struct global_data* init_device_monitor();
struct global_data* data;

void run() {
    data = init_device_monitor();

    if (data == NULL) {
        fprintf(stderr, "Failed to initialize device monitor\n");
        return;
    }
    run_device_monitor(data);
}

void stop() {
    stop_device_monitor(data);
}
*/
import "C"
import "time"

func Run() {
	go C.run()
}

func Stop() {
	C.stop()
	time.Sleep(1 * time.Second)
}

func RestartDeviceMonitor() {
	time.Sleep(2 * time.Second)
	C.stop()
	time.Sleep(1 * time.Second)
	go C.run()
}
