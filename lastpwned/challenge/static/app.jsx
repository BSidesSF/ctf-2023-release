"use strict";

const Routes = ReactRouterDOM.Routes;
const Route = ReactRouterDOM.Route;
const RouterProvider = ReactRouterDOM.RouterProvider;
const Outlet = ReactRouterDOM.Outlet;
const Link = ReactRouterDOM.Link;
const Redirect = ReactRouterDOM.Redirect;
const Navigate = ReactRouterDOM.Navigate;

const GlobalContext = React.createContext();

class PageLayout extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      loggedIn: false,
      globalContextData: {
        username: null,
      },
    };
    this.updateLoggedIn = this.updateLoggedIn.bind(this);
    this.state.globalContextData.updateLoggedIn = this.updateLoggedIn;
    const authInfo = function() {
      try {
        return getAuthInfo();
      } catch(e) {
        console.log('error getting auth info: ' + e);
        clearStorage();
        return null;
      }
    }();
    if (authInfo !== null) {
      this.state.loggedIn = true;
      this.state.globalContextData.username = authInfo.username;
    }
  }

  updateLoggedIn(username) {
    console.log('updateLoggedIn: ' + username);
    if (username === null) {
      this.setState({
        loggedIn: false,
        globalContextData: Object.assign(this.state.globalContextData, {username: null}),
      });
    } else {
      this.setState({
        loggedIn: true,
        globalContextData: Object.assign(this.state.globalContextData, {username: username}),
      });
    }
  }

  render() {
    return (
      <GlobalContext.Provider value={this.state.globalContextData}>
        <div>
          <Navbar loggedIn={this.state.loggedIn} />
          <div id="page-column">
            <div id="page-contents">
              <Outlet />
            </div>
          </div>
        </div>
      </GlobalContext.Provider>
    );
  }
}

// Horizontal navbar
class Navbar extends React.Component {
  render() {
    return (
      <nav className="navbar">
        <div className="navbar-brand navbar-item">
          Lastpwned
        </div>
        <div className="navbar-menu is-active">
          <div className="navbar-start">
            <Link className="navbar-item" to="/">Home</Link>
      {this.props.loggedIn &&
          <Link className="navbar-item" to="/passwords">Passwords</Link>}
      {this.props.loggedIn &&
          <Link className="navbar-item" to="/history">History</Link>}
      {this.props.loggedIn &&
          <Link className="navbar-item" to="/logout">Logout</Link>}
      {!this.props.loggedIn &&
          <Link className="navbar-item" to="/login">Login</Link>}
      {!this.props.loggedIn &&
          <Link className="navbar-item" to="/register">Register</Link>}
            <Link className="navbar-item" to="/about">About</Link>
          </div>
        </div>
      </nav>
    );
  }
}

// Notification bar
// props:
// - type (error, success, info)
// - message
class Notification extends React.Component {
  render() {
    const classMap = {
      'error': 'is-danger',
      'success': 'is-success',
    };
    const levelClass = classMap[this.props.type] || 'is-info';
    if (!this.props.message) {
      return '';
    }
    return (
      <div className={`notification ${levelClass}`}>
        {this.props.message}
      </div>
    );
  }
}

