#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define KEY "73f46e5e-7c23-11ed-a1eb-0242ac120002"

char destination[100];

int isAuthenticated();
int run();
int authenticate();
int promptForTakeOff();
int takeOff();
int promptToLand();


int main(int argc, char **argv){
	return run();
}

int run(){
	char input[100];
	char password[30];
	char *temp;

	setbuf(stdout, NULL);

	if(!authenticate()){
		printf("Incorrect password\n");
		return 1;
	}

	printf("Password accepted\n");

	if(!promptForTakeOff()){
		printf("OK, maybe next time\n");
		return 1;
	}


	do{
		takeOff();

		printf("You have arrived at %s\n", destination);



	}while(!promptToLand());

	printf("Landing at %s\nThe local time is: ", destination);

	system("date");
	return 0;	
}

int authenticate(){
	char password[40];	
	printf("Welcome to the ISS Destructor navigation computer\nPlease enter the password\n> ");
	scanf("%40s", password);

	return !strcmp(password, KEY);
}

int promptForTakeOff(){
	char input[5];
	do{
		printf("Are you ready to take off? (Y/N)\n> ");
		scanf("%1s",input);
	} while( strcmp(input,"Y") && strcmp(input, "N"));

	return !strcmp(input, "Y");
}


int takeOff(){
	printf("Where would you like to go?\n> ");
	scanf("%100s", destination);
	printf("Beginning journey to %s\n", destination);
}

int promptToLand(){
	char input[100];
	do {
		printf("Would you like to land or continue flying? (land/continue)\n> ");
		scanf("%s", input);
	} while( strcmp(input,"land") && strcmp(input, "continue"));

	return !strcmp(input, "land");
}
