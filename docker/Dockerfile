FROM openjdk:jre-alpine
MAINTAINER Kristopher Cost <me@krisco.st>

COPY closure-compiler-v20170521.jar /gccapi/closure-compiler.jar
COPY service /gccapi/service

WORKDIR /gccapi

CMD ["./service"]
