dev:
	RELOAD_TEMPLATES=1 DEV_JSX=1 go run .

prod:
	go build .
	./lastpwned

devsql:
	podman run -d -p 3306:3306 -e MARIADB_ROOT_PASSWORD=foobarbaz docker.io/mariadb
	mysql -pfoobarbaz -uroot -P3306 -h 127.0.0.1 < ./schema.sql
