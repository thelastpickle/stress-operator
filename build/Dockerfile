FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/stress-operator \
    USER_UID=1001 \
    USER_NAME=stress-operator \
    TEMPLATE_PATH=/usr/local/bin/templates

# install operator binary
COPY build/_output/bin/stress-operator ${OPERATOR}

ADD templates/*.yaml /usr/local/bin/templates/

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
