IMAGE = faraway-app
NAMESPACE = faraway
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

build:
	docker build -t $(IMAGE) .

minikube-up: build
	minikube start --wait all
	minikube image load $(IMAGE)

	kustomize build env | kubectl apply -f -

	helm upgrade --install --create-namespace --namespace $(NAMESPACE) --set image.repository=$(IMAGE) faraway-app ./chart

minikube-test:
	bash ./minikube-test.sh	

minikube-down:
	minikube stop
	minikube delete