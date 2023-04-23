"use strict";

let authInfo = null;
let masterKey = null;
let userCreds = null;
const DEFAULT_ITER = 100;
const IV_LEN = 96/8;

async function getData(url) {
  console.log("GET " + url);
  const headers = {
    'Content-Type': 'application/json',
  };
  if (authInfo !== null) {
    headers['X-Auth-Token'] = authInfo['token'];
  }
  const response = await fetch(url, {
    method: 'GET',
    headers: headers,
  });
  if (!response.ok) {
    throw Error(response.statusText);
  }
  return response.json();
};

async function postData(url, data) {
  console.log("POST " + url);
  const headers = {
    'Content-Type': 'application/json',
  };
  if (authInfo !== null) {
    headers['X-Auth-Token'] = authInfo['token'];
  }
  const response = await fetch(url, {
    method: 'POST',
    headers: headers,
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw Error(response.statusText);
  }
  const body = await response.json();
  if (url.endsWith('/register') || url.endsWith('/login')) {
    if ('token' in body && body['token'] != '') {
      userCreds = data;
      authInfo = {
        username: data['username'],
        token: body['token'],
      };
      setAuthInfo(authInfo);
    }
  }
  return body;
};

async function deriveKey(salt, password, iter) {
  const utf8encode = new TextEncoder();
  const rawSalt = utf8encode.encode(salt);
  const rawPassword = utf8encode.encode(password);
  const saltedData = [];
  let i = 0;
  while (true) {
    if (i >= rawSalt.length || i >= rawPassword.length) {
      break;
    }
    saltedData.push(rawSalt[i] ^ rawPassword[i]);
    i++;
  }
  let keyData = new Uint8Array(saltedData);
  for (let i=0; i<iter; i++) {
    keyData = await window.crypto.subtle.digest("SHA-256", keyData);
  }
  let keyHash = await window.crypto.subtle.digest("SHA-256", keyData);
  const keyHashHex = toHexString(new Uint8Array(keyHash));
  return {
    keyData: keyData,
    keyHash: keyHashHex,
  };
};

let loadLatestKeybag = function() {
  let loading = false;
  const pending = [];
  return async function() {
    if (loading) {
      return await new Promise((resolve, reject) => {
        pending.push({
          resolve: resolve,
          reject: reject,
        });
      });
    }
    loading = true;
    try {
      const keybagMeta = await getData("/api/keybag");
      const rv = await loadKeybag(keybagMeta);
      console.log("Keybag result: ", rv);
      while (pending.length > 0) {
        const p = pending.pop();
        p.resolve(rv);
      }
      return rv;
    } catch(err) {
      console.error("Error loading keybag: ", err);
      while (pending.length > 0) {
        const p = pending.pop();
        p.reject(err);
      }
      throw err;
    } finally {
      loading = false;
    }
  };
}();

async function loadKeybagGeneration(username, generation) {
  const historyURL = "/api/keybag/history/" + username + "/" + generation.toString();
  const keybagMeta = await getData(historyURL);
  return loadKeybag(keybagMeta);
};

async function loadHistory() {
  const keybagHistory = await getData("/api/keybag/history");
  if (keybagHistory.success) {
    return keybagHistory.entries;
  }
  if (keybagHistory.message) {
    throw new Error(keybagHistory.message);
  }
  throw new Error("unknown error loading history");
};

// keybag entries are objects like {uid, title, url, username, password}
async function loadKeybag(keybagData) {
  let keyData = getKeyData();
  if (keyData === null || keyData.keyHash !== keybagData.keyhash) {
    // okay saved keys are not what we need
    if (userCreds === null) {
      throw new Error("No valid credentials for keybag!");
    }
    const newkey = await deriveKey(
      userCreds.username, userCreds.password, keybagData.iterations);
    if (keybagData.keyhash.length == 0) {
      // create a new keybag
      console.log('Creating new keybag.');
      keybagData.keyhash = newkey.keyHash;
      keybagData.keys = [];
      await saveKeybag(keybagData, newkey);
    } else if (newkey.keyHash !== keybagData.keyhash) {
      throw new Error("Could not derive valid key -- wrong password?");
    }
    keyData = newkey;
    setKeyData(newkey);
  }
  keybagData.keys = await loadRawKeybag(keybagData.keybag, keyData);
  if (keybagData.iterations < DEFAULT_ITER && !!userCreds) {
    console.log('Updating key!');
    const updatedKey = await deriveKey(
      userCreds.username, userCreds.password, DEFAULT_ITER);
    keybagData.keyhash = updatedKey.keyHash;
    keybagData.iterations = DEFAULT_ITER;
    setKeyData(updatedKey);
  }
  return keybagData;
};

async function loadRawKeybag(keybagRaw, keyData) {
  if (keybagRaw === null || keybagRaw.length == 0) {
    console.error('Asked to decrypt empty keybag!');
    return [];
  }
  const keybagBytes = decodeBase64ToArray(keybagRaw);
  const key = await getKeyFromBytes(keyData.keyData);
  const plainText = await window.crypto.subtle.decrypt(
    {
      name: "AES-GCM",
      iv: keybagBytes.slice(0, IV_LEN),
    },
    key,
    keybagBytes.slice(IV_LEN),
  );
  const dec = new TextDecoder();
  return JSON.parse(dec.decode(plainText));
};

async function encryptKeybag(keys, keyData) {
  const enc = new TextEncoder();
  const keybagPlainRaw = enc.encode(JSON.stringify(keys));
  const key = await getKeyFromBytes(keyData.keyData);
  const iv = new Uint8Array(IV_LEN);
  window.crypto.getRandomValues(iv);
  const cipherText = new Uint8Array(await window.crypto.subtle.encrypt(
    {
      name: "AES-GCM",
      iv: iv,
    },
    key,
    keybagPlainRaw
  ));
  const fullText = new Uint8Array(iv.length + cipherText.length);
  fullText.set(iv);
  fullText.set(cipherText, iv.length);
  return encodeArrayBase64(fullText);
}

async function saveKeybag(keybagData, keyData) {
  keyData = keyData || getKeyData();
  if (!('keys' in keybagData)) {
    keybagData.keys = [];
  }
  keybagData.keybag = await encryptKeybag(keybagData.keys, keyData);
  const saveData = {
    keybag: keybagData.keybag,
    generation: keybagData.generation,
    iterations: keybagData.iterations,
    keyhash: keyData.keyHash,
  };
  const resp = await postData('/api/keybag', saveData);
  if (resp.success && resp.updated) {
    const gen = resp.updated.generation;
    console.log("Saved keybag, new generation: " + gen);
    keybagData.generation = gen;
  }
  return resp;
}

async function getKeyFromBytes(keyBytes) {
  return window.crypto.subtle.importKey("raw", keyBytes, "AES-GCM", false, ["encrypt", "decrypt"]);
}

function toHexString(byteArray) {
  return Array.prototype.map.call(byteArray, function(byte) {
    return ('0' + (byte & 0xFF).toString(16)).slice(-2);
  }).join('');
}

function randomInt() {
  while (true) {
    const v = Math.floor(Math.random() * 2**48);
    if (Number.isSafeInteger(v)) {
      return v;
    }
  }
}

const {setAuthInfo, getAuthInfo, setKeyData, getKeyData, clearStorage,
  constantCompareArrays, decodeBase64ToArray, encodeArrayBase64} = (function() {
  const authTokenKey = "authToken";
  const keyDataKey = "keyData";

  const encodeArray = function(arbuf) {
    const arr = new Uint8Array(arbuf);
    return btoa(Array(arr.length).fill('').map(
      (_, i) => String.fromCharCode(arr[i])).join(''));
  };

  const decodeToArray = function(str) {
    const b = atob(str);
    const arr = new Uint8Array(b.length);
    for (let i=0; i < b.length; i++) {
      arr[i] = b.charCodeAt(i);
    }
    return arr.buffer;
  };

  // temporary for debugging
  window.encodeArray = encodeArray;
  window.decodeToArray = decodeToArray;

  const rv = {
    decodeBase64ToArray: decodeToArray,
    encodeArrayBase64: encodeArray,

    setAuthInfo: function(token=null) {
      if (token === null) {
        window.localStorage.removeItem(authTokenKey);
      } else {
        window.localStorage.setItem(authTokenKey, JSON.stringify(token));
      }
    },

    getAuthInfo: function() {
      const rv = window.localStorage.getItem(authTokenKey);
      if (rv === null) {
        return null;
      }
      return JSON.parse(rv);
    },

    setKeyData: function(keyPair=null) {
      if (keyPair === null) {
        window.localStorage.removeItem(keyDataKey);
      } else {
        const serData = {
          keyData: encodeArray(keyPair.keyData),
          keyHash: keyPair.keyHash,
        };
        window.localStorage.setItem(keyDataKey, JSON.stringify(serData));
      }
    },

    getKeyData: function() {
      const val = window.localStorage.getItem(keyDataKey);
      if (val === null) {
        return null;
      }
      const serData = JSON.parse(val);
      const keyPair = {
        keyData: decodeToArray(serData.keyData),
        keyHash: serData.keyHash,
      };
      return keyPair;
    },

    clearStorage: function() {
      window.localStorage.clear();
    },

    constantCompareArrays: function(a, b) {
      const arra = new Uint8Array(a);
      const arrb = new Uint8Array(b);
      if (arra.length != arrb.length)
        return false;
      let v = 0;
      for (let i=0; i<arra.length; i++) {
        v |= (arra[i] ^ arrb[i]);
      }
      return v == 0;
    },
  };

  try {
    authInfo = rv.getAuthInfo();
    masterKey = rv.getKeyData();
  } catch(e) {
    console.error('error loading storage: ' + e);
  }

  return rv;
}());
