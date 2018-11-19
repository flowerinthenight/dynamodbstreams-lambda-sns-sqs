ENVIRONMENT := development
NAME        := teststreams-$(ENVIRONMENT)
BUCKET      := lambda-deploy-$(ENVIRONMENT)
OUT_DIR     := workspace
REGION      := $(AWS_REGION)

build: clean
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -v -a -installsuffix cgo -o $(OUT_DIR)/teststreams

package: build
	aws cloudformation package \
		--template-file template/$(ENVIRONMENT).yml \
		--s3-bucket $(BUCKET) \
		--s3-prefix $(NAME) \
		--output-template-file $(OUT_DIR)/.template.yml

deploy: package
	aws cloudformation deploy \
		--template-file $(OUT_DIR)/.template.yml \
		--stack-name $(NAME) \
		--capabilities CAPABILITY_IAM \
		--region $(REGION) 

clean: 
	rm -rf $(OUT_DIR)/*
	rm -rf $(OUT_DIR)/.template.yml
	
remove: 
	aws cloudformation delete-stack \
		--stack-name $(NAME) \
		--region $(REGION) 
	
.PHONY: clean build package deploy remove
