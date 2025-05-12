package main

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
import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

var (
	version   = "dev"
	buildTime = "unknown"
	commit    = "none"
)

func main() {
	fmt.Printf("Version: %s\nBuild time: %s\nCommit: %s\n", version, buildTime, commit)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go C.run()

	<-sigChan
	fmt.Println("\nReceived shutdown signal. Stopping device monitor...")

	C.stop()

	fmt.Println("Device monitor stopped. Exiting.")
}
