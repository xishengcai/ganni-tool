#!/bin/bash

set -e
echo "run mysqldump source $SOURCE_HOST:$SOURCE_PORT"
mysqldump --host=$SOURCE_HOST --port=$SOURCE_PORT \
--user=$SOURCE_USER --password=$SOURCE_PASSWORD --databases $BACKUP_DB_LIST > dump.sql

echo "run mysql restore target $TARGET_HOST:$TARGET_PORT"
mysql -f -h $TARGET_HOST --port=$TARGET_PORT \
--user=$TARGET_USER --password=$TARGET_PASSWORD < dump.sql

echo "success"

