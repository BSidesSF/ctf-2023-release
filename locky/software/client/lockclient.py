import datetime
import os
import os.path
import sys
import threading
import time

import paho.mqtt.client as mqtt
from PIL import Image
import RPi.GPIO as GPIO

import printer
import nonce
import lcd


LOG_FILE = '/root/locklog'
STATE_TOPIC = 'ctf/lock/lock'
HEARTBEAT_TOPIC = 'ctf/lock/heartbeat'
RESET_TOPIC = 'ctf/reset'
MQTT_HOST = 'localhost'
LOGO = 'bsides_logo_gs.png'
KEY = 'make_me_a_secret'
UNLOCK_TIME = 2
HEARTBEAT_TIMEOUT = 10

# GPIO
RESET_GPIO = 25
LED_RED = 12
LED_GREEN = 13

LOCKED = 0
UNLOCKED = 1
PRINTED = 2
FAILED = 3

ALIGN_LEFT = 0
ALIGN_CENTER = 1
ALIGN_RIGHT = 2
LCD_DISP_WIDTH = 16


_ALIGN_FUNCS = {
        ALIGN_LEFT: lambda x: x.ljust(LCD_DISP_WIDTH, " "),
        ALIGN_CENTER: lambda x: x.center(LCD_DISP_WIDTH, " "),
        ALIGN_RIGHT: lambda x: x.rjust(LCD_DISP_WIDTH, " "),
}


def state_str(state):
    return ({
            LOCKED: "LOCKED",
            UNLOCKED: "UNLOCKED",
            PRINTED: "PRINTED",
            FAILED: "FAILED",
            }).get(state, "UNKNOWN")


