package com.watfaq.garrow;

import android.content.Intent;
import android.net.VpnService;
import android.os.Handler;
import android.os.Message;
import android.os.ParcelFileDescriptor;
import android.util.Log;
import android.widget.Toast;

import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.DatagramChannel;

/**
 * Created by wtf on 2016/12/6.
 */
public class GArrowVPNService extends VpnService implements Handler.Callback, Runnable {

    private static final String TAG = "GArrowVPNService";

    private Handler mHandler;
    private Thread mThread;

    private ParcelFileDescriptor mInterface;

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if (mHandler == null) {
            mHandler = new Handler(this);
        }

        if (mThread != null) {
            mThread.interrupt();
        }

        mThread = new Thread(this, "GArrowVPNThread");
        mThread.start();

        return START_STICKY;
    }

    @Override
    public boolean handleMessage(Message message) {
        if (message != null) {
            Toast.makeText(this, message.what, Toast.LENGTH_SHORT).show();
        }
        return true;
    }

    @Override
    public void onDestroy() {
        if (mThread != null) {
            mThread.interrupt();
        }
    }

    @Override
    public synchronized void run() {
        try {
            Log.i(TAG, "Starting");
            InetSocketAddress server = new InetSocketAddress("127.0.0.1", 1080);
            for (int attempt = 0; attempt < 10; ++attempt) {
                mHandler.sendEmptyMessage(R.string.connecting);

                if (run(server)) {
                    attempt = 0;
                }

                Thread.sleep(3000);
            }

            Log.i(TAG, "Give up");
        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            try {
                mInterface.close();
            } catch (Exception e) {

            }
            mInterface = null;
            mHandler.sendEmptyMessage(R.string.disconnected);
            Log.i(TAG, "Exiting");
        }
    }

    private boolean run(InetSocketAddress server) throws Exception {
        configure();
        DatagramChannel tunnel = null;
        boolean connected = false;

        try {
            tunnel = DatagramChannel.open();

            if (!protect(tunnel.socket())) {
                throw new IllegalStateException("Cannot protect the tunnel");
            }

            tunnel.connect(server);

            tunnel.configureBlocking(false);

            connected = true;

            mHandler.sendEmptyMessage(R.string.connected);

            FileInputStream in = new FileInputStream(mInterface.getFileDescriptor());

            FileOutputStream out = new FileOutputStream(mInterface.getFileDescriptor());

            ByteBuffer packet = ByteBuffer.allocate(32767);

            while (true) {

                int length = in.read(packet.array());

                Log.i(TAG, "Got " + length + " from remote");

                if (length > 0) {
                    packet.limit(length);
                    tunnel.write(packet);
                }

                length = tunnel.read(packet);
                if (length > 0) {
                    out.write(packet.array(), 0, length);
                    Log.i(TAG, "Sending " + length + " to local");
                }
                packet.clear();
            }
        } catch (Exception e) {
            e.printStackTrace();
        } finally {
            try {
                tunnel.close();
            } catch (Exception e) {
            }
        }

        return connected;
    }

    private void configure() {
        Builder builder = new Builder();
        builder.setMtu(1500)
                .addAddress("172.24.0.1", 24)
                .addRoute("0.0.0.0", 24)
                .addDnsServer("114.114.114.114");
        try {
            mInterface.close();
        } catch (Exception e) {
        }

        mInterface = builder.setSession("GArrow").establish();

        int fd = mInterface.getFd();

        String[] cmd = {
                getApplicationInfo().dataDir + "/tun2socks",
                "--netif-ipaddr", "172.0.0.2",
                "--netif-netmask", "255.255.255.0",
                "--socks-server-addr", "127.0.0.1:" + 1080,
                "--tunfd", String.valueOf(fd),
                "--tunmtc", "1500",
                "--sock-path", getApplicationInfo().dataDir + "/sock_path",
                "--loglevel", "3"
        };

        Log.i(TAG, cmd.toString());
        new GuardedProcess(cmd).start();
    }
}
