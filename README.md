# DockerHub Feed

Atom feed for Docker Hub images

Docker Hubの各リポジトリの更新情報を出すRSSが配信されていないのでテキトウに作成

```bash
go get github.com/gorilla/feeds
go build dockerhub-feed.go
./dockerhub-feed
```

Getting official repository tags
```bash
curl -i "http://localhost:8080/atom?name=docker
```

Getting user repository tags
```bash
curl -i "http://localhost:8080/atom?user=${DOCKER_USER}&name=${REPOSITORY_NAME}"
```

