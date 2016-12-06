package com.watfaq.garrow;

import android.util.Log;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;

/**
 * Created by wtf on 2016/12/6.
 */

class StreamLogger extends Thread {
    private InputStream mIn;
    private String mTag;

    public StreamLogger(InputStream in, String tag) {
        mIn = in;
        mTag = tag;
    }

    @Override
    public void run() {
        InputStreamReader r = new InputStreamReader(mIn);
        BufferedReader br = new BufferedReader(r);
        try {
            while (true) {
                String line = br.readLine();
                Log.i(mTag, line);
            }
        } catch (IOException e) {
            e.printStackTrace();
        } finally {
            try {
                r.close();
                br.close();
            } catch (Exception e) {

            }
        }
    }
}

public class GuardedProcess {
    private static final String TAG = GuardedProcess.class.getSimpleName();
    volatile private Thread guardThread;
    volatile private boolean destroyed;
    volatile private Process process;
    volatile private boolean restart;

    private String[] mCmd;

    public GuardedProcess(String[] cmd) {
        mCmd = cmd;
    }

    public void start() {
        guardThread = new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    process = new ProcessBuilder(mCmd).redirectErrorStream(true).start();

                    InputStream is = process.getInputStream();
                    new StreamLogger(is, TAG).start();
                } catch (IOException e) {
                    e.printStackTrace();
                }
            }
        });
        guardThread.start();
    }
}
