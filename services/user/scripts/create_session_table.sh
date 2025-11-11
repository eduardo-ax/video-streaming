#!/bin/bash
docker exec -i postgres_user psql -U users_user -d usersdb < ../infrastructure/schemas/session_table.up.sql
