.PRECIOUS: %.list %.index %.manifest

all: kopia.fullindex kopia.filenames kopia.suffixarray

kopia.filenames: $(shell kopia manifest list | grep type:snapshot | awk '{print $$1".filenames"}')
	sort --merge $^ > $@

kopia.fullindex: $(shell kopia manifest list | grep type:snapshot | awk '{print $$1".index"}')
	sort --merge $^ > $@

kopia.suffixarray: $(shell kopia manifest list | grep type:snapshot | awk '{print $$1".suffixarray"}')
	sort --merge $^ > $@

%.suffixarray: %.index
	cat $^ | suffixarray > $@

%.filenames: %.index
	cat $^ | filenamearray > $@

%.index: %.list
	sort $< > $@

%.list: %.manifest
	kopiaindexer --manifest=$* > $@

%.manifest:
	kopia manifest show $* > $@
