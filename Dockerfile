FROM centurylink/ca-certs
MAINTAINER karolis.rusenas@gmail.com
COPY       webhookrelayd /webhookrelayd

ENTRYPOINT ["/webhookrelayd"]