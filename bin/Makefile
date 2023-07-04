.PRECIOUS: %.list %.index %.manifest

all: kopia.fullindex


kopia.fullindex: $(shell kopia manifest list | grep type:snapshot | awk '{print $$1".index"}')

%.index: %.list
	sort $< > $@

%.list: %.manifest
	kopiaindexer --manifest=$* > $@

%.manifest:
	kopia manifest show $* > $@