class LockClient(object):

    def __init__(self):
        self.logf = open(LOG_FILE, 'a')

        # GPIO setup
        GPIO.setmode(GPIO.BCM)
        GPIO.setup(RESET_GPIO, GPIO.IN, pull_up_down=GPIO.PUD_UP)
        GPIO.setup(LED_RED, GPIO.OUT)
        GPIO.setup(LED_GREEN, GPIO.OUT)

        # Printer setup
        printer.ThermalPrinter.BAUDRATE = 9600
        try:
            self.printer = printer.ThermalPrinter(serialport='/dev/ttyUSB0')
        except Exception as ex:
            self.printer = None
            self.log("Printer failure!!")
        self.client = mqtt.Client()
        self.set_key(KEY)
        self.state = PRINTED  # Require reset
        self.state_time = 0
        self.heartbeat_time = time.time()
        self.lock = threading.Lock()
        self.red_led()

        # LCD setup
        self.lcd = lcd.LCD(i2c_addr=0x27)
        self.lcd_message("Please Reset", align=ALIGN_CENTER)
        if not self.printer:
            self.lcd_message("Printer offline!", line=2, clear=False)

    def set_key(self, key):
        self.validator = nonce.Nonce_24_56_Base32_Validator(key)

    def print_image(self, impath):
        """Print an image."""
        try:
            impath = os.path.join(os.path.dirname(__file__), impath)
        except NameError:
            pass
        try:
            im = Image.open(impath)
        except IOError:
            print('Unable to load image: %s' % impath)
            return
        data = list(im.getdata())
        self.printer.print_bitmap(data, *im.size)

    def print_flag(self, flag, timestamp):
        """Print the whole flag output."""
        if not self.printer:
            return
        self.print_image(LOGO)
        self.printer.print_text("\"Locky\" solved!\n")
        self.printer.print_text("Nicely done.  How about a flag\n"
                                "for your troubles?\n")
        self.printer.print_text("%s\n" % flag.decode())
        self.printer.print_text("\n")
        self.printer.print_text("Challenge lock picked at:\n")
        self.printer.print_text("%s\n" % timestamp)
        self.printer.linefeed(4)

    def handle_unlock(self):
        self.green_led()
        self.lcd_message("Unlocked!", align=ALIGN_CENTER)
        timestamp = self.timestamp
        flag = self.flag
        self.log("%s: Lock Unlocked: %s\n" % (timestamp, flag))
        try:
            self.print_flag(flag, timestamp)
        except Exception as ex:
            self.log("%s: Error while printing: %s\n",
                    timestamp, str(ex))
        self.lcd_message("Reset when done", line=2, clear=False)

    @property
    def timestamp(self):
        return datetime.datetime.now().strftime("%Y-%m-%d %H:%I:%S")

    @property
    def nonce(self):
        return int(time.time()) & 0xFFFFFF

    @property
    def flag(self):
        return self.validator.make_answer(self.nonce)

    def mqtt_on_connect(self, client, userdata, flags, rc):
        def subscribe(topic):
            self.client.subscribe(topic)
            self.log("%s: connected, subscribed to %s.\n" %
                    (self.timestamp, topic))
        subscribe(STATE_TOPIC)
        subscribe(HEARTBEAT_TOPIC)

    def mqtt_on_message(self, client, userdata, msg):
        if msg.topic == STATE_TOPIC:
            self.lock_state_handler(msg)
        elif msg.topic == HEARTBEAT_TOPIC:
            self.heartbeat_time = time.time()

    def lock_state_handler(self, msg):
        """State machine for lock."""
        with self.lock:
            self.log('Payload: %s, state: %s',
                    msg.payload, state_str(self.state))
            if msg.payload == b"open":
                if self.state == UNLOCKED:
                    if time.time() > (self.state_time + UNLOCK_TIME):
                        self.log('Saw full unlock')
                        self.handle_unlock()
                        self.state = PRINTED
                    else:
                        self.log('Saw unlock, in debounce')
                    return
                if self.state == LOCKED:
                    self.log('Saw unlock, entering debounce.')
                    self.state = UNLOCKED
                    self.state_time = time.time()
                    return
            elif msg.payload == b"closed":
                self.log('Saw lock closed')
                self.state = LOCKED
            else:
                self.log('Unknown lock message: %s' % msg.payload)

    def connect(self):
        """Connect to MQTT broker."""
        self.client.on_connect = self.mqtt_on_connect
        self.client.on_message = self.mqtt_on_message
        self.client.connect(MQTT_HOST, 1883, 60)

    def run(self):
        """Run mqtt loop and threads."""
        self._heartbeat_thread = threading.Thread(target=self.heartbeat_watcher)
        self._heartbeat_thread.daemon = True
        self._heartbeat_thread.start()

        GPIO.add_event_detect(RESET_GPIO, GPIO.FALLING,
                callback=self.reset_handler)

        # Noreturn
        self.client.loop_forever()

    def log(self, msg, *args):
        if args:
            msg = msg % args
        if msg[-1] != "\n":
            msg += "\n"
        self.logf.write(msg)
        self.logf.flush()
        sys.stdout.write(msg)
        os.fsync(self.logf.fileno())

    def heartbeat_watcher(self):
        """Keep an eye on the heartbeat."""
        while True:
            time.sleep(HEARTBEAT_TIMEOUT)
            with self.lock:
                if time.time() > (self.heartbeat_time + HEARTBEAT_TIMEOUT):
                    self.log('%s: heartbeat timeout.' % self.timestamp)
                    self.clear_led()
                    self.red_led()
                    self.lcd_message("Heartbeat tmout", line=1)
                    self.lcd_message("Check lock", line=2, clear=False)
                    if self.state != FAILED:
                        if self.printer:
                            self.printer.print_text(
                                    'ERROR\nHeartbeat timeout\nERROR\n')
                            self.printer.linefeed(4)
                    self.state = FAILED

    def reset_handler(self, unused_channel):
        """Called on reset button press."""
        with self.lock:
            self.log('%s: reset button pressed.' % self.timestamp)
            self.client.publish(RESET_TOPIC, "reset", qos=1, retain=True)
            if self.state != LOCKED:
                self.red_led()
                self.lcd_message("Close lock!!!", align=ALIGN_CENTER)
            else:
                self.clear_led()
                self.lcd_message("Locked!", align=ALIGN_CENTER)

    @classmethod
    def red_led(cls):
        cls.clear_led()
        GPIO.output(LED_RED, True)

    @classmethod
    def green_led(cls):
        cls.clear_led()
        GPIO.output(LED_GREEN, True)

    @classmethod
    def amber_led(cls):
        cls.clear_led()
        GPIO.output((LED_RED, LED_GREEN), True)

    @staticmethod
    def clear_led():
        GPIO.output((LED_RED, LED_GREEN), False)

    def lcd_message(self, msg, line=1, align=ALIGN_LEFT, clear=True):
        try:
            if clear:
                self.lcd.clear()
            msg = _ALIGN_FUNCS[align](msg)
            self.lcd.message(msg, line)
        except Exception as ex:
            self.log("Exception in LCD: %s", str(ex))


if __name__ == '__main__':
    client = None
    try:
        client = LockClient()
        print('Testing printer...')
        flg = client.flag[4:-4]
        client.print_flag(b'TEST' + flg + b'TEST', client.timestamp)
        if 'test' in sys.argv:
            sys.exit(0)
        print('Going for main.')
        # DEBUG
        client.connect()
        client.run()
    except KeyboardInterrupt:
        if client is not None:
            client.log('Saw keyboard interrupt, exiting cleanly.')
        sys.exit(0)
    finally:
        GPIO.cleanup()
