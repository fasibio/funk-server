buildBinary: 
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -installsuffix cgo -o funk_server .

buildDocker: 
	docker build --build-arg buildNumber=${CI_PIPELINE_IID} -t fasibio/funk_server:${CI_PIPELINE_IID} .

publishDocker: 
	docker login -u ${dockerhubuser} -p ${dockerhubpassword}
	docker push fasibio/funk_server:${CI_PIPELINE_IID}