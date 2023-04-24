from werkzeug.middleware import proxy_fix
from flask import Flask, render_template, request

app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)

# Application routes
## Home
@app.route('/')
def home():
    flag = "CTF{C3rt1f1edOS1nt}"
    return render_template('home.html', flag=flag)


if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8000)
    app._static_folder = ''
