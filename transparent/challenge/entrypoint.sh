#!/bin/sh

# start gunicorn as daemon
/app/venv/bin/gunicorn \
  --bind  127.0.0.1:8000 \
  --workers 4 \
  --access-logfile /dev/null \
  --chdir /app \
  --daemon \
  --reuse-port \
  --user ctf \
  --group ctf \
  server:app

# now start nginx with alternative config *not* as daemon
exec /usr/sbin/nginx -c /config/nginx.conf -g 'daemon off;'
