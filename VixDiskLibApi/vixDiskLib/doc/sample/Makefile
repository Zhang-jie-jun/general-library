INCLUDEDIR=../../../include
LIBDIR=../../../lib64
LIBS=-ldl -lz -lsqlite3 -lcurl -lssl -lcrypto -lcares

ifdef VIX_AIO_BUFPOOL_SIZE
CXXFLAGS+= -DVIX_AIO_BUFPOOL_SIZE=$(VIX_AIO_BUFPOOL_SIZE)
endif

GCC_MAJOR_VER_GTEQ_4_8 := $(shell expr `$(CXX) -dumpversion | cut -f1,2 -d.` \>= 4.8)

ifeq "$(GCC_MAJOR_VER_GTEQ_4_8)" "1"
# this version support C++11
CXXFLAGS+= -std=c++11 -lpthread
else
# use boost libs
CXXFLAGS+= -lboost_system-gcc41-mt-1_42 -lboost_thread-gcc41-mt-1_42
endif

all: vix-disklib-sample vix-mntapi-sample

vix-disklib-sample: vixDiskLibSample.cpp
	$(CXX) $(CXXFLAGS) -o $@ -I$(INCLUDEDIR) -L$(LIBDIR) $? $(LIBS) -lvixDiskLib

vix-mntapi-sample:  vixDiskLibSample.cpp
	$(CXX) $(CXXFLAGS) -o $@ -DFOR_MNTAPI -I$(INCLUDEDIR) -L$(LIBDIR) $? $(LIBS) -lvixDiskLib -lvixMntapi

clean:
	$(RM) -f vix-disklib-sample vix-mntapi-sample

