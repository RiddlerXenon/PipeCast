#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdbool.h>
#include <pipewire/pipewire.h>
#include <json-c/json.h>

struct global_data {
    struct pw_core *core;
    struct pw_registry *registry;
    struct pw_context *context;
    struct spa_hook registry_listener;
    struct pw_main_loop *loop;
    struct json_object *json_array;
    FILE *file;
};

static bool id_exists_in_json(const struct json_object *json_array, uint32_t id) {
    size_t array_len = json_object_array_length(json_array);
    for (size_t i = 0; i < array_len; i++) {
        struct json_object *obj = json_object_array_get_idx(json_array, i);
        struct json_object *obj_id;
        if (json_object_object_get_ex(obj, "id", &obj_id) && json_object_get_int(obj_id) == (int)id) {
            return true;
        }
    }
    return false;
}

static void remove_id_from_json(struct json_object *json_array, uint32_t id) {
    size_t array_len = json_object_array_length(json_array);
    for (size_t i = 0; i < array_len; i++) {
        struct json_object *obj = json_object_array_get_idx(json_array, i);
        struct json_object *obj_id;
        if (json_object_object_get_ex(obj, "id", &obj_id) && json_object_get_int(obj_id) == (int)id) {
            json_object_array_del_idx(json_array, i, 1);
            break;
        }
    }
}

static void write_json_to_file(struct global_data *global) {
    if (global->file) {
        rewind(global->file);
        fprintf(global->file, "%s\n", json_object_to_json_string_ext(global->json_array, JSON_C_TO_STRING_PRETTY));
        fflush(global->file);
    }
}

static void print_properties(const struct spa_dict *props) {
    const struct spa_dict_item *item;
    spa_dict_for_each(item, props) {
        printf("  %s = %s\n", item->key, item->value);
    }
}

static void registry_event_global(void *data, uint32_t id, uint32_t permissions, const char *type, uint32_t version, const struct spa_dict *props) {
    struct global_data *global = data;

    if (strcmp(type, "PipeWire:Interface:Node") == 0 && props) {
        const char *media_class = NULL;
        const struct spa_dict_item *item;
        spa_dict_for_each(item, props) {
            if (strcmp(item->key, "media.class") == 0) {
                media_class = item->value;
                break;
            }
        }

        if (media_class && strcmp(media_class, "Audio/Sink") == 0) {
            if (!id_exists_in_json(global->json_array, id)) {
                struct json_object *json_obj = json_object_new_object();
                struct json_object *json_props = json_object_new_object();

                json_object_object_add(json_obj, "id", json_object_new_int(id));
                json_object_object_add(json_obj, "type", json_object_new_string(type));
                json_object_object_add(json_obj, "version", json_object_new_int(version));

                spa_dict_for_each(item, props) {
                    json_object_object_add(json_props, item->key, json_object_new_string(item->value));
                }

                json_object_object_add(json_obj, "properties", json_props);
                json_object_array_add(global->json_array, json_obj);

                write_json_to_file(global);

                // printf("Object ID: %u\n", id);
                // printf("  Type: %s\n", type);
                // printf("  Version: %u\n", version);
                // print_properties(props);
                // printf("\n");
            }
        }
    }
}

static void registry_event_global_remove(void *data, uint32_t id) {
    struct global_data *global = data;
    remove_id_from_json(global->json_array, id);
    write_json_to_file(global);
}

static const struct pw_registry_events registry_events = {
    PW_VERSION_REGISTRY_EVENTS,
    .global = registry_event_global,
    .global_remove = registry_event_global_remove,
};

static int create_mon_list_file(struct global_data *global) {
    const char *home = getenv("HOME");
    if (!home) {
        perror("getenv");
        return -1;
    }

    char file_path[1024];
    snprintf(file_path, sizeof(file_path), "%s/.cache/pipecast/mon-list.json", home);

    global->file = fopen(file_path, "w+");
    if (!global->file) {
        fprintf(stderr, "Failed to open %s\n", file_path);
        return -1;
    }

    return 0;
}

static void cleanup(struct global_data *global) {
    if (global->context) pw_context_destroy(global->context);
    if (global->loop) pw_main_loop_destroy(global->loop);
    if (global->core) pw_core_disconnect(global->core);
    if (global->json_array) json_object_put(global->json_array);
    if (global->file) fclose(global->file);
    spa_hook_remove(&global->registry_listener);
    pw_deinit();
}

static int create_pw_module(struct global_data *global) {
    global->loop = pw_main_loop_new(NULL);
    if (!global->loop) {
        fprintf(stderr, "Failed to create main loop\n");
        return -1;
    }

    global->context = pw_context_new(pw_main_loop_get_loop(global->loop), NULL, 0);
    if (!global->context) {
        fprintf(stderr, "Failed to create context\n");
        return -1;
    }

    global->core = pw_context_connect(global->context, NULL, 0);
    if (!global->core) {
        fprintf(stderr, "Failed to connect to PipeWire core\n");
        return -1;
    }

    global->registry = pw_core_get_registry(global->core, PW_VERSION_REGISTRY, 0);
    if (!global->registry) {
        fprintf(stderr, "Failed to get registry\n");
        return -1;
    }

    return 0;
}

struct global_data* init_device_monitor() {
    struct global_data* global = malloc(sizeof(struct global_data));
    if (!global) return NULL;

    pw_init(NULL, NULL);
    global->json_array = json_object_new_array();

    if (create_mon_list_file(global) != 0) {
        cleanup(global);
        free(global);
        return NULL;
    }

    if (create_pw_module(global) != 0) {
        cleanup(global);
        free(global);
        return NULL;
    }

    pw_registry_add_listener(global->registry, &global->registry_listener, &registry_events, global);

    return global;
}

void run_device_monitor(struct global_data* global) {
    pw_main_loop_run(global->loop);
}

void stop_device_monitor(struct global_data* global) {
    pw_main_loop_quit(global->loop);
    cleanup(global);
}
