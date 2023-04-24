import requests
import json
import hmac
import hashlib
import sys
import base64
import functools
import math

import dec8b10b


REQUEST_KEY = b"13375dea789b1337"


class Solution:

    def __init__(self, endpoint):
        if endpoint[-1] != '/':
            endpoint = endpoint + '/'
        self.endpoint = endpoint

    def get_wu(self, client_id, units_done):
        body = json.dumps(
                {'client_id': client_id, 'units_finished': units_done}
                ).encode("utf-8")
        hasher = hmac.new(key=REQUEST_KEY, digestmod=hashlib.sha256)
        hasher.update(body)
        sig = hasher.hexdigest()
        resp = requests.request(
                method='POST',
                url=self.endpoint+'api/workUnit',
                data=body,
                headers={
                    'Content-type': 'application/json',
                    'X-Request-Signature': sig,
                })
        resp.raise_for_status()
        return resp.json()['work_unit']

    def get_wus(self, client_id):
        key_counts = {}
        chunks = {}
        i = 0
        while True:
            resp = self.get_wu(client_id, i)
            i += 1
            ts = resp['start_time']
            key_counts[ts] = key_counts.get(ts, 0) + 1
            if ts not in chunks:
                chunks[ts] = base64.b64decode(resp['samples'])
            if len(key_counts) < 2:
                continue
            if all(k >= 3 for k in key_counts.values()):
                break
        return chunks

    def get_channel_samples(self, client_id):
        units = self.get_wus(client_id)
        allunits = b""
        for k in sorted(units.keys()):
            allunits += units[k]
        # find all samples mod 1337
        res = []
        i = 1337
        while i < len(allunits):
            res.append(allunits[i])
            i += 2048
        return res

    def normalize_samples(self, samples):
        return [1 if s>128 else 0 for s in samples]

    def make_runs(self, samples):
        """returns a list of 2-tuples with each being a (value, length) pair"""
        rv = []
        ct = 0
        st = 0
        for s in samples:
            if s != st:
                if ct > 0:
                    rv.append((st, ct))
                st = s
                ct = 0
            ct += 1
        if ct > 0:
            rv.append((st, ct))
        return rv

    def high_run_len(self, runs):
        high_runs = [a[1] for a in runs if a[0] == 1]
        return min(high_runs)

    def alt_run_len(self, runs):
        lens = [a[1] for a in runs]
        return functools.reduce(math.gcd, lens)

    def extract_bits(self, samples, width):
        rv = []
        for s in samples:
            v = s[0]
            for _ in range(s[1]//width):
                rv.append(v)
        return rv


if __name__ == '__main__':
    s = Solution(sys.argv[1])
    samples = s.normalize_samples(s.get_channel_samples("foo"))
    runs = s.make_runs(samples)
    print(runs)
    bitlen = s.high_run_len(runs)
    print("{} samples, {} bit len".format(len(samples), bitlen))
    altlen = s.alt_run_len(runs)
    if altlen != bitlen:
        print("length mismatch {} vs {}".format(altlen, bitlen))
        raise ValueError("length mismatch")
    bits = s.extract_bits(runs, bitlen)
    print(bits)
    res = dec8b10b.dec8b10b(bits)
    print(''.join(chr(c) for c in res))