// Registration page
class Registration extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      username: '',
      password: '',
      confirmation: '',
      notification: '',
    };
    this.handleUsername = this.handleUsername.bind(this);
    this.handlePassword = this.handlePassword.bind(this);
    this.handleConfirm = this.handleConfirm.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleUsername(event) {
    this.setState({username: event.target.value});
  }

  handlePassword(event) {
    this.setState({password: event.target.value});
  }

  handleConfirm(event) {
    this.setState({confirmation: event.target.value});
  }

  async handleSubmit(event) {
    event.preventDefault();
    try {
      const result = await postData('/api/register', {
        'username': this.state.username.toLowerCase(),
        'password': this.state.password.toLowerCase(),
        'confirm': this.state.confirmation.toLowerCase(),
      });
      console.log(result);
      if (result.success) {
        this.setState({notification: (<Notification type="success" message={result.message} />)});
        this.updateLoggedIn(result.username);
      } else {
        this.setState({notification: (<Notification type="error" message={result.message} />)});
        this.updateLoggedIn(null);
      }
    } catch(error) {
      console.log(error);
      this.setState({notification: (<Notification type="error" message="Unknown error." />)});
      this.updateLoggedIn(null);
    }
  }

  render() {
    this.updateLoggedIn = this.context.updateLoggedIn;
    const loggedIn = this.context.username !== null;
    return (
      <div id="registration-form" className="auth-form-container">
        <CenteredDialog>
          {this.state.notification}
          {loggedIn && (
            <Navigate to="/passwords" />
          )}
          <form onSubmit={this.handleSubmit} className="auth-form box">
            <h2 className="is-size-4 has-text-weight-semibold">Register</h2>
            <div className="field">
              <label className="label">Username</label>
              <UsernameField value={this.state.username} onChange={this.handleUsername} />
            </div>
            <div className="field">
              <label className="label">Password</label>
                <PasswordField placeholder="password"
                  value={this.state.password} onChange={this.handlePassword} />
            </div>
            <div className="field">
              <label className="label">Confirm Password</label>
                <PasswordField placeholder="confirm"
                  value={this.state.confirmation} onChange={this.handleConfirm} />
            </div>
            <div className="field">
              <p className="control">
                <button className="button is-success" onClick={this.handleSubmit}>Register</button>
              </p>
            </div>
          </form>
        </CenteredDialog>
      </div>
    );
  }
}
Registration.contextType = GlobalContext;

// Login page
class LoginPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      username: '',
      password: '',
      notification: '',
    };
    this.handleUsername = this.handleUsername.bind(this);
    this.handlePassword = this.handlePassword.bind(this);
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleUsername(event) {
    this.setState({username: event.target.value});
  }

  handlePassword(event) {
    this.setState({password: event.target.value});
  }

  async handleSubmit(event) {
    event.preventDefault();
    try {
      const result = await postData('/api/login', {
        'username': this.state.username.toLowerCase(),
        'password': this.state.password.toLowerCase(),
      });
      console.log(result);
      if (result.success) {
        this.setState({notification: (<Notification type="success" message={result.message} />)});
        this.updateLoggedIn(result.username);
      } else {
        this.setState({notification: (<Notification type="error" message={result.message} />)});
        this.updateLoggedIn(null);
      }
    } catch(error) {
      console.log(error);
      this.setState({notification: (<Notification type="error" message="Unknown error." />)});
      this.updateLoggedIn(null);
    }
  }

  render() {
    this.updateLoggedIn = this.context.updateLoggedIn;
    const loggedIn = this.context.username !== null;
    return (
      <div id="login-form" className="auth-form-container">
        <CenteredDialog>
          {this.state.notification}
          {loggedIn && (
            <Navigate to="/passwords" />
          )}
          <form onSubmit={this.handleSubmit} className="auth-form box">
            <h2 className="is-size-4 has-text-weight-semibold">Login</h2>
            <div className="field">
              <label className="label">Username</label>
              <UsernameField value={this.state.username} onChange={this.handleUsername} />
            </div>
            <div className="field">
              <label className="label">Password</label>
                <PasswordField placeholder="password"
                  value={this.state.password} onChange={this.handlePassword} />
            </div>
            <div className="field">
              <p className="control">
                <button className="button is-success" onClick={this.handleSubmit}>Login</button>
              </p>
            </div>
          </form>
        </CenteredDialog>
      </div>
    );
  }
}
LoginPage.contextType = GlobalContext;

/* Main password page and components */

class PasswordPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      success: null,
      keybag: null,
      loaded: false,
      current: null,
    };
    this.loadInProgress = false;
    this.createNewEntry = this.createNewEntry.bind(this);
    this.savePassword = this.savePassword.bind(this);
    this.selectItem = this.selectItem.bind(this);
  }

  componentDidMount() {
    if (this.context.username === null) {
      return;
    }
    if (!this.loadInProgress) {
      this.loadInProgress = true;
      const keybagPromise = function(generation) {
        if (generation === undefined) {
          return loadLatestKeybag();
        }
        const authInfo = getAuthInfo();
        return loadKeybagGeneration(authInfo.username, generation);
      }(this.props.generation);
      keybagPromise.then((kbmeta) => {
        this.setState({
          keybag: kbmeta,
          loaded: true,
        });
      }).catch((e) => {
        console.error("error loading keybag: " + e);
        // check if forbidden
        if (/Forbidden/.test(e.message)) {
          // need to redirect to login
          this.context.updateLoggedIn(null);
        } else {
          this.setState({
            error: "error occurred loading",
          });
        }
      }).finally(() => {this.loadInProgress = false});
    }
  }

  componentWillUnmount() {
    this.setState({
      loaded: false,
      keybag: null,
      current: null,
      success: null,
      error: null,
      needLogin: false,
    });
    this.loadInProgress = false;
  }

  createNewEntry() {
    console.log('Creating new entry');
    if (this.state.keybag === null) {
      return;
    }
    const newItem = {
      uid: randomInt(),
      title: "",
      url: "",
      username: "",
      password: "",
    };
    this.setState({
      current: newItem,
    });
  }

  hasThisEntry(uid) {
    const keys = this.state.keybag.keys;
    for (let i = 0; i < keys.length; i++) {
      if (keys[i].uid == uid) {
        return true;
      }
    }
    return false;
  }

  savePassword(pwinfo) {
    if (this.props.readOnly) {
      return;
    }
    const keybag = this.state.keybag;
    if (!this.hasThisEntry(pwinfo.uid)) {
      keybag.keys.push(pwinfo);
    }
    try {
      saveKeybag(this.state.keybag)
        .then((res) => {
          console.log(res);
          if (res.success) {
            this.setState((state, props) => {
              if (res.updated) {
                keybag.generation = res.updated.generation;
              }
              return {
                keybag: keybag,
                success: res.message || "Saved.",
              };
            });
          } else {
            console.error("error in saving " + res.message);
            this.setState({
              error: res.message || "Unknown error",
            });
          }
        }).catch((e) => {
          console.error("error in saving: " + e);
          this.setState({
            error: "Unknown error saving.  Logged in?",
          });
        });
    } catch(e) {
      console.error("error in saving: " + e);
      this.setState({
        error: "Unknown error saving.  Logged in?",
      });
    }
  }

  selectItem(uid) {
    console.log(uid);
    const keys = this.state.keybag.keys;
    for (let i=0; i<keys.length; i++) {
      if (keys[i].uid == uid) {
        this.setState({
          current: keys[i],
          error: null,
          success: null,
        });
        return;
      }
    }
    this.setState({
      success: null,
      error: "Error selecting password entry.",
    });
  }

  render() {
    if (this.context.username === null || this.state.needLogin) {
      return (
        <Navigate to="/login" />
      );
    }
    if (!this.state.loaded) {
      return (
        <LoadingPage />
      );
    }
    const entryIndex = this.state.keybag.keys.map((kinfo) => {
      return (
        <PasswordIndexItem
          key={kinfo.uid.toString()}
          title={kinfo.title}
          url={kinfo.url}
          onClick={() => this.selectItem(kinfo.uid)}
        />);
    });
    return (
      <div className="columns">
        <div className="column is-one-third password-index">
          {this.props.readOnly ||
          <button className="box button is-fullwidth btn-new mx-auto"
            onClick={this.createNewEntry}>
            <i className="fa-solid fa-square-plus mx-auto is-size-3"></i>
          </button>}
          {entryIndex}
        </div>
        <div className="column is-two-thirds password-edit-col">
          {this.state.error &&
            <Notification type="error" message={this.state.error} />
          }
          {this.state.success &&
            <Notification type="success" message={this.state.success} />
          }
          {this.state.current &&
            <PasswordEditDialog
              key={this.state.current.uid}
              pwinfo={this.state.current}
              savepwinfo={this.savePassword}
              readOnly={this.props.readOnly || false}
              />}
        </div>
      </div>
    );
  }
}
PasswordPage.contextType = GlobalContext;

