LOCAL_PATH := $(call my-dir)
ROOT_PATH := $(LOCAL_PATH)

########################################################
## libancillary
########################################################

include $(CLEAR_VARS)

ANCILLARY_SOURCE := fd_recv.c fd_send.c

LOCAL_MODULE := libancillary
LOCAL_CFLAGS += -O2 -I$(LOCAL_PATH)/libancillary

LOCAL_SRC_FILES := $(addprefix libancillary/, $(ANCILLARY_SOURCE))

include $(BUILD_STATIC_LIBRARY)

########################################################
## tun2socks
########################################################

include $(CLEAR_VARS)

LOCAL_CFLAGS := -std=gnu99
LOCAL_CFLAGS += -DBADVPN_THREADWORK_USE_PTHREAD -DBADVPN_LINUX -DBADVPN_BREACTOR_BADVPN -D_GNU_SOURCE
LOCAL_CFLAGS += -DBADVPN_USE_SELFPIPE -DBADVPN_USE_EPOLL
LOCAL_CFLAGS += -DBADVPN_LITTLE_ENDIAN -DBADVPN_THREAD_SAFE
LOCAL_CFLAGS += -DNDEBUG -DANDROID
# LOCAL_CFLAGS += -DTUN2SOCKS_JNI

LOCAL_STATIC_LIBRARIES := libancillary

LOCAL_C_INCLUDES:= \
		$(LOCAL_PATH)/libancillary \
        $(LOCAL_PATH)/badvpn/lwip/src/include/ipv4 \
        $(LOCAL_PATH)/badvpn/lwip/src/include/ipv6 \
        $(LOCAL_PATH)/badvpn/lwip/src/include \
        $(LOCAL_PATH)/badvpn/lwip/custom \
        $(LOCAL_PATH)/badvpn/

TUN2SOCKS_SOURCES := \
        base/BLog_syslog.c \
        system/BReactor_badvpn.c \
        system/BSignal.c \
        system/BConnection_common.c \
        system/BConnection_unix.c \
        system/BTime.c \
        system/BUnixSignal.c \
        system/BNetwork.c \
        flow/StreamRecvInterface.c \
        flow/PacketRecvInterface.c \
        flow/PacketPassInterface.c \
        flow/StreamPassInterface.c \
        flow/SinglePacketBuffer.c \
        flow/BufferWriter.c \
        flow/PacketBuffer.c \
        flow/PacketStreamSender.c \
        flow/PacketPassConnector.c \
        flow/PacketProtoFlow.c \
        flow/PacketPassFairQueue.c \
        flow/PacketProtoEncoder.c \
        flow/PacketProtoDecoder.c \
        socksclient/BSocksClient.c \
        tuntap/BTap.c \
        lwip/src/core/timers.c \
        lwip/src/core/udp.c \
        lwip/src/core/memp.c \
        lwip/src/core/init.c \
        lwip/src/core/pbuf.c \
        lwip/src/core/tcp.c \
        lwip/src/core/tcp_out.c \
        lwip/src/core/netif.c \
        lwip/src/core/def.c \
        lwip/src/core/mem.c \
        lwip/src/core/tcp_in.c \
        lwip/src/core/stats.c \
        lwip/src/core/inet_chksum.c \
        lwip/src/core/ipv4/icmp.c \
        lwip/src/core/ipv4/igmp.c \
        lwip/src/core/ipv4/ip4_addr.c \
        lwip/src/core/ipv4/ip_frag.c \
        lwip/src/core/ipv4/ip4.c \
        lwip/src/core/ipv4/autoip.c \
        lwip/src/core/ipv6/ethip6.c \
        lwip/src/core/ipv6/inet6.c \
        lwip/src/core/ipv6/ip6_addr.c \
        lwip/src/core/ipv6/mld6.c \
        lwip/src/core/ipv6/dhcp6.c \
        lwip/src/core/ipv6/icmp6.c \
        lwip/src/core/ipv6/ip6.c \
        lwip/src/core/ipv6/ip6_frag.c \
        lwip/src/core/ipv6/nd6.c \
        lwip/custom/sys.c \
        tun2socks/tun2socks.c \
        base/DebugObject.c \
        base/BLog.c \
        base/BPending.c \
		system/BDatagram_unix.c \
        flowextra/PacketPassInactivityMonitor.c \
        tun2socks/SocksUdpGwClient.c \
        udpgw_client/UdpGwClient.c

LOCAL_MODULE := tun2socks

LOCAL_LDLIBS := -ldl -llog

LOCAL_SRC_FILES := $(addprefix badvpn/, $(TUN2SOCKS_SOURCES))

include $(BUILD_EXECUTABLE)

########################################################
## system
########################################################

include $(CLEAR_VARS)

LOCAL_MODULE:= system

LOCAL_C_INCLUDES:= $(LOCAL_PATH)/libancillary

LOCAL_SRC_FILES:= system.cpp

LOCAL_LDLIBS := -ldl -llog

LOCAL_STATIC_LIBRARIES := cpufeatures libancillary

include $(BUILD_SHARED_LIBRARY)

$(call import-module, android/cpufeatures)