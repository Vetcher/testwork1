set path=%path%;C:\Program Files\PostgreSQL\9.6\bin
set PGPASSWORD=postgres
set PGUSER=postgres
dropdb -e --host=localhost --port=5432 --username=postgres --no-password --if-exists mouse && echo "dropdb OK" ^
&& createdb -e --owner=postgres --host=localhost --port=5432 --username=postgres --no-password mouse && echo "createdb OK" ^
&& psql --host=localhost --port=5432 --username=postgres --no-password --file=db.sql mouse && echo "psql OK" ^
&& echo "All is OK"

