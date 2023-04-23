#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sqlite3.h>



void initDB();
void upload(char *,char *);
int download(char*);
void search(char*);
void processCommand();
void printUnknownCommand();

int main(int argc, char** argv){
	initDB();
	processCommand();
	return 0;
}

void processCommand(){
	char input[100000] = "";
	char command[100] = "";
	char arg1[2000] = "";
	char arg2[2000] = "";
	char *token;
	setbuf(stdout, NULL);
	while(strcmp(command, "exit")){
		printf("Please enter a valid command \n");
		printf("> ");
		scanf("%10000s", input);
		token = strtok(input, ",");
		strncpy(command, token, 100);
		token = strtok(NULL, ",");
		if(token){
			strncpy(arg1, token, 2000);
			token = strtok(NULL, ",");
			if(token){
				strncpy(arg2, token, 2000);
			}
		}
		if(!strcmp(command, "upload")){
			upload(arg1, arg2);
		}
		else if(!strcmp(command, "download")){
			download(arg1);
		}
		else if(strcmp(command, "exit")){
			printUnknownCommand(command);	
		}
	}
}

void printUnknownCommand(char *command){
	printf("Unkown command: ");
	printf(command);
	printf("\n");
}

void initDB(){
	char *err = 0;
	int rc;
	sqlite3 *db;
	char *sql;

	rc = sqlite3_open("sqlite.db", &db);

	sql = "DROP TABLE files";

	rc = sqlite3_exec(db, sql, NULL, 0, &err);

	sql =  "CREATE TABLE files(" \
	       "id INTEGER PRIMARY KEY AUTOINCREMENT," \
	       "name 		TEXT			NOT NULL," \
	       "data		BLOB			NOT NULL);";

	rc = sqlite3_exec(db, sql, NULL, 0, &err);	

	sqlite3_close(db);
}

void upload(char *name, char *data){
	char *err;
	int rc;
	sqlite3 *db;
	sqlite3_stmt *stmt;
	sqlite3_open("sqlite.db", &db);
	printf("Uploading file with the following name: ");
	printf(name);
	printf("\n");
	char *sql = "INSERT INTO files (name, data) VALUES (?, ?);";
	rc = sqlite3_prepare_v2(db, sql, -1, &stmt, 0);
	rc = sqlite3_bind_text(stmt, 1, name, strlen(name), SQLITE_STATIC);
	rc = sqlite3_bind_blob(stmt, 2, data, strlen(data), SQLITE_STATIC);
	rc = sqlite3_step(stmt);
	rc = sqlite3_finalize(stmt);
	sqlite3_close(db);
}

int download(char *name){
	char *err;
	char out[1000] = "";
	int rc;
	sqlite3_stmt *stmt = 0;
	sqlite3 *db;
	sqlite3_open("sqlite.db", &db);
	rc = sqlite3_prepare_v2(db, "SELECT data FROM files WHERE name=?;", -1, &stmt, 0);
	rc = sqlite3_bind_text(stmt, 1, name, -1, SQLITE_STATIC);
	rc = sqlite3_step(stmt);
	char *data = (char *)sqlite3_column_blob(stmt, 0);
	strcpy(out, name);
	strcat(out, " - ");
	strcat(out, data);
	fprintf(stdout, "%s\n", out);
	sqlite3_close(db);
	return 1;
}
