$env:DB_HOST="127.0.0.1"
$env:DB_PORT="8080"
$env:DB_NAME="name"
$env:DB_USER="user"
$env:DB_PASSWORD="password"

$env:REDIS_HOST="127.0.0.1"
$env:REDIS_PORT="6379"
$env:REDIS_DB="0"
$env:REDIS_PASSWORD="password"

go test -v .\app\...
