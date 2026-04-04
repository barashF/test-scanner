#!/bin/sh

while ! nc -z $POSTGRES_HOST 5432; do
  sleep 1
done

goose -dir ./migrations postgres "postgres://$POSTGRES_USER:$POSTGRES_PASSWORD@$POSTGRES_HOST:5432/$POSTGRES_DB?sslmode=disable" up

exec ./booking-app