function PasswordIndexItem(props) {
  return (
    <div className="password-index-box box" onClick={props.onClick}>
      <div className="idx-title is-size-4 has-text-weight-semibold">{props.title}</div>
      <div className="idx-url">{props.url}</div>
    </div>
  );
}

function PasswordHistoryPage(props) {
  const params = ReactRouterDOM.useParams();
  return (
    <PasswordPage readOnly={true} generation={params.generation} />
  );
}

// pwinfo passed in, savepwinfo callback
// error, success props
class PasswordEditDialog extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      uid: props.pwinfo.uid,
      pwinfo: props.pwinfo,
    };
    this.handleSave = this.handleSave.bind(this);
    this.fieldUpdateHandler = this.fieldUpdateHandler.bind(this);
    this.copyPassword = this.copyPassword.bind(this);
  }

  handleSave(e) {
    e.preventDefault();
    if (this.props.readOnly) {
      return;
    }
    this.props.savepwinfo(this.state.pwinfo);
  }

  copyPassword() {
    const password = this.state.pwinfo.password;
  }

  fieldUpdateHandler(name) {
    return (function(e) {
      if (this.props.readOnly) {
        return;
      }
      const value = e.target.value;
      this.setState((state, props) => {
        const pwinfo = state.pwinfo;
        pwinfo[name] = value;
        return {pwinfo: pwinfo};
      });
    }).bind(this);
  }

  render() {
    return (
      <div className="box pw-edit-box">
        <form onSubmit={this.handleSave}>
          <div className="field">
            <label className="label">Title</label>
            <div className="control">
              <input className="input" type="text"
                placeholder="Entry Name"
                value={this.state.pwinfo.title}
                onChange={this.fieldUpdateHandler("title")}
                readOnly={this.props.readOnly}
              />
            </div>
          </div>
          <div className="field">
            <label className="label">URL</label>
            <div className="control">
              <input className="input" type="text"
                placeholder="https://www.example.com/"
                value={this.state.pwinfo.url}
                onChange={this.fieldUpdateHandler("url")}
                readOnly={this.props.readOnly}
              />
            </div>
          </div>
          {/* u/p side by side */}
          <div className="columns">
            <div className="column is-half">
              <div className="field">
                <label className="label">Username</label>
                <div className="control">
                  <input className="input" type="text"
                    placeholder="username"
                    value={this.state.pwinfo.username}
                    onChange={this.fieldUpdateHandler("username")}
                    readOnly={this.props.readOnly}
                  />
                </div>
              </div>
            </div>
            <div className="column is-half">
              <div className="field">
                <label className="label">Password</label>
                <div className="field has-addons has-addons-fullwidth">
                  <PasswordField
                    value={this.state.pwinfo.password}
                    onChange={this.fieldUpdateHandler("password")}
                    extraClasses="is-expanded"
                    readOnly={this.props.readOnly}
                  />
                  <div className="control">
                    <button className="button is-info" onClick={this.copyPassword}>
                      <i className="fa-solid fa-copy"></i>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
      {this.props.readOnly ||
          <div className="field">
            <p className="control">
              <button className="button is-success" onClick={this.handleSave}>
                Save
              </button>
            </p>
          </div>
      }
        </form>
      </div>
    );
  }
};

function IndexPage(props) {
  return (
    <div className="index-page columns">
      <div className="column is-two-thirds is-offset-one-third">
        <h2 className="is-size-3">Lastpwned</h2>
        <p className="block">
          Lastpwned is a secure password manager using zero-knowledge encryption to encrypt
          your passwords entirely on the client side.  The W3C Crypto API is used to ensure
          ensure that only the latest and greatest in security is applied to your data.
          Passwords you store are never sent to the server in plaintext.
        </p>
        <p className="block">
          Our servers also use state-of-the-art security mechanisms to protect your data.
        </p>
        <p className="block">
          We recommend making a passphrase by selecting
          <a href="https://xkcd.com/936/">random common words</a>.
        </p>
        <p className="block">
          We trust our password managers so much that our <span
            className="is-family-code">admin</span> even uses it themself.
        </p>
      </div>
    </div>
  );
};

