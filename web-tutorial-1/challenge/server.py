from flask import Flask, render_template, request, redirect, send_from_directory
from flask_csp.csp import csp_header
from werkzeug.middleware import proxy_fix
import requests
import urllib

app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)

# csp one (data uri) use cookie d635e1e1bf8b714212158fd569bfa2631565f6e959b226c7a7e0e1211dcd6d36


@app.after_request
def apply_csp(response):
    response.headers["Content-Security-Policy"] = "default-src 'self' 'unsafe-inline';" \
        "script-src 'self' 'unsafe-inline';" \
        "connect-src *;" \
        "style-src-elem 'self' fonts.googleapis.com fonts.gstatic.com;" \
        "font-src 'self' fonts.gstatic.com fonts.googleapis.com"
    return response


@app.route('/')
@app.route('/xss-one')
def xssOne():
    return render_template('xss-one.html')


@app.route('/xss-one-result', methods=['POST', 'GET'])
def xssOneResult():
    payload = "None"
    if request.method == 'POST':
        payload = request.form['payload']
        r = requests.post('http://127.0.0.1:3000/submit', data={
                          'url': request.base_url + "?payload=" + urllib.parse.quote(payload)})
    if request.method == 'GET' and 'admin' in request.cookies and request.cookies.get("admin") == u"d635e1e1bf8b714212158fd569bfa2631565f6e959b226c7a7e0e1211dcd6d36":
        payload = request.args.get('payload')
    elif request.method == 'GET':
        app.logger.warning('GET request without valid admin cookie.')
    return render_template('xss-one-result.html', payload=payload)


@app.route('/xss-one-flag', methods=['GET'])
def xssOneFlag():
    if 'admin' in request.cookies and request.cookies.get("admin") == u"d635e1e1bf8b714212158fd569bfa2631565f6e959b226c7a7e0e1211dcd6d36":
        return "CTF{XS5-1s-fun}"
    else:
        return "Sorry, admins only!"


app.run(host='0.0.0.0', port=8000)
