# lint: no-cmd
FROM tomcat:11.0-jdk11

# Create a non-root user
RUN useradd -r tomcat -u 1000

# Provide some utilities
RUN apt update
RUN apt install -y ncat socat

# Add files
ADD flag.txt /flag.txt
ADD request.xml /var/share/request.xml

# Upload the application
ADD app/webapp /usr/local/tomcat/webapps/ROOT

# Create a secret folder for our uptime checks
RUN chown tomcat.tomcat /usr/local/tomcat/webapps/ROOT/.uptime

# Drop privileges
USER 1000
