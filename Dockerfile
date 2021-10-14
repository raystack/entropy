FROM alpine:3.13

COPY entropy /usr/bin/entropy

RUN apk --no-cache add ca-certificates bash

CMD ["entropy"]
