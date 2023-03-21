default:
	go run cmd/main.go -i

client: 
	go run cmd/flight_client/main.go

gen_release:
	docker build -f Dockerfile.build -t flight-builder . 
	docker run --rm -v $(pwd)/release:/release flight-builder

build:
	docker build  -t flight-sys .

docker_run:
	docker run -p 8080:8080/udp flight-sys 
	
docker_clean:
	docker image prune --all --filter "dangling=true"