/* Lower level components */
class PasswordField extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      hidden: true,
    };
    this.toggleHidden = this.toggleHidden.bind(this);
  }

  toggleHidden() {
    const newState = !this.state.hidden;
    this.setState({hidden: newState});
  }

  render() {
    return (
      <div 
        className={"control password-wrapper has-icons-right has-icons-left " + (this.props.extraClasses || "")}>
        <span className="icon is-left"><i className="fa fa-lock"></i></span>
        <input className={"input is-expanded " + (this.props.extraInputClasses || "")}
          type={this.state.hidden && "password" || "text"}
          placeholder={this.props.placeholder}
          value={this.props.value}
          onChange={this.props.onChange}
          readOnly={this.props.readOnly}
        />
        <span className="password-eye icon is-right"
            onClick={this.toggleHidden}>
          <i
            className={"password-eye fa-regular " + (this.state.hidden && "fa-eye" || "fa-eye-slash")}
          ></i>
        </span>
        {this.props.children}
      </div>
    );
  }
}

class HistoryPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      history: null,
      generation: null,
    };
    this.makeClickHistoryHandler = this.makeClickHistoryHandler.bind(this);
  }

  makeClickHistoryHandler(hist) {
    return function() {
      this.setState({
        generation: hist.generation,
      });
    }.bind(this);
  };

  componentDidMount() {
    loadHistory().then((history) => {
      console.log('history: ', history);
      this.setState({
        history: history,
      });
    });
  };

  render() {
    if (this.state.generation != null) {
      return (<Navigate to={"/history/" + this.state.generation.toString()} />);
    }
    if (this.state.history == null) {
      return (<LoadingPage />);
    }
    const entries = this.state.history.map((e) => {
      return (
        <div className="box" onClick={this.makeClickHistoryHandler(e)} key={e.generation}>
          <b>{e.generation}</b>: {e.created}
        </div>
      );
    });
    return (
      <div className="columns">
        <div className="column is-one-third is-offset-one-third">
          {entries}
        </div>
      </div>
    );
  }
};
HistoryPage.contextType = GlobalContext;

class LogoutPage extends React.Component {
  componentDidMount() {
    clearStorage();
    this.context.updateLoggedIn(null);
  }

  render() {
    return(
      <Navigate to="/login" />
    );
  }
};
LogoutPage.contextType = GlobalContext;

function UsernameField(props) {
  return(
    <div className="control has-icons-left">
      <span className="icon is-left"><i className="fa-regular fa-user"></i></span>
      <input className="input" type="text" placeholder="username"
        value={props.value} onChange={props.onChange} />
    </div>
  );
};

function CenteredDialog(props) {
  return (
      <div className="columns is-centered">
        <div className="column is-6-tablet is-5-desktop is-4-widescreen">
          {props.children}
        </div>
      </div>
  );
};

function LoadingPage(props) {
  return (
    <CenteredDialog>
      <div className="box is-size-1 has-text-centered">
        <div className="loader"></div>
      </div>
    </CenteredDialog>
  );
};

function appMain() {
  const root = ReactDOM.createRoot(
    document.getElementById('root')
  );
  const routes = (
    <Route path="/" element={<PageLayout />}>
      <Route index element={<IndexPage />} />
      <Route path="about" element={<IndexPage />} />
      <Route path="register" element={<Registration />} />
      <Route path="login" element={<LoginPage />} />
      <Route path="passwords" element={<PasswordPage />} />
      <Route path="logout" element={<LogoutPage />} />
      <Route path="history" element={<HistoryPage />} />
      <Route path="history/:generation" element={<PasswordHistoryPage />} />
    </Route>
  );
  const router = ReactRouterDOM.createBrowserRouter(
    ReactRouterDOM.createRoutesFromElements(routes),
  );
  const element = (
    <React.StrictMode>
      <RouterProvider router={router} />
    </React.StrictMode>
  );
  root.render(element);
  console.log('App started');
};

(function(){
  // already loaded
  console.log("readyState: " + document.readyState);
  if (/loaded|interactive|complete/.test(document.readyState)) {
    appMain();
  } else {
    document.addEventListener('DOMContentLoaded', function(e) {
      console.log('DOMContentLoaded');
      appMain();
    })
  }
})();
