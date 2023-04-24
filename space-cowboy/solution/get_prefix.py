from datetime import datetime, timezone
import calendar
import sys
import hashlib

def get_coupon_prefix(uid):
    part1 = uid[0:3]
    now = datetime.utcnow().replace(tzinfo=timezone.utc)
    today = now.date()
    part2 = str(today.day).rjust(2, '0')
    part3 = calendar.month_name[today.month]
    part3 = part3[0:2]
    data = part1 + part2 + part3
    data = data.upper()
    result = hashlib.md5(data.encode())
    resulthex = result.hexdigest().upper()
    return resulthex[0:8]

if len(sys.argv) != 2:
    print('Usage: solution.py <firebase_userid>')

# Variables 
uid = sys.argv[1]

print("Use the coupon prefix:",get_coupon_prefix(uid))

