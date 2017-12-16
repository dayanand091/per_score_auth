set -e
source .env

sudo printf "CREATE USER $DEV_USERNAME WITH PASSWORD '$DEV_PASSWORD';\nCREATE DATABASE $DEV_DBNAME WITH OWNER $DEV_USERNAME;" > setupPostgres.sql
sudo -u postgres psql -f setupPostgres.sql
