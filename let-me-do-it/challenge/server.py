from google.oauth2 import id_token
from google.auth.transport import requests as r
from flask import Flask, render_template, request, redirect
from flask_csp.csp import csp_header
from werkzeug.middleware import proxy_fix
import requests
import json

# Flask App initialization 
app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)


# Apply the CSP header to all requests
@app.after_request
def apply_csp(response):
    response.headers["Content-Security-Policy"] = "default-src 'self'" \
        "style-src-elem 'self' fonts.googleapis.com fonts.gstatic.com;" \
        "font-src 'self' fonts.gstatic.com fonts.googleapis.com"
    return response

# App routes
## Home page
@app.route('/', methods=['GET'])
def home():
    return render_template('home.html')

## Login 
@app.route('/login', methods=['GET'])
def login():
    return render_template('login.html')

## Flag
@app.route('/flag', methods=['POST','GET'])
def flag():
    logged_in = False
    user = ""
    flag = ""
    error = ""
    if request.method == 'GET':
        error = "Needs to be POST request"
        return render_template('flag.html',flag=flag, logged_in=logged_in, user=user, error=error)
    try:
        token = request.form.get('id_token')
        idinfo = id_token.verify_oauth2_token(token, r.Request())
        response = requests.get('https://oauth2.googleapis.com/tokeninfo?id_token=' + token)
        response_json = json.loads(response.text)
        if ('email' in response_json.keys()):
            user = response_json['email']
        else:
            error = "Require scope - https://www.googleapis.com/auth/userinfo.email"
    except ValueError:
        # Invalid token
        error = "Invalid id_token"
        pass
    else:
        if error == "":
            logged_in = True
            flag = "CTF{Val1dat3Cl13ntId}"
    return render_template('flag.html',flag=flag, logged_in=logged_in, user=user, error=error)

app.run(host='0.0.0.0', port=8000)
app._static_folder = ''
