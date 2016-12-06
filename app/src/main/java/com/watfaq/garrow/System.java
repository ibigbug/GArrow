package com.watfaq.garrow;

/**
 * Created by wtf on 2016/12/6.
 */


public class System {
    static {
        java.lang.System.loadLibrary("system");
    }

    public static native int exec(String cmd);

    public static native String getABI();

    public static native int sendfd(int fd, String path);

    public static native void jniclose(int fd);
}
