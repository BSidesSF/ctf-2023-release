import random
from datetime import datetime, timezone
from werkzeug.middleware import proxy_fix
from flask import Flask, render_template, request, redirect, flash
from flask_csp.csp import csp_header

# Form related
from flask_wtf import FlaskForm, CSRFProtect
from wtforms import StringField, PasswordField, SubmitField, TextAreaField
from wtforms.validators import DataRequired, EqualTo, ValidationError, Regexp, Length
from flask_wtf.csrf import CSRFError


# Login/Registration related
from flask_login import UserMixin, logout_user, login_user, LoginManager, login_required, current_user
from werkzeug.security import generate_password_hash, check_password_hash

# Backend
from flask_sqlalchemy import SQLAlchemy

# Flask App initialization
app = Flask(__name__)
app.wsgi_app = proxy_fix.ProxyFix(app.wsgi_app)

# Flask_login initialization
login_manager = LoginManager()
login_manager.init_app(app)


# Secret key, also used for CSRF token
app.secret_key = b'ZS8Jb7dTQW!'
csrf = CSRFProtect(app)

# Database setup
app.config["SQLALCHEMY_DATABASE_URI"] = "sqlite:///database.sqlite"
db = SQLAlchemy(app)

# User model
class User(db.Model, UserMixin):
    __tablename__ = 'user'
    id = db.Column(db.Integer, primary_key=True, index=True)
    username = db.Column(db.String(50), nullable=False, unique=True)
    password_hash = db.Column(db.String(255), nullable=False)
    score = db.Column(db.Integer, default=0, nullable=False)

    def set_password(self, password):
        self.password_hash = generate_password_hash(password)

    def check_password(self, password):
        return check_password_hash(self.password_hash, password)

    def increment_score(self):
        self.score = self.score + 1

    def reset_score(self):
        self.score = 0

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
@csrf.exempt
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
def login():
    form = LoginForm()
    if request.method == 'POST':
        if form.validate_on_submit():
            user = User.query.filter_by(username=form.username.data).first()
            if user is None or not user.check_password(form.password.data):
                flash('Invalid username or password', 'error')
                return redirect('/login')
            login_user(user)
            return redirect('/home')
    return render_template('login.html', form=form)

# Registration


@app.route('/register', methods=['GET', 'POST'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
@csrf.exempt
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


# Home page
@app.route('/home', methods=['GET'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
@login_required
def home():
    msg = ""
    if 'msg' in request.args:
        msg = request.args.get('msg')
        print(msg)
    return render_template('home.html', user=current_user, msg=msg)


@app.route('/round', methods=['GET'])
@csp_header({'default-src': "'self'", 'style-src-elem': "'self' https://fonts.googleapis.com", 'font-src': "https://fonts.gstatic.com"})
@login_required
def home_round():
    msg = ""
    if 'pick' in request.args:
        user_pick = request.args.get('pick')
        msg = round(user_pick, bot_play())
    else:
        msg = "Please pick an option!"
    return redirect('/home?msg=' + msg)

# Flag
@app.route('/flag', methods=['GET'])
@login_required
def flag():
    flagStr = 'CTF{t1me4ftert1me}'
    user = User.query.filter_by(id=current_user.id).one_or_none()
    if user.score >= 50:
        return render_template('flag.html', flag=flagStr)
    else:
        error = "You need to beat the bot 50 times to get the flag!"
        return render_template('error.html', error=error)

# Logout
@app.route('/logout', methods=['GET'])
def logout():
    logout_user()
    return redirect('/login')


# Game logic
def round(user_pick, bot_pick):
    msg = ""
    if user_pick == bot_pick:
        msg = "It's a tie!"
        loss()
    elif user_pick == "rock":
        if bot_pick == "scissors":
            msg = "Rock breaks scissors, you win!"
            win()
        else:
            msg = "Paper covers rock, you lose!"
            loss()
    elif user_pick == "paper":
        if bot_pick == "scissors":
            msg = "Scissors cuts paper, you lose!"
            loss()
        else:
            msg = "Paper covers rock, you win!"
            win()
    elif user_pick == "scissors":
        if bot_pick == "paper":
            msg = "Scissors cuts paper, you win!"
            win()
        else:
            msg = "Rock breaks scissors, you lose!"
            loss()
    else:
        msg = "Invalid input"
        loss()
    return msg


# Win
def win():
    user = User.query.filter_by(id=current_user.id).one_or_none()
    user.increment_score()
    db.session.commit()

# Lose
def loss():
    user = User.query.filter_by(id=current_user.id).one_or_none()
    user.reset_score()
    db.session.commit()

# Pick what the Bot plays
# Bot plays paper during even minutes
def bot_play():
    minute = datetime.now(tz=timezone.utc).minute
    if minute in  range(0, 62, 2):
        return "paper"
    options = ["rock", "scissors"]
    return random.choice(options)

# Helper functions
@login_manager.user_loader
def load_user(id):
    return db.session.get(User, id)


@app.errorhandler(CSRFError)
def handle_error(e):
    return render_template('error.html', error=e.description), 400


app.run(host='0.0.0.0', port=8000)
app._static_folder = ''
