# luke cotnoir — terminal portfolio
# one content model, two front doors: web (angular) + ssh (go/wish)

.PHONY: sync web ssh dev-web dev-ssh build test clean

# single source of truth: content/content.json → each app gets a copy
sync:
	cp content/content.json web/public/content.json
	cp content/content.json ssh/content.json

web: sync
	cd web && ng build

ssh: sync
	cd ssh && go build -o lcotnoir .

dev-web: sync
	cd web && ng serve

dev-ssh: sync
	cd ssh && go run .

build: web ssh

test:
	cd web && ng test --watch=false --browsers=ChromeHeadless
	cd ssh && go test ./...

clean:
	rm -rf web/dist ssh/lcotnoir
