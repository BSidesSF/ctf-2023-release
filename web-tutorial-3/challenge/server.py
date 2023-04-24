from flask import Flask, render_template, request, redirect, send_from_directory
from flask_csp.csp import csp_header
from werkzeug.middleware import proxy_fix
import requests
import urllib

app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)

# csp three use cookie 74106ecde0f7bd94e9c921c0e9c1acbe56748331e3a740f86aa455fa29b08202


@app.after_request
def apply_csp(response):
    response.headers["Content-Security-Policy"] = "default-src 'self' 'unsafe-inline';" \
        "script-src https://* 'nonce-corgi';" \
        "connect-src *;" \
        "style-src-elem 'self' fonts.googleapis.com fonts.gstatic.com;" \
        "font-src 'self' fonts.gstatic.com fonts.googleapis.com"
    return response


@app.route('/')
@app.route('/xss-three')
def xssThree():
    return render_template('xss-three.html')


@app.route('/xss-three-result', methods=['POST', 'GET'])
def xssThreeResult():
    payload = "None"
    if request.method == 'POST':
        payload = request.form['payload']
        r = requests.post('http://127.0.0.1:3000/submit', data={
                          'url': request.base_url + "?payload=" + urllib.parse.quote(payload)})
    if request.method == 'GET' and 'admin' in request.cookies and request.cookies.get("admin") == u"74106ecde0f7bd94e9c921c0e9c1acbe56748331e3a740f86aa455fa29b08202":
        payload = request.args.get('payload')
    elif request.method == 'GET':
        app.logger.warning('GET request without valid admin cookie.')
    return render_template('xss-three-result.html', payload=payload)


@app.route('/xss-three-flag', methods=['GET'])
def xssThreeFlag():
    if 'admin' in request.cookies and request.cookies.get("admin") == u"74106ecde0f7bd94e9c921c0e9c1acbe56748331e3a740f86aa455fa29b08202":
        app.logger.warning('GET request with valid admin cookie.')
        return "CTF{CSP-St4t1c-N0nc3}"
    else:
        return "Sorry, admins only!"


app.run(host='0.0.0.0', port=8000)
