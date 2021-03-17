FROM registry.access.redhat.com/ubi8/ubi:latest AS build

# golang 1.14.12 as it's the only version available on yum at time of writing.
ENV GOLANG_VERSION=1.14.12
ENV GOLANGCI_VERSION=v1.32.2

ENV TNF_DIR=/usr/tnf
ENV TNF_SRC_DIR=${TNF_DIR}/tnf-src
ENV TNF_BIN_DIR=${TNF_DIR}/test-network-function

ENV TEMP_DIR=/tmp
# Most recent version of 4.4 available at time of writing.
ENV OC_SRC_ARCHIVE=openshift-clients-4.4.0-202006211643.p0.tar.gz
ENV OC_SRC_URL=https://github.com/openshift/oc/archive/${OC_SRC_ARCHIVE}
ENV OC_SRC_DIR=${TEMP_DIR}/oc-client-src

# Install dependencies
RUN yum install -y golang-${GOLANG_VERSION} jq make git

# Build oc from source
ADD ${OC_SRC_URL} ${TEMP_DIR}
RUN mkdir ${OC_SRC_DIR} && \ 
	tar -xf ${TEMP_DIR}/${OC_SRC_ARCHIVE} -C ${OC_SRC_DIR} --strip-components=1 && \
	cd ${OC_SRC_DIR} && \
	make oc && \
	mv ./oc /usr/bin/

# golangci-lint
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /usr/bin ${GOLANGCI_VERSION}

# Add go binary directory to $PATH
ENV PATH="/root/go/bin:${PATH}"

# Git identifier to checkout
ARG TNF_VERSION
# Pull the required version of TNF
RUN git clone --depth=1 --branch=${TNF_VERSION} https://github.com/test-network-function/test-network-function ${TNF_SRC_DIR}

# Build TNF binary
WORKDIR ${TNF_SRC_DIR}
# TODO: RUN make install-tools
RUN make install-tools && \
	make mocks && \
	make update-deps && \
	make build-cnf-tests

#  Extract what's needed to run at a seperate location
RUN mkdir ${TNF_BIN_DIR} && \
	cp run-cnf-suites.sh ${TNF_DIR} && \
	cp version.json ${TNF_DIR} && \
	cp test-network-function/test-network-function.test ${TNF_BIN_DIR}

WORKDIR ${TNF_DIR}

# ENV TEST_CONFIGURATION_PATH=${TNF_DIR}/config
# Currently not able to get it working with the env var config location, using symlinks until config files are rebuilt.
RUN ln -s ${TNF_DIR}/config/cnf_test_configuration.yml ${TNF_DIR}/test-network-function/cnf_test_configuration.yml && \
	ln -s ${TNF_DIR}/config/generic_test_configuration.yml ${TNF_DIR}/test-network-function/generic_test_configuration.yml && \
	ln -s ${TNF_DIR}/config/testconfigure.yml ${TNF_DIR}/test-network-function/testconfigure.yml

# Remove most of the build artefacts
RUN yum remove -y golang make git && \
	yum clean all && \
	rm -rf ${TNF_SRC_DIR} && \
	rm -rf ${TEMP_DIR} && \
	rm -rf /root/.cache && \
	rm -rf /root/go/pkg && \
	rm -rf /root/go/src && \
	rm -rf /usr/lib/golang/pkg && \
	rm -rf /usr/lib/golang/src

# Copy the state into a new flattened image to reduce size.
# TODO run as non-root
FROM scratch
COPY --from=build / /
ENV KUBECONFIG=/usr/tnf/kubeconfig/config
WORKDIR /usr/tnf
CMD ["./run-cnf-suites.sh", "-o", "claim", "diagnostic", "generic"]
