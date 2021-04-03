FROM alpine
COPY twitter-cleaner_*.apk /tmp/
RUN apk add --allow-untrusted /tmp/twitter-cleaner_*.apk
ENTRYPOINT ["/usr/local/bin/twitter-cleaner"]
