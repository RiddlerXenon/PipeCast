#include <stdio.h>

#ifndef DEVICE_MONITOR_H
#define DEVICE_MONITOR_H

// global_data не объявлен в этом заголовочном файле, поэтому мы должны включить его в этот заголовочный файл для использования его в качестве аргумента функции и возвращаемого значения

// добавим структуру global_data в заголовочный файл
struct global_data {
    struct json_object *json_array;
    FILE *file;
    struct pw_main_loop *loop;
    struct pw_context *context;
    struct pw_core *core;
};

struct global_data* init_device_monitor();
void run_device_monitor(struct global_data* data);
void stop_device_monitor(struct global_data* data);

#endif

