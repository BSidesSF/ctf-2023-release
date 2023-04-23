import random
import requests
from functools import wraps
from werkzeug.middleware import proxy_fix
from flask import Flask, render_template, request, redirect
from flask import flash, make_response, jsonify
from flask_csp.csp import csp_header

# Date related
from datetime import datetime, timedelta, timezone

# Form related
from flask_wtf import FlaskForm
from wtforms import StringField, PasswordField, SubmitField
from wtforms.validators import DataRequired, EqualTo, ValidationError, Regexp
from flask_wtf.csrf import CSRFError


# Login/Registration related
import jwt
from werkzeug.security import generate_password_hash, check_password_hash

# Backend
from flask_sqlalchemy import SQLAlchemy
from sqlalchemy.orm import relationship
from sqlalchemy.exc import InterfaceError

# Flask App initialization
app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)

# JWT related
app.secret_key = b'PL5b2wNKq!'
secret = "a2KnOcpKWNfzZJtf7ldV"


# Database setup
app.config["SQLALCHEMY_DATABASE_URI"] = "sqlite:///database.sqlite"
db = SQLAlchemy(app)

# User model


class User(db.Model):
    __tablename__ = 'user'
    id = db.Column(db.Integer, primary_key=True, index=True)
    username = db.Column(db.String(50), nullable=False, unique=True)
    password_hash = db.Column(db.String(255), nullable=False)

    def set_password(self, password):
        self.password_hash = generate_password_hash(password)

    def check_password(self, password):
        return check_password_hash(self.password_hash, password)

    def __repr__(self):
        return self.username


with app.app_context():
    db.create_all()

# Forms used by the application


class LoginForm(FlaskForm):
    class Meta:
        csrf = False
    username = StringField('Username', validators=[DataRequired(), Regexp(
        '^\w+$', message="Username must be AlphaNumeric")])
    password = PasswordField('Password', validators=[DataRequired()])
    submit = SubmitField('Login')


class RegistrationForm(FlaskForm):
    class Meta:
        csrf = False
    username = StringField('Username', validators=[DataRequired(), Regexp(
        '^\w+$', message="Username must be AlphaNumeric")])
    # email = StringField('Email Address', validators=[DataRequired(), Email()])
    password = PasswordField('New Password',
                             validators=[DataRequired()])
    confirm = PasswordField('Repeat Password', validators=[
                            DataRequired(), EqualTo('password', message='Passwords must match')])
    submit = SubmitField('Register')

    def validate_username(self, username):
        user = User.query.filter_by(username=username.data).first()
        if user is not None:
            raise ValidationError('Please use a different username.')

# Application routes

# Login


