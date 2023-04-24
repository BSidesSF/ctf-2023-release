import sqlite3
import argparse
import os
import json
import uuid

def create_connection(db_file):
    conn = None
    try:
        conn = sqlite3.connect(db_file)
    except Error as e:
        print(e)
    return conn

def create_db(conn):
    createRegistrationTable="""CREATE TABLE IF NOT EXISTS registrationRequests(
            id integer PRIMARY KEY AUTOINCREMENT,
            rperiod int NOT NULL,
            rtype text NOT NULL,
			name text NOT NULL,
			address text NOT NULL,
            year int NOT NULL,
            make text NOT NULL,
            model text NOT NULL,
            color text NOT NULL);"""
    c = conn.cursor()
    c.execute(createRegistrationTable)

    deleteUsers = "DROP TABLE users;"
    
    c = conn.cursor()
    c.execute(deleteUsers)

    createUsersTable="""CREATE TABLE IF NOT EXISTS users(
            id integer PRIMARY KEY AUTOINCREMENT,
            username text NOT NULL,
            password text NOT NULL,
            session text);"""
    c = conn.cursor()

    c.execute(createUsersTable)

    insertAdmin = "INSERT INTO users (username, password, session) VALUES('admin', 'b30ca19e79822f0bbaeae1aadac45040', 'cc5ef4ea-46b8-450a-b6ee-8a62680e3a7d');"

    c = conn.cursor()

    c.execute(insertAdmin)

    conn.commit()

def insertRegistration(conn, args):
    print("Running Insert");
    insertRequest="""INSERT INTO registrationRequests (rperiod, rtype, name, address, year, make, model, color)
    VALUES(?, ?, ?, ?, ?, ?, ?, ?);"""
    cursor=conn.execute(insertRequest,args)

def getRequests(conn):
    getRequestsQuery="""SELECT rperiod, rtype, name, address, year, make, model, color FROM RegistrationRequests;
    """    

    cursor=conn.execute(getRequestsQuery)
    
    columnNames = [d[0] for d in cursor.description]

    out = []

    for row in cursor:
        info = dict(zip(columnNames, row))
        info['address'] = json.loads(info['address'])
        out.append(info)

    return out

def validateUser(conn, args):
    getRequestsQuery="""SELECT * FROM users WHERE username=? AND password=?;
    """    
    cursor=conn.execute(getRequestsQuery, args)

    data=cursor.fetchall()

    if len(data)==0:
        return None
    else:
        insertSessionCookie = """UPDATE users SET session=?
        WHERE username=? AND password=?;
        """
        session=str(uuid.uuid4())
        cursor=conn.execute(insertSessionCookie, (session, args[0], args[1]))
        return session

def validateSession(conn, args):
    print(args)
    getSessionQuery="""SELECT * FROM users WHERE session=?;
    """

    cursor=conn.execute(getSessionQuery, args)

    data=cursor.fetchall()

    if len(data)==0:
        return False

    return True

if __name__ == "__main__":
    database = r"sqlite.db"

    conn = create_connection(database)

    print("[+] Creating Database")
    create_db(conn)
