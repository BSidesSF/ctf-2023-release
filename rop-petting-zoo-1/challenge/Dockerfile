FROM ruby:3.1.3

RUN apt-get update
RUN apt-get install -y net-tools ruby-bundler

# Configuration
ENV HOST 0.0.0.0
ENV PORT 8080

# This is where everything lives
ENV APP_HOME /app
WORKDIR $APP_HOME

# Copy the app
COPY ui/ $APP_HOME
RUN find $APP_HOME -type f -exec chmod 0644 "{}" \;
RUN find $APP_HOME -type d -exec chmod 0755 "{}" \;

# Install gems
RUN bundle install

# Copy up the binaries we need
COPY target/target $APP_HOME/target
ENV TARGET $APP_HOME/target
RUN chmod 0755 $APP_HOME/target

COPY mandrake $APP_HOME/mandrake
ENV MANDRAKE $APP_HOME/mandrake
RUN chmod 0755 $APP_HOME/mandrake

RUN useradd -r ctf -u 1000
USER 1000

# Start server
CMD ["ruby", "./server.rb"]
