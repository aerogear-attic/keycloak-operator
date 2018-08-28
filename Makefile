ORG=witmicko
NAMESPACE=rhsso
PROJECT=keycloak-operator
SHELL = /bin/bash
TAG = 0.0.3
PKG = github.com/aerogear/keycloak-operator
TEST_DIRS     ?= $(shell sh -c "find $(TOP_SRC_DIRS) -name \\*_test.go -exec dirname {} \\; | sort | uniq")

.PHONY: check-gofmt
check-gofmt:
	diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)



.PHONY: test-unit
test-unit:
	@echo Running tests:
	go test -v -race -cover ./pkg/...

.PHONY: setup
setup:
	@echo Installing operator-sdk cli
	cd vendor/github.com/operator-framework/operator-sdk/commands/operator-sdk/ && go install .
	@echo Installing dep
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	@echo Installing errcheck
	@go get github.com/kisielk/errcheck
	@echo setup complete run make build deploy to build and deploy the operator to a local cluster


.PHONY: build
build-image:
	operator-sdk build docker.io/${ORG}/${PROJECT}:${TAG}

.PHONY: run
run:
	operator-sdk up local --namespace=${NAMESPACE} --operator-flags="--resync=8"

.PHONY: generate
generate:
	operator-sdk generate k8s

compile:
	go build -o=keycloak-operator ./cmd/keycloak-operator

.PHONY: check
check: check-gofmt test-unit
	@echo errcheck
	@errcheck -ignoretests $$(go list ./...)
	@echo go vet
	@go vet ./...

.PHONY: install
install: install_crds
	-oc new-project $(NAMESPACE)
	-kubectl create -f deploy/rbac.yaml -n $(NAMESPACE)

.PHONY: install_crds
install_crds:
	-kubectl create -f deploy/Keycloak_crd.yaml
	-kubectl create -f deploy/SharedService_crd.yaml
	-kubectl create -f deploy/SharedServicePlan_crd.yaml
	-kubectl create -f deploy/SharedServiceAction_crd.yaml
	-kubectl create -f deploy/SharedServiceSlice_crd.yaml


.PHONY: uninstall
uninstall:
	-kubectl delete role keycloak-operator -n $(NAMESPACE)
	-kubectl delete rolebinding default-account-keycloak-operator -n $(NAMESPACE)
	-kubectl delete crd keycloaks.aerogear.org
	-kubectl delete crd sharedservices.aerogear.org
	-kubectl delete crd sharedserviceplans.aerogear.org
	-kubectl delete crd sharedserviceactions.aerogear.org
	-kubectl delete crd sharedserviceslices.aerogear.org
	-kubectl delete namespace $(NAMESPACE)


.PHONY: create-examples
create-examples:
		-kubectl create -f deploy/examples/sharedservice.json -n $(NAMESPACE)
		-kubectl create -f deploy/examples/sharedserviceslice.json -n $(NAMESPACE)
		-kubectl create -f deploy/examples/sharedserviceaction.json -n $(NAMESPACE)
		-kubectl create -f deploy/examples/keycloak.json -n $(NAMESPACE)
