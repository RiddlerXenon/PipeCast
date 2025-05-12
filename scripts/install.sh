#!/bin/bash

SOURCE_FILE="/usr/share/pipewire/pipewire.conf"
DEST_DIR="$HOME/.config/pipewire"
DEST_FILE="$DEST_DIR/pipewire.conf"
BACKUP_FILE="$DEST_FILE.bak_$(date +%Y%m%d_%H%M%S)"

VIRTUAL_SINK_BLOCK=$(cat << EOF
{ factory = adapter
        args = {
            node.name = "Virtual_Sink"
            media.class = "Audio/Sink"
            factory.name = support.null-audio-sink
            node.description = "Virtual Audio Output"
            node.latency = 1024/48000
        }
    }
EOF
)

if [ -f "$DEST_FILE" ]; then
    echo "Конфигурационный файл $DEST_FILE уже существует."
else
    if [ -f "$SOURCE_FILE" ]; then
        echo "Файл $SOURCE_FILE найден. Копирую в $DEST_FILE."
        mkdir -p "$DEST_DIR"
        cp "$SOURCE_FILE" "$DEST_FILE"
    else
        echo "Исходный файл $SOURCE_FILE не найден. Завершение работы."
        exit 1
    fi
fi

if grep -q 'node.name = "Virtual_Sink"' "$DEST_FILE"; then
    echo "node.name = \"Virtual_Sink\" уже присутствует в $DEST_FILE."
else
    echo "Создаю резервную копию: $BACKUP_FILE"
    cp "$DEST_FILE" "$BACKUP_FILE"

    echo "Добавляю блок Virtual_Sink в $DEST_FILE."
    echo -e "\ncontext.objects = $VIRTUAL_SINK_BLOCK" >> "$DEST_FILE"
    echo "Блок добавлен."
fi

sudo cp bin/pipecast /usr/bin/