@app.route('/')
@app.route('/login', methods=['GET', 'POST'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
def login():
    form = LoginForm()
    if request.method == 'POST':
        if form.validate_on_submit():
            username = form.username.data
            user = User.query.filter_by(username=username).first()
            if user is None or not user.check_password(form.password.data):
                flash('Invalid username or password', 'error')
                return redirect('/login')
            jwt_token = create_jwt(user.username, 0)
            response = make_response(redirect('/home'))
            response.set_cookie("token", jwt_token)
            return response
    return render_template('login.html', form=form)

# Registration


@app.route('/register', methods=['GET', 'POST'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
def register():
    form = RegistrationForm()
    if form.validate_on_submit():
        user = User(username=form.username.data)
        user.set_password(form.password.data)
        db.session.add(user)
        db.session.commit()
        flash('Thanks for registering')
        return redirect('/login')
    return render_template('register.html', form=form)

# Protected routes
# Decode JWT


def decode_jwt():
    jwt_token = request.cookies.get("token")
    jwt_header = jwt.get_unverified_header(jwt_token)
    if jwt_header["alg"] == "none":
        jwt_str = None
        try:
            # To accept None Algo turn off verification
            jwt_str = jwt.decode(jwt_token, options={
                                 "verify_signature": False})
            print("None algorithm JWT:", jwt_str)
            return jwt_str
        except Exception as e:
            return None
    else:
        try:
            jwt_str = jwt.decode(jwt_token, secret, "HS256")
        except Exception as e:
            return None
        else:
            return jwt_str

# JWT required decorator


def login_required(f):
    @wraps(f)
    def decorator(*args, **kwargs):
        jwt_token = ""
        if "token" in request.cookies:
            jwt_token = request.cookies["token"]
        else:
            return handle_error("No JWT token")
        jwt_str = decode_jwt()
        if jwt_str == None:
            return handle_error("Invalid JWT token")
        return f(*args, **kwargs)
    return decorator


# Home page
@app.route('/home', methods=['GET'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
@login_required
def home():
    return render_template('home.html', user=get_user(), score=get_jwt_score(), msg="")

# Game round


@app.route('/round', methods=['GET'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
@login_required
def home_round():
    msg = ""
    if 'pick' in request.args:
        user_pick = request.args.get('pick')
        score, msg = round(user_pick, bot_play())
        response = make_response(render_template(
            'home.html', user=get_user(), score=score, msg=msg))
        response.set_cookie("token", set_jwt_score(score))
        return response
    else:
        msg = "Please pick an option!"
        return render_template('home.html', user=get_user(), score=get_jwt_score(), msg=msg)

# Flag
# You don't need a registered user to get the flag :)


@app.route('/flag', methods=['GET'])
@login_required
def flag():
    flagStr = 'CTF{W3W1llR0ckY0u}'
    score = get_jwt_score()
    if score >= 1000000:
        return render_template('flag.html', flag=flagStr)
    else:
        error = "You need to beat the bot 1 million times to get the flag!"
        return render_template('error.html', error=error)

# Logout


@app.route('/logout', methods=['GET'])
def logout():
    response = make_response(redirect('/login'))
    response.delete_cookie("token")
    return response

# Game logic


def round(user_pick, bot_pick):
    msg = ""
    score = 0
    if user_pick == bot_pick:
        msg = "It's a tie!"
        score = loss()
    elif user_pick == "rock":
        if bot_pick == "scissors":
            msg = "Rock breaks scissors, you win!"
            score = win()
        else:
            msg = "Paper covers rock, you lose!"
            score = loss()
    elif user_pick == "paper":
        if bot_pick == "scissors":
            msg = "Scissors cuts paper, you lose!"
            score = loss()
        else:
            msg = "Paper covers rock, you win!"
            score = win()
    elif user_pick == "scissors":
        if bot_pick == "paper":
            msg = "Scissors cuts paper, you win!"
            score = win()
        else:
            msg = "Rock breaks scissors, you lose!"
            score = loss()
    else:
        msg = "Invalid input"
        score = loss()
    return score, msg


# Win
def win():
    score = get_jwt_score()
    return score+1

# Lose


def loss():
    return 0

# Pick what the Bot plays


def bot_play():
    options = ["rock", "paper", "scissors"]
    return random.choice(options)


# Hendle the score
def get_jwt_score():
    print("In jwt score")
    jwt_token = decode_jwt()
    if "score" in jwt_token:
        return jwt_token["score"]
    else:
        set_jwt_score(0)
        return 0


def set_jwt_score(num):
    print("Setting jwt to:", num)
    username = get_user().username
    return create_jwt(username, num)


# Create JWT
def create_jwt(username, score):
    time_exp = datetime.now(tz=timezone.utc) + timedelta(hours=1)
    time_now = datetime.now(tz=timezone.utc)
    payload = {"sub": username,
               "score": score,
               "exp": time_exp,
               "iat": time_now
               }
    return jwt.encode(payload, secret, "HS256")


# Helper functions
def get_user():
    jwt_token = decode_jwt()
    if "sub" in jwt_token:
        user = User.query.filter_by(username=jwt_token["sub"]).one_or_none()
        return user
    return None

# Other helper functions


def handle_error(str):
    return render_template('error.html', error=str), 400


app.run(host='0.0.0.0', port=8000)
app._static_folder = ''
