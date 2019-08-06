setup: modules extern

modules:
	git submodule update --init --recursive

extern: minballPy

minballPy:
	make -C extern/minballPy
	cp extern/minballPy/build/minball.so .

minballGO:
	make -C minball
