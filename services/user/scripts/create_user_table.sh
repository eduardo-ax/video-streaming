#!/bin/bash
docker exec -i postgres_user psql -U users_user -d usersdb < ../schemas/user_table.up.sql
