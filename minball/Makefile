CXX ?= g++

LDFLAGS   += -lCGAL -lgmp
LIBS      =
INCLUDES  =
WARNINGS  = -Wall -Wextra

CXXFLAGS  += -std=c++11 -x c++ $(INCLUDES) $(WARNINGS) -fPIC \
            -g3 -O2

all: libminball.so

setup:
	git submodule update --init --recursive

build_dir:
	mkdir -p build

build/libminball.so: build_dir
	$(CXX) $(CXXFLAGS) *.cpp $(LDFLAGS) -shared -o build/libminball.so

libminball.so: build/libminball.so
	cp build/libminball.so .

clean:
	rm -rf build
