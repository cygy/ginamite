GIT=git
DEP=go mod

bump_major: version_major archive

bump_minor: version_minor archive

bump_patch: version_patch archive

archive: dep git_push

version_major:
	$(eval VERSION=$(shell cat VERSION | awk -F. '{OFS="."; $$1+=1; $$2=0; $$3=0; print $0}'))
	echo $(VERSION) > VERSION

version_minor:
	$(eval VERSION=$(shell cat VERSION | awk -F. '{OFS="."; $$2+=1; $$3=0; print $0}'))
	echo $(VERSION) > VERSION

version_patch:
	$(eval VERSION=$(shell cat VERSION | awk -F. '{OFS="."; $$3+=1; print $0}'))
	echo $(VERSION) > VERSION

dep:
	$(DEP) tidy

git_push: vars
	$(GIT) add .
	$(GIT) commit -a -s -m "Version v$(CURRENT_VERSION)"
	$(GIT) tag -a v$(CURRENT_VERSION) -m "Version v$(CURRENT_VERSION)"
	$(GIT) push origin
	$(GIT) push --tags origin

vars:
	$(eval CURRENT_VERSION=$(shell cat VERSION))