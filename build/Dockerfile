FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/tlp-stress-operator \
    USER_UID=1001 \
    USER_NAME=tlp-stress-operator

WORKDIR

# install operator binary
COPY /home/circleci/project/build/_output/bin/tlp-stress-operator /usr/local/bin/

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
