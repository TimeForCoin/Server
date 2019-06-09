su
git checkout .
git clean -df
git pull
sudo docker-compose build
sudo docker-compose down
sudo docker-compose up -d
