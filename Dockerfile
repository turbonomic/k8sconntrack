# Set the base image
FROM ubuntu

# Set the file maintainer
MAINTAINER Dongyi Yang <dongyi.yang@vmturbo.com>

ADD ./_output/conntracker /bin/conntracker

ENTRYPOINT ["/bin/conntracker"]
