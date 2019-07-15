FROM alpine:latest
RUN apk update
RUN apk add libgcc libstdc++ libx11 glib libxrender libxext libintl 

WORKDIR /app
RUN mkdir cert
ADD ./simpletracker_server /app/simpletracker_server
COPY ./cert/ /app/cert/
EXPOSE 3000
CMD [ "/app/simpletracker_server" ]
