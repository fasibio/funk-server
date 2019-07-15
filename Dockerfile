FROM alpine:latest
RUN apk update
RUN apk add libgcc libstdc++ libx11 glib libxrender libxext libintl 
WORKDIR /app
ADD ./funk_server /app/funk_server
EXPOSE 3000
CMD [ "/app/funk_server" ]
