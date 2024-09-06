#!/usr/bin/env bash

# Runs praga with the nginx user, giving the nginx user access to the unix socket for connections

# Typically www-data or nginx, check your /etc/nginx/nginx.conf
NGINX_USER=www-data

mkdir /run/praga
chown "$NGINX_USER" /run/praga
exec su "$NGINX_USER" -s /bin/bash -c 'exec praga --config=/etc/praga.yaml'
