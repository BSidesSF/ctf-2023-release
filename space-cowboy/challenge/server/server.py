import hashlib
from flask import Flask, request
import requests
from werkzeug.middleware import proxy_fix
import firebase_admin
from firebase_admin import credentials, auth, firestore
from datetime import date
import calendar

app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)

# Application Routes
## Home 
@app.route('/', methods=['GET', 'POST'])
@app.route('/redeem-coupon', methods=['GET', 'POST'])
def redeemCoupon():
    # To handle load balancer 
    if request.method == 'GET':
        return "Server is up, challenge requires post", 200
    # Post request
    id_token = request.form.get('token')
    coupon = request.form.get('coupon')
    if id_token:
        decoded_token = auth.verify_id_token(id_token)
        # Get the user id by decoding the token
        uid = decoded_token['uid']
        # Fetch all coupons redeemed by user till date
        coupons = fetchCoupons(uid)
        coupon = coupon.upper()
        # Already redeemed coupon 
        if coupon in coupons:
            return "You've already redeemed this coupon", 403
        # Validate the coupon and update the coins balance
        else:
            if validateCoupon(uid,coupon):
                updateCoins(uid,coupon)
                return "Valid coupon", 200
            else:
                return "Invalid coupon code", 403

    else:
        return "Missing Access token, forbidden", 403

## Get the flag
@app.route('/get-flag', methods=['GET','POST'])
def getFlag():
    if request.method == 'GET':
        return "Server is up, challenge requires post", 200
    id_token = request.form.get('token')
    flag = "CTF{C0up0nC011ect10n}"
    if id_token:
        decoded_token = auth.verify_id_token(id_token)
        uid = decoded_token['uid']
        coins = fetchCoins(uid)
        # If user has more than 500 coins, return the flag
        if coins >= 500:
            return flag, 200
        # 403 error, not enough coins
        else:
            return "You need 500 coins to get the flag", 403
    else:
        return "Missing Access token, forbidden", 403

# Helper functions
## Fetch user's current coins amount 
def fetchCoins(uid):
    firestore_db = firestore.client()
    coin = "0"
    user_ref = firestore_db.collection(u'users').document(uid)
    user = user_ref.get()
    if user.exists:
        coin_ref = firestore_db.collection(u'users').document(uid)
        coin = coin_ref.get()
        if coin.exists:
            coin = coin.to_dict()['coins']
        else:
            print(u'No coins entry found')
    else:
        print(u'No such user!')
    return int(coin)

## Increment the user's coins by 100 for coupon redemption
def updateCoins(uid, coupon):
    coins = fetchCoins(uid) + 100
    firestore_db = firestore.client()
    data = {
    u'redeemed':u'true'
    }
    firestore_db.collection(u'users').document(uid).collection(u'coupons').document(coupon).set(data)
    data = {
    u'coins': str(coins)
    }
    firestore_db.collection(u'users').document(uid).set(data)
    
## Fetch all coupons redeemed by the user
def fetchCoupons(uid):
    result = set()
    firestore_db = firestore.client()
    coupons = firestore_db.collection(u'users').document(uid).collection(u'coupons').stream()
    for coupon in coupons:
        result.add(coupon.id)
    print(result)
    return result

## Validate the coupon 
def validateCoupon(uid, coupon):
    # First 3 character of the Firebase user id
    part1 = uid[0:3]
    # 2 digit format of current day
    today = date.today()
    part2 = str(today.day).rjust(2, '0')
    # First 2 character of current month's name 
    part3 = calendar.month_name[today.month]
    part3 = part3[0:2]
    # Combine parts and convert to uppercase
    data = part1 + part2 + part3
    data = data.upper()
    print(data)
    # MD5 hash of data 
    result = hashlib.md5(data.encode())
    resulthex = result.hexdigest().upper()
    print(resulthex)
    # Convert to uppercase and compare first 8 characters
    # Last 2 characters don't matter 
    coupon = coupon.upper()
    return (resulthex[0:8] == coupon[0:8])

# Todo - update to production Firebase service account 
cred = credentials.Certificate("bsidessf-ctf-2023-firebase-adminsdk.json")
firebase_admin.initialize_app(cred)
app.run(host='0.0.0.0', port=8000)